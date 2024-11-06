package qant_api_bybit

import (
	"fmt"
	"testing"
	"time"

	"github.com/goccy/go-json"
)

type WSResponse struct {
	Topic string      `json:"topic"`
	Ts    int64       `json:"ts"`
	Type  string      `json:"type"`
	Data  interface{} `json:"data"`
	Cts   int64       `json:"cts"`
}

type PongResponse struct {
	Success bool   `json:"success"`
	RetMsg  string `json:"ret_msg"`
	ConnId  string `json:"conn_id"`
	ReqId   string `json:"req_id,omitempty"`
	Op      string `json:"op"`
}

func getApi() *BybitApi {
	api := NewBybitApi("QU0G8RSs5aSsoGVir2", "IHmT3wcDaI7TBo0WlJaPlJj8JTMtdb5KQrZR")
	api.ConfigureWsUrls(
		TESTNET_PRIVATE_WS,
		TESTNET_SPOT_WS,
		TESTNET_LINEAR_WS,
		TESTNET_INVERSE_WS,
		TESTNET_TRADE_WS,
	)
	return api
}

func TestGetReqId(t *testing.T) {
	api := NewBybitApi("QU0G8RSs5aSsoGVir", "IHmT3wcDaI7TBo0WlJaPlJj8JTMtdb5KQrZR")

	wsm := newWSManager(api, WS_LINEAR, TESTNET_SPOT_WS)
	reqId := wsm.getReqId("subscribe")
	if reqId == "" {
		t.Errorf("Error: reqId is empty")
	}
	if reqId != "subscribe_1" {
		t.Errorf("Error: wrong reqId")
	}
	reqId = wsm.getReqId("unsubscribe")
	if reqId != "unsubscribe_1" {
		t.Errorf("Error: wrong reqId")
	}
	reqId = wsm.getReqId("subscribe")
	if reqId != "subscribe_2" {
		t.Errorf("Error: wrong reqId")
	}
}

func TestWsManagerSubscribe(t *testing.T) {
	api := getApi()
	api.ConfigureTestNetUrls()
	api.Spot.Subscribe("orderbook.1.BTCUSDT")
	for i := 0; i < 1000; i++ {
		fmt.Println(<-api.Spot.DataCh)
	}
	api.Spot.Unsubscribe("orderbook.1.BTCUSDT")
}

func TestWsManagerTradeSubscribe(t *testing.T) {
	api := getApi()
	api.ConfigureTestNetUrls()
	api.Spot.Subscribe("publicTrade.BTCUSDT")
	for i := 0; i < 10; i++ {
		fmt.Println(<-api.Spot.DataCh)
	}
	api.Spot.Unsubscribe("publicTrade.BTCUSDT")
}

func TestWsManagerKlineSubscribe(t *testing.T) {
	api := getApi()
	api.ConfigureTestNetUrls()
	api.Spot.Subscribe("kline.1.BTCUSDT.1m")
	for i := 0; i < 10; i++ {
		fmt.Println(<-api.Spot.DataCh)
	}
	api.Spot.Unsubscribe("kline.1.BTCUSDT.1m")
}

func TestWsManagerLiquidationSubscribe(t *testing.T) {
	api := getApi()
	api.ConfigureTestNetUrls()
	api.Linear.Subscribe("liquidation.BTCUSDT")
	for i := 0; i < 10; i++ {
		fmt.Println(<-api.Spot.DataCh)
	}
	api.Spot.Unsubscribe("liquidation.BTCUSDT")
}

func TestWsManagerTickerSubscribe(t *testing.T) {
	api := getApi()
	api.ConfigureTestNetUrls()
	api.Spot.Subscribe("ticker.BTCUSDT")
	for i := 0; i < 10; i++ {
		fmt.Println(<-api.Spot.DataCh)
	}
	api.Spot.Unsubscribe("ticker.BTCUSDT")
}

func TestWsConnectionPingPong(t *testing.T) {
	api := getApi()
	api.ConfigureTestNetUrls()
	api.Private.Subscribe("position")
	for i := 0; i < 5; i++ {
		data := <-api.Private.DataCh
		var pm interface{}
		if err := json.Unmarshal(data, &pm); err != nil {
			t.Errorf("Error: %v", err)
		}
		PrettyPrint(pm)
	}
}

func TestWsMultiConnection(t *testing.T) {
	api := getApi()
	api.ConfigureTestNetUrls()
	api.Spot.Subscribe("orderbook.1.BTCUSDT")
	api.Linear.Subscribe("orderbook.1.BTCUSDT")
	for i := 0; i < 10; i++ {
		select {
		case data := <-api.Spot.DataCh:
			fmt.Println(data)
		case data := <-api.Linear.DataCh:
			fmt.Println(data)
		}

	}
}

func TestPositionWs(t *testing.T) {
	api := getApi()
	api.ConfigureTestNetUrls()
	api.Private.Subscribe("position")
	for i := 0; i < 10; i++ {
		data := <-api.Private.DataCh
		fmt.Println(data)
	}
}

func TestExecutionWs(t *testing.T) {
	api := getApi()
	api.ConfigureTestNetUrls()
	api.Private.Subscribe("execution")
	data := <-api.Private.DataCh
	fmt.Println(data)
}

func TestTradeAuthWs(t *testing.T) {
	api := getApi()
	api.ConfigureTestNetUrls()
	api.Trade.Subscribe("trade")
	data := <-api.Trade.DataCh
	fmt.Println(data)
}

func TestSendRequest(t *testing.T) {
	api := getApi()
	api.ConfigureTestNetUrls()
	header := map[string]interface{}{
		"X-BAPI-TIMESTAMP": time.Now().UnixNano() / 1000000,
	}

	params := map[string]interface{}{
		"category":  "linear",
		"orderType": "Limit",
		"price":     "60000",
		"qty":       "0.001",
		"side":      "Buy",
		"symbol":    "BTCUSDT",
	}

	if err := api.Trade.sendRequest("order.create", params, header); err != nil {
		t.Fatalf("Error: %v", err)
	}
	for i := 0; i < 2; i++ {
		data := <-api.Trade.DataCh
		var pm interface{}
		if err := json.Unmarshal(data, &pm); err != nil {
			t.Errorf("Error: %v", err)
		}
		PrettyPrint(pm)
	}
}
