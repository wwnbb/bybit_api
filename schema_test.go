package bybit_api

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/gorilla/schema"
)

func TestEncodeOpenOrderRequest(t *testing.T) {
	// TestEncodeOpenOrderRequest tests the encoding of an open order request
	// with the given parameters.
	// The test should pass if the request is encoded correctly.
	oor := OpenOrderRequest{
		Category: "linear",
		Symbol:   "BTCUSDT",
	}
	u := url.Values{}
	enc := schema.NewEncoder()
	enc.Encode(oor, u)
	fmt.Println(u)
}
