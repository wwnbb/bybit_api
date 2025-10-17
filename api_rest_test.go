package bybit_api

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	pp "github.com/wwnbb/pprint"
	"github.com/wwnbb/ptr"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"
)

func ConvertTsToHumanReadable(ts string) (time.Time, error) {
	ts64, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("Error while converting ts string: %v to Time %v", ts, err)
	}
	ts64 = ts64 / 1000
	t := time.Unix(ts64, 0)
	return t, nil
}

func GetApi() *BybitApi {
	apiKey := os.Getenv("BYBIT_API_KEY")
	secretKey := os.Getenv("BYBIT_SECRET_KEY")
	ctx := context.Background()
	api := NewBybitApi(apiKey, secretKey, ctx)
	api.ConfigureMainNetUrls()
	return api
}

func GetApiProd() *BybitApi {
	apiKey := os.Getenv("BYBIT_API_KEY_PROD")
	secretKey := os.Getenv("BYBIT_SECRET_KEY_PROD")
	ctx := context.Background()
	api := NewBybitApi(apiKey, secretKey, ctx)
	api.ConfigureMainNetUrls()
	return api
}

func TestGetServerTime(t *testing.T) {
	api := GetApi()
	api.ConfigureMainNetUrls()
	r, err := api.REST.GetServerTime()
	if err != nil {
		t.Error(err)
	}
	if r.RetMsg != "OK" {
		t.Errorf("Error: %s", r.RetMsg)
	}
	pp.PrettyPrint(r)
	fmt.Printf("Current timestamp: %d \n", time.Now().UnixMilli())
}

func TestGetOrderRT(t *testing.T) {
	api := GetApi()
	limit := 10
	openOnly := 0
	r, err := api.REST.GetOrders(OpenOrderRequest{
		Category: "inverse",
		Symbol:   "LTCUSD",
		OpenOnly: &openOnly,
		Limit:    &limit,
		OrderId:  "7c1aac82-b3ac-4b97-8375-6a446e269c05",
	})
	if err != nil {
		t.Error(err)
	}
	if r.RetMsg != "OK" {
		t.Errorf("Error: %s", r.RetMsg)
	}
	pp.PrettyPrint(r)
}

func TestGetOrders(t *testing.T) {
	api := GetApi()
	limit := 10
	openOnly := 0
	r, err := api.REST.GetOrders(OpenOrderRequest{
		Category: "linear",
		Symbol:   "BTCUSDT",
		OpenOnly: &openOnly,
		Limit:    &limit,
	})
	if err != nil {
		t.Fatalf("Cannot get orders %s", err)
	}
	if r.RetMsg != "OK" {
		t.Errorf("Error: %s", r.RetMsg)
	}
	pp.PrettyPrint(r)
}

func TestGetKline(t *testing.T) {
	api := GetApi()
	// 1724789575703
	// 1730837575703

	endt := time.Now()
	end := endt.UnixMilli()
	end = 1748304000000 + 1000*60*60*24
	// 1747526400000
	// 1747440000000

	r, err := api.REST.GetKline(GetKlineParams{
		Category: "spot",
		Symbol:   "BTCUSDT",
		Interval: "D",
		// Start:    &start,
		End:   &end,
		Limit: ptr.Ptr(11),
	})
	if err != nil {
		t.Error(err)
	}
	if r.RetMsg != "OK" {
		t.Errorf("Error: %s", r.RetMsg)
	}
	for _, v := range r.Result.List {
		ts, err := ConvertTsToHumanReadable(v.StartTime())
		if err != nil {
			t.Error(err)
			continue
		}
		fmt.Printf("ts: %d, ", ts.UnixMilli())
		fmt.Printf("t: %s, ", ts.UTC().Format(time.DateTime))
		fmt.Printf("o: %s, ", v.OpenPrice())
		fmt.Printf("h: %s, ", v.HighPrice())
		fmt.Printf("c: %s, ", v.ClosePrice())
		fmt.Printf("l: %s\n", v.LowPrice())
	}
}

