/*
 * Copyright (c) 2024 QuantHill <info@quanthill.ae>
 */
package qant_api_bybit

import (
	"errors"
	"fmt"
	"strings"
)

// https://bybit-exchange.github.io/docs/v5/market/time#response-parameters
type GetServerTimeResponse struct {
	RetCode int    `json:"retCode"`
	RetMsg  string `json:"retMsg"`
	Result  struct {
		TimeSecond string `json:"timeSecond"`
		TimeNano   string `json:"timeNano"`
	} `json:"result"`
	RetExtInfo struct{} `json:"retExtInfo"`
	Time       int      `json:"time"`
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
	RetCode    int         `json:"retCode"`
	RetMsg     string      `json:"retMsg"`
	Result     OrderResult `json:"result"`
	RetExtInfo struct{}    `json:"retExtInfo"`
	Time       int64       `json:"time"`
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

func (p *OrderRequestParams) Validate() error {
	if p.Category == "" {
		return errors.New("category is required")
	}

	if p.Symbol == "" {
		return errors.New("symbol is required")
	}

	if p.OrderID == "" && p.OrderLinkID == "" {
		return errors.New("either orderId or orderLinkId must be provided")
	}

	validCategories := map[string]bool{"linear": true, "inverse": true, "spot": true, "option": true}
	if !validCategories[p.Category] {
		return fmt.Errorf("invalid category: %s", p.Category)
	}

	if p.Category == "spot" && p.OrderFilter != "" {
		validFilters := map[string]bool{"Order": true, "tpslOrder": true, "StopOrder": true}
		if !validFilters[p.OrderFilter] {
			return fmt.Errorf("invalid orderFilter: %s", p.OrderFilter)
		}
	}

	if p.Symbol != strings.ToUpper(p.Symbol) {
		return errors.New("symbol must be uppercase")
	}

	return nil
}

// https://bybit-exchange.github.io/docs/v5/order/cancel-order#response-parameters
type OrderResponse struct {
	OrderID     string `json:"orderId"`
	OrderLinkID string `json:"orderLinkId"`
}

// Get Open & Closed Orders
// https://bybit-exchange.github.io/docs/v5/order/open-order#request-parameters
type OpenOrderRequest struct {
	Category    string `json:"category"`
	Symbol      string `json:"symbol,omitempty"`
	BaseCoin    string `json:"baseCoin,omitempty"`
	SettleCoin  string `json:"settleCoin,omitempty"`
	OrderID     string `json:"orderId,omitempty"`
	OrderLinkID string `json:"orderLinkId,omitempty"`
	OpenOnly    *int   `json:"openOnly,omitempty"`
	OrderFilter string `json:"orderFilter,omitempty"`
	Limit       *int   `json:"limit,omitempty"`
	Cursor      string `json:"cursor,omitempty"`
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

func (r *OpenOrderRequest) Validate() error {
	if r.Category == "" {
		return errors.New("category is required")
	}

	validCategories := map[string]bool{"linear": true, "inverse": true, "spot": true, "option": true}
	if !validCategories[r.Category] {
		return fmt.Errorf("invalid category: %s", r.Category)
	}

	if r.Category == "linear" && r.Symbol == "" && r.BaseCoin == "" && r.SettleCoin == "" {
		return errors.New("linear category requires either symbol, baseCoin, or settleCoin")
	}

	if r.Symbol != "" && r.Symbol != strings.ToUpper(r.Symbol) {
		return errors.New("symbol must be uppercase")
	}

	if r.BaseCoin != "" && r.BaseCoin != strings.ToUpper(r.BaseCoin) {
		return errors.New("baseCoin must be uppercase")
	}

	if r.SettleCoin != "" && r.SettleCoin != strings.ToUpper(r.SettleCoin) {
		return errors.New("settleCoin must be uppercase")
	}

	if r.Limit != nil && (*r.Limit < 1 || *r.Limit > 50) {
		return errors.New("limit must be between 1 and 50")
	}

	if r.OrderFilter != "" {
		validFilters := map[string]bool{
			"Order":                  true,
			"StopOrder":              true,
			"tpslOrder":              true,
			"OcoOrder":               true,
			"BidirectionalTpslOrder": true,
		}
		if !validFilters[r.OrderFilter] {
			return fmt.Errorf("invalid orderFilter: %s", r.OrderFilter)
		}
	}

	return nil
}

// https://bybit-exchange.github.io/docs/v5/order/open-order#response-parameters
type GetOrdersResponse struct {
	Category       string      `json:"category"`
	NextPageCursor string      `json:"nextPageCursor"`
	List           []OrderInfo `json:"list"`
}
