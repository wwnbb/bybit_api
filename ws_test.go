package bybit_api

import (
	"context"
	"fmt"
	"testing"
	"time"

	pp "github.com/wwnbb/pprint"
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
	ctx := context.Background()
	api := NewBybitApi("xH43Zpsk5MMvoAYz0J", "89aQHZd9P2xH3oTQDOPMuKIi2LvLgu1TcP6Z", ctx)
	api.ConfigureMainNetDemoUrls()
	return api
}

func TestGetReqId(t *testing.T) {
	api := getApi()

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
	err := api.Spot.Subscribe("orderbook.1.BTCUSDT")
	if err != nil {
		t.Fatalf("Colud not connect to ws %v", err)
	}
	for i := 0; i < 1000; i++ {
		fmt.Println(<-api.Spot.DataCh)
	}
	api.Spot.Unsubscribe("orderbook.1.BTCUSDT")
}

func TestWsManagerTradeSubscribe(t *testing.T) {
	api := getApi()
	api.Spot.Subscribe("publicTrade.BTCUSDT")
	for i := 0; i < 10; i++ {
		fmt.Println(<-api.Spot.DataCh)
	}
	api.Spot.Unsubscribe("publicTrade.BTCUSDT")
}

func TestWsManagerKlineSubscribe(t *testing.T) {
	api := getApi()
	api.Inverse.Subscribe("kline.1.BTCUSD")
	for i := 0; i < 100; i++ {
		fmt.Println(<-api.Inverse.DataCh)
	}
	api.Inverse.Unsubscribe("kline.1.BTCUSD")
}

func TestWsManagerKlineSubscribeInverseBTCUSD(t *testing.T) {
	api := getApi()
	api.Inverse.Subscribe("kline.1.BTCUSD")
	for i := 0; i < 100; i++ {
		fmt.Println(<-api.Inverse.DataCh)
	}
	api.Inverse.Unsubscribe("kline.1.BTCUSD")
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
	api.Spot.Subscribe("tickers.BTCUSDT")
	for i := 0; i < 100; i++ {
		fmt.Println(<-api.Spot.DataCh)
	}
	api.Spot.Unsubscribe("ticker.BTCUSDT")
}

func TestWsManagerTickerSubscribeLinear(t *testing.T) {
	api := getApi()
	api.Linear.Subscribe("tickers.BTCUSDT")
	for i := 0; i < 100; i++ {
		fmt.Println(<-api.Linear.DataCh)
	}
	api.Spot.Unsubscribe("ticker.BTCUSDT")
}

func TestWsConnectionPingPong(t *testing.T) {
	api := getApi()
	api.Private.Subscribe("wallet")
	for i := 0; i < 5; i++ {
		data := <-api.Private.DataCh
		pp.PrettyPrint(data)
	}
}

func TestWsMultiConnection(t *testing.T) {
	api := getApi()
	api.Spot.Subscribe("orderbook.1.BTCUSDT")
	api.Linear.Subscribe("orderbook.1.BTCUSDT")
	for i := 0; i < 100; i++ {
		select {
		case data := <-api.Spot.DataCh:
			fmt.Println("Spot")
			pp.PrettyPrint(data)
		case data := <-api.Linear.DataCh:
			fmt.Println("Linear")
			pp.PrettyPrint(data)
		}

	}
}

func TestPositionWs(t *testing.T) {
	api := getApi()
	api.Private.Subscribe("position")
	for i := 0; i < 100; i++ {
		data := <-api.Private.DataCh
		fmt.Println(data)
	}
}

func TestBalanceWs(t *testing.T) {
	api := getApi()
	api.Private.Subscribe("wallet")
	for i := 0; i < 1000; i++ {
		data := <-api.Private.DataCh
		fmt.Println("*")
		pp.PrettyPrint(data)
		fmt.Println("*")
	}
}

func TestExecutionWs(t *testing.T) {
	api := getApi()
	api.Private.Subscribe("execution")
	data := <-api.Private.DataCh
	fmt.Println(data)
}

func TestTradeAuthWs(t *testing.T) {
	api := getApi()
	err := api.Trade.Subscribe("trade")
	if err != nil {
		t.Fatalf("Could not subscribe to trade channel: %v", err)
	}
	data := <-api.Trade.DataCh
	fmt.Println(data)
}

func TestSendRequest(t *testing.T) {
	api := getApi()
	header := map[string]interface{}{
		"X-BAPI-TIMESTAMP": time.Now().UnixNano() / 1000000,
	}

	params := map[string]interface{}{
		"category":  "linear",
		"orderType": "Market",
		// "price":     "60000",
		"qty":    "0.001",
		"side":   "Buy",
		"symbol": "BTCUSDT",
	}

	if err := api.Trade.sendRequest("order.create", params, header); err != nil {
		t.Fatalf("Error: %v", err)
	}
	for i := 0; i < 2; i++ {
		data := <-api.Trade.DataCh
		pp.PrettyPrint(data.Data)
	}
}

func TestWsDisconnect(t *testing.T) {
	api := getApi()
	api.Spot.Subscribe("orderbook.1.BTCUSDT")
	api.Linear.Subscribe("orderbook.1.BTCUSDT")
	api.Disconnect()
	if api.Spot.Conn.GetState() != StatusDisconnected && api.Linear.Conn.GetState() != StatusDisconnected {
		t.Errorf("Error: spot connection not disconnected")
	}

}
