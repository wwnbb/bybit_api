package bybit_api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/dustinxie/lockfree"
	json "github.com/goccy/go-json"
	"github.com/google/uuid"
	pp "github.com/wwnbb/pprint"
	"github.com/wwnbb/wsmanager"
)

type ResponseHeader struct {
	Topic string `json:"topic"`
	Op    string `json:"op"`
}

type WsMsg struct {
	Topic string
	Op    string
	Data  interface{}
}

type AuthCredentials struct {
	ApiKey    string
	ApiSecret string
}

type WSBybit struct {
	*wsmanager.WSManager
	api    *BybitApi
	wsType WebSocketT

	subscriptions   map[string]int32
	subscriptionsMu sync.RWMutex

	requestIds lockfree.HashMap

	authCredentials AuthCredentials

	DataCh chan WsMsg
}

func (m *WSBybit) GetReqId(topic string) string {
	if n, exist := m.requestIds.Get(topic); exist {
		id := n.(int) + 1
		m.requestIds.Set(topic, id)
		return fmt.Sprintf("%s_%d", topic, id)
	}

	m.requestIds.Set(topic, 1)
	return fmt.Sprintf("%s_%d", topic, 1)
}

func newWSBybit(api *BybitApi, wsType WebSocketT, url string) *WSBybit {
	wsm := wsmanager.NewWSManager(url, api.context)
	wsm.SetLogger(api.Logger)

	ws := &WSBybit{
		WSManager:     wsm,
		api:           api,
		wsType:        wsType,
		subscriptions: make(map[string]int32),
		requestIds:    lockfree.NewHashMap(),
		DataCh:        make(chan WsMsg, 100),
		authCredentials: AuthCredentials{
			ApiKey:    api.ApiKey,
			ApiSecret: api.ApiSecret,
		},
	}

	go ws.processMessages()

	return ws
}

func (m *WSBybit) processMessages() {
	for {
		select {
		case <-m.api.context.Done():
			return
		case rawMsg := <-m.WSManager.DataCh:
			// Extract topic and op fields without using interface{}
			var header ResponseHeader
			if err := json.Unmarshal(rawMsg, &header); err != nil {
				m.Logger.Error("failed to unmarshal message header", "error", err)
				continue
			}

			if header.Op == "ping" || header.Op == "pong" {
				continue
			}

			// rawMsg is already raw JSON bytes, use it directly
			serialized, err := m.serializeWsResponse(header.Topic, header.Op, rawMsg)
			if err != nil {
				m.Logger.Error("failed to serialize message", "error", err, "topic", header.Topic, "op", header.Op)
				continue
			}

			msg := WsMsg{
				Topic: header.Topic,
				Op:    header.Op,
				Data:  serialized,
			}

			select {
			case m.DataCh <- msg:
				m.Logger.Debug("processed message", "topic", header.Topic, "op", header.Op, "data", pp.PrettyFormat(serialized))
			default:
				m.Logger.Error("message buffer full, dropping message")
			}
		}
	}
}

func (m *WSBybit) SetAuthCredentials(apiKey, apiSecret string) {
	m.authCredentials = AuthCredentials{
		ApiKey:    apiKey,
		ApiSecret: apiSecret,
	}
}

func (m *WSBybit) sendAuth() error {
	expires := time.Now().UnixMilli() + 10000
	val := fmt.Sprintf("GET/realtime%d", expires)

	h := hmac.New(sha256.New, []byte(m.authCredentials.ApiSecret))
	h.Write([]byte(val))
	signature := hex.EncodeToString(h.Sum(nil))

	m.Logger.Debug("auth args", "apiKey", m.authCredentials.ApiKey, "expires", expires, "signature", signature)

	authMessage := map[string]interface{}{
		"req_id": m.GetReqId("auth"),
		"op":     "auth",
		"args":   []interface{}{m.authCredentials.ApiKey, expires, signature},
	}

	return m.WSManager.SendRequest(authMessage)
}

func (m *WSBybit) Connect() error {
	err := m.WSManager.Connect()
	if err != nil {
		return err
	}

	if m.wsType == WS_PRIVATE || m.wsType == WS_TRADE {
		if m.authCredentials.ApiKey != "" && m.authCredentials.ApiSecret != "" {
			time.Sleep(500 * time.Millisecond)
			if err := m.sendAuth(); err != nil {
				m.Logger.Error("failed to authenticate", "error", err)
				return err
			}
		}
	}

	return nil
}

