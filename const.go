package qant_api_bybit

type ConnectionState int32

const (
	StatusDisconnected ConnectionState = iota // connection error occurred, can be reconnected
	StatusConnected
	StatusClosed // socket was permanently closed, exit all routines
)

type WebSocketT int

const (
	WS_SPOT WebSocketT = iota
	WS_LINEAR
	WS_INVERSE
	WS_PRIVATE
	WS_TRADE
)

var (
	AUTH_REQUIRED_TYPES = map[WebSocketT]bool{
		WS_PRIVATE: true,
		WS_TRADE:   true,
	}
)

const (
	WS_URL_TRADE   = "wss://stream.bybit.com/realtime"
	WS_URL_PRIVATE = "wss://stream.bybit.com/v5/private"
	WS_URL_SPOT    = "wss://stream.bybit.com/v5/public/spot"
	WS_URL_LINEAR  = "wss://stream.bybit.com/v5/public/linear"
	WS_URL_INVERSE = "wss://stream.bybit.com/v5/public/inverse"
)

const (
	BASE_URL    = "https://api.bybit.com"
	TESTNET_URL = "https://api-testnet.bybit.com"
	RECV_WINDOW = "5000"
)

const (
	MAINNET_BASE_URL = "wss://stream.bybit.com"
	TESTNET_BASE_URL = "wss://stream-testnet.bybit.com"

	API_VERSION = "v5"

	MAINNET_SPOT_WS    = MAINNET_BASE_URL + "/" + API_VERSION + "/public/spot"
	MAINNET_LINEAR_WS  = MAINNET_BASE_URL + "/" + API_VERSION + "/public/linear"
	MAINNET_INVERSE_WS = MAINNET_BASE_URL + "/" + API_VERSION + "/public/inverse"
	MAINNET_OPTION_WS  = MAINNET_BASE_URL + "/" + API_VERSION + "/public/option"

	MAINNET_PRIVATE_WS = MAINNET_BASE_URL + "/" + API_VERSION + "/private"
	MAINNET_TRADE_WS   = MAINNET_BASE_URL + "/" + API_VERSION + "/trade"

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
