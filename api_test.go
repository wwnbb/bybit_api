/*
 * Copyright (c) 2024 QuantHill <info@quanthill.ae>
 */
package qant_api_bybit

import (
	// "testing"
	"time"
)

const (
	api_test_key        = "QU0G8RSs5aSsoGVir2"
	api_test_secret_key = "IHmT3wcDaI7TBo0WlJaPlJj8JTMtdb5KQrZR"
)

type TimeMock struct {
	mockedTime time.Time
}

func (t *TimeMock) Now() time.Time {
	return t.mockedTime
}

// func TestGenSignature(t *testing.T) {
// 	// "1730158649877QU0G8RSs5aSsoGVir25000{\"category\":\"linear\",\"orderType\":\"Market\",\"qty\":\"0.001\",\"side\":\"Buy\",\"symbol\":\"BTCUSDT\"}"
// 	// "1730158649877QU0G8RSs5aSsoGVir25000{\"category\":\"linear\",\"symbol\":\"BTCUSDT\",\"side\":\"Buy\",\"orderType\":\"Market\",\"qty\":\"0.001\"}"
// 	// 55b80c34e25c28bd1640e1d1c78cc6d232d987273ddd8783a4208e9063e00cad
// 	mockedTime := time.UnixMilli(1730158649877)
// 	timeMock := &TimeMock{
// 		mockedTime: mockedTime,
// 	}
// 	params := PlaceOrderParams{
// 		Category:  "linear",
// 		OrderType: "Market",
// 		Qty:       "0.001",
// 		Side:      "Buy",
// 		Symbol:    "BTCUSDT",
// 	}
// 	signature, ts, err := genSignature(params, api_test_key, api_test_secret_key, RECV_WINDOW, timeMock)
// 	if err != nil {
// 		t.Error(err)
// 	}
//
// 	if signature != "55b80c34e25c28bd1640e1d1c78cc6d232d987273ddd8783a4208e9063e00cad" {
// 		t.Error("signature mismatch")
// 	}
// 	if ts != 1730158649877 {
// 		t.Error("timestamp mismatch")
// 	}
// }
