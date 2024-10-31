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
	"strconv"
	// "sync"
	json "github.com/goccy/go-json"
	"time"
)

type StringOrInterface interface {
	string | interface{}
}

type TimeGetter interface {
	Now() time.Time
}

type TimeProvider struct{}

func (tp *TimeProvider) Now() time.Time {
	return time.Now()
}

type BybitApi struct {
	ApiKey    string
	ApiSecret string
	context   context.Context

	BASE_REST_URL  string
	WS_URL_PRIVATE string
	WS_URL_SPOT    string
	WS_URL_LINEAR  string
	WS_URL_INVERSE string
	WS_URL_TRADE   string

	wsPool map[string]chan []byte
	logger Logger

	Spot    *WSManager
	Linear  *WSManager
	Inverse *WSManager
	Trade   *WSManager
	Private *WSManager
}

func (b *BybitApi) SetLogger(logger Logger) {
	b.logger = logger
}

func (b *BybitApi) ConfigureRestUrl(restUrl string) {
	b.BASE_REST_URL = restUrl
}

func (b *BybitApi) ConfigureWsUrls(privateUrl, spotUrl, linearUrl, inverseUrl, tradeUrl string) {
	b.WS_URL_PRIVATE = privateUrl
	b.WS_URL_SPOT = spotUrl
	b.WS_URL_LINEAR = linearUrl
	b.WS_URL_INVERSE = inverseUrl
	b.WS_URL_TRADE = tradeUrl
}

func (b *BybitApi) ConfigureMainNetUrls() {
	b.ConfigureWsUrls(MAINNET_PRIVATE_WS, MAINNET_SPOT_WS, MAINNET_LINEAR_WS, MAINNET_INVERSE_WS, MAINNET_TRADE_WS)
}

func (b *BybitApi) ConfigureTestNetUrls() {
	b.ConfigureWsUrls(TESTNET_PRIVATE_WS, TESTNET_SPOT_WS, TESTNET_LINEAR_WS, TESTNET_INVERSE_WS, TESTNET_TRADE_WS)
}

func NewBybitApi(apiKey, apiSecret string) *BybitApi {
	api := &BybitApi{
		ApiKey:    apiKey,
		ApiSecret: apiSecret,
		context:   context.Background(),
		wsPool:    make(map[string]chan []byte),
		logger:    BasicLogger("BybitApi"),

		WS_URL_PRIVATE: TESTNET_PRIVATE_WS,
		WS_URL_SPOT:    TESTNET_SPOT_WS,
		WS_URL_LINEAR:  TESTNET_LINEAR_WS,
		WS_URL_INVERSE: TESTNET_INVERSE_WS,
		WS_URL_TRADE:   TESTNET_TRADE_WS,
	}

	api.Spot = newWSManager(api, WS_SPOT, api.WS_URL_SPOT)
	api.Linear = newWSManager(api, WS_LINEAR, api.WS_URL_LINEAR)
	api.Inverse = newWSManager(api, WS_INVERSE, api.WS_URL_INVERSE)
	api.Trade = newWSManager(api, WS_TRADE, api.WS_URL_TRADE)
	api.Private = newWSManager(api, WS_PRIVATE, api.WS_URL_PRIVATE)

	return api

}

// GenSignature generates an HMAC SHA256 signature for API authentication using the provided parameters.
// It accepts a generic type T that can be either a string or an interface{}, along with API credentials
// and timing parameters.
// https://github.com/bybit-exchange/api-usage-examples
func GenSignature[T StringOrInterface](
	params T,
	apiKey, apiSecret string,
	recvWindow string,
) (string, int64, error) {
	return genSignature(params, apiKey, apiSecret, recvWindow, &TimeProvider{})
}

