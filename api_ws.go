/*
 * Copyright (c) 2024 QuantHill <info@quanthill.ae>
 */
package qant_api_bybit

func getApi() *BybitApi {
	api := NewBybitApi("QU0G8RSs5aSsoGVir2", "IHmT3wcDaI7TBo0WlJaPlJj8JTMtdb5KQrZR")
	api.ConfigureWsUrls(
		TESTNET_PRIVATE_WS,
		TESTNET_SPOT_WS,
		TESTNET_LINEAR_WS,
		TESTNET_INVERSE_WS,
		TESTNET_TRADE_WS,
	)
	return api
}

