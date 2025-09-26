package bybit_api

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/dustinxie/lockfree"
	json "github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/wwnbb/bybit_api/bpool"
	pp "github.com/wwnbb/pprint"
)

const (
	wsInitialReconnectDelay  = 5 * time.Second
	wsMaxReconnectDelay      = 5 * time.Minute
	wsReconnectBackoffFactor = 2.0
	wsMaxSilentPeriod        = 30 * time.Second
	wsPingInterval           = 10 * time.Second
)

type ResponseHeader struct {
	Topic string `json:"topic"`
	Op    string `json:"op"`
}

// WSConnection represents a websocket connection,
// it embeds the gorilla websocket connection
// creation done in connect method of WSManager
type WSConnection struct {
	*websocket.Conn

	ctx    context.Context
	cancel context.CancelFunc

	lastPing time.Time

	subscribed map[string]struct{}

	apiKey    string
	apiSecret string

	writeMu sync.Mutex
	pingMu  sync.Mutex
}

func (c *WSConnection) getLastPing() time.Time {
	c.pingMu.Lock()
	defer c.pingMu.Unlock()
	return c.lastPing
}

func (c *WSConnection) setLastPing(t time.Time) {
	c.pingMu.Lock()
	defer c.pingMu.Unlock()
	c.lastPing = t
}

func (c *WSManager) GetConnState() ConnectionState {
	return ConnectionState(atomic.LoadInt32((*int32)(&c.connState)))
}

func (c *WSConnection) WriteJSON(v any) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	return wsjson.Write(ctx, c.Conn, v)
}

type WsMsg struct {
	Topic string
	Data  interface{}
}

// WSManager manages the websocket connection
type WSManager struct {
	api       *BybitApi
	wsType    WebSocketT
	connState ConnectionState
	Conn      *WSConnection
	DataCh    chan WsMsg
	url       string

	subscriptions   map[string]int32
	subscriptionsMu sync.RWMutex

	requestIds lockfree.HashMap

	reconnectOnce sync.Once

	connMu sync.RWMutex
}

func (m *WSManager) getConn() *WSConnection {
	m.connMu.RLock()
	defer m.connMu.RUnlock()
	return m.Conn
}

func (m *WSManager) setConn(conn *WSConnection) {
	m.connMu.Lock()
	defer m.connMu.Unlock()
	m.Conn = conn
}

/*
 ******************************
* Connection State Transitions *
 ******************************
*/

// Transition from New to Connecting
func (m *WSManager) SetConnecting() bool {
	m.api.Logger.Debug("SetConnecting: %s -> %s", m.GetConnState(), StateConnecting)
	return atomic.CompareAndSwapInt32((*int32)(&m.connState), int32(StateNew), int32(StateConnecting))
}

func (m *WSManager) SetReconnecting() bool {
	m.api.Logger.Debug("SetReconnecting: %s -> %s", m.GetConnState(), StateConnecting)
	return atomic.CompareAndSwapInt32((*int32)(&m.connState), int32(StateReconnecting), int32(StateConnecting))
}

func (m *WSManager) SetReconnectingFromConnected(msg ...string) bool {
	if m == nil {
		return false
	}
	msgStr := ""
	for _, s := range msg {
		msgStr += s + " "
	}
	m.api.Logger.Debug("SetReconnectingFromConnected: %s -> %s ()", m.GetConnState(), StateReconnecting, msgStr)
	return atomic.CompareAndSwapInt32((*int32)(&m.connState), int32(StateConnected), int32(StateReconnecting))
}

func (m *WSManager) SetReconnectingFromConnecting() bool {
	if m == nil {
		return false
	}
	m.api.Logger.Debug("SetReconnectingFromConnecting: %s -> %s", m.GetConnState(), StateReconnecting)
	return atomic.CompareAndSwapInt32((*int32)(&m.connState), int32(StateConnecting), int32(StateReconnecting))
}

func (m *WSManager) SetConnected() bool {
	if m == nil {
		return false
	}
	m.api.Logger.Debug("SetConnected: %s -> %s", m.GetConnState(), StateConnected)
	return atomic.CompareAndSwapInt32((*int32)(&m.connState), int32(StateConnecting), int32(StateConnected))
}

