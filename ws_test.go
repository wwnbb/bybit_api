package bybit_api

import (
	"context"
	"fmt"
	"os"
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
	api := NewBybitApi(os.Getenv("BYBIT_API_KEY"), os.Getenv("BYBIT_SECRET_KEY"), ctx)
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

func TestWsSubscribeUnsubscribe(t *testing.T) {
	api := getApi()
	api.Logger.SetLogLevel(LogLevelDebug)
	err := api.Spot.Subscribe("orderbook.1.BTCUSDT")
	if err != nil {
		t.Fatalf("Colud not connect to ws %v", err)
	}
	err = api.Spot.Subscribe("orderbook.1.BTCUSDT")
	if err != nil {
		t.Fatalf("Colud not connect to ws %v", err)
	}
	api.Spot.Unsubscribe("orderbook.1.BTCUSDT")
	subscribedTopics := api.Spot.GetSubscribedTopics()
	if len(subscribedTopics) != 1 {
		t.Errorf("Should leave one topic subscribed, got %d", len(subscribedTopics))
	}
	api.Spot.Unsubscribe("orderbook.1.BTCUSDT")
	if len(api.Spot.GetSubscribedTopics()) != 0 {
		t.Errorf("Should leave no topic subscribed")
	}
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
	api.Spot.Subscribe("kline.1.BTCUSDT")
	for i := 0; i < 100; i++ {
		fmt.Println(<-api.Inverse.DataCh)
	}
	api.Spot.Unsubscribe("kline.1.BTCUSDT")
}

func TestWsManagerKlineSubscribeInverseBTCUSD(t *testing.T) {
	api := getApi()
	api.Linear.Subscribe("kline.1.BTCUSDT")
	for i := 0; i < 100; i++ {
		fmt.Println(<-api.Linear.DataCh)
	}
	api.Linear.Unsubscribe("kline.1.BTCUSDT")
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
	api.Spot.Subscribe("orderbook.30.BTCUSDT")
	api.Linear.Subscribe("orderbook.30.BTCUSDT")
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

func TestOrdersWs(t *testing.T) {
	api := getApi()
	api.Private.Subscribe("order")
	for i := 0; i < 100; i++ {
		_ = <-api.Private.DataCh
		fmt.Println("")
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
	var counter int = 0
	for {
		data := <-api.Private.DataCh
		if data.Topic == "execution" {
			counter++
			fmt.Println(data)
			fmt.Println("Counter:", counter)
		}
	}
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
		"category":  LinearCategory,
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
	if api.Spot.GetConnState() != StateDisconnected && api.Linear.api.Linear.GetConnState() != StateDisconnected {
		t.Errorf("Error: spot connection not disconnected")
	}
	time.Sleep(20 * time.Second)
}

func TestWsSubscribeAll(t *testing.T) {
	api := getApi()
	api.Logger.SetLogLevel(LogLevelInfo)
	api.Private.Subscribe("execution")
	api.Private.Subscribe("position")
	api.Private.Subscribe("wallet")
	api.Private.Subscribe("order")
	api.Linear.Subscribe("orderbook.50.BTCUSDT")
	api.Linear.Subscribe("publicTrade.BTCUSDT")
	api.Linear.Subscribe("kline.1.BTCUSDT")
	api.Linear.Subscribe("tickers.BTCUSDT")

	api.Linear.Subscribe("orderbook.50.ASTERUSDT")
	api.Linear.Subscribe("publicTrade.ASTERUSDT")
	api.Linear.Subscribe("kline.1.ASTERUSDT")
	api.Linear.Subscribe("tickers.ASTERUSDT")
	api.Linear.Subscribe("allLiquidation.ASTERUSDT")
	for {
		select {
		case data := <-api.Linear.DataCh:
			pp.PrettyPrint(data)
		case data := <-api.Private.DataCh:
			pp.PrettyPrint(data)
		}
	}
}
