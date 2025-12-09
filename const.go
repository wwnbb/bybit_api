package bybit_api

type WebSocketT int

const (
	WS_SPOT WebSocketT = iota
	WS_LINEAR
	WS_INVERSE
	WS_OPTION
	WS_PRIVATE
	WS_TRADE
)

var (
	AUTH_REQUIRED_TYPES = map[WebSocketT]bool{
		WS_PRIVATE: true,
		WS_TRADE:   true,
	}
)

const API_VERSION = "v5"

const (
	WSS_URL_BASE = "wss://stream.bybit.com"

	WS_URL_TRADE   = WSS_URL_BASE + "/realtime"
	WS_URL_PRIVATE = WSS_URL_BASE + "/" + API_VERSION + "/" + "private"
	WS_URL_SPOT    = WSS_URL_BASE + "/" + API_VERSION + "/" + "public/spot"
	WS_URL_LINEAR  = WSS_URL_BASE + "/" + API_VERSION + "/" + "public/linear"
	WS_URL_INVERSE = WSS_URL_BASE + "/" + API_VERSION + "/" + "public/inverse"
	WS_URL_OPTION  = WSS_URL_BASE + "/" + API_VERSION + "/" + "public/option"
)

const (
	BASE_URL              = "https://api.bybit.com"
	BASE_DEMO_URL         = "https://api-demo.bybit.com"
	TESTNET_URL           = "https://api-testnet.bybit.com"
	BASE_DEMO_TESTNET_URL = "https://api-demo-testnet.bybit.com"
	RECV_WINDOW           = "5000"
)

const (
	MAINNET_BASE_URL      = "wss://stream.bybit.com"
	MAINNET_DEMO_BASE_URL = "wss://stream-demo.bybit.com"
	TESTNET_BASE_URL      = "wss://stream-testnet.bybit.com"

	MAINNET_SPOT_WS    = MAINNET_BASE_URL + "/" + API_VERSION + "/public/spot"
	MAINNET_LINEAR_WS  = MAINNET_BASE_URL + "/" + API_VERSION + "/public/linear"
	MAINNET_INVERSE_WS = MAINNET_BASE_URL + "/" + API_VERSION + "/public/inverse"
	MAINNET_OPTION_WS  = MAINNET_BASE_URL + "/" + API_VERSION + "/public/option"

	MAINNET_PRIVATE_WS = MAINNET_BASE_URL + "/" + API_VERSION + "/private"
	MAINNET_TRADE_WS   = MAINNET_BASE_URL + "/" + API_VERSION + "/trade"

	MAINNET_DEMO_SPOT_WS    = MAINNET_DEMO_BASE_URL + "/" + API_VERSION + "/public/spot"
	MAINNET_DEMO_LINEAR_WS  = MAINNET_DEMO_BASE_URL + "/" + API_VERSION + "/public/linear"
	MAINNET_DEMO_INVERSE_WS = MAINNET_DEMO_BASE_URL + "/" + API_VERSION + "/public/inverse"
	MAINNET_DEMO_OPTION_WS  = MAINNET_DEMO_BASE_URL + "/" + API_VERSION + "/public/option"

	MAINNET_DEMO_PRIVATE_WS = MAINNET_DEMO_BASE_URL + "/" + API_VERSION + "/private"
	MAINNET_DEMO_TRADE_WS   = MAINNET_DEMO_BASE_URL + "/" + API_VERSION + "/trade"

	TESTNET_SPOT_WS    = TESTNET_BASE_URL + "/" + API_VERSION + "/public/spot"
	TESTNET_LINEAR_WS  = TESTNET_BASE_URL + "/" + API_VERSION + "/public/linear"
	TESTNET_INVERSE_WS = TESTNET_BASE_URL + "/" + API_VERSION + "/public/inverse"
	TESTNET_OPTION_WS  = TESTNET_BASE_URL + "/" + API_VERSION + "/public/option"

	TESTNET_PRIVATE_WS = TESTNET_BASE_URL + "/" + API_VERSION + "/private"
	TESTNET_TRADE_WS   = TESTNET_BASE_URL + "/" + API_VERSION + "/trade"
)

const (
	timestampKey  = "X-BAPI-TIMESTAMP"
	signatureKey  = "X-BAPI-SIGN"
	apiRequestKey = "X-BAPI-API-KEY"
	recvWindowKey = "X-BAPI-RECV-WINDOW"
	signTypeKey   = "X-BAPI-SIGN-TYPE"
)

type OrderStatusT string

const (
	OrderStatusNew             OrderStatusT = "New"
	OrderStatusPartiallyFilled OrderStatusT = "PartiallyFilled"
	OrderStatusUntriggered     OrderStatusT = "Untriggered"

	OrderStatusRejected                OrderStatusT = "Rejected"
	OrderStatusPartiallyFilledCanceled OrderStatusT = "PartiallyFilledCanceled"
	OrderStatusFilled                  OrderStatusT = "Filled"
	OrderStatusCancelled               OrderStatusT = "Cancelled"
	OrderStatusTriggered               OrderStatusT = "Triggered"
	OrderStatusDeactivated             OrderStatusT = "Deactivated"
)

type OrderTypeT string

const (
	OrderTypeLimit   OrderTypeT = "Limit"
	OrderTypeMarket  OrderTypeT = "Market"
	OrderTypeUnknown OrderTypeT = "Unknown"
)

type TopicType string

const (
	// Order topics
	TopicOrder        TopicType = "order"
	TopicOrderSpot    TopicType = "order.spot"
	TopicOrderLinear  TopicType = "order.linear"
	TopicOrderInverse TopicType = "order.inverse"
	TopicOrderOption  TopicType = "order.option"

	// Position topics
	TopicPosition        TopicType = "position"
	TopicPositionSpot    TopicType = "position.spot"
	TopicPositionLinear  TopicType = "position.linear"
	TopicPositionInverse TopicType = "position.inverse"
	TopicPositionOption  TopicType = "position.option"

	// Wallet topics
	TopicWallet        TopicType = "wallet"
	TopicWalletSpot    TopicType = "wallet.spot"
	TopicWalletLinear  TopicType = "wallet.linear"
	TopicWalletInverse TopicType = "wallet.inverse"
)

type CategoryType string

func (c CategoryType) String() string {
	return string(c)
}

// "spot", "linear", "inverse", "option"
const (
	SpotCategory    CategoryType = "spot"
	LinearCategory  CategoryType = "linear"
	InverseCategory CategoryType = "inverse"
	OptionCategory  CategoryType = "option"
)

var Categories []CategoryType = []CategoryType{
	SpotCategory,
	LinearCategory,
	InverseCategory,
	OptionCategory,
}