func (m *WSManager) SetDisconnected() bool {
	if m == nil {
		return false
	}
	m.api.Logger.Debug("SetConnected: %s -> %s", m.GetConnState(), StateConnected)
	atomic.StoreInt32((*int32)(&m.connState), int32(StateDisconnected))
	return true
}

// getReqId generates a request id for a given topic
func (m *WSManager) getReqId(topic string) string {
	if n, exist := m.requestIds.Get(topic); exist {
		id := n.(int) + 1
		m.requestIds.Set(topic, id)
		return fmt.Sprintf("%s_%d", topic, id)
	}

	m.requestIds.Set(topic, 1)
	return fmt.Sprintf("%s_%d", topic, 1)
}

func newWSManager(api *BybitApi, wsType WebSocketT, url string) *WSManager {
	wsm := &WSManager{
		api:           api,
		wsType:        wsType,
		subscriptions: make(map[string]int32),
		requestIds:    lockfree.NewHashMap(),
		url:           url,
		connState:     StateNew,

		DataCh: make(chan WsMsg, 100),
	}
	return wsm
}

// ensureConnected will establish a connection if it is not already established
func (m *WSManager) ensureConnected() error {
	if m.GetConnState() == StateNew {
		m.api.Logger.Info("WebSocket not connected, connecting...")
		if err := m.connect(); err != nil {
			return fmt.Errorf("failed to establish connection: %w", err)
		}
		m.reconnectOnce.Do(func() { go m.reconnectLoop() })
	} else {
		fmt.Printf("WebSocket already connected, state: %s\n", m.GetConnState())
	}

	return nil
}

func (m *WSManager) sendAuth() error {
	expires := time.Now().UnixNano()/1e6 + 10000
	val := fmt.Sprintf("GET/realtime%d", expires)

	h := hmac.New(sha256.New, []byte(m.api.ApiSecret))
	h.Write([]byte(val))
	signature := hex.EncodeToString(h.Sum(nil))

	m.api.Logger.Debug("auth args: [%s, %d, %s]", m.api.ApiKey, expires, signature)

	authMessage := map[string]interface{}{
		"req_id": m.getReqId("auth"),
		"op":     "auth",
		"args":   []interface{}{m.api.ApiKey, expires, signature},
	}

	return m.getConn().WriteJSON(authMessage)
}

func (m *WSManager) connect() error {
	currentState := m.GetConnState()
	var transitionOk bool

	switch currentState {
	case StateNew:
		transitionOk = m.SetConnecting()
	case StateReconnecting:
		transitionOk = m.SetConnecting()
	default:
		return fmt.Errorf("cannot connect from state %s", currentState)
	}

	if !transitionOk {
		return fmt.Errorf("state transition failed from %s", currentState)
	}

	m.api.Logger.Info("Connecting to %s", m.url)

	conn := m.getConn()
	if conn != nil {
		if conn.cancel != nil {
			conn.cancel()
		}
		if conn.Conn != nil {
			conn.Conn.CloseNow()
		}
	}

	dialCtx, dialCancel := context.WithTimeout(m.api.context, 15*time.Second)
	defer dialCancel()

	opts := &websocket.DialOptions{
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				TLSHandshakeTimeout:   10 * time.Second,
				ResponseHeaderTimeout: 10 * time.Second,
				DisableKeepAlives:     false,
				MaxIdleConnsPerHost:   -1,
			},
		},
	}

	wsConn, _, err := websocket.Dial(dialCtx, m.url, opts)
	if err != nil {
		m.SetReconnectingFromConnecting()
		return fmt.Errorf("websocket dial failed: %w", err)
	}

	connCtx, cancel := context.WithCancel(m.api.context)

	m.setConn(&WSConnection{
		Conn:       wsConn,
		ctx:        connCtx,
		cancel:     cancel,
		lastPing:   time.Now(),
		subscribed: make(map[string]struct{}),
		writeMu:    sync.Mutex{},
		apiKey:     m.api.ApiKey,
		apiSecret:  m.api.ApiSecret,
	})
	conn = m.getConn()

	defer func() {
		if err != nil && conn != nil {
			if conn.cancel != nil {
				conn.cancel()
			}
			if conn.Conn != nil {
				conn.Conn.CloseNow()
			}
		}
	}()

	if _, exists := AUTH_REQUIRED_TYPES[m.wsType]; exists {
		if conn.apiKey == "" || conn.apiSecret == "" {
			return fmt.Errorf("api key and secret required for private websocket")
		}
		if err := m.sendAuth(); err != nil {
			m.SetReconnectingFromConnected("Send Auth error")
			return fmt.Errorf("failed to authenticate: %w", err)
		}
	}

	m.SetConnected()

	go m.readMessages()
	go m.pingLoop()

	return nil
}