func TestGetPositions(t *testing.T) {
	api := GetApi()

	testCases := []struct {
		name    string
		params  GetPositionParams
		wantErr bool
	}{
		{
			name: "Get All Linear Positions",
			params: GetPositionParams{
				Category: "inverse",
			},
			wantErr: false,
		},
		{
			name: "Get Positions with Limit",
			params: GetPositionParams{
				Category: "option",
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := api.REST.GetPositions(tc.params)

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if resp.RetMsg != "OK" {
				t.Errorf("unexpected response message: got %s, want OK", resp.RetMsg)
			}

			// Verify that the response contains the expected data structure
			if resp.Result.Category != tc.params.Category {
				t.Errorf("unexpected category: got %s, want %s", resp.Result.Category, tc.params.Category)
			}

			// If we specified a symbol, check that the positions are for that symbol
			if tc.params.Symbol != "" {
				for _, position := range resp.Result.List {
					if position.Symbol != tc.params.Symbol {
						t.Errorf("position symbol mismatch: got %s, want %s", position.Symbol, tc.params.Symbol)
					}
				}
			}

			// Check if limit parameter is respected
			if tc.params.Limit != nil && len(resp.Result.List) > *tc.params.Limit {
				t.Errorf("response contains more positions than the limit: got %d, limit %d",
					len(resp.Result.List), *tc.params.Limit)
			}

			// Print the response for debugging
			pp.PrettyPrint(resp)
		})
	}
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
	pp.PrettyPrint(r)
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
	pp.PrettyPrint(r)
}

func GetDefaultLimit(category CategoryType) int {
	switch category {
	case "spot":
		return 1
	case LinearCategory, InverseCategory:
		return 25
	case OptionCategory:
		return 1
	default:
		return 1
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

	pp.PrettyPrint(resp)
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
				Price:      ptr.Ptr("50000"),
				OrderType:  "Limit",
				Side:       "Buy",
				MarketUnit: ptr.Ptr(QuoteCoin),
			},
		},
		{
			name: "Place a market order",
			params: PlaceOrderParams{
				Category:   "linear",
				Symbol:     "BTCUSDT",
				Qty:        "0.001",
				MarketUnit: ptr.Ptr(QuoteCoin),
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
				pp.PrettyPrint(resp)
			},
		)
	}
}

