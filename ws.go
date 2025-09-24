package bybit_api

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dustinxie/lockfree"
	json "github.com/goccy/go-json"
	"github.com/google/uuid"
	"nhooyr.io/websocket"
	pp "github.com/wwnbb/pprint"
)

const (
	wsInitialReconnectDelay  = 5 * time.Second
	wsMaxReconnectDelay      = 5 * time.Minute
	wsReconnectBackoffFactor = 2.0
	wsMaxSilentPeriod        = 30 * time.Second
	wsPingInterval           = 10 * time.Second
)

// WSConnection represents a websocket connection,
// it embeds the coder websocket connection  
// creation done in connect method of WSManager
type WSConnection struct {
	*websocket.Conn
	ctx        context.Context
	cancel     context.CancelFunc
	lastPing   time.Time
	subscribed map[string]struct{}
	apiKey     string
	apiSecret  string
	writeMu    sync.Mutex

	pingMu sync.Mutex
}

func (c *WSConnection) GetLastPing() time.Time {
	c.pingMu.Lock()
	defer c.pingMu.Unlock()
	return c.lastPing
}

func (c *WSConnection) SetLastPing(t time.Time) {
	c.pingMu.Lock()
	defer c.pingMu.Unlock()
	c.lastPing = t
}

func (c *WSManager) GetConnState() ConnectionState {
	return ConnectionState(atomic.LoadInt32((*int32)(&c.connState)))
}

func (c *WSConnection) WriteJSONThreadSafe(v interface{}) error {
	c.writeMu.Lock()  
	defer c.writeMu.Unlock()
	
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	
	return c.Write(c.ctx, websocket.MessageText, data)
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
}

/*
 ******************************
* Connection State Transitions *
 ******************************
*/

// Transition from New to Connecting
func (c *WSManager) SetConnecting() bool {
	return atomic.CompareAndSwapInt32((*int32)(&c.connState), int32(StateNew), int32(StateConnecting))
}

func (c *WSManager) SetReconnecting() bool {
	return atomic.CompareAndSwapInt32((*int32)(&c.connState), int32(StateDisconnected), int32(StateConnecting))
}

func (c *WSManager) SetDisconnectedFromConnected() bool {
	if c == nil {
		return false
	}
	return atomic.CompareAndSwapInt32((*int32)(&c.connState), int32(StateConnected), int32(StateDisconnected))
}

func (c *WSManager) SetDisconnectedFromConnecting() bool {
	if c == nil {
		return false
	}
	return atomic.CompareAndSwapInt32((*int32)(&c.connState), int32(StateConnecting), int32(StateDisconnected))
}

func (c *WSManager) SetConnected() bool {
	if c == nil {
		return false
	}
	return atomic.CompareAndSwapInt32((*int32)(&c.connState), int32(StateConnecting), int32(StateConnected))
}

func (c *WSManager) SetReadyForConnecting() bool {
	atomic.StoreInt32((*int32)(&c.connState), int32(StateNew))
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

		DataCh: make(chan WsMsg, 5000),
	}
	return wsm
}

// ensureConnected will establish a connection if it is not already established
func (m *WSManager) ensureConnected(ctx context.Context) error {
	if m.GetConnState() == StateNew {
		if err := m.connect(ctx); err != nil {
			return fmt.Errorf("failed to establish connection: %w", err)
		}
		go m.reconnectLoop(ctx)
		go m.readMessages(ctx)
		go m.pingLoop(ctx)
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

	return m.Conn.WriteJSONThreadSafe(authMessage)
}

// connect to the websocket server
func (m *WSManager) connect(ctx context.Context) error {
	if ok := m.SetConnecting() || m.SetReconnecting(); !ok {
		m.api.Logger.Error("Connection State transition error: cannot set to connecting from state %s", m.GetConnState())
		return nil
	}

	m.api.Logger.Info("Connecting to %s", m.url)
	connCtx, cancel := context.WithCancel(ctx)
	wsConn, _, err := websocket.Dial(connCtx, m.url, nil)
	if err != nil {
		m.SetDisconnectedFromConnecting()
		cancel()
		return fmt.Errorf("websocket dial failed: %w", err)
	}
	m.Conn = &WSConnection{
		Conn:       wsConn,
		ctx:        connCtx,
		cancel:     cancel,
		lastPing:   time.Now(),
		subscribed: make(map[string]struct{}),
		writeMu:    sync.Mutex{},
		apiKey:     m.api.ApiKey,
		apiSecret:  m.api.ApiSecret,
	}

	if _, exists := AUTH_REQUIRED_TYPES[m.wsType]; exists {
		if m.Conn.apiKey == "" || m.Conn.apiSecret == "" {
			return fmt.Errorf("api key and secret required for private websocket")
		}
		if err := m.sendAuth(); err != nil {
			m.SetDisconnectedFromConnected()
			return fmt.Errorf("failed to authenticate: %w", err)
		}
	}

	m.SetConnected()

	return nil
}

// pingLoop will send a ping message to the server
func (m *WSManager) pingLoop(ctx context.Context) {
	ticker := time.NewTicker(wsPingInterval)

	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Skip ping if disconnected
			if m.GetConnState() == StateDisconnected {
				m.api.Logger.Debug("Disconnected, skipping ping")
				time.Sleep(1 * time.Second)
				continue
			}
			payload := map[string]interface{}{
				"req_id": m.getReqId("ping"),
				"op":     "ping",
			}
			err := m.Conn.WriteJSONThreadSafe(payload)
			if err != nil {
				m.SetDisconnectedFromConnected()
				m.api.Logger.Error("failed to send ping message: %v", err)
				continue
			} else {
				m.api.Logger.Debug("Ping message sent")
			}
		case <-ctx.Done():
			m.api.Logger.Debug("Ping loop context done, exiting")
			return
		}
	}
}

