/*
 * Copyright (c) 2024 QuantHill <info@quanthill.ae>
 */
package qant_api_bybit

import (
	"os"
	"testing"
)

func GetApi() *BybitApi {
	apiKey := os.Getenv("BYBIT_API_KEY")
	secretKey := os.Getenv("BYBIT_SECRET_KEY")
	api := NewBybitApi(apiKey, secretKey)
	api.ConfigureRestUrl(TESTNET_URL)
	return api
}

func TestGetServerTime(t *testing.T) {
	api := GetApi()
	r, err := api.GetServerTime()
	if err != nil {
		t.Error(err)
	}
	PrettyPrint(r)
}

func TestPlaceOrder(t *testing.T) {
	api := GetApi()
	marketUnit := QuoteCoin
	r, err := api.PlaceOrder(PlaceOrderParams{
		Category:   "linear",
		Symbol:     "BTCUSDT",
		Qty:        "0.001",
		MarketUnit: &marketUnit,
		OrderType:  "Market",
		Side:       "Buy",
	})
	if err != nil {
		t.Error(err)
	}
	PrettyPrint(r)
}

func TestGetOrderRT(t *testing.T) {
	api := GetApi()
	r, err := api.GetOrders(OpenOrderRequest{
		Symbol: "BTCUSDT",
	})
	if err != nil {
		t.Error(err)
	}
	PrettyPrint(r)
}
