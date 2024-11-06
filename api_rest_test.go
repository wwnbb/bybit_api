/*
 * Copyright (c) 2024 QuantHill <info@quanthill.ae>
 */
package qant_api_bybit

import (
	"os"
	"testing"
	"time"
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
	r, err := api.REST.GetServerTime()
	if err != nil {
		t.Error(err)
	}
	if r.RetMsg != "OK" {
		t.Errorf("Error: %s", r.RetMsg)
	}
	PrettyPrint(r)
}

func TestGetOrderRT(t *testing.T) {
	api := GetApi()
	limit := 10
	openOnly := 0
	r, err := api.REST.GetOrders(OpenOrderRequest{
		Category: "linear",
		Symbol:   "ETHUSDT",
		OpenOnly: &openOnly,
		Limit:    &limit,
	})
	if err != nil {
		t.Error(err)
	}
	if r.RetMsg != "OK" {
		t.Errorf("Error: %s", r.RetMsg)
	}
	PrettyPrint(r)
}

func TestGetKline(t *testing.T) {
	api := GetApi()
	// 1724789575703
	// 1730837575703

	endt := time.Now()
	end := endt.UnixMilli()
	start := endt.Add(-10 * 7 * 24 * time.Hour).UnixMilli() // 2 weeks ago

	r, err := api.REST.GetKline(GetKlineParams{
		Category: "linear",
		Symbol:   "BTCUSDT",
		Interval: "60",
		Start:    &start,
		End:      &end,
	})
	if err != nil {
		t.Error(err)
	}
	if r.RetMsg != "OK" {
		t.Errorf("Error: %s", r.RetMsg)
	}
	PrettyPrint(r)
}

func TestGetMarkPriceKline(t *testing.T) {
	api := GetApi()
	// 1724789575703
	// 1730837575703

	endt := time.Now()
	end := endt.UnixMilli()
	start := endt.Add(-10 * 7 * 24 * time.Hour).UnixMilli() // 2 weeks ago

	r, err := api.REST.GetMarkPriceKline(GetKlineParams{
		Category: "linear",
		Symbol:   "BTCUSDT",
		Interval: "60",
		Start:    &start,
		End:      &end,
	})
	if err != nil {
		t.Error(err)
	}
	if r.RetMsg != "OK" {
		t.Errorf("Error: %s", r.RetMsg)
	}
	PrettyPrint(r)
}

func TestGetIndexPriceKline(t *testing.T) {
	api := GetApi()

	endt := time.Now()
	end := endt.UnixMilli()
	start := endt.Add(-10 * 7 * 24 * time.Hour).UnixMilli() // 2 weeks ago

	r, err := api.REST.GetIndexPriceKline(GetKlineParams{
		Category: "linear",
		Symbol:   "BTCUSDT",
		Interval: "60",
		Start:    &start,
		End:      &end,
	})
	if err != nil {
		t.Error(err)
	}
	if r.RetMsg != "OK" {
		t.Errorf("Error: %s", r.RetMsg)
	}
	PrettyPrint(r)
}

func GetDefaultLimit(category string) int {
	switch category {
	case "spot":
		return 1
	case "linear", "inverse":
		return 25
	case "option":
		return 1
	default:
		return 1
	}
}

func NewOrderbookParams(category, symbol string) GetOrderbookParams {
	limit := GetDefaultLimit(category)
	return GetOrderbookParams{
		Category: category,
		Symbol:   symbol,
		Limit:    &limit,
	}
}

func TestGetOrderbook(t *testing.T) {
	api := GetApi()

	limit := 25
	params := GetOrderbookParams{
		Category: "spot",
		Symbol:   "BTCUSDT",
		Limit:    &limit,
	}

	resp, err := api.REST.GetOrderbook(params)
	if err != nil {
		t.Fatal(err)
	}

	if resp.RetMsg != "OK" {
		t.Errorf("unexpected response message: got %s, want OK", resp.RetMsg)
	}

	if len(resp.Result.Bids) == 0 || len(resp.Result.Asks) == 0 {
		t.Error("expected non-empty orderbook")
	}

	if len(resp.Result.Bids) > 1 {
		for i := 1; i < len(resp.Result.Bids); i++ {
			prevPrice := resp.Result.Bids[i-1].Price()
			currPrice := resp.Result.Bids[i].Price()
			if prevPrice < currPrice {
				t.Error("bids not sorted in descending order")
				break
			}
		}
	}

	if len(resp.Result.Asks) > 1 {
		for i := 1; i < len(resp.Result.Asks); i++ {
			prevPrice := resp.Result.Asks[i-1].Price()
			currPrice := resp.Result.Asks[i].Price()
			if prevPrice > currPrice {
				t.Error("asks not sorted in ascending order")
				break
			}
		}
	}

	PrettyPrint(resp)
}

/*
TEST CASES FOR TRADE
*/
func TestPlaceOrder(t *testing.T) {
	// TestPlaceOrder tests the PlaceOrder endpoint
	// Cases:
	// 1. Place a limit order
	// 2. Place a market order
	api := GetApi()

	marketUnit := QuoteCoin

	testCases := []struct {
		name    string
		params  PlaceOrderParams
		wantErr bool
	}{
		{
			name: "Place a limit order",
			params: PlaceOrderParams{
				Category:   "linear",
				Symbol:     "BTCUSDT",
				Qty:        "0.001",
				Price:      ptr("50000"),
				OrderType:  "Limit",
				Side:       "Buy",
				MarketUnit: &marketUnit,
			},
		},
		{
			name: "Place a market order",
			params: PlaceOrderParams{
				Category:   "linear",
				Symbol:     "BTCUSDT",
				Qty:        "0.001",
				MarketUnit: &marketUnit,
				OrderType:  "Market",
				Side:       "Buy",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name,
			func(t *testing.T) {
				resp, err := api.REST.PlaceOrder(tc.params)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if resp.RetMsg != "OK" {
					t.Errorf("unexpected response message: got %s, want OK", resp.RetMsg)
				}
				PrettyPrint(resp)
			},
		)
	}
}

func TestGetInstrumentsInfo(t *testing.T) {
	// TestGetInstrumentsInfo tests the GetInstrumentsInfo endpoint
	// Cases:
	// 1. Get BTCUSDT Spot info
	// 2. Get Linear Instruments
	// 3. Invalid Category
	api := GetApi()
	limit := 500

	testCases := []struct {
		name    string
		params  GetInstrumentsInfoParams
		wantErr bool
	}{
		{
			name: "Get BTCUSDT Spot Info",
			params: GetInstrumentsInfoParams{
				Category: "spot",
				Symbol:   "BTCUSDT",
			},
			wantErr: false,
		},
		{
			name: "Get Linear Instruments",
			params: GetInstrumentsInfoParams{
				Category: "linear",
				Limit:    &limit,
			},
			wantErr: false,
		},
		{
			name: "Invalid Category",
			params: GetInstrumentsInfoParams{
				Category: "invalid",
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := api.REST.GetInstrumentsInfo(tc.params)

			if tc.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if resp.RetMsg != "OK" {
				t.Errorf("unexpected response message: got %s, want OK", resp.RetMsg)
			}

			if len(resp.Result.List) == 0 {
				t.Error("expected non-empty result list")
			}

			for _, instrument := range resp.Result.List {
				if instrument.Symbol == "" {
					t.Error("instrument symbol is empty")
				}

			}
			PrettyPrint(resp)
		})
	}
}
