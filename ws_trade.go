package bybit_api

import (
	"fmt"
	"strconv"
	"time"

	"github.com/dustinxie/lockfree"
	pp "github.com/wwnbb/pprint"
)

type TradeWSBybit struct {
	WSBybit
}

func newTradeWSManager(api *BybitApi, url string) *TradeWSBybit {
	fmt.Printf("Trade WS URL: %s\n", url)
	wsm := &TradeWSBybit{
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

func (m *TradeWSBybit) PlaceOrder(params PlaceOrderParams) (string, error) {
	reqId := m.getReqId("order.create")
	placeOrderMsg := PlaceOrderWsSchema{
		ReqId:  reqId,
		Header: GenerateAPIHeaders(),
		Op:     "order.create",
		Args:   []PlaceOrderParams{params},
	}
	m.Logger.Debug("Placing order", "reqId", reqId, "order", pp.PrettyFormat(placeOrderMsg))
	return reqId, m.SendRequest(placeOrderMsg)
}

func (m *TradeWSBybit) CancelOrder(params CancelOrderParams) error {
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
	m.api.Logger.Debug("Canceling order", "reqId", reqId, "order", pp.PrettyFormat(placeOrderMsg))
	return m.getConn().WriteJSON(placeOrderMsg)
}

func (m *TradeWSBybit) AmendOrder(params AmendOrderParams) error {
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
	m.api.Logger.Debug("Amending order", "reqId", reqId, "order", pp.PrettyFormat(placeOrderMsg))
	return m.getConn().WriteJSON(placeOrderMsg)
}