func (m *WSManager) pingLoop() {
	m.api.Logger.Debug("Starting pingLoop")

	ticker := time.NewTicker(wsPingInterval)

	defer ticker.Stop()

	for {
		conn := m.getConn()
		select {

		case <-conn.ctx.Done():
			m.api.Logger.Debug("Ping loop context done, exiting")
			return
		case <-ticker.C:
			if m.GetConnState() == StateReconnecting {
				m.api.Logger.Debug("Disconnected, skipping ping")
				time.Sleep(1 * time.Second)
				continue
			}
			payload := map[string]interface{}{
				"req_id": m.getReqId("ping"),
				"op":     "ping",
			}
			err := conn.WriteJSON(payload)
			if err != nil {
				m.SetReconnectingFromConnected("PingLoop error")
				conn.cancel()
				m.api.Logger.Error("failed to send ping message: %v", err)
				continue
			} else {
				m.api.Logger.Debug("Ping message sent")
			}
		}
	}
}

func (m *WSManager) ResubscribeAll() error {
	m.api.Logger.Info("Resubscribing to all topics")

	time.Sleep(500 * time.Millisecond)
	topics := m.GetSubscribedTopics()
	for _, topic := range topics {
		if m.GetConnState() != StateConnected {
			return fmt.Errorf("not connected, current state: %s", m.GetConnState())
		}

		err := m.sendSubscribe(topic)
		if err != nil {
			m.api.Logger.Error("failed to resubscribe to %s: %v", topic, err)
			return err
		}
		m.api.Logger.Info("Resubscribed to %s", topic)
		time.Sleep(300 * time.Millisecond)
	}
	return nil
}

func (m *WSManager) reconnectLoop() {
	log := m.api.Logger
	log.Debug("Starting reconnect loop")
	defer log.Debug("Exiting reconnect loop")

	backoff := wsInitialReconnectDelay
	ticker := time.NewTicker(1 * time.Second)

	defer ticker.Stop()

	for {
		select {
		case <-m.api.context.Done():
			log.Debug("Reconnect loop context done, exiting")
			return
		case <-ticker.C:

			conn := m.getConn()
			if conn == nil {
				continue
			}
			if delta := time.Now().Sub(conn.getLastPing()); m.GetConnState() == StateConnected && delta > wsMaxSilentPeriod {
				m.api.Logger.Error("No ping received for %v, %v, reconnecting...", wsMaxSilentPeriod, delta)
				m.SetReconnectingFromConnected("ReconnectLoop: ping timeout")
			}

			if m.GetConnState() != StateReconnecting {
				continue
			}

			m.api.Logger.Debug("reconnecting websocket")
			err := m.connect()
			if err != nil {
				m.api.Logger.Error("reconnect loop: failed to reconnect %v", err)

				select {
				case <-time.After(backoff):
					backoff = time.Duration(float64(backoff) * wsReconnectBackoffFactor)
					if backoff > wsMaxReconnectDelay {
						backoff = wsMaxReconnectDelay
					}
				}
				continue
			}

			backoff = wsInitialReconnectDelay

			err = m.ResubscribeAll()
			if err != nil {
				m.SetReconnectingFromConnected("ReconnectLoop: ResubscribeAll failed")
				m.getConn().cancel()
				log.Error("failed to resubscribe after reconnect: %v", err)
				continue
			}
		}
	}
}

