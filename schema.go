package bybit_api

import (
	"fmt"
	"strconv"

	json "github.com/goccy/go-json"
)

type BaseResponse struct {
	RetCode    int      `json:"retCode"`
	RetMsg     string   `json:"retMsg"`
	RetExtInfo struct{} `json:"retExtInfo"`
	Time       int      `json:"time"`
}

func (b BaseResponse) GetRetCode() int {
	return b.RetCode
}

func (b BaseResponse) GetRetMsg() string {
	return b.RetMsg
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

// GetAccountInfoResponse represents the response from the Get Account Info endpoint
// https://bybit-exchange.github.io/docs/v5/account/account-info#response-parameters
type GetAccountInfoResponse struct {
	BaseResponse
	Result AccountInfo `json:"result"`
}

// AccountInfo represents the account information data
type AccountInfo struct {
	UnifiedMarginStatus int    `json:"unifiedMarginStatus"` // Account status
	MarginMode          string `json:"marginMode"`          // ISOLATED_MARGIN, REGULAR_MARGIN, PORTFOLIO_MARGIN
	IsMasterTrader      bool   `json:"isMasterTrader"`      // Whether this account is a leader (copytrading). true, false
	SpotHedgingStatus   string `json:"spotHedgingStatus"`   // Whether the unified account enables Spot hedging. ON, OFF
	UpdatedTime         string `json:"updatedTime"`         // Account data updated timestamp (ms)
	DcpStatus           string `json:"dcpStatus"`           // deprecated, always OFF. Please use Get DCP Info
	TimeWindow          int    `json:"timeWindow"`          // deprecated, always 0. Please use Get DCP Info
	SmpGroup            int    `json:"smpGroup"`            // deprecated, always 0. Please query Get SMP Group ID endpoint
}

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
	OrderId     string `json:"orderId,omitempty"`
	OrderLinkId string `json:"orderLinkId,omitempty"`
	OrderFilter string `json:"orderFilter,omitempty"`
}

// https://bybit-exchange.github.io/docs/v5/order/cancel-order#response-parameters
type OrderResponse struct {
	OrderId     string `json:"orderId"`
	OrderLinkId string `json:"orderLinkId"`
}

// CancelOrderParams represents the parameters for canceling an order
// https://bybit-exchange.github.io/docs/v5/order/cancel-order#request-parameters
type CancelOrderParams struct {
	Category    string  `json:"category"`
	Symbol      string  `json:"symbol"`
	OrderId     *string `json:"orderId,omitempty"`
	OrderLinkId *string `json:"orderLinkId,omitempty"`
	OrderFilter *string `json:"orderFilter,omitempty"`
}

// CancelOrderResponse represents the response from the cancel order endpoint
// https://bybit-exchange.github.io/docs/v5/order/cancel-order#request-parameters
type CancelOrderResponse struct {
	BaseResponse
	Result struct {
		OrderId     string `json:"orderId"`
		OrderLinkId string `json:"orderLinkId"`
	} `json:"result"`
}

// Get Open & Closed Orders
// https://bybit-exchange.github.io/docs/v5/order/open-order#request-parameters
type OpenOrderRequest struct {
	Category    string `schema:"category"`
	Symbol      string `schema:"symbol,omitempty"`
	BaseCoin    string `schema:"baseCoin,omitempty"`
	SettleCoin  string `schema:"settleCoin,omitempty"`
	OrderId     string `schema:"orderId,omitempty"`
	OrderLinkId string `schema:"orderLinkId,omitempty"`
	OpenOnly    *int   `schema:"openOnly,omitempty"`
	OrderFilter string `schema:"orderFilter,omitempty"`
	Limit       *int   `schema:"limit,omitempty"`
	Cursor      string `schema:"cursor,omitempty"`
}

