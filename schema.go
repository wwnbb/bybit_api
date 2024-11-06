/*
 * Copyright (c) 2024 QuantHill <info@quanthill.ae>
 */
package qant_api_bybit

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type BaseResponse struct {
	RetCode    int      `json:"retCode"`
	RetMsg     string   `json:"retMsg"`
	RetExtInfo struct{} `json:"retExtInfo"`
	Time       int      `json:"time"`
}

// https://bybit-exchange.github.io/docs/v5/market/time#response-parameters
type GetServerTimeResponse struct {
	BaseResponse
	Result struct {
		TimeSecond string `json:"timeSecond"`
		TimeNano   string `json:"timeNano"`
	} `json:"result"`
}

type MarketUnit string

const (
	QuoteCoin MarketUnit = "quoteCoin"
	BaseCoin  MarketUnit = "baseCoin"
)

// https://bybit-exchange.github.io/docs/v5/order/create-order#request-parameters
type PlaceOrderParams struct {
	Category         string      `json:"category"`
	Symbol           string      `json:"symbol"`
	IsLeverage       *int        `json:"isLeverage,omitempty"`
	Side             string      `json:"side"`
	OrderType        string      `json:"orderType"`
	Qty              string      `json:"qty"`
	MarketUnit       *MarketUnit `json:"marketUnit,omitempty"`
	Price            *string     `json:"price,omitempty"`
	TriggerDirection *int        `json:"triggerDirection,omitempty"`
	OrderFilter      *string     `json:"orderFilter,omitempty"`
	TriggerPrice     *string     `json:"triggerPrice,omitempty"`
	TriggerBy        *string     `json:"triggerBy,omitempty"`
	OrderIv          *string     `json:"orderIv,omitempty"`
	TimeInForce      *string     `json:"timeInForce,omitempty"`
	PositionIdx      *int        `json:"positionIdx,omitempty"`
	OrderLinkId      *string     `json:"orderLinkId,omitempty"`
	TakeProfit       *string     `json:"takeProfit,omitempty"`
	StopLoss         *string     `json:"stopLoss,omitempty"`
	TpTriggerBy      *string     `json:"tpTriggerBy,omitempty"`
	SlTriggerBy      *string     `json:"slTriggerBy,omitempty"`
	ReduceOnly       *bool       `json:"reduceOnly,omitempty"`
	CloseOnTrigger   *bool       `json:"closeOnTrigger,omitempty"`
	SmpType          *string     `json:"smpType,omitempty"`
	Mmp              *bool       `json:"mmp,omitempty"`
	TpslMode         *string     `json:"tpslMode,omitempty"`
	TpLimitPrice     *string     `json:"tpLimitPrice,omitempty"`
	SlLimitPrice     *string     `json:"slLimitPrice,omitempty"`
	TpOrderType      *string     `json:"tpOrderType,omitempty"`
	SlOrderType      *string     `json:"slOrderType,omitempty"`
}

// https://bybit-exchange.github.io/docs/v5/order/create-order#response-parameters
type PlaceOrderResponse struct {
	BaseResponse
	Result OrderResult `json:"result"`
}

type OrderResult struct {
	OrderId     string `json:"orderId"`
	OrderLinkId string `json:"orderLinkId"`
}

// https://bybit-exchange.github.io/docs/v5/order/amend-order#request-parameters
type AmendOrderParams struct {
	Category     string  `json:"category"`
	Symbol       string  `json:"symbol"`
	OrderId      *string `json:"orderId,omitempty"`
	OrderLinkId  *string `json:"orderLinkId,omitempty"`
	OrderIv      *string `json:"orderIv,omitempty"`
	TriggerPrice *string `json:"triggerPrice,omitempty"`
	Qty          *string `json:"qty,omitempty"`
	Price        *string `json:"price,omitempty"`
	TpslMode     *string `json:"tpslMode,omitempty"`
	TakeProfit   *string `json:"takeProfit,omitempty"`
	StopLoss     *string `json:"stopLoss,omitempty"`
	TpTriggerBy  *string `json:"tpTriggerBy,omitempty"`
	SlTriggerBy  *string `json:"slTriggerBy,omitempty"`
	TriggerBy    *string `json:"triggerBy,omitempty"`
	TpLimitPrice *string `json:"tpLimitPrice,omitempty"`
	SlLimitPrice *string `json:"slLimitPrice,omitempty"`
}

