package bybit_api

import (
	"strconv"
	"time"

	pp "github.com/wwnbb/pprint"
)

type TradeWSBybit struct {
	*WSBybit
}

func newTradeWSManager(api *BybitApi, url string) *TradeWSBybit {
	ws := newWSBybit(api, WS_TRADE, url)
	return &TradeWSBybit{
		WSBybit: ws,
	}
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

type BatchPlaceOrderWsSchema struct {
	ReqId  string                  `json:"reqId"`
	Header HeaderData              `json:"header"`
	Op     string                  `json:"op"`
	Args   []BatchPlaceOrderParams `json:"args"`
}

func (m *TradeWSBybit) PlaceOrder(params PlaceOrderParams) (string, error) {
	reqId := m.GetReqId("order.create")
	placeOrderMsg := PlaceOrderWsSchema{
		ReqId:  reqId,
		Header: GenerateAPIHeaders(),
		Op:     "order.create",
		Args:   []PlaceOrderParams{params},
	}
	m.Logger.Debug("Placing order", "reqId", reqId, "order", pp.PrettyFormat(placeOrderMsg))
	return reqId, m.WSManager.SendRequest(placeOrderMsg)
}

func (m *TradeWSBybit) CancelOrder(params CancelOrderParams) error {
	reqId := m.GetReqId("order.cancel")
	cancelOrderMsg := CancelOrderWsSchema{
		ReqId:  reqId,
		Header: GenerateAPIHeaders(),
		Op:     "order.cancel",
		Args:   []CancelOrderParams{params},
	}
	m.Logger.Debug("Canceling order", "reqId", reqId, "order", pp.PrettyFormat(cancelOrderMsg))
	return m.WSManager.SendRequest(cancelOrderMsg)
}

func (m *TradeWSBybit) AmendOrder(params AmendOrderParams) error {
	reqId := m.GetReqId("order.amend")
	amendOrderMsg := AmendOrderWsSchema{
		ReqId:  reqId,
		Header: GenerateAPIHeaders(),
		Op:     "order.amend",
		Args:   []AmendOrderParams{params},
	}
	m.Logger.Debug("Amending order", "reqId", reqId, "order", pp.PrettyFormat(amendOrderMsg))
	return m.WSManager.SendRequest(amendOrderMsg)
}

func (m *TradeWSBybit) BatchPlaceOrder(params BatchPlaceOrderParams) (string, error) {
	reqId := m.GetReqId("order.create-batch")
	batchPlaceOrderMsg := BatchPlaceOrderWsSchema{
		ReqId:  reqId,
		Header: GenerateAPIHeaders(),
		Op:     "order.create-batch",
		Args:   []BatchPlaceOrderParams{params},
	}
	m.Logger.Debug("Placing batch orders", "reqId", reqId, "order", pp.PrettyFormat(batchPlaceOrderMsg))
	return reqId, m.WSManager.SendRequest(batchPlaceOrderMsg)
}
