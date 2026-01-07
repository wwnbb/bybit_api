package bybit_api

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"log/slog"
	"net/http"
	_ "net/http/pprof"

	pp "github.com/wwnbb/pprint"
	"github.com/wwnbb/wsmanager/states"
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

func TestGetReqId(t *testing.T) {
	api := GetApi()

	wsm := newWSBybit(api, WS_LINEAR, TESTNET_SPOT_WS)
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
	api := GetApi()
	err := api.Spot.Connect()
	if err != nil {
		t.Fatalf("Could not connect to ws %v", err)
	}
	err = api.Spot.Subscribe("orderbook.1.BTCUSDT")
	if err != nil {
		t.Fatalf("Could not subscribe %v", err)
	}
	for i := 0; i < 10; i++ {
		fmt.Println(<-api.Spot.DataCh)
	}
	for len(api.Spot.DataCh) > 0 {
		<-api.Spot.DataCh
	}
	api.Spot.Unsubscribe("orderbook.1.BTCUSDT")
	time.Sleep(10 * time.Second)
	api.Spot.Close()
	time.Sleep(10 * time.Second)
	api.Spot.Connect()
	time.Sleep(10 * time.Second)
	api.Spot.Close()
	time.Sleep(10 * time.Second)
	api.Spot.Connect()
	time.Sleep(10 * time.Second)
	api.Spot.Close()
	time.Sleep(10 * time.Second)
	api.Spot.Connect()
	time.Sleep(10 * time.Second)
	api.Spot.Subscribe("orderbook.1.BTCUSDT")
	for len(api.Spot.DataCh) > 0 {
		<-api.Spot.DataCh
	}
	time.Sleep(10 * time.Second)
}

