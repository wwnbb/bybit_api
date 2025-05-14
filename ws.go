/*
 * Copyright (c) 2024 QuantHill <info@quanthill.ae>
 */
package qant_api_bybit

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
	gws "github.com/gorilla/websocket"
	pp "github.com/wwnbb/pprint"
)

const (
	wsInitialReconnectDelay  = 1 * time.Second
	wsMaxReconnectDelay      = 5 * time.Minute
	wsReconnectBackoffFactor = 2.0
	wsMaxSilentPeriod        = 10 * time.Second
	wsPingInterval           = 20 * time.Second
)

// WSConnection represents a websocket connection,
// it embeds the gorilla websocket connection
// creation done in connect method of WSManager
type WSConnection struct {
	*gws.Conn
	state      atomic.Int32
	mu         sync.Mutex
	ctx        context.Context
	cancel     context.CancelFunc
	lastPing   time.Time
	subscribed map[string]struct{}
	apiKey     string
	apiSecret  string
	writeMu    sync.Mutex
}

func (c *WSConnection) GetState() ConnectionState {
	return ConnectionState(c.state.Load())
}

func (c *WSConnection) setState(state ConnectionState) {
	c.state.Store(int32(state))
}

// Add thread-safe write method
func (c *WSConnection) WriteJSONThreadSafe(v interface{}) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	return c.WriteJSON(v)
}

type WsMsg struct {
	Topic string
	Data  interface{}
}

// WSManager manages the websocket connection
type WSManager struct {
	api    *BybitApi
	wsType WebSocketT
	Conn   *WSConnection
	DataCh chan WsMsg
	url    string

	subscriptions   map[string]struct{}
	subscriptionsMu sync.RWMutex

	reconnectDelay time.Duration

	requestIds lockfree.HashMap
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
		api:            api,
		wsType:         wsType,
		subscriptions:  make(map[string]struct{}),
		reconnectDelay: wsInitialReconnectDelay,
		requestIds:     lockfree.NewHashMap(),
		url:            url,

		DataCh: make(chan WsMsg, 5000),
	}
	return wsm
}

// ensureConnected will establish a connection if it is not already established
func (m *WSManager) ensureConnected(ctx context.Context) error {
	if m.Conn == nil {
		if err := m.connect(ctx); err != nil {
			return fmt.Errorf("failed to establish connection: %w", err)
		}
		go m.reconnectLoop(ctx)
		if _, exists := AUTH_REQUIRED_TYPES[m.wsType]; exists {
			if m.Conn.apiKey == "" || m.Conn.apiSecret == "" {
				return fmt.Errorf("api key and secret required for private websocket")
			}
			if err := m.sendAuth(); err != nil {
				m.Conn.setState(StatusDisconnected)
				return fmt.Errorf("failed to authenticate: %w", err)
			}
		}
	}

	if m.Conn.GetState() != StatusConnected {
		return fmt.Errorf("websocket connection not in connected state")
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
	if m.Conn != nil && m.Conn.GetState() == StatusConnected {
		return fmt.Errorf("websocket connection already established")
	}

	connCtx, cancel := context.WithCancel(ctx)
	wsConn, _, err := gws.DefaultDialer.DialContext(connCtx, m.url, nil)
	if err != nil {
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

	m.Conn.state.Store(int32(StatusConnected))

	go m.readMessages(ctx)
	go m.pingLoop(ctx)

	return nil
}

// pingLoop will send a ping message to the server
func (m *WSManager) pingLoop(ctx context.Context) {
	ticker := time.NewTicker(wsPingInterval)

	defer ticker.Stop()
	fmt.Println("ping loop", ctx)

	for {
		select {
		case <-ticker.C:
			if m.Conn.GetState() == StatusDisconnected {
				time.Sleep(1 * time.Second)
				continue
			}
			payload := map[string]interface{}{
				"req_id": m.getReqId("ping"),
				"op":     "ping",
			}
			if err := m.Conn.WriteJSONThreadSafe(payload); err != nil {
				m.Conn.setState(StatusDisconnected)
				m.api.Logger.Error("failed to send ping message: %v", err)
				continue
			} else {
				m.api.Logger.Debug("Ping message sent")
			}
		case <-ctx.Done():
			return
		}
	}
}

// reconnectLoop will attempt to reconnect the websocket connection
func (m *WSManager) reconnectLoop(ctx context.Context) {
	backoff := wsInitialReconnectDelay
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if m.Conn == nil || m.Conn.GetState() == StatusDisconnected {
				m.api.Logger.Debug("reconnecting websocket")

				if err := m.connect(ctx); err != nil {
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

				topics := m.GetSubscribedTopics()
				for _, topic := range topics {
					if err := m.Subscribe(topic); err != nil {
						m.api.Logger.Error("failed to resubscribe to %s: %v", topic, err)
					}
					time.Sleep(300 * time.Millisecond)
				}
			}
		}
	}
}

func (m *WSManager) readMessages(ctx context.Context) {
	defer func() {
		m.Conn.setState(StatusDisconnected)
		if m.Conn != nil {
			m.Conn.state.Store(int32(StatusDisconnected))
			m.Conn.Close()
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			if m.Conn == nil {
				return
			}

			_, message, err := m.Conn.ReadMessage()
			if err != nil {
				if gws.IsUnexpectedCloseError(err,
					gws.CloseGoingAway,
					gws.CloseAbnormalClosure) {
					m.api.Logger.Error("read error: %v", err)
				}
				return
			}

			m.Conn.lastPing = time.Now()

			topicStruct := struct {
				Topic string `json:"topic"`
			}{}

			if err := json.Unmarshal(message, &topicStruct); err != nil {
				m.api.Logger.Error("failed to get topic: %v", err)
			}
			topic := topicStruct.Topic

			if topic == "pong" {
				m.api.Logger.Debug("received pong message")
				continue
			}

			// Convert byte array to serialized struct
			serialized, err := m.serializeWsResponse(topic, message)
			if err != nil {
				m.api.Logger.Error("failed to serialize message: %v", err)
				continue
			}

			select {
			case m.DataCh <- WsMsg{Topic: topic, Data: serialized}:
				m.api.Logger.Debug("received message: %s", pp.PrettyFormat(serialized))
			default:
				m.api.Logger.Error("message buffer full, dropping message")
			}
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
		// Add position response struct when available
		var response PositionWebsocketResponse
		if err := json.Unmarshal(data, &response); err != nil {
			return nil, fmt.Errorf("failed to unmarshal position: %w", err)
		}
		return response, nil

	case "execution":
		// Add execution response struct when available
		var response map[string]interface{}
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
		m.Conn.cancel()
		if m.Conn != nil {
			return m.Conn.Close()
		}
	}

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

	if err := m.ensureConnected(m.api.context); err != nil {
		return err
	}

	if err := m.sendSubscribe(topic); err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}
	m.subscriptions[topic] = struct{}{}

	return nil
}

func (m *WSManager) Unsubscribe(topic string) error {
	m.subscriptionsMu.Lock()
	defer m.subscriptionsMu.Unlock()

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

// TODO: replace with CancelOrderParams
func (m *WSManager) CancelOrder(request PlaceOrderParams) error {
	return m.sendRequest("order.amend", request, nil)
}

func (m *WSManager) CancelAllOrders(request AmendOrderParams) error {
	return m.sendRequest("order.cancel", request, nil)
}