func genSignature[T StringOrInterface](
	params T, apiKey, apiSecret string,
	recvWindow string,
	tg TimeGetter,
) (string, int64, error) {

	now := tg.Now()
	unixNano := now.UnixNano()
	timestamp := unixNano / int64(time.Millisecond)
	var paramStr string

	switch v := any(params).(type) {
	case string:
		paramStr = v
	case interface{}:
		jsonData, err := json.Marshal(v)
		if err != nil {
			return "", timestamp, fmt.Errorf("failed to marshal params: %w", err)
		}
		paramStr = string(jsonData[:])
	}
	signatureStr := []byte(strconv.FormatInt(timestamp, 10) + apiKey + recvWindow + paramStr)
	hmac256 := hmac.New(sha256.New, []byte(apiSecret))
	hmac256.Write([]byte(signatureStr))
	signature := hex.EncodeToString(hmac256.Sum(nil))

	return signature, timestamp, nil
}

//
// func (b *BybitApi) ConnectWebSocket(endpoints []string) error {
// 	var wg sync.WaitGroup
// 	errChan := make(chan error, len(endpoints))
//
// 	for _, endpoint := range endpoints {
// 		wg.Add(1)
// 		go func(url string) {
// 			defer wg.Done()
// 			if err := b.connectEndpoint(url); err != nil {
// 				errChan <- fmt.Errorf("failed to connect to %s: %w", url, err)
// 			}
// 		}(endpoint)
// 	}
//
// 	wg.Wait()
// 	close(errChan)
//
// 	var errors []error
// 	for err := range errChan {
// 		errors = append(errors, err)
// 	}
//
// 	if len(errors) > 0 {
// 		return fmt.Errorf("connection errors: %v", errors)
// 	}
//
// 	return nil
// }
//
// func (b *BybitApi) connectEndpoint(endpoint string) error {
// 	if _, exists := b.wsConnections[endpoint]; exists {
// 		return fmt.Errorf("connection to %s already exists", endpoint)
// 	}
//
// 	ctx, cancel := context.WithCancel(b.context)
// 	wsConn := &WSConnection{
// 		ctx:        ctx,
// 		cancel:     cancel,
// 		url:        endpoint,
// 		subscribed: make(map[string]struct{}),
// 	}
// 	b.wsConnections[endpoint] = wsConn
//
// 	if err := b.connectWebSocket(wsConn); err != nil {
// 		return fmt.Errorf("failed to connect to %s: %w", endpoint, err)
// 	}
//
// 	go b.reconnectRoutine(wsConn)
// 	go b.readRoutine(wsConn)
// 	go b.pingRoutine(wsConn)
//
// 	return nil
// }
//
// func (b *BybitApi) connectPrivateEndpoint(url string) error {
// 	if _, exists := b.wsPrivateConnections[url]; exists {
// 		return fmt.Errorf("connection to %s already exists", url)
// 	}
//
// 	ctx, cancel := context.WithCancel(b.context)
// 	wsConn := &WSConnection{
// 		ctx:        ctx,
// 		cancel:     cancel,
// 		url:        url,
// 		subscribed: make(map[string]struct{}),
// 		apiKey:     b.ApiKey,
// 		apiSecret:  b.ApiSecret,
// 	}
// 	b.wsPrivateConnections[url] = wsConn
//
// 	if err := b.connectWebSocket(wsConn); err != nil {
// 		return fmt.Errorf("failed to connect to %s: %w", url, err)
// 	}
//
// 	go b.reconnectRoutine(wsConn)
// 	go b.readRoutine(wsConn)
// 	go b.pingRoutine(wsConn)
//
// 	return nil
// }
//
// func (b *BybitApi) ConnectPrivateWebSocket() error {
// 	endpoints := []string{b.WS_URL_PRIVATE}
// 	var wg sync.WaitGroup
// 	errChan := make(chan error, len(endpoints))
//
// 	for _, endpoint := range endpoints {
// 		wg.Add(1)
// 		go func(url string) {
// 			defer wg.Done()
// 			if err := b.connectPrivateEndpoint(url); err != nil {
// 				errChan <- fmt.Errorf("failed to connect to %s: %w", url, err)
// 			}
// 		}(endpoint)
// 	}
//
// 	wg.Wait()
// 	close(errChan)
//
// 	var errors []error
// 	for err := range errChan {
// 		errors = append(errors, err)
// 	}
//
// 	if len(errors) > 0 {
// 		return fmt.Errorf("connection errors: %v", errors)
// 	}
//
// 	return nil
// }
//
// func (b *BybitApi) ConnectTradeWs() error {
// 	endpoints := []string{b.WS_URL_TRADE}
// 	var wg sync.WaitGroup
// 	errChan := make(chan error, len(endpoints))
//
// 	for _, endpoint := range endpoints {
// 		wg.Add(1)
// 		go func(url string) {
// 			defer wg.Done()
// 			if err := b.connectPrivateEndpoint(url); err != nil {
// 				errChan <- fmt.Errorf("failed to connect to %s: %w", url, err)
// 			}
// 		}(endpoint)
// 	}
//
// 	wg.Wait()
// 	close(errChan)
//
// 	var errors []error
// 	for err := range errChan {
// 		errors = append(errors, err)
// 	}
//
// 	if len(errors) > 0 {
// 		return fmt.Errorf("connection errors: %v", errors)
// 	}
//
// 	return nil
//
// }
//
// func (b *BybitApi) reconnectRoutine(wsConn *WSConnection) {
// 	for {
// 		select {
// 		case <-wsConn.ctx.Done():
// 			return
// 		default:
// 			if wsConn.state.Load() == int32(StatusDisconnected) {
// 				b.logger.Info("Reconnecting to %s", wsConn.url)
// 				if err := b.connectWebSocket(wsConn); err != nil {
// 					b.logger.Error("Failed to reconnect to %s: %v", wsConn.url, err)
// 					time.Sleep(wsInitialReconnectDelay)
// 				}
// 			}
//
// 			time.Sleep(wsMaxSilentPeriod)
// 		}
// 	}
// }
//
// func (b *BybitApi) pingRoutine(wsConn *WSConnection) {
// 	ticker := time.NewTicker(wsPingInterval)
// 	defer ticker.Stop()
//
// 	for {
// 		select {
// 		case <-wsConn.ctx.Done():
// 			return
// 		case <-ticker.C:
// 			currentTime := time.Now().UnixNano()
// 			if wsConn.state.Load() == int32(StatusConnected) {
// 				wsConn.mu.Lock()
// 				if time.Since(wsConn.lastPing) >= wsPingInterval {
// 					pingMessage := map[string]string{
// 						"op":     "ping",
// 						"req_id": fmt.Sprintf("%d", currentTime),
// 					}
// 					jsonPingMessage, _ := json.Marshal(pingMessage)
// 					if err := wsConn.conn.WriteMessage(gws.PingMessage, jsonPingMessage); err != nil {
// 						b.logger.Error("Error sending ping to %s: %v", wsConn.url, err)
// 						wsConn.state.Store(int32(StatusDisconnected))
// 					}
// 					wsConn.lastPing = time.Now()
// 				}
// 				wsConn.mu.Unlock()
// 			}
// 		}
// 	}
// }
//
// func (b *BybitApi) connectWebSocket(wsConn *WSConnection) error {
// 	dialer := gws.DefaultDialer
// 	conn, resp, err := dialer.DialContext(wsConn.ctx, wsConn.url, nil)
// 	if err != nil {
// 		return fmt.Errorf("dial failed: %w", err)
// 	}
//
// 	wsConn.mu.Lock()
// 	wsConn.conn = conn
// 	wsConn.state.Store(int32(StatusConnected))
// 	wsConn.lastPing = time.Now()
// 	wsConn.mu.Unlock()
//
// 	if b.ApiKey != "" && b.ApiSecret != "" {
// 		if err := b.authenticate(wsConn); err != nil {
// 			return fmt.Errorf("authentication failed: %w", err)
// 		}
// 	}
//
// 	b.logger.Info("Connected to %s: response status %d", wsConn.url, resp.StatusCode)
// 	return nil
// }
//
// func (b *BybitApi) CloseWebSocket() error {
// 	var wg sync.WaitGroup
// 	errChan := make(chan error, len(b.wsConnections))
//
// 	for endpoint, conn := range b.wsConnections {
// 		wg.Add(1)
// 		go func(ep string, wsConn *WSConnection) {
// 			defer wg.Done()
// 			if err := b.closeEndpoint(ep, wsConn); err != nil {
// 				errChan <- fmt.Errorf("error closing %s: %w", ep, err)
// 			}
// 		}(endpoint, conn)
// 	}
//
// 	wg.Wait()
// 	close(errChan)
//
// 	var errors []error
// 	for err := range errChan {
// 		errors = append(errors, err)
// 	}
//
// 	b.wsConnections = make(map[string]*WSConnection)
//
// 	if len(errors) > 0 {
// 		return fmt.Errorf("closure errors: %v", errors)
// 	}
//
// 	return nil
// }
//
// func (b *BybitApi) closeEndpoint(endpoint string, wsConn *WSConnection) error {
// 	wsConn.mu.Lock()
// 	defer wsConn.mu.Unlock()
//
// 	if wsConn.conn != nil {
// 		wsConn.state.Store(int32(StatusClosed))
// 		wsConn.cancel()
//
// 		err := wsConn.conn.WriteMessage(
// 			gws.CloseMessage,
// 			gws.FormatCloseMessage(gws.CloseNormalClosure, "client closed connection"),
// 		)
// 		if err != nil {
// 			b.logger.Error("Error sending close message to %s: %v", endpoint, err)
// 		}
//
// 		if err := wsConn.conn.Close(); err != nil {
// 			return fmt.Errorf("error closing connection: %w", err)
// 		}
// 		wsConn.conn = nil
// 	}
//
// 	return nil
// }
//
// func (b *BybitApi) Subscribe(topic string) error {
// 	if ch, exists := b.wsPool[topic]; exists {
// 		return ch, nil
// 	}
//
// 	ch := make(chan []byte, 10)
// 	b.wsPool[topic] = ch
//
// 	return ch, nil
// }
//
// func (b *BybitApi) readRoutine(wsConn *WSConnection) {
// 	for {
// 		select {
// 		case <-wsConn.ctx.Done():
// 			return
// 		default:
// 			if wsConn.state.Load() != int32(StatusConnected) {
// 				time.Sleep(time.Second)
// 				continue
// 			}
//
// 			wsConn.mu.Lock()
// 			_, message, err := wsConn.conn.ReadMessage()
// 			wsConn.mu.Unlock()
//
// 			if err != nil {
// 				b.logger.Error("Read error from %s: %v", wsConn.url, err)
// 				wsConn.state.Store(int32(StatusDisconnected))
// 				continue
// 			}
//
// 			if err := b.processMessage(wsConn.url, message); err != nil {
// 				b.logger.Error("Error processing message from %s: %v", wsConn.url, err)
// 			}
// 		}
// 	}
// }
//
// func (b *BybitApi) processMessage(endpoint string, message []byte) error {
// 	var msg struct {
// 		Topic string      `json:"topic"`
// 		Type  string      `json:"type"`
// 		Data  interface{} `json:"data"`
// 	}
//
// 	if err := json.Unmarshal(message, &msg); err != nil {
// 		return fmt.Errorf("error unmarshaling message: %w", err)
// 	}
//
// 	if ch, exists := b.wsPool[msg.Topic]; exists {
// 		select {
// 		case ch <- message:
// 		default:
// 			b.logger.Info("Channel full for topic %s, dropping message", msg.Topic)
// 		}
// 	}
//
// 	return nil
// }
//
// func (b *BybitApi) authenticate(wsConn *WSConnection) error {
// 	return nil
// }