func TestWsSubscribeUnsubscribe(t *testing.T) {
	api := GetApi()
	err := api.Spot.Connect()
	if err != nil {
		t.Fatalf("Could not connect to ws %v", err)
	}
	err = api.Spot.Subscribe("orderbook.1.BTCUSDT")
	if err != nil {
		t.Fatalf("Could not subscribe %v", err)
	}
	err = api.Spot.Subscribe("orderbook.1.BTCUSDT")
	if err != nil {
		t.Fatalf("Could not subscribe %v", err)
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
	api := GetApi()
	api.Spot.Connect()
	err := api.Spot.Subscribe("publicTrade.BTCUSDT")
	if err != nil {
		t.Fatalf("Could not subscribe to publicTrade channel: %v", err)
	}
	for i := 0; i < 10; i++ {
		msg := <-api.Spot.DataCh
		fmt.Printf("\n%s\n", pp.PrettyFormat(msg.Data))
	}
	api.Spot.Unsubscribe("publicTrade.BTCUSDT")
}

func TestWsManagerKlineSubscribe(t *testing.T) {
	api := GetApi()
	api.Spot.Subscribe("kline.1.BTCUSDT")
	for i := 0; i < 100; i++ {
		fmt.Println(<-api.Inverse.DataCh)
	}
	api.Spot.Unsubscribe("kline.1.BTCUSDT")
}

func TestWsManagerKlineSubscribeInverseBTCUSD(t *testing.T) {
	api := GetApi()
	api.Linear.Subscribe("kline.1.BTCUSDT")
	for i := 0; i < 100; i++ {
		fmt.Println(<-api.Linear.DataCh)
	}
	api.Linear.Unsubscribe("kline.1.BTCUSDT")
}

func TestWsManagerLiquidationSubscribe(t *testing.T) {
	api := GetApi()
	api.ConfigureTestNetUrls()
	api.Linear.Subscribe("liquidation.BTCUSDT")
	for i := 0; i < 10; i++ {
		fmt.Println(<-api.Spot.DataCh)
	}
	api.Spot.Unsubscribe("liquidation.BTCUSDT")
}

func TestWsManagerTickerSubscribe(t *testing.T) {
	api := GetApi()
	api.Spot.Subscribe("tickers.BTCUSDT")
	for i := 0; i < 100; i++ {
		fmt.Println(<-api.Spot.DataCh)
	}
	api.Spot.Unsubscribe("ticker.BTCUSDT")
}

func TestWsManagerTickerSubscribeLinear(t *testing.T) {
	api := GetApi()
	api.Linear.Subscribe("tickers.BTCUSDT")
	for i := 0; i < 100; i++ {
		fmt.Println(<-api.Linear.DataCh)
	}
	api.Spot.Unsubscribe("ticker.BTCUSDT")
}

func TestWsConnectionPingPong(t *testing.T) {
	api := GetApi()
	api.Private.Subscribe("wallet")
	for i := 0; i < 5; i++ {
		data := <-api.Private.DataCh
		pp.PrettyPrint(data)
	}
}

func TestWsMultiConnection(t *testing.T) {
	api := GetApi()
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
	api := GetApi()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	api.SetLogger(logger)
	api.Private.Subscribe("position")
	for i := 0; i < 100; i++ {
		data := <-api.Private.DataCh
		fmt.Println(data)
	}
}

func TestOrdersWs(t *testing.T) {
	api := GetApi()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	api.SetLogger(logger)
	api.ConfigureMainNetUrls()
	api.Private.Subscribe("execution")
	for {
		fmt.Println(pp.PrettyFormat(<-api.Private.DataCh))
	}
}

func TestBalanceWs(t *testing.T) {
	api := GetApi()
	api.Private.Subscribe("wallet")
	for i := 0; i < 1000; i++ {
		data := <-api.Private.DataCh
		fmt.Println("*")
		pp.PrettyPrint(data)
		fmt.Println("*")
	}
}

func TestExecutionWs(t *testing.T) {
	api := GetApi()
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
	api := GetApi()
	err := api.Trade.Subscribe("trade")
	if err != nil {
		t.Fatalf("Could not subscribe to trade channel: %v", err)
	}
	data := <-api.Trade.DataCh
	fmt.Println(data)
}

func TestSendRequest(t *testing.T) {
	api := GetApi()
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
	api := GetApi()
	api.Spot.Subscribe("orderbook.1.BTCUSDT")
	api.Linear.Subscribe("orderbook.1.BTCUSDT")
	api.Disconnect()
	if api.Spot.GetConnState() != states.StateDisconnected && api.Linear.GetConnState() != states.StateDisconnected {
		t.Errorf("Error: spot connection not disconnected")
	}
	time.Sleep(20 * time.Second)
}

func TestWsSubscribeAll(t *testing.T) {
	api := GetApi()
	defer api.Disconnect()
	api2 := GetApi()
	api3 := GetApi()
	api4 := GetApi()
	api5 := GetApi()
	api6 := GetApi()

	go func() {
		for {
			fmt.Println(<-api.Private.DataCh)
		}
	}()

	go func() {
		for {
			fmt.Println(<-api2.Spot.DataCh)
		}
	}()
	go func() {
		for {
			fmt.Println(<-api2.Linear.DataCh)
		}
	}()
	go func() {
		for {
			fmt.Println(<-api3.Linear.DataCh)
		}
	}()
	go func() {
		for {
			fmt.Println(<-api4.Spot.DataCh)
		}
	}()
	go func() {
		for {
			fmt.Println(<-api5.Spot.DataCh)
		}
	}()
	go func() {
		for {
			fmt.Println(<-api6.Spot.DataCh)
		}
	}()

	go http.ListenAndServe("localhost:6060", nil)
	error := api.Private.Subscribe("execution")
	fmt.Println("Subscribe execution errkr:", error)
	api.Private.Subscribe("position")
	api.Private.Subscribe("wallet")
	api.Private.Subscribe("order")

	api2.Spot.Subscribe("orderbook.50.BTCUSDT")
	api2.Spot.Subscribe("publicTrade.BTCUSDT")
	api2.Spot.Subscribe("kline.1.BTCUSDT")
	api2.Spot.Subscribe("tickers.BTCUSDT")

	api2.Linear.Subscribe("orderbook.50.BTCUSDT")
	api2.Linear.Subscribe("publicTrade.BTCUSDT")
	api2.Linear.Subscribe("kline.1.BTCUSDT")
	api2.Linear.Subscribe("tickers.BTCUSDT")

	api3.Linear.Subscribe("orderbook.50.ASTERUSDT")
	api3.Linear.Subscribe("publicTrade.ASTERUSDT")
	api3.Linear.Subscribe("kline.1.ASTERUSDT")
	api3.Linear.Subscribe("tickers.ASTERUSDT")

	api4.Spot.Subscribe("orderbook.50.XPLUSDT")
	api4.Spot.Subscribe("publicTrade.XPLUSDT")
	api4.Spot.Subscribe("kline.1.XPLUSDT")
	api4.Spot.Subscribe("tickers.XPLUSDT")

	api5.Spot.Subscribe("orderbook.50.ASTERUSDT")
	api5.Spot.Subscribe("publicTrade.ASTERUSDT")
	api5.Spot.Subscribe("kline.1.ASTERUSDT")
	api5.Spot.Subscribe("tickers.ASTERUSDT")

	api6.Spot.Subscribe("kline.1.DOGEUSDT")
	api6.Spot.Subscribe("tickers.DOGEUSDT")

	time.Sleep(100 * time.Second)
}