func (m *WSBybit) serializeWsResponse(topic, op string, data []byte) (interface{}, error) {
	if len(topic) == 0 && len(op) != 0 {
		switch op {
		case "order.create":
			var response OrderWebsocketCreateResponse
			if err := json.Unmarshal(data, &response); err != nil {
				return nil, fmt.Errorf("failed to unmarshal wallet response: %w", err)
			}
			return response, nil
		case "order.cancel":
			var response OrderWebsocketCancelResponse
			if err := json.Unmarshal(data, &response); err != nil {
				return nil, fmt.Errorf("failed to unmarshal wallet response: %w", err)
			}
			return response, nil
		case "order.amend":
			var response OrderWebsocketAmendResponse
			if err := json.Unmarshal(data, &response); err != nil {
				return nil, fmt.Errorf("failed to unmarshal wallet response: %w", err)
			}
			return response, nil

		}
	}
	parts := strings.Split(topic, ".")
	mainTopic := parts[0]

	switch mainTopic {
	case "publicTrade":
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid publicTrade topic format: %s", topic)
		}

		var response PublicTradeWebsocketResponse
		if err := json.Unmarshal(data, &response); err != nil {
			return nil, fmt.Errorf("failed to unmarshal publicTrade for %s: %w", parts[1], err)
		}
		return response, nil
	case "orderbook":
		if len(parts) < 3 {
			return nil, fmt.Errorf("invalid orderbook topic format: %s", topic)
		}

		var response OrderbookWebsocketResponse
		if err := json.Unmarshal(data, &response); err != nil {
			return nil, fmt.Errorf("failed to unmarshal orderbook for %s (depth: %s): %w", parts[2], parts[1], err)
		}
		return response, nil

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

func (m *WSBybit) sendSubscribe(topic string) error {
	m.Logger.Debug("Subscribing", "topic", topic)
	return m.WSManager.SendRequest(map[string]interface{}{
		"req_id": m.GetReqId("subscribe"),
		"op":     "subscribe",
		"args":   []string{topic},
	})
}

func (m *WSBybit) sendUnsubscribe(topic string) error {
	return m.WSManager.SendRequest(map[string]interface{}{
		"req_id": m.GetReqId("unsubscribe"),
		"op":     "unsubscribe",
		"args":   []string{topic},
	})
}

func (m *WSBybit) Subscribe(topic string) error {
	m.subscriptionsMu.Lock()
	defer m.subscriptionsMu.Unlock()

	if m.subscriptions[topic] >= 1 {
		m.subscriptions[topic]++
		return nil
	}

	if err := m.sendSubscribe(topic); err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}
	m.subscriptions[topic] = 1
	return nil
}

func (m *WSBybit) Unsubscribe(topic string) error {
	m.subscriptionsMu.Lock()
	defer m.subscriptionsMu.Unlock()
	if m.subscriptions[topic] > 1 {
		m.subscriptions[topic]--
		return nil
	} else if m.subscriptions[topic] == 1 {
		delete(m.subscriptions, topic)
	} else {
		m.Logger.Error("Not subscribed", "topic", topic)
		return nil
	}

	if err := m.sendUnsubscribe(topic); err != nil {
		m.Logger.Error("Failed to unsubscribe", "topic", topic, "error", err)
	}

	return nil
}

func (m *WSBybit) GetSubscribedTopics() []string {
	m.subscriptionsMu.RLock()
	defer m.subscriptionsMu.RUnlock()

	topics := make([]string, 0, len(m.subscriptions))
	for topic := range m.subscriptions {
		topics = append(topics, topic)
	}
	return topics
}

func (m *WSBybit) Close() error {
	return m.WSManager.Close()
}

func (m *WSBybit) sendRequest(operation string, request interface{}, headers map[string]interface{}) error {
	reqId := uuid.New().String()
	message := map[string]interface{}{
		"reqId":  reqId,
		"header": headers,
		"op":     operation,
		"args":   []interface{}{request},
	}
	return m.WSManager.SendRequest(message)
}
