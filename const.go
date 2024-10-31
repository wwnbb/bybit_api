package qant_api_bybit

type ConnectionState int32

const (
	StatusDisconnected ConnectionState = iota // connection error occurred, can be reconnected
	StatusConnected
	StatusClosed // socket was permanently closed, exit all routines
)

type WSType int

const (
	WS_SPOT WSType = iota
	WS_LINEAR
	WS_INVERSE
	WS_PRIVATE
	WS_TRADE
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