func (m *WSManager) readMessages() {
	defer func() {
		if r := recover(); r != nil {
			m.api.Logger.Error("recovered from panic in readMessages: %v", r)
		}
	}()
	conn := m.getConn()

	if conn == nil || conn.ctx == nil {
		m.api.Logger.Error("readMessages: connection or context is nil")
		return
	}

	m.api.Logger.Debug("Starting ReadMessages")
	defer m.api.Logger.Debug("Exiting ReadMessages")

	for {
		select {
		case <-conn.ctx.Done():
			m.api.Logger.Debug("Read messages context done, exiting")
			return
		default:
			if m.GetConnState() != StateConnected {
				time.Sleep(1 * time.Second)
				continue
			}

			func() {
				defer func() {
					if r := recover(); r != nil {
						m.api.Logger.Error("recovered from panic in ReadMessage: %v", r)
						m.SetReconnectingFromConnected("readMessages: Panic recover")
					}
				}()

				readDataCtx, cancelData := context.WithCancel(conn.ctx)
				defer cancelData()

				_, reader, err := conn.Conn.Reader(readDataCtx)
				if err != nil {
					if strings.Contains(err.Error(), "context deadline exceeded") {
						return
					} else if strings.Contains(err.Error(), "context canceled") {
						return
					}
					m.api.Logger.Error("failed to get reader: %v", err)
					return
				}

				b := bpool.Get()
				defer bpool.Put(b)

				_, err = b.ReadFrom(reader)
				if err != nil {
					if strings.Contains(err.Error(), "context deadline exceeded") {
						m.api.Logger.Error("read timeout")
						m.SetReconnectingFromConnected("Read timeout")
						return
					}
					m.api.Logger.Error("failed to read message: %v", err)
					return
				}

				heardersSerialized := struct {
					Topic string `json:"topic"`
					Op    string `json:"op"`
				}{}

				if err := json.Unmarshal(b.Bytes(), &heardersSerialized); err != nil {
					m.api.Logger.Error("failed to get topic: %v", err)
					return
				}
				serialized, err := m.serializeWsResponse(heardersSerialized.Topic, b.Bytes())
				if err != nil {
					m.api.Logger.Error("failed to serialize message: %v", err)
					return
				}

				serializedMsg := WsMsg{Topic: heardersSerialized.Topic, Data: serialized}

				if heardersSerialized.Op == "ping" || heardersSerialized.Op == "pong" {
					conn.setLastPing(time.Now())
					return
				}

				select {
				case m.DataCh <- serializedMsg:
					m.api.Logger.Debug("received message: %s", pp.PrettyFormat(serialized))
				default:
					m.api.Logger.Error("message buffer full, dropping message")
				}
			}()
		}
	}
}

func (m *WSManager) serializeWsResponse(topic string, data []byte) (interface{}, error) {
	parts := strings.Split(topic, ".")
	mainTopic := parts[0]

	switch mainTopic {
	case "wallet":
		var response WalletWebsocketResponse
		if err := json.Unmarshal(data, &response); err != nil {
			return nil, fmt.Errorf("failed to unmarshal wallet response: %w", err)
		}
		return response, nil

	case "tickers":
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid ticker topic format: %s", topic)
		}

		var response LinearInverseTicker
		if err := json.Unmarshal(data, &response); err != nil {
			return nil, fmt.Errorf("failed to unmarshal %s ticker: %w", parts[1], err)
		}
		return response, nil

	case "orderbook":
		var response GetOrderbookResponse
		if err := json.Unmarshal(data, &response); err != nil {
			return nil, fmt.Errorf("failed to unmarshal orderbook: %w", err)
		}
		return response, nil

	case "kline":
		var response KlineWsResponse
		if err := json.Unmarshal(data, &response); err != nil {
			return nil, fmt.Errorf("failed to unmarshal kline: %w", err)
		}
		return response, nil

	case "order":
		var response OrderWebsocketResponse
		if err := json.Unmarshal(data, &response); err != nil {
			return nil, fmt.Errorf("failed to unmarshal order: %w", err)
		}
		return response, nil

	case "position":
		var response PositionWebsocketResponse
		if err := json.Unmarshal(data, &response); err != nil {
			return nil, fmt.Errorf("failed to unmarshal position: %w", err)
		}
		return response, nil

	case "execution":
		var response ExecutionWebsocketResponse
		if err := json.Unmarshal(data, &response); err != nil {
			return nil, fmt.Errorf("failed to unmarshal execution: %w", err)
		}
		return response, nil

	case "pong":
		return "pong", nil

	default:
		var response map[string]interface{}
		if err := json.Unmarshal(data, &response); err != nil {
			return nil, fmt.Errorf("failed to unmarshal unknown topic %s: %w", topic, err)
		}
		return response, nil
	}
}