func TestPlaceInverseOrder(t *testing.T) {
	api := GetApi()

	wg := sync.WaitGroup{}
	for range 10 {
		for range 5 {
			wg.Add(1)
			go func() {
				id := uuid.New()
				params := PlaceOrderParams{
					Category:    "inverse",
					Symbol:      "DOGEUSD",
					Qty:         "5",
					OrderType:   "Market",
					Side:        "Buy",
					OrderLinkId: ptr.Ptr(id.String()),
					MarketUnit:  ptr.Ptr(QuoteCoin),
				}

				resp, err := api.REST.PlaceOrder(params)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if resp.RetMsg != "OK" {
					t.Errorf("unexpected response message: got %s, want OK", resp.RetMsg)
				}
				pp.PrettyPrint(resp)
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func TestPlaceAndCancelOrder(t *testing.T) {
	// 1. Place a limit order
	// 2. Cancel the order we just placed

	api := GetApi()

	placeParams := PlaceOrderParams{
		Category:  "linear",
		Symbol:    "BTCUSDT",
		Side:      "Buy",
		OrderType: "Limit",
		Qty:       "0.001",
		Price:     ptr.Ptr("35000"),
	}

	placeResp, err := api.REST.PlaceOrder(placeParams)
	if err != nil {
		t.Fatalf("Failed to place order: %v", err)
	}

	if placeResp.RetMsg != "OK" {
		t.Errorf("Error placing order: %s", placeResp.RetMsg)
	}

	pp.PrettyPrint(placeResp)

	cancelParams := CancelOrderParams{
		Category: "linear",
		Symbol:   "BTCUSDT",
		OrderId:  &placeResp.Result.OrderId,
	}

	cancelResp, err := api.REST.CancelOrder(cancelParams)
	if err != nil {
		t.Fatalf("Failed to cancel order: %v", err)
	}

	if cancelResp.RetMsg != "OK" {
		t.Errorf("Error canceling order: %s", cancelResp.RetMsg)
	}

	pp.PrettyPrint(cancelResp)
}

func TestGetInstrumentsInfoCall(t *testing.T) {
	// TestGetInstrumentsInfo tests the GetInstrumentsInfo endpoint
	// Cases:
	// 1. Get BTCUSDT Spot info
	// 2. Get Linear Instruments
	api := GetApi()

	params := GetInstrumentsInfoParams{
		Category: "linear",
	}

	// для инверсного
	// BTC -> funding currency
	// USD -> quote coin

	//  Для линейного
	// usdt - И quote и funding

	resp, err := api.REST.GetInstrumentsInfo(params)

	fmt.Printf("BaseCoin: %s, QuoteCoin: %s\n", resp.Result.List[0].BaseCoin, resp.Result.List[0].QuoteCoin)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	time.Sleep(1 * time.Second) // Sleep to avoid hitting rate limits
	fmt.Print("\n")
}

func TestGetInstrumentsInfo(t *testing.T) {
	// TestGetInstrumentsInfo tests the GetInstrumentsInfo endpoint
	// Cases:
	// 1. Get BTCUSDT Spot info
	// 2. Get Linear Instruments
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
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := api.REST.GetInstrumentsInfo(tc.params)

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
			pp.PrettyPrint(resp)
		})
	}
}

func TestGetWalletBalance(t *testing.T) {
	api := GetApi()
	r, err := api.REST.GetWalletBalance(GetWalletBalanceParams{
		AccountType: "UNIFIED",
	})
	if err != nil {
		t.Error(err)
	}
	if r.RetMsg != "OK" {
		t.Errorf("Error: %s", r.RetMsg)
	}
	pp.PrettyPrint(r)
}

func TestGetWalletBalanceProd(t *testing.T) {
	api := GetApiProd()
	r, err := api.REST.GetWalletBalance(GetWalletBalanceParams{
		AccountType: "UNIFIED",
	})
	if err != nil {
		t.Error(err)
	}
	if r.RetMsg != "OK" {
		t.Errorf("Error: %s", r.RetMsg)
	}
	pp.PrettyPrint(r)
}

type InstrumentType struct {
	baseCoin     string
	quoteCoin    string
	contractType string
}

func TestGetInstruments(t *testing.T) {
	api := GetApi()
	// limit := 1000
	// var instrMap = make(map[string]InstrumentType)

	params := GetInstrumentsInfoParams{
		Category: "linear",
		Symbol:   "BTCUSDT",
	}

	resp, _ := api.REST.GetInstrumentsInfo(params)
	pp.PrettyPrint(resp)
}

func TestGetPositionsCall(t *testing.T) {
	api := GetApi()

	params := GetPositionParams{
		Category: "spot",
		// SettleCoin: "USDT",
	}

	resp, err := api.REST.GetPositions(params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.RetMsg != "OK" {
		t.Errorf("unexpected response message: got %s, want OK", resp.RetMsg)
	}

	pp.PrettyPrint(resp)
}

func TestRecentPublicTradesCall(t *testing.T) {
	api := GetApi()
	params := GetRecentPublicTradesParams{
		Symbol:   "BTCUSDT",
		Category: "spot",
	}
	resp, err := api.REST.GetRecentPublicTrades(params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.RetMsg != "OK" {
		t.Errorf("unexpected response message: got %s, want OK", resp.RetMsg)
	}

	pp.PrettyPrint(resp)
}

func TestGetTicker(t *testing.T) {
	api := GetApi()
	api.ConfigureMainNetUrls()
	params := GetTickerParams{
		Category: "linear",
		Symbol:   "BTCUSDT",
	}
	resp, err := api.REST.GetTicker(params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	pp.PrettyPrint(resp)
}
