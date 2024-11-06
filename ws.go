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
	"sync"
	"sync/atomic"
	"time"

	"github.com/dustinxie/lockfree"
	"github.com/google/uuid"
	gws "github.com/gorilla/websocket"
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

func (c *WSConnection) GetWsState() ConnectionState {
	return ConnectionState(c.state.Load())
}

func (c *WSConnection) SetState(state ConnectionState) {
	c.state.Store(int32(state))
}

// Add thread-safe write method
func (c *WSConnection) WriteJSONThreadSafe(v interface{}) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	return c.WriteJSON(v)
}

// WSManager manages the websocket connection
type WSManager struct {
	api    *BybitApi
	wsType WSType
	conn   *WSConnection
	DataCh chan []byte
	url    string

	subscriptions   map[string]chan []byte
	subscriptionsMu sync.RWMutex

	isConnected    atomic.Bool
	reconnectDelay time.Duration

	done       chan struct{}
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

func newWSManager(api *BybitApi, wsType WSType, url string) *WSManager {
	wsm := &WSManager{
		api:            api,
		wsType:         wsType,
		subscriptions:  make(map[string]chan []byte),
		reconnectDelay: wsInitialReconnectDelay,
		done:           make(chan struct{}),
		requestIds:     lockfree.NewHashMap(),
		DataCh:         make(chan []byte, 5000),
		url:            url,
	}

	return wsm
}

// EnsureConnected will establish a connection if it is not already established
func (m *WSManager) EnsureConnected(ctx context.Context) error {
	if m.conn == nil {
		if err := m.Connect(ctx); err != nil {
			return fmt.Errorf("failed to establish connection: %w", err)
		}
		if _, exists := AUTH_REQUIRED_TYPES[m.wsType]; exists {
			if m.conn.apiKey == "" || m.conn.apiSecret == "" {
				return fmt.Errorf("api key and secret required for private websocket")
			}
			if err := m.sendAuth(); err != nil {
				m.conn.SetState(StatusDisconnected)
				return fmt.Errorf("failed to authenticate: %w", err)
			}
		}
	}

	if m.conn.GetWsState() != StatusConnected {
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

	m.api.logger.Debug("auth args: [%s, %d, %s]", m.api.ApiKey, expires, signature)

	authMessage := map[string]interface{}{
		"req_id": m.getReqId("auth"),
		"op":     "auth",
		"args":   []interface{}{m.api.ApiKey, expires, signature},
	}

	return m.conn.WriteJSONThreadSafe(authMessage)
}

// Connect to the websocket server
func (m *WSManager) Connect(ctx context.Context) error {
	if !m.isConnected.CompareAndSwap(false, true) {
		return fmt.Errorf("websocket connection already established")
	}

	connCtx, cancel := context.WithCancel(ctx)
	wsConn, _, err := gws.DefaultDialer.DialContext(connCtx, m.url, nil)
	if err != nil {
		cancel()
		m.isConnected.Store(false)
		return fmt.Errorf("websocket dial failed: %w", err)
	}

	m.conn = &WSConnection{
		Conn:       wsConn,
		ctx:        connCtx,
		cancel:     cancel,
		lastPing:   time.Now(),
		subscribed: make(map[string]struct{}),
		writeMu:    sync.Mutex{},
		apiKey:     m.api.ApiKey,
		apiSecret:  m.api.ApiSecret,
	}

	m.conn.state.Store(int32(StatusConnected))

	go m.readMessages(ctx)
	go m.pingLoop(ctx)
	go m.reconnectLoop(ctx)

	return nil
}

// pingLoop will send a ping message to the server
func (m *WSManager) pingLoop(ctx context.Context) {
	ticker := time.NewTicker(wsPingInterval)

	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if m.conn.GetWsState() == StatusDisconnected {
				time.Sleep(1 * time.Second)
				continue
			}
			payload := map[string]interface{}{
				"req_id": m.getReqId("ping"),
				"op":     "ping",
			}
			if err := m.conn.WriteJSONThreadSafe(payload); err != nil {
				m.conn.SetState(StatusDisconnected)
				m.api.logger.Error("failed to send ping message: %v", err)
				continue
			} else {
				m.api.logger.Debug("Ping message sent")
			}
		case <-ctx.Done():
			return
		}
	}
}

// reconnectLoop will attempt to reconnect the websocket connection
func (m *WSManager) reconnectLoop(ctx context.Context) {
	backoff := wsInitialReconnectDelay

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.done:
			return
		default:
			if m.conn == nil || m.conn.GetWsState() == StatusDisconnected {
				m.api.logger.Debug("reconnecting websocket")

				if err := m.Connect(ctx); err != nil {
					m.api.logger.Error("failed to reconnect: %v", err)

					select {
					case <-ctx.Done():
						return
					case <-m.done:
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
						m.api.logger.Error("failed to resubscribe to %s: %v", topic, err)
					}
					time.Sleep(300 * time.Millisecond)
				}
			}
		}
	}
}

func (m *WSManager) readMessages(ctx context.Context) {
	defer func() {
		m.isConnected.Store(false)
		if m.conn != nil {
			m.conn.state.Store(int32(StatusDisconnected))
			m.conn.Close()
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.done:
			return
		default:
			if m.conn == nil {
				return
			}

			_, message, err := m.conn.ReadMessage()
			if err != nil {
				if gws.IsUnexpectedCloseError(err,
					gws.CloseGoingAway,
					gws.CloseAbnormalClosure) {
					m.api.logger.Error("read error: %v", err)
				}
				return
			}

			m.conn.lastPing = time.Now()

			select {
			case m.DataCh <- message:
				m.api.logger.Debug("received message: %s", message)
			default:
				m.api.logger.Error("message buffer full, dropping message")
			}
		}
	}
}

func (m *WSManager) Close() error {
	close(m.done)

	m.subscriptionsMu.Lock()
	defer m.subscriptionsMu.Unlock()

	for topic, ch := range m.subscriptions {
		close(ch)
		delete(m.subscriptions, topic)
	}

	if m.conn != nil {
		m.conn.cancel()
		if m.conn != nil {
			return m.conn.Close()
		}
	}

	return nil
}

func (m *WSManager) sendSubscribe(topic string) error {
	m.api.logger.Debug("Subscribing to %s", topic)
	return m.conn.WriteJSONThreadSafe(map[string]interface{}{
		"req_id": m.getReqId("subscribe"),
		"op":     "subscribe",
		"args":   []string{topic},
	})
}

func (m *WSManager) sendUnsubscribe(topic string) error {
	return m.conn.WriteJSONThreadSafe(map[string]interface{}{
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

	if err := m.EnsureConnected(context.Background()); err != nil {
		return err
	}

	if err := m.sendSubscribe(topic); err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	return nil
}

func (m *WSManager) Unsubscribe(topic string) error {
	m.subscriptionsMu.Lock()
	defer m.subscriptionsMu.Unlock()

	if err := m.sendUnsubscribe(topic); err != nil {
		m.api.logger.Error("Failed to unsubscribe from %s: %v", topic, err)
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
	if err := m.EnsureConnected(context.Background()); err != nil {
		return err
	}

	reqId := uuid.New().String()
	message := map[string]interface{}{

		"reqId":  reqId,
		"header": headers,
		"op":     operation,
		"args":   []interface{}{request},
	}
	PrettyPrint(message)

	return m.conn.WriteJSONThreadSafe(message)
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
