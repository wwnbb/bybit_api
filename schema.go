package api_bybit

import (
	"fmt"
	"strconv"
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