// https://bybit-exchange.github.io/docs/v5/order/open-order#response-parameters
type GetOrdersResponse struct {
	BaseResponse
	Result struct {
		Category       string      `json:"category"`
		NextPageCursor string      `json:"nextPageCursor"`
		List           []OrderData `json:"list"`
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

// Subscribe to User's Position updates response
// https://bybit-exchange.github.io/docs/v5/websocket/private/position#response-parameters
type PositionWebsocketResponse struct {
	Id           string         `json:"id"`
	Topic        string         `json:"topic"`
	CreationTime int64          `json:"creationTime"`
	Data         []PositionData `json:"data"`
}

type PositionData struct {
	Category               string `json:"category"`
	Symbol                 string `json:"symbol"`
	Side                   string `json:"side"`
	Size                   string `json:"size"`
	PositionIdx            int    `json:"positionIdx"`
	TradeMode              int    `json:"tradeMode"`
	PositionValue          string `json:"positionValue"`
	RiskId                 int    `json:"riskId"`
	RiskLimitValue         string `json:"riskLimitValue"`
	EntryPrice             string `json:"entryPrice"`
	MarkPrice              string `json:"markPrice"`
	Leverage               string `json:"leverage"`
	PositionBalance        string `json:"positionBalance"`
	AutoAddMargin          int    `json:"autoAddMargin"`
	PositionIM             string `json:"positionIM"`
	PositionMM             string `json:"positionMM"`
	LiqPrice               string `json:"liqPrice"`
	BustPrice              string `json:"bustPrice"`
	TpslMode               string `json:"tpslMode"`
	TakeProfit             string `json:"takeProfit"`
	StopLoss               string `json:"stopLoss"`
	TrailingStop           string `json:"trailingStop"`
	UnrealisedPnl          string `json:"unrealisedPnl"`
	CurRealisedPnl         string `json:"curRealisedPnl"`
	SessionAvgPrice        string `json:"sessionAvgPrice"`
	Delta                  string `json:"delta,omitempty"`
	Gamma                  string `json:"gamma,omitempty"`
	Vega                   string `json:"vega,omitempty"`
	Theta                  string `json:"theta,omitempty"`
	CumRealisedPnl         string `json:"cumRealisedPnl"`
	PositionStatus         string `json:"positionStatus"`
	AdlRankIndicator       int    `json:"adlRankIndicator"`
	IsReduceOnly           bool   `json:"isReduceOnly"`
	MMRSysUpdatedTime      string `json:"mmrSysUpdatedTime"`
	LeverageSysUpdatedTime string `json:"leverageSysUpdatedTime"`
	CreatedTime            string `json:"createdTime"`
	UpdatedTime            string `json:"updatedTime"`
	Seq                    int64  `json:"seq"`
}

type OrderWebsocketCreateResponse struct {
	ConnId     string                 `json:"connId"`
	Data       OrderResponse          `json:"data"`
	Header     map[string]string      `json:"header"`
	Op         string                 `json:"op"`
	ReqId      string                 `json:"reqId"`
	RetCode    int                    `json:"retCode"`
	RetExtInfo map[string]interface{} `json:"retExtInfo"`
	RetMsg     string                 `json:"retMsg"`
}

type OrderWebsocketCancelResponse struct {
	ConnId     string                 `json:"connId"`
	Data       OrderResponse          `json:"data"`
	Header     map[string]string      `json:"header"`
	Op         string                 `json:"op"`
	ReqId      string                 `json:"reqId"`
	RetCode    int                    `json:"retCode"`
	RetExtInfo map[string]interface{} `json:"retExtInfo"`
	RetMsg     string                 `json:"retMsg"`
}

type OrderWebsocketAmendResponse struct {
	ConnId     string                 `json:"connId"`
	Data       OrderResponse          `json:"data"`
	Header     map[string]string      `json:"header"`
	Op         string                 `json:"op"`
	ReqId      string                 `json:"reqId"`
	RetCode    int                    `json:"retCode"`
	RetExtInfo map[string]interface{} `json:"retExtInfo"`
	RetMsg     string                 `json:"retMsg"`
}

type OrderWebsocketBatchCreateResponse struct {
	ConnId string            `json:"connId"`
	Data   BatchPlaceOrderResponse `json:"data"`
	Header map[string]string `json:"header"`
	Op     string            `json:"op"`
	ReqId  string            `json:"reqId"`
	RetCode int              `json:"retCode"`
	RetExtInfo struct {
		List []struct {
			Code int    `json:"code"`
			Msg  string `json:"msg"`
		} `json:"list"`
	} `json:"retExtInfo"`
	RetMsg string `json:"retMsg"`
}

type PublicTradeWebsocketResponse struct {
	Topic string            `json:"topic"`
	Type  string            `json:"type"`
	Ts    int64             `json:"ts"`
	Data  []PublicTradeData `json:"data"`
}

// PublicTradeData represents individual trade data
type PublicTradeData struct {
	T    int64  `json:"T"`   // Trade timestamp in milliseconds
	S    string `json:"s"`   // Symbol
	Side string `json:"S"`   // Side: Buy or Sell
	V    string `json:"v"`   // Volume
	P    string `json:"p"`   // Price
	L    string `json:"L"`   // Tick direction: PlusTick, ZeroPlusTick, MinusTick, ZeroMinusTick
	I    string `json:"i"`   // Trade ID
	BT   bool   `json:"BT"`  // Whether it's a block trade
	Seq  int64  `json:"seq"` // Sequence number
}

// OrderbookWebsocketResponse represents the websocket response for orderbook updates
type OrderbookWebsocketResponse struct {
	Topic string        `json:"topic"`
	Type  string        `json:"type"`
	Ts    int64         `json:"ts"`
	Data  OrderbookData `json:"data"`
	Cts   int64         `json:"cts"` // Client timestamp
}

// OrderbookData represents the orderbook data
type OrderbookData struct {
	S   string           `json:"s"`   // Symbol
	B   []OrderbookLevel `json:"b"`   // Bids
	A   []OrderbookLevel `json:"a"`   // Asks
	U   int64            `json:"u"`   // Update ID
	Seq int64            `json:"seq"` // Sequence number
}

// OrderbookLevel represents a single price level with price and quantity
type OrderbookLevel struct {
	Price    string
	Quantity string
}

// UnmarshalJSON custom unmarshaler for OrderbookLevel
func (o *OrderbookLevel) UnmarshalJSON(data []byte) error {
	var arr [2]string
	if err := json.Unmarshal(data, &arr); err != nil {
		return err
	}
	o.Price = arr[0]
	o.Quantity = arr[1]
	return nil
}

// Add this wrapper struct to match what you're actually receiving
type WebsocketMessage struct {
	Topic string                       `json:"topic"`
	Data  OrderWebsocketCreateResponse `json:"data"`
}

// https://bybit-exchange.github.io/docs/v5/websocket/private/order#response-parameters
type OrderWebsocketResponse struct {
	Id           string      `json:"id"`
	Topic        string      `json:"topic"`
	CreationTime int64       `json:"creationTime"`
	Data         []OrderData `json:"data"`
}

type OrderData struct {
	AvgPrice       string `json:"avgPrice"`
	BlockTradeId   string `json:"blockTradeId"`
	CancelType     string `json:"cancelType"`
	Category       string `json:"category"`
	CloseOnTrigger bool   `json:"closeOnTrigger"`
	// Only in websocket
	ClosedPnl string `json:"closedPnl"`

	CreateType   string `json:"createType"`
	CreatedTime  string `json:"createdTime"`
	CumExecFee   string `json:"cumExecFee"`
	CumExecQty   string `json:"cumExecQty"`
	CumExecValue string `json:"cumExecValue"`
	// Only in websocket
	FeeCurrency        string       `json:"feeCurrency"`
	IsLeverage         string       `json:"isLeverage"`
	LastPriceOnCreated string       `json:"lastPriceOnCreated"`
	LeavesQty          string       `json:"leavesQty"`
	LeavesValue        string       `json:"leavesValue"`
	MarketUnit         string       `json:"marketUnit"`
	OcoTriggerBy       string       `json:"ocoTriggerBy"`
	OrderId            string       `json:"orderId"`
	OrderIv            string       `json:"orderIv"`
	OrderLinkId        string       `json:"orderLinkId"`
	OrderStatus        OrderStatusT `json:"orderStatus"`
	OrderType          OrderTypeT   `json:"orderType"`
	PlaceType          string       `json:"placeType"`
	PositionIdx        int          `json:"positionIdx"`
	Price              string       `json:"price"`
	Qty                string       `json:"qty"`
	ReduceOnly         bool         `json:"reduceOnly"`
	RejectReason       string       `json:"rejectReason"`
	Side               string       `json:"side"`
	SlLimitPrice       string       `json:"slLimitPrice"`
	SlTriggerBy        string       `json:"slTriggerBy"`
	SmpGroup           int          `json:"smpGroup"`
	SmpOrderId         string       `json:"smpOrderId"`
	SmpType            string       `json:"smpType"`
	StopLoss           string       `json:"stopLoss"`
	StopOrderType      string       `json:"stopOrderType"`
	Symbol             string       `json:"symbol"`
	TakeProfit         string       `json:"takeProfit"`
	TimeInForce        string       `json:"timeInForce"`
	TpLimitPrice       string       `json:"tpLimitPrice"`
	TpTriggerBy        string       `json:"tpTriggerBy"`
	TpslMode           string       `json:"tpslMode"`
	TriggerBy          string       `json:"triggerBy"`
	TriggerDirection   int          `json:"triggerDirection"`
	TriggerPrice       string       `json:"triggerPrice"`
	UpdatedTime        string       `json:"updatedTime"`
}

// https://bybit-exchange.github.io/docs/v5/websocket/private/execution
type ExecutionWebsocketResponse struct {
	Id           string          `json:"id"`
	Topic        string          `json:"topic"`
	CreationTime int64           `json:"creationTime"`
	Data         []ExecutionData `json:"data"`
}

type ExecutionData struct {
	BlockTradeId    string `json:"blockTradeId"`
	Category        string `json:"category"`
	ClosedSize      string `json:"closedSize"`
	CreateType      string `json:"createType"`
	ExecFee         string `json:"execFee"`
	ExecId          string `json:"execId"`
	ExecPnl         string `json:"execPnl"`
	ExecPrice       string `json:"execPrice"`
	ExecQty         string `json:"execQty"`
	ExecTime        string `json:"execTime"`
	ExecType        string `json:"execType"`
	ExecValue       string `json:"execValue"`
	FeeRate         string `json:"feeRate"`
	IndexPrice      string `json:"indexPrice"`
	IsLeverage      string `json:"isLeverage"`
	IsMaker         bool   `json:"isMaker"`
	LeavesQty       string `json:"leavesQty"`
	MarkIv          string `json:"markIv"`
	MarkPrice       string `json:"markPrice"`
	MarketUnit      string `json:"marketUnit"`
	OrderId         string `json:"orderId"`
	OrderLinkId     string `json:"orderLinkId"`
	OrderPrice      string `json:"orderPrice"`
	OrderQty        string `json:"orderQty"`
	OrderType       string `json:"orderType"`
	Seq             int64  `json:"seq"`
	Side            string `json:"side"`
	StopOrderType   string `json:"stopOrderType"`
	Symbol          string `json:"symbol"`
	TradeIv         string `json:"tradeIv"`
	UnderlyingPrice string `json:"underlyingPrice"`
}

type KlineItem [7]string

func (k KlineItem) StartTime() string  { return k[0] }
func (k KlineItem) OpenPrice() string  { return k[1] }
func (k KlineItem) HighPrice() string  { return k[2] }
func (k KlineItem) LowPrice() string   { return k[3] }
func (k KlineItem) ClosePrice() string { return k[4] }
func (k KlineItem) Volume() string     { return k[5] }
func (k KlineItem) Turnover() string   { return k[6] }

type GetKlineResponse struct {
	BaseResponse
	Result KlineResult `json:"result"`
}

type KlineResult struct {
	Category string      `json:"category"`
	Symbol   string      `json:"symbol"`
	List     []KlineItem `json:"list"`
}

type KlineWsItem struct {
	Start     int64  `json:"start"`
	End       int64  `json:"end"`
	Interval  string `json:"interval"`
	Open      string `json:"open"`
	Close     string `json:"close"`
	High      string `json:"high"`
	Low       string `json:"low"`
	Volume    string `json:"volume"`
	Turnover  string `json:"turnover"`
	Confirm   bool   `json:"confirm"`
	Timestamp int64  `json:"timestamp"`
}

type KlineWsResponse struct {
	Ts   int64         `json:"ts"`
	Type string        `json:"type"`
	Data []KlineWsItem `json:"data"`
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
	UpdateId      int64           `json:"u"`
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
	PriceScale     string         `json:"priceScale"`
	ContractType   string         `json:"contractType"`
}

type GetInstrumentsInfoResponse struct {
	BaseResponse
	Result struct {
		Category       string           `json:"category"`
		List           []InstrumentInfo `json:"list"`
		NextPageCursor string           `json:"nextPageCursor,omitempty"`
	} `json:"result"`
}

type CoinBalance struct {
	Coin                string `json:"coin"`
	Equity              string `json:"equity"`
	UsdValue            string `json:"usdValue"`
	WalletBalance       string `json:"walletBalance"`
	Free                string `json:"free"`
	Locked              string `json:"locked"`
	SpotHedgingQty      string `json:"spotHedgingQty"`
	BorrowAmount        string `json:"borrowAmount"`
	AvailableToWithdraw string `json:"availableToWithdraw"`
	AccruedInterest     string `json:"accruedInterest"`
	TotalOrderIM        string `json:"totalOrderIM"`
	TotalPositionIM     string `json:"totalPositionIM"`
	TotalPositionMM     string `json:"totalPositionMM"`
	UnrealisedPnl       string `json:"unrealisedPnl"`
	CumRealisedPnl      string `json:"cumRealisedPnl"`
	Bonus               string `json:"bonus"`
	MarginCollateral    bool   `json:"marginCollateral"`
	CollateralSwitch    bool   `json:"collateralSwitch"`
	AvailableToBorrow   string `json:"availableToBorrow"`
}

type GetWalletBalanceParams struct {
	AccountType string `schema:"accountType"`
	Coin        string `schema:"coin,omitempty"`
}

type GetRecentPublicTradesParams struct {
	Category   string `schema:"category"`
	Symbol     string `schema:"symbol,omitempty"`
	BaseCoin   string `schema:"baseCoin,omitempty"`
	OptionType string `schema:"optionType,omitempty"`
	Limit      *int   `schema:"limit,omitempty"`
}

type GetRecentPublicTradesResponse struct {
	BaseResponse
	Result RecentPublicTradesResult `json:"result"`
}

type RecentPublicTradesResult struct {
	Category string              `json:"category"`
	List     []RecentPublicTrade `json:"list"`
}

type RecentPublicTrade struct {
	ExecId       string `json:"execId"`
	Symbol       string `json:"symbol"`
	Price        string `json:"price"`
	Size         string `json:"size"`
	Side         string `json:"side"`
	Time         string `json:"time"`
	IsBlockTrade bool   `json:"isBlockTrade"`
	IsRPITrade   bool   `json:"isRPITrade"`
	MarkPrice    string `json:"mP"`
	IndexPrice   string `json:"iP"`
	MarkIv       string `json:"mIv"`
	Iv           string `json:"iv"`
	Seq          string `json:"seq"`
}

type AccountBalance struct {
	AccountType            string        `json:"accountType"`
	AccountLTV             string        `json:"accountLTV"` // Deprecated
	AccountIMRate          string        `json:"accountIMRate"`
	AccountMMRate          string        `json:"accountMMRate"`
	TotalEquity            string        `json:"totalEquity"`
	TotalWalletBalance     string        `json:"totalWalletBalance"`
	TotalMarginBalance     string        `json:"totalMarginBalance"`
	TotalAvailableBalance  string        `json:"totalAvailableBalance"`
	TotalPerpUPL           string        `json:"totalPerpUPL"`
	TotalInitialMargin     string        `json:"totalInitialMargin"`
	TotalMaintenanceMargin string        `json:"totalMaintenanceMargin"`
	Coin                   []CoinBalance `json:"coin"`
}

// GetWalletBalanceResponse represents the response
type GetWalletBalanceResponse struct {
	BaseResponse
	Result struct {
		List []AccountBalance `json:"list"`
	} `json:"result"`
}

// https://bybit-exchange.github.io/docs/v5/websocket/private/wallet
type WalletWebsocketResponse struct {
	Id           string           `json:"id"`
	Topic        string           `json:"topic"`
	CreationTime int64            `json:"creationTime"`
	Data         []AccountBalance `json:"data"`
}

// Ticker
// https://bybit-exchange.github.io/docs/v5/websocket/public/ticker#response-parameters
type CommonTicker struct {
	Topic string `json:"topic"`
	Type  string `json:"type"`
	Ts    int64  `json:"ts"`
}

type LinearInverseTicker struct {
	CommonTicker
	Cs   int64                   `json:"cs"`
	Data LinearInverseTickerData `json:"data"`
}

type GetTickerParams struct {
	Category string `schema:"category"`
	Symbol   string `schema:"symbol,omitempty"`
	BaseCoin string `schema:"baseCoin,omitempty"`
	ExpDate  string `schema:"expDate,omitempty"`
}

type GetTickerResponseList struct {
	Category string                    `json:"category"`
	List     []LinearInverseTickerData `json:"list"`
}

type GetTickerResponse struct {
	BaseResponse
	Result GetTickerResponseList `json:"result"`
}

type LinearInverseTickerData struct {
	Symbol                 string `json:"symbol"`
	TickDirection          string `json:"tickDirection"`
	Price24hPcnt           string `json:"price24hPcnt"`
	LastPrice              string `json:"lastPrice"`
	PrevPrice24h           string `json:"prevPrice24h"`
	HighPrice24h           string `json:"highPrice24h"`
	LowPrice24h            string `json:"lowPrice24h"`
	PrevPrice1h            string `json:"prevPrice1h"`
	MarkPrice              string `json:"markPrice"`
	IndexPrice             string `json:"indexPrice"`
	OpenInterest           string `json:"openInterest"`
	OpenInterestValue      string `json:"openInterestValue"`
	Turnover24h            string `json:"turnover24h"`
	Volume24h              string `json:"volume24h"`
	NextFundingTime        string `json:"nextFundingTime"`
	FundingRate            string `json:"fundingRate"`
	Bid1Price              string `json:"bid1Price"`
	Bid1Size               string `json:"bid1Size"`
	Ask1Price              string `json:"ask1Price"`
	Ask1Size               string `json:"ask1Size"`
	DeliveryTime           string `json:"deliveryTime,omitempty"`
	BasisRate              string `json:"basisRate,omitempty"`
	DeliveryFeeRate        string `json:"deliveryFeeRate,omitempty"`
	PredictedDeliveryPrice string `json:"predictedDeliveryPrice,omitempty"`
	PreOpenPrice           string `json:"preOpenPrice,omitempty"`
	PreQty                 string `json:"preQty,omitempty"`
	CurPreListingPhase     string `json:"curPreListingPhase,omitempty"`
}

type OptionTicker struct {
	CommonTicker
	Id   string           `json:"id"`
	Data OptionTickerData `json:"data"`
}

type OptionTickerData struct {
	Symbol                 string `json:"symbol"`
	BidPrice               string `json:"bidPrice"`
	BidSize                string `json:"bidSize"`
	BidIv                  string `json:"bidIv"`
	AskPrice               string `json:"askPrice"`
	AskSize                string `json:"askSize"`
	AskIv                  string `json:"askIv"`
	LastPrice              string `json:"lastPrice"`
	HighPrice24h           string `json:"highPrice24h"`
	LowPrice24h            string `json:"lowPrice24h"`
	MarkPrice              string `json:"markPrice"`
	IndexPrice             string `json:"indexPrice"`
	MarkPriceIv            string `json:"markPriceIv"`
	UnderlyingPrice        string `json:"underlyingPrice"`
	OpenInterest           string `json:"openInterest"`
	Turnover24h            string `json:"turnover24h"`
	Volume24h              string `json:"volume24h"`
	TotalVolume            string `json:"totalVolume"`
	TotalTurnover          string `json:"totalTurnover"`
	Delta                  string `json:"delta"`
	Gamma                  string `json:"gamma"`
	Vega                   string `json:"vega"`
	Theta                  string `json:"theta"`
	PredictedDeliveryPrice string `json:"predictedDeliveryPrice"`
	Change24h              string `json:"change24h"`
}

type SpotTicker struct {
	CommonTicker
	Cs   int64          `json:"cs"`
	Data SpotTickerData `json:"data"`
}

type SpotTickerData struct {
	Symbol        string `json:"symbol"`
	LastPrice     string `json:"lastPrice"`
	HighPrice24h  string `json:"highPrice24h"`
	LowPrice24h   string `json:"lowPrice24h"`
	PrevPrice24h  string `json:"prevPrice24h"`
	Volume24h     string `json:"volume24h"`
	Turnover24h   string `json:"turnover24h"`
	Price24hPcnt  string `json:"price24hPcnt"`
	UsdIndexPrice string `json:"usdIndexPrice"`
}

type AuthResponse struct {
}

type GetPositionParams struct {
	Category   string `schema:"category"`             // Required. Product type: linear, inverse, option
	Symbol     string `schema:"symbol,omitempty"`     // Symbol name, like BTCUSDT, uppercase only
	BaseCoin   string `schema:"baseCoin,omitempty"`   // Base coin, uppercase only. option only
	SettleCoin string `schema:"settleCoin,omitempty"` // Settle coin. For linear: either symbol or settleCoin is required
	Limit      *int   `schema:"limit,omitempty"`      // Limit for data size per page. [1, 200]. Default: 20
	Cursor     string `schema:"cursor,omitempty"`     // Cursor for pagination
}

type GetPositionResponse struct {
	BaseResponse
	Result struct {
		Category       string         `json:"category"`
		NextPageCursor string         `json:"nextPageCursor"`
		List           []PositionItem `json:"list"`
	} `json:"result"`
}

type PositionItem struct {
	PositionIdx            int    `json:"positionIdx"`            // Position idx, used to identify positions in different position modes
	RiskId                 int    `json:"riskId"`                 // Risk tier ID
	RiskLimitValue         string `json:"riskLimitValue"`         // Risk limit value
	Symbol                 string `json:"symbol"`                 // Symbol name
	Side                   string `json:"side"`                   // Position side. Buy: long, Sell: short
	Size                   string `json:"size"`                   // Position size, always positive
	AvgPrice               string `json:"avgPrice"`               // Average entry price
	PositionValue          string `json:"positionValue"`          // Position value
	TradeMode              int    `json:"tradeMode"`              // Trade mode
	AutoAddMargin          int    `json:"autoAddMargin"`          // Whether to add margin automatically when using isolated margin mode
	PositionStatus         string `json:"positionStatus"`         // Position status. Normal, Liq, Adl
	Leverage               string `json:"leverage"`               // Position leverage
	MarkPrice              string `json:"markPrice"`              // Mark price
	LiqPrice               string `json:"liqPrice"`               // Position liquidation price
	BustPrice              string `json:"bustPrice"`              // Bankruptcy price
	PositionIM             string `json:"positionIM"`             // Initial margin
	PositionMM             string `json:"positionMM"`             // Maintenance margin
	PositionBalance        string `json:"positionBalance"`        // Position margin
	TakeProfit             string `json:"takeProfit"`             // Take profit price
	StopLoss               string `json:"stopLoss"`               // Stop loss price
	TrailingStop           string `json:"trailingStop"`           // Trailing stop (The distance from market price)
	SessionAvgPrice        string `json:"sessionAvgPrice"`        // USDC contract session avg price
	Delta                  string `json:"delta,omitempty"`        // Delta (options only)
	Gamma                  string `json:"gamma,omitempty"`        // Gamma (options only)
	Vega                   string `json:"vega,omitempty"`         // Vega (options only)
	Theta                  string `json:"theta,omitempty"`        // Theta (options only)
	UnrealisedPnl          string `json:"unrealisedPnl"`          // Unrealised PnL
	CurRealisedPnl         string `json:"curRealisedPnl"`         // The realised PnL for the current holding position
	CumRealisedPnl         string `json:"cumRealisedPnl"`         // Cumulative realised pnl
	AdlRankIndicator       int    `json:"adlRankIndicator"`       // Auto-deleverage rank indicator
	CreatedTime            string `json:"createdTime"`            // Timestamp of the first time position was created (ms)
	UpdatedTime            string `json:"updatedTime"`            // Position updated timestamp (ms)
	Seq                    int64  `json:"seq"`                    // Cross sequence
	IsReduceOnly           bool   `json:"isReduceOnly"`           // Useful when Bybit lower the risk limit
	MMRSysUpdatedTime      string `json:"mmrSysUpdatedTime"`      // Timestamp of MMR system update
	LeverageSysUpdatedTime string `json:"leverageSysUpdatedTime"` // Timestamp of leverage system update
	TpslMode               string `json:"tpslMode"`               // TP/SL mode, deprecated, always "Full"
}

type OrderRealtimeRequest struct {
	Category string `json:"category"`
	OrderId  string `json:"orderId"`
}

type SetLeverageParams struct {
	Category     string `json:"category"`     // Contract type, only support "linear" and "inverse"
	Symbol       string `json:"symbol"`       // Trading pair
	BuyLeverage  string `json:"buyLeverage"`  // Buy leverage
	SellLeverage string `json:"sellLeverage"` // Sell leverage
}

type SetLeverageResponse struct {
	BaseResponse
}

// Batch Place Order
// https://bybit-exchange.github.io/docs/v5/order/batch-place
type BatchOrderItem struct {
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

type BatchPlaceOrderParams struct {
	Category string           `json:"category"`
	Request  []BatchOrderItem `json:"request"`
}

type BatchPlaceOrderResponse struct {
	BaseResponse
	Result struct {
		List []struct {
			Category    string `json:"category"`
			Symbol      string `json:"symbol"`
			OrderId     string `json:"orderId"`
			OrderLinkId string `json:"orderLinkId"`
			CreateAt    string `json:"createAt"`
		} `json:"list"`
	} `json:"result"`
	RetExtInfo struct {
		List []struct {
			Code int    `json:"code"`
			Msg  string `json:"msg"`
		} `json:"list"`
	} `json:"retExtInfo"`
}

// Batch operations
// Request Parameters
// Parameter	Required	Type	Comments
// category	true	string	Product type linear, option, spot, inverse
// request	true	array	Object
// > symbol	true	string	Symbol name, like BTCUSDT, uppercase only
// > isLeverage	false	integer	Whether to borrow, spot** only. 0(default): false then spot trading, 1: true then margin trading
// > side	true	string	Buy, Sell
// > orderType	true	string	Market, Limit
// > qty	true	string	Order quantity
//
//     Spot: set marketUnit for market order qty unit, quoteCoin for market buy by default, baseCoin for market sell by default
//     Perps, Futures & Option: always use base coin as unit.
//     Perps & Futures: If you pass qty="0" and specify reduceOnly=true&closeOnTrigger=true, you can close the position up to maxMktOrderQty or maxOrderQty shown on Get Instruments Info of current symbol
//
// > marketUnit	false	string	The unit for qty when create Spot market orders, orderFilter="tpslOrder" and "StopOrder" are supported as well.
// baseCoin: for example, buy BTCUSDT, then "qty" unit is BTC
// quoteCoin: for example, sell BTCUSDT, then "qty" unit is USDT
// > price	false	string	Order price
//
//     Market order will ignore this field
//     Please check the min price and price precision from instrument info endpoint
//     If you have position, price needs to be better than liquidation price
//
// > triggerDirection	false	integer	Conditional order param. Used to identify the expected direction of the conditional order.
//
//     1: triggered when market price rises to triggerPrice
//     2: triggered when market price falls to triggerPrice
//
// Valid for linear
// > orderFilter	false	string	If it is not passed, Order by default.
//
//     Order
//     tpslOrder: Spot TP/SL order, the assets are occupied even before the order is triggered
//     StopOrder: Spot conditional order, the assets will not be occupied until the price of the underlying asset reaches the trigger price, and the required assets will be occupied after the Conditional order is triggered
//
// Valid for spot only
// > triggerPrice	false	string
//
//     For Perps & Futures, it is the conditional order trigger price. If you expect the price to rise to trigger your conditional order, make sure:
//     triggerPrice > market price
//     Else, triggerPrice < market price
//     For spot, it is the orderFilter="tpslOrder", or "StopOrder" trigger price
//
// > triggerBy	false	string	Conditional order param (Perps & Futures). Trigger price type. LastPrice, IndexPrice, MarkPrice
// > orderIv	false	string	Implied volatility. option only. Pass the real value, e.g for 10%, 0.1 should be passed. orderIv has a higher priority when price is passed as well
// > timeInForce	false	string	Time in force
//
//     Market order will use IOC directly
//     If not passed, GTC is used by default
//
// > positionIdx	false	integer	Used to identify positions in different position modes. Under hedge-mode, this param is required
//
//     0: one-way mode
//     1: hedge-mode Buy side
//     2: hedge-mode Sell side
//
// > orderLinkId	false	string	User customised order ID. A max of 36 characters. Combinations of numbers, letters (upper and lower cases), dashes, and underscores are supported.
// Futures, Perps & Spot: orderLinkId rules:
//
//     optional param
//     always unique
//     option orderLinkId rules:
//     required param
//     always unique
//
// > takeProfit	false	string	Take profit price
// > stopLoss	false	string	Stop loss price
// > tpTriggerBy	false	string	The price type to trigger take profit. MarkPrice, IndexPrice, default: LastPrice.
// Valid for linear, inverse
// > slTriggerBy	false	string	The price type to trigger stop loss. MarkPrice, IndexPrice, default: LastPrice
// Valid for linear, inverse
// > reduceOnly	false	boolean	What is a reduce-only order? true means your position can only reduce in size if this order is triggered.
//
//     You must specify it as true when you are about to close/reduce the position
//     When reduceOnly is true, take profit/stop loss cannot be set
//
// Valid for linear, inverse & option
// > closeOnTrigger	false	boolean	What is a close on trigger order? For a closing order. It can only reduce your position, not increase it. If the account has insufficient available balance when the closing order is triggered, then other active orders of similar contracts will be cancelled or reduced. It can be used to ensure your stop loss reduces your position regardless of current available margin.
// Valid for linear, inverse
// > smpType	false	string	Smp execution type. What is SMP?
// > mmp	false	boolean	Market maker protection. option only. true means set the order as a market maker protection order. What is mmp?
// > tpslMode	false	string	TP/SL mode
//
//     Full: entire position for TP/SL. Then, tpOrderType or slOrderType must be Market
//     Partial: partial position tp/sl (as there is no size option, so it will create tp/sl orders with the qty you actually fill). Limit TP/SL order are supported. Note: When create limit tp/sl, tpslMode is required and it must be Partial
//
// Valid for linear, inverse
// > tpLimitPrice	false	string	The limit order price when take profit price is triggered
//
//     linear&inverse: only works when tpslMode=Partial and tpOrderType=Limit
//     Spot: it is required when the order has takeProfit and tpOrderType=Limit
//
// > slLimitPrice	false	string	The limit order price when stop loss price is triggered
//
//     linear&inverse: only works when tpslMode=Partial and slOrderType=Limit
//     Spot: it is required when the order has stopLoss and slOrderType=Limit
//
// > tpOrderType	false	string	The order type when take profit is triggered
//
//     linear&inverse: Market(default), Limit. For tpslMode=Full, it only supports tpOrderType=Market
//     Spot:
//     Market: when you set "takeProfit",
//     Limit: when you set "takeProfit" and "tpLimitPrice"
//
// > slOrderType	false	string	The order type when stop loss is triggered
//
//     linear&inverse: Market(default), Limit. For tpslMode=Full, it only supports slOrderType=Market
//     Spot:
//     Market: when you set "stopLoss",
//     Limit: when you set "stopLoss" and "slLimitPrice"
//
// Response Parameters
// Parameter	Type	Comments
// result	Object
// > list	array	Object
// >> category	string	Product type
// >> symbol	string	Symbol name
// >> orderId	string	Order ID
// >> orderLinkId	string	User customised order ID
// >> createAt	string	Order created time (ms)
// retExtInfo	Object
// > list	array	Object
// >> code	number	Success/error code
// >> msg	string	Success/error message
// info
//
// The acknowledgement of an place order request indicates that the request was sucessfully accepted. This request is asynchronous so please use the websocket to confirm the order status.
// import (
//     "context"
//     "fmt"
//     bybit "https://github.com/bybit-exchange/bybit.go.api")
// client := bybit.NewBybitHttpClient("YOUR_API_KEY", "YOUR_API_SECRET", bybit.WithBaseURL(bybit.TESTNET))
// params := map[string]interface{}{"category": "option",
//     "request": []map[string]interface{}{
//         {
//             "category":    "option",
//             "symbol":      "BTC-10FEB23-24000-C",
//             "orderType":   "Limit",
//             "side":        "Buy",
//             "qty":         "0.1",
//             "price":       "5",
//             "orderIv":     "0.1",
//             "timeInForce": "GTC",
//             "orderLinkId": "9b381bb1-401",
//             "mmp":         false,
//             "reduceOnly":  false,
//         },
//         {
//             "category":    "option",
//             "symbol":      "BTC-10FEB23-24000-C",
//             "orderType":   "Limit",
//             "side":        "Buy",
//             "qty":         "0.1",
//             "price":       "5",
//             "orderIv":     "0.1",
//             "timeInForce": "GTC",
//             "orderLinkId": "82ee86dd-001",
//             "mmp":         false,
//             "reduceOnly":  false,
//         },
//     },
// }
// client.NewUtaBybitServiceWithParams(params).PlaceBatchOrder(context.Background())
// {
//     "retCode": 0,
//     "retMsg": "OK",
//     "result": {
//         "list": [
//             {
//                 "category": "spot",
//                 "symbol": "BTCUSDT",
//                 "orderId": "1666800494330512128",
//                 "orderLinkId": "spot-btc-03",
//                 "createAt": "1713434102752"
//             },
//             {
//                 "category": "spot",
//                 "symbol": "ATOMUSDT",
//                 "orderId": "1666800494330512129",
//                 "orderLinkId": "spot-atom-03",
//                 "createAt": "1713434102752"
//             }
//         ]
//     },
//     "retExtInfo": {
//         "list": [
//             {
//                 "code": 0,
//                 "msg": "OK"
//             },
//             {
//                 "code": 0,
//                 "msg": "OK"
//             }
//         ]
//     },
//     "time": 1713434102753
// }
// Batch Create/Amend/Cancel Order
// info
//
//     A maximum of 20 orders (option), 20 orders (inverse), 20 orders (linear), 10 orders (spot) can be placed per request. The returned data list is divided into two lists. The first list indicates whether or not the order creation was successful and the second list details the created order information. The structure of the two lists are completely consistent.
//
//     Option rate limt instruction: its rate limit is count based on the actual number of request sent, e.g., by default, option trading rate limit is 10 reqs per sec, so you can send up to 20 * 10 = 200 orders in one second.
//     Perpetual, Futures, Spot rate limit instruction, please check here
//
//     The account rate limit is shared between websocket and http batch orders
//     The acknowledgement of batch create/amend/cancel order requests indicates that the request was sucessfully accepted. The request is asynchronous so please use the websocket to confirm the order status.
//
// Request Parameters
// Parameter	Required	Type	Comments
// reqId	false	string	Used to identify the uniqueness of the request, the response will return it when passed. The length cannot exceed 36 characters.
// If passed, it can't be duplicated, otherwise you will get "20006"
// header	true	object	Request headers
// > X-BAPI-TIMESTAMP	true	string	Current timestamp
// > X-BAPI-RECV-WINDOW	false	string	5000(ms) by default. Request will be rejected when not satisfy this rule: Bybit_server_time - X-BAPI-RECV-WINDOW <= X-BAPI-TIMESTAMP < Bybit_server_time + 1000
// > Referer	false	string	The referer identifier for API broker user
// op	true	string	Op type
// order.create-batch: batch create orders
// order.amend-batch: batch amend orders
// order.cancel-batch: batch cancel orders
// args	true	array<object>	Args array
// order.create-batch: refer to Batch Place Order request
// order.amend-batch: refer to Batch Amend Order request
// order.cancel-batch: refer to Batch Cancel Order request
// Response Parameters
// Parameter	Type	Comments
// reqId	string
// If it is passed on the request, then it is returned in the response
// If it is not passed, then it is not returned in the response
// retCode	integer
// 0: success
// 10403: exceed IP rate limit. 3000 requests per second per IP
// 10404: 1. op type is not found; 2. category is not correct/supported
// 10429: System level frequency protection
// 20006: reqId is duplicated
// 10016: 1. internal server error; 2. Service is restarting
// 10019: ws trade service is restarting, do not accept new request, but the request in the process is not affected. You can build new connection to be routed to normal service
// retMsg	string
// OK
// ""
// Error message
// op	string	Op type
// data	object	Business data, keep the same as result on rest api response
// order.create-batch: refer to Batch Place Order response
// order.amend-batch: refer to Batch Amend Order response
// order.cancel-batch: refer to Batch Cancel Order response
// retExtInfo	object
// > list	array<object>
// >> code	number	Success/error code
// >> msg	string	Success/error message
// header	object	Header info
// > TraceId	string	Trace ID, used to track the trip of request
// > Timenow	string	Current timestamp
// > X-Bapi-Limit	string	The total rate limit of the current account for this op type
// > X-Bapi-Limit-Status	string	The remaining rate limit of the current account for this op type
// > X-Bapi-Limit-Reset-Timestamp	string	The timestamp indicates when your request limit resets if you have exceeded your rate limit. Otherwise, this is just the current timestamp (it may not exactly match timeNow)
// connId	string	Connection id, the unique id for the connection
// Request Example
//
//
// {
//     "op": "order.create-batch",
//     "header": {
//         "X-BAPI-TIMESTAMP": "1740453381256",
//         "X-BAPI-RECV-WINDOW": "1000"
//     },
//     "args": [
//         {
//             "category": "linear",
//             "request": [
//                 {
//                     "symbol": "SOLUSDT",
//                     "qty": "10",
//                     "price": "500",
//                     "orderType": "Limit",
//                     "timeInForce": "GTC",
//                     "orderLinkId": "-batch-000",
//                     "side": "Buy"
//                 },
//                 {
//                     "symbol": "SOLUSDT",
//                     "qty": "20",
//                     "price": "1000",
//                     "orderType": "Limit",
//                     "timeInForce": "GTC",
//                     "orderLinkId": "batch-001",
//                     "side": "Buy"
//                 },
//                 {
//                     "symbol": "SOLUSDT",
//                     "qty": "30",
//                     "price": "1500",
//                     "orderType": "Limit",
//                     "timeInForce": "GTC",
//                     "orderLinkId": "batch-002",
//                     "side": "Buy"
//                 }
//             ]
//         }
//     ]
// }
//
// Response Example
//
// {
//     "retCode": 0,
//     "retMsg": "OK",
//     "op": "order.create-batch",
//     "data": {
//         "list": [
//             {
//                 "category": "linear",
//                 "symbol": "SOLUSDT",
//                 "orderId": "",
//                 "orderLinkId": "batch-000",
//                 "createAt": ""
//             },
//             {
//                 "category": "linear",
//                 "symbol": "SOLUSDT",
//                 "orderId": "",
//                 "orderLinkId": "batch-001",
//                 "createAt": ""
//             },
//             {
//                 "category": "linear",
//                 "symbol": "SOLUSDT",
//                 "orderId": "",
//                 "orderLinkId": "batch-002",
//                 "createAt": ""
//             }
//         ]
//     },
//     "retExtInfo": {
//         "list": [
//             {
//                 "code": 10001,
//                 "msg": "position idx not match position mode"
//             },
//             {
//                 "code": 10001,
//                 "msg": "position idx not match position mode"
//             },
//             {
//                 "code": 10001,
//                 "msg": "position idx not match position mode"
//             }
//         ]
//     },
//     "header": {
//         "Timenow": "1740453408556",
//         "X-Bapi-Limit": "150",
//         "X-Bapi-Limit-Status": "147",
//         "X-Bapi-Limit-Reset-Timestamp": "1740453408555",
//         "Traceid": "0e32b551b3e17aae77651aadf6a5be80"
//     },
//     "connId": "cupviqn88smf24t2kpb0-536o"
// }
