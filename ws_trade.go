package bybit_api

import (
	"fmt"
	"strconv"
	"time"

	"github.com/dustinxie/lockfree"
	pp "github.com/wwnbb/pprint"
)

type TradeWSManager struct {
	WSManager
}

func newTradeWSManager(api *BybitApi, url string) *TradeWSManager {
	fmt.Printf("Trade WS URL: %s\n", url)
	wsm := &TradeWSManager{
		WSManager: WSManager{
			api:           api,
			wsType:        WS_TRADE,
			subscriptions: make(map[string]int32),
			requestIds:    lockfree.NewHashMap(),
			url:           url,
			connState:     StateNew,

			DataCh: make(chan WsMsg, 100),
		}}
	return wsm
}

type HeaderData struct {
	Timestamp  string `json:"X-BAPI-TIMESTAMP"`
	RecvWindow string `json:"X-BAPI-RECV-WINDOW"`
}

func GenerateAPIHeaders() HeaderData {
	timestamp := time.Now().UnixMilli()

	return HeaderData{
		Timestamp:  strconv.FormatInt(timestamp, 10),
		RecvWindow: "3000",
	}
}

//	{
//	    "reqId": "test-005",
//	    "header": {
//	        "X-BAPI-TIMESTAMP": "1711001595207",
//	        "X-BAPI-RECV-WINDOW": "8000",
//	        "Referer": "bot-001" // for api broker
//	    },
//	    "op": "order.create",
//	    "args": [
//	        {
//	            "symbol": "ETHUSDT",
//	            "side": "Buy",
//	            "orderType": "Limit",
//	            "qty": "0.2",
//	            "price": "2800",
//	            "category": "linear",
//	            "timeInForce": "PostOnly"
//	        }
//	    ]
//	}
type PlaceOrderWsSchema struct {
	ReqId  string             `json:"reqId"`
	Header HeaderData         `json:"header"`
	Op     string             `json:"op"`
	Args   []PlaceOrderParams `json:"args"`
}

type CancelOrderWsSchema struct {
	ReqId  string              `json:"reqId"`
	Header HeaderData          `json:"header"`
	Op     string              `json:"op"`
	Args   []CancelOrderParams `json:"args"`
}

type AmendOrderWsSchema struct {
	ReqId  string             `json:"reqId"`
	Header HeaderData         `json:"header"`
	Op     string             `json:"op"`
	Args   []AmendOrderParams `json:"args"`
}

func (m *TradeWSManager) PlaceOrder(params PlaceOrderParams) (string, error) {
	if err := m.ensureConnected(); err != nil {
		return "", err
	}
	reqId := m.getReqId("order.create")
	placeOrderMsg := PlaceOrderWsSchema{
		ReqId:  reqId,
		Header: GenerateAPIHeaders(),
		Op:     "order.create",
		Args:   []PlaceOrderParams{params},
	}
	m.api.Logger.Info("Placing order", "reqId", reqId, "order", pp.PrettyFormat(placeOrderMsg))
	return reqId, m.getConn().WriteJSON(placeOrderMsg)
}

func (m *TradeWSManager) CancelOrder(params CancelOrderParams) error {
	if err := m.ensureConnected(); err != nil {
		return err
	}
	reqId := m.getReqId("order.cancel")
	placeOrderMsg := CancelOrderWsSchema{
		ReqId:  reqId,
		Header: GenerateAPIHeaders(),
		Op:     "order.cancel",
		Args:   []CancelOrderParams{params},
	}
	m.api.Logger.Info("Canceling order", "reqId", reqId, "order", pp.PrettyFormat(placeOrderMsg))
	return m.getConn().WriteJSON(placeOrderMsg)
}

func (m *TradeWSManager) AmendOrder(params AmendOrderParams) error {
	if err := m.ensureConnected(); err != nil {
		return err
	}
	reqId := m.getReqId("order.amend")
	placeOrderMsg := AmendOrderWsSchema{
		ReqId:  reqId,
		Header: GenerateAPIHeaders(),
		Op:     "order.amend",
		Args:   []AmendOrderParams{params},
	}
	m.api.Logger.Info("Amending order", "reqId", reqId, "order", pp.PrettyFormat(placeOrderMsg))
	return m.getConn().WriteJSON(placeOrderMsg)
}