type AmendOrderResponse struct {
	RetCode    int         `json:"retCode"`
	RetMsg     string      `json:"retMsg"`
	Result     OrderResult `json:"result"`
	RetExtInfo struct{}    `json:"retExtInfo"`
	Time       int64       `json:"time"`
}

// https://bybit-exchange.github.io/docs/v5/order/cancel-order#request-parameters
type OrderRequestParams struct {
	Category    string `json:"category"`
	Symbol      string `json:"symbol"`
	OrderID     string `json:"orderId,omitempty"`
	OrderLinkID string `json:"orderLinkId,omitempty"`
	OrderFilter string `json:"orderFilter,omitempty"`
}

// https://bybit-exchange.github.io/docs/v5/order/cancel-order#response-parameters
type OrderResponse struct {
	OrderID     string `json:"orderId"`
	OrderLinkID string `json:"orderLinkId"`
}

// Get Open & Closed Orders
// https://bybit-exchange.github.io/docs/v5/order/open-order#request-parameters
type OpenOrderRequest struct {
	Category    string `schema:"category"`
	Symbol      string `schema:"symbol,omitempty"`
	BaseCoin    string `schema:"baseCoin,omitempty"`
	SettleCoin  string `schema:"settleCoin,omitempty"`
	OrderID     string `schema:"orderId,omitempty"`
	OrderLinkID string `schema:"orderLinkId,omitempty"`
	OpenOnly    *int   `schema:"openOnly,omitempty"`
	OrderFilter string `schema:"orderFilter,omitempty"`
	Limit       *int   `schema:"limit,omitempty"`
	Cursor      string `schema:"cursor,omitempty"`
}

type OrderInfo struct {
	OrderID            string     `json:"orderId"`
	OrderLinkID        string     `json:"orderLinkId"`
	BlockTradeID       string     `json:"blockTradeId"`
	Symbol             string     `json:"symbol"`
	Price              string     `json:"price"`
	Qty                string     `json:"qty"`
	Side               string     `json:"side"`
	IsLeverage         string     `json:"isLeverage"`
	PositionIdx        int        `json:"positionIdx"`
	OrderStatus        string     `json:"orderStatus"`
	CreateType         string     `json:"createType,omitempty"`
	CancelType         string     `json:"cancelType"`
	RejectReason       string     `json:"rejectReason"`
	AvgPrice           string     `json:"avgPrice"`
	LeavesQty          string     `json:"leavesQty"`
	LeavesValue        string     `json:"leavesValue"`
	CumExecQty         string     `json:"cumExecQty"`
	CumExecValue       string     `json:"cumExecValue"`
	CumExecFee         string     `json:"cumExecFee"`
	TimeInForce        string     `json:"timeInForce"`
	OrderType          string     `json:"orderType"`
	StopOrderType      string     `json:"stopOrderType"`
	OrderIv            string     `json:"orderIv"`
	MarketUnit         MarketUnit `json:"marketUnit"`
	TriggerPrice       string     `json:"triggerPrice"`
	TakeProfit         string     `json:"takeProfit"`
	StopLoss           string     `json:"stopLoss"`
	TpslMode           string     `json:"tpslMode"`
	OcoTriggerBy       string     `json:"ocoTriggerBy"`
	TpLimitPrice       string     `json:"tpLimitPrice"`
	SlLimitPrice       string     `json:"slLimitPrice"`
	TpTriggerBy        string     `json:"tpTriggerBy"`
	SlTriggerBy        string     `json:"slTriggerBy"`
	TriggerDirection   int        `json:"triggerDirection"`
	TriggerBy          string     `json:"triggerBy"`
	LastPriceOnCreated string     `json:"lastPriceOnCreated"`
	ReduceOnly         bool       `json:"reduceOnly"`
	CloseOnTrigger     bool       `json:"closeOnTrigger"`
	PlaceType          string     `json:"placeType"`
	SmpType            string     `json:"smpType"`
	SmpGroup           int        `json:"smpGroup"`
	SmpOrderID         string     `json:"smpOrderId"`
	CreatedTime        string     `json:"createdTime"`
	UpdatedTime        string     `json:"updatedTime"`
}

// https://bybit-exchange.github.io/docs/v5/order/open-order#response-parameters
type GetOrdersResponse struct {
	BaseResponse
	Result struct {
		Category       string      `json:"category"`
		NextPageCursor string      `json:"nextPageCursor"`
		List           []OrderInfo `json:"list"`
	} `json:"result"`
}

