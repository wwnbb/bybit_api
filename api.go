/*
 * Copyright (c) 2024 QuantHill <info@quanthill.ae>
 */
package qant_api_bybit

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"

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

var ERR_URLS_NOT_CONFIGURED = errors.New("Configure api urls")
var ERR_TRADING_STREAMS_NOT_SUPPORTED = errors.New("Trading streams not supported on demo account")

type BybitApi struct {
	ApiKey    string
	ApiSecret string
	context   context.Context
	timeout   time.Duration
	UrlSet    bool

	BASE_REST_URL  string
	WS_URL_PRIVATE string
	WS_URL_SPOT    string
	WS_URL_LINEAR  string
	WS_URL_INVERSE string
	WS_URL_OPTION  string
	WS_URL_TRADE   string

	wsPool  map[string]chan []byte
	Logger  Logger
	encoder *schema.Encoder

	REST    *RESTManager
	Spot    *WSManager
	Linear  *WSManager
	Inverse *WSManager
	Option  *WSManager
	Trade   *WSManager
	Private *WSManager
}

func (b *BybitApi) SetLogger(logger Logger) {
	b.Logger = logger
}

func (b *BybitApi) ConfigureRestUrl(restUrl string) {
	b.BASE_REST_URL = restUrl
	b.REST = NewRESTManager(b)
}

func (b *BybitApi) ConfigureWsUrls(privateUrl, spotUrl, linearUrl, inverseUrl, optionUrl, tradeUrl string) {
	b.UrlSet = true
	b.WS_URL_PRIVATE = privateUrl
	b.WS_URL_SPOT = spotUrl
	b.WS_URL_LINEAR = linearUrl
	b.WS_URL_INVERSE = inverseUrl
	b.WS_URL_OPTION = optionUrl
	b.WS_URL_TRADE = tradeUrl

	b.Spot = newWSManager(b, WS_SPOT, b.WS_URL_SPOT)
	b.Linear = newWSManager(b, WS_LINEAR, b.WS_URL_LINEAR)
	b.Inverse = newWSManager(b, WS_INVERSE, b.WS_URL_INVERSE)
	b.Option = newWSManager(b, WS_OPTION, b.WS_URL_OPTION)
	b.Trade = newWSManager(b, WS_TRADE, b.WS_URL_TRADE)
	b.Private = newWSManager(b, WS_PRIVATE, b.WS_URL_PRIVATE)
	b.REST = NewRESTManager(b)
}

func (b *BybitApi) ConfigureMainNetUrls() {
	b.ConfigureRestUrl(BASE_URL)
	b.ConfigureWsUrls(
		MAINNET_PRIVATE_WS,
		MAINNET_SPOT_WS,
		MAINNET_LINEAR_WS,
		MAINNET_INVERSE_WS,
		MAINNET_OPTION_WS,
		MAINNET_TRADE_WS,
	)
}

func (b *BybitApi) ConfigureMainNetDemoUrls() {
	b.ConfigureRestUrl(BASE_DEMO_URL)
	b.ConfigureWsUrls(
		MAINNET_DEMO_PRIVATE_WS,
		MAINNET_SPOT_WS,
		MAINNET_LINEAR_WS,
		MAINNET_INVERSE_WS,
		MAINNET_OPTION_WS,
		"",
	)
}

func (b *BybitApi) ConfigureTestNetUrls() {
	b.ConfigureRestUrl(TESTNET_URL)
	b.ConfigureWsUrls(
		TESTNET_PRIVATE_WS,
		TESTNET_SPOT_WS,
		TESTNET_LINEAR_WS,
		TESTNET_INVERSE_WS,
		TESTNET_OPTION_WS,
		TESTNET_TRADE_WS,
	)
}

func NewBybitApi(apiKey, apiSecret string, ctx context.Context) *BybitApi {
	api := &BybitApi{
		ApiKey:    apiKey,
		ApiSecret: apiSecret,
		context:   ctx,
		wsPool:    make(map[string]chan []byte),
		Logger:    BasicLogger("BybitApi"),
		timeout:   10 * time.Second,
	}
	api.encoder = schema.NewEncoder()
	return api

}

func (b *BybitApi) Disconnect() {
	for _, m := range []*WSManager{
		b.Spot,
		b.Linear,
		b.Inverse,
		b.Option,
		b.Trade,
		b.Private,
	} {
		err := m.close()
		if err != nil {
			// TODO: return errors instead of logging out
			b.Logger.Error("Failed to disconnect: %v", err)
		}
	}
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

// SetTimeout sets a custom timeout for the API
func SetTimeout(timeout time.Duration) func(*BybitApi) {
	return func(api *BybitApi) {
		api.timeout = timeout
	}
}

// SetContext sets a custom context for the API
func SetContext(ctx context.Context) func(*BybitApi) {
	return func(api *BybitApi) {
		api.context = ctx
	}
}
