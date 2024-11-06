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
	"time"

	json "github.com/goccy/go-json"
	"github.com/gorilla/schema"
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

	wsPool  map[string]chan []byte
	logger  Logger
	encoder *schema.Encoder

	REST    *RESTManager
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
	b.ConfigureRestUrl(BASE_URL)
	b.ConfigureWsUrls(
		MAINNET_PRIVATE_WS,
		MAINNET_SPOT_WS,
		MAINNET_LINEAR_WS,
		MAINNET_INVERSE_WS,
		MAINNET_TRADE_WS,
	)
}

func (b *BybitApi) ConfigureTestNetUrls() {
	b.ConfigureRestUrl(TESTNET_URL)
	b.ConfigureWsUrls(
		TESTNET_PRIVATE_WS,
		TESTNET_SPOT_WS,
		TESTNET_LINEAR_WS,
		TESTNET_INVERSE_WS,
		TESTNET_TRADE_WS,
	)
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
	api.encoder = schema.NewEncoder()

	api.Spot = newWSManager(api, WS_SPOT, api.WS_URL_SPOT)
	api.Linear = newWSManager(api, WS_LINEAR, api.WS_URL_LINEAR)
	api.Inverse = newWSManager(api, WS_INVERSE, api.WS_URL_INVERSE)
	api.Trade = newWSManager(api, WS_TRADE, api.WS_URL_TRADE)
	api.Private = newWSManager(api, WS_PRIVATE, api.WS_URL_PRIVATE)
	api.REST = NewRESTManager(api, api.BASE_REST_URL)

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