func (m *WSManager) ResubscribeAll() error {
	m.api.Logger.Info("Resubscribing to all topics")
	topics := m.GetSubscribedTopics()
	for _, topic := range topics {
		err := m.Subscribe(topic)
		if err != nil {
			m.api.Logger.Error("failed to resubscribe to %s: %v", topic, err)
			return err
		}
		m.api.Logger.Info("Resubscribed to %s", topic)
		time.Sleep(300 * time.Millisecond)
	}
	return nil
}

// reconnectLoop will attempt to reconnect the websocket connection
func (m *WSManager) reconnectLoop(ctx context.Context) {
	log := m.api.Logger
	log.Debug("Starting reconnect loop")

	backoff := wsInitialReconnectDelay
	ticker := time.NewTicker(1 * time.Second)

	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Debug("Reconnect loop context done, exiting")
			return
		case <-ticker.C:
			if delta := time.Now().Sub(m.Conn.GetLastPing()); delta > wsMaxSilentPeriod {
				m.api.Logger.Error("No ping received for %v, %v, reconnecting...", wsMaxSilentPeriod, delta)
				m.SetDisconnectedFromConnected()
			}

			if m.GetConnState() != StateDisconnected {
				continue
			}

			m.api.Logger.Debug("reconnecting websocket")
			err := m.connect(ctx)
			if err != nil {
				m.api.Logger.Error("failed to reconnect: %v", err)

				select {
				case <-ctx.Done():
					return
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
				m.SetDisconnectedFromConnected()
				log.Error("failed to resubscribe after reconnect: %v", err)
				continue
			}
		}
	}
}

func (m *WSManager) readMessages(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			m.api.Logger.Error("recovered from panic in readMessages: %v", r)
		}
		if m.Conn != nil {
			m.Conn.Close(websocket.StatusInternalError, "read loop error")
		}
		m.SetDisconnectedFromConnected()
	}()

	for {
		select {
		case <-ctx.Done():
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
						m.SetDisconnectedFromConnected()
					}
				}()

				_, message, err := m.Conn.Read(ctx)
				if err != nil {
					if websocket.CloseStatus(err) != websocket.StatusNormalClosure &&
						websocket.CloseStatus(err) != websocket.StatusGoingAway {
						m.api.Logger.Error("read error: %v", err)
					}
					m.SetDisconnectedFromConnected()
					return
				}

				heardersSerialized := struct {
					Topic string `json:"topic"`
					Op    string `json:"op"`
				}{}

				if err := json.Unmarshal(message, &heardersSerialized); err != nil {
					m.api.Logger.Error("failed to get topic: %v", err)
					return
				}
				serialized, err := m.serializeWsResponse(heardersSerialized.Topic, message)
				if err != nil {
					m.api.Logger.Error("failed to serialize message: %v", err)
					return
				}

				serializedMsg := WsMsg{Topic: heardersSerialized.Topic, Data: serialized}

				if heardersSerialized.Op == "ping" || heardersSerialized.Op == "pong" {
					m.Conn.SetLastPing(time.Now())
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
	// For topics with prefixes, extract the main topic and subtopic
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

	for topic := range m.subscriptions {
		delete(m.subscriptions, topic)
	}

	if m.Conn != nil {
		// Cancel the context first to signal shutdown
		m.Conn.cancel()
		// Close the websocket connection with proper status
		err := m.Conn.Close(websocket.StatusNormalClosure, "closing connection")
		m.SetReadyForConnecting()
		return err
	}
	m.SetReadyForConnecting()
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
	return m.Conn.WriteJSONThreadSafe(map[string]interface{}{
		"req_id": m.getReqId("subscribe"),
		"op":     "subscribe",
		"args":   []string{topic},
	})
}

func (m *WSManager) sendUnsubscribe(topic string) error {
	if m.api.UrlSet != true {
		return ERR_URLS_NOT_CONFIGURED
	}
	return m.Conn.WriteJSONThreadSafe(map[string]interface{}{
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

	if err := m.ensureConnected(m.api.context); err != nil {
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
	if err := m.ensureConnected(m.api.context); err != nil {
		return err
	}

	reqId := uuid.New().String()
	message := map[string]interface{}{

		"reqId":  reqId,
		"header": headers,
		"op":     operation,
		"args":   []interface{}{request},
	}
	return m.Conn.WriteJSONThreadSafe(message)
}

func (m *WSManager) PlaceOrder(request PlaceOrderParams) error {
	return m.sendRequest("order.create", request, nil)
}
