module github.com/wwnbb/bybit_api

go 1.23

toolchain go1.24.7

require (
	github.com/dustinxie/lockfree v0.0.0-20210712051436-ed0ed42fd0d6
	github.com/goccy/go-json v0.10.5
	github.com/google/uuid v1.6.0
	github.com/gorilla/mux v1.8.1
	github.com/gorilla/schema v1.4.1
	github.com/gorilla/websocket v1.5.3
	github.com/wwnbb/pprint v0.0.4
	github.com/wwnbb/ptr v1.0.2
	github.com/wwnbb/wsmanager v0.0.0-20251207061014-f0c776d51419
)

require (
	github.com/coder/websocket v1.8.14 // indirect
	github.com/dchest/siphash v1.2.3 // indirect
)

replace github.com/wwnbb/wsmanager => ../wsmanager