func (m *WSManager) close() error {
	m.subscriptionsMu.Lock()
	defer m.subscriptionsMu.Unlock()

	conn := m.getConn()
	if conn != nil {
		if conn.cancel != nil {
			conn.cancel()
		}
		if conn.Conn != nil {
			conn.Conn.CloseNow()
		}
	}
	m.SetDisconnected()
	return nil
}

func (m *WSManager) sendSubscribe(topic string) error {
	if m.api.UrlSet != true {
		return ERR_URLS_NOT_CONFIGURED
	}
	if m.url == "" {
		return ERR_TRADING_STREAMS_NOT_SUPPORTED
	}
	m.api.Logger.Debug("Subscribing to %s", topic)
	return m.getConn().WriteJSON(map[string]interface{}{
		"req_id": m.getReqId("subscribe"),
		"op":     "subscribe",
		"args":   []string{topic},
	})
}

func (m *WSManager) sendUnsubscribe(topic string) error {
	if m.api.UrlSet != true {
		return ERR_URLS_NOT_CONFIGURED
	}
	return m.getConn().WriteJSON(map[string]interface{}{
		"req_id": m.getReqId("unsubscribe"),
		"op":     "unsubscribe",
		"args":   []string{topic},
	})
}

// Subscribe to a topic
// https://bybit-exchange.github.io/docs/v5/ws/connect#how-to-subscribe-to-topics
func (m *WSManager) Subscribe(topic string) error {
	m.subscriptionsMu.Lock()
	defer m.subscriptionsMu.Unlock()

	if m.subscriptions[topic] >= 1 {
		m.subscriptions[topic]++
		return nil
	}

	if err := m.ensureConnected(); err != nil {
		return err
	}

	if err := m.sendSubscribe(topic); err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}
	m.subscriptions[topic] = 1
	return nil
}

func (m *WSManager) Unsubscribe(topic string) error {
	m.subscriptionsMu.Lock()
	defer m.subscriptionsMu.Unlock()
	if m.subscriptions[topic] > 1 {
		m.subscriptions[topic]--
		return nil
	} else if m.subscriptions[topic] == 1 {
		delete(m.subscriptions, topic)
	} else {
		m.api.Logger.Error("Not subscribed to %s", topic)
		return nil
	}

	if err := m.sendUnsubscribe(topic); err != nil {
		m.api.Logger.Error("Failed to unsubscribe from %s: %v", topic, err)
	}

	return nil
}

// GetSubscribedTopics returns a list of topics
// that the client is currently subscribed to
func (m *WSManager) GetSubscribedTopics() []string {
	m.subscriptionsMu.RLock()
	defer m.subscriptionsMu.RUnlock()

	topics := make([]string, 0, len(m.subscriptions))
	for topic := range m.subscriptions {
		topics = append(topics, topic)
	}
	return topics
}

func (m *WSManager) sendRequest(operation string, request interface{}, headers map[string]interface{}) error {
	if m.url == "" {
		return ERR_URLS_NOT_CONFIGURED
	}
	if err := m.ensureConnected(); err != nil {
		return err
	}

	reqId := uuid.New().String()
	message := map[string]interface{}{

		"reqId":  reqId,
		"header": headers,
		"op":     operation,
		"args":   []interface{}{request},
	}
	return m.getConn().WriteJSON(message)
}