// https://bybit-exchange.github.io/docs/v5/market/kline#request-parameters
type GetKlineParams struct {
	Category string `schema:"category,omitempty"`
	Symbol   string `schema:"symbol"`
	Interval string `schema:"interval"`
	Start    *int64 `schema:"start,omitempty"`
	End      *int64 `schema:"end,omitempty"`
	Limit    *int   `schema:"limit,omitempty"`
}

type KlineItem [7]string

func (k KlineItem) StartTime() string  { return k[0] }
func (k KlineItem) OpenPrice() string  { return k[1] }
func (k KlineItem) HighPrice() string  { return k[2] }
func (k KlineItem) LowPrice() string   { return k[3] }
func (k KlineItem) ClosePrice() string { return k[4] }
func (k KlineItem) Volume() string     { return k[5] }
func (k KlineItem) Turnover() string   { return k[6] }

// GetKlineResponse represents the response from the /v5/market/kline endpoint
type GetKlineResponse struct {
	BaseResponse
	Result KlineResult `json:"result"`
}

type KlineResult struct {
	Category string      `json:"category"`
	Symbol   string      `json:"symbol"`
	List     []KlineItem `json:"list"`
}

func (k KlineItem) ToFloat64() (startTime, open, high, low, close, volume, turnover float64, err error) {
	values := make([]float64, 7)
	for i, v := range k {
		values[i], err = strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, 0, 0, 0, 0, 0, 0, fmt.Errorf("failed to parse value at index %d: %w", i, err)
		}
	}
	return values[0], values[1], values[2], values[3], values[4], values[5], values[6], nil
}

const (
	Interval1Min    = "1"
	Interval3Min    = "3"
	Interval5Min    = "5"
	Interval15Min   = "15"
	Interval30Min   = "30"
	Interval60Min   = "60"
	Interval120Min  = "120"
	Interval240Min  = "240"
	Interval360Min  = "360"
	Interval720Min  = "720"
	IntervalDaily   = "D"
	IntervalWeekly  = "W"
	IntervalMonthly = "M"
)

type GetOrderbookParams struct {
	Category string `schema:"category"`
	Symbol   string `schema:"symbol"`
	Limit    *int   `schema:"limit,omitempty"`
}

type OrderbookItem [2]string

func (i OrderbookItem) Price() string { return i[0] }
func (i OrderbookItem) Size() string  { return i[1] }

type GetOrderbookResponse struct {
	BaseResponse
	Result OrderbookResult `json:"result"`
}

type OrderbookResult struct {
	Symbol        string          `json:"s"`
	Bids          []OrderbookItem `json:"b"`
	Asks          []OrderbookItem `json:"a"`
	Timestamp     int64           `json:"ts"`
	UpdateID      int64           `json:"u"`
	CrossSequence int64           `json:"seq"`
	CreateTime    int64           `json:"cts"`
}

type GetInstrumentsInfoParams struct {
	Category string `schema:"category"`
	Symbol   string `schema:"symbol,omitempty"`
	Status   string `schema:"status,omitempty"`
	BaseCoin string `schema:"baseCoin,omitempty"`
	Limit    *int   `schema:"limit,omitempty"`
	Cursor   string `schema:"cursor,omitempty"`
}

type LotSizeFilter struct {
	BasePrecision  string `json:"basePrecision"`
	QuotePrecision string `json:"quotePrecision"`
	MinOrderQty    string `json:"minOrderQty"`
	MaxOrderQty    string `json:"maxOrderQty"`
	MinOrderAmt    string `json:"minOrderAmt"`
	MaxOrderAmt    string `json:"maxOrderAmt"`
}

type PriceFilter struct {
	TickSize string `json:"tickSize"`
}

type RiskParameters struct {
	LimitParameter  string `json:"limitParameter"`
	MarketParameter string `json:"marketParameter"`
}

type InstrumentInfo struct {
	Symbol         string         `json:"symbol"`
	BaseCoin       string         `json:"baseCoin"`
	QuoteCoin      string         `json:"quoteCoin"`
	Innovation     string         `json:"innovation"`
	Status         string         `json:"status"`
	MarginTrading  string         `json:"marginTrading"`
	LotSizeFilter  LotSizeFilter  `json:"lotSizeFilter"`
	PriceFilter    PriceFilter    `json:"priceFilter"`
	RiskParameters RiskParameters `json:"riskParameters"`
}

type GetInstrumentsInfoResponse struct {
	BaseResponse
	Result struct {
		Category       string           `json:"category"`
		List           []InstrumentInfo `json:"list"`
		NextPageCursor string           `json:"nextPageCursor,omitempty"`
	} `json:"result"`
}
