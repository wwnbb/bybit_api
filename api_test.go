package bybit_api

import (
	"context"
	"fmt"
	"os"
	"testing"
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

func TestNewApi(t *testing.T) {
	apiKey := os.Getenv("BYBIT_API_KEY")
	secretKey := os.Getenv("BYBIT_SECRET_KEY")
	ctx := context.Background()
	api := NewBybitApi(apiKey, secretKey, ctx)
	api.ConfigureMainNetUrls()
	if api.REST.api.BASE_REST_URL != BASE_URL {
		t.Fatal("WRONG BASE URL")
	}
	if api.Spot.url != WS_URL_SPOT {
		t.Fatal("WRONG SPOT URL")
	}
	if api.Spot.url != WS_URL_SPOT {
		t.Fatalf("URls don't match %s != %s", api.Inverse.url, WS_URL_INVERSE)
	}
	if api.Linear.url != WS_URL_LINEAR {
		t.Fatalf("URls don't match %s != %s", api.Linear.url, WS_URL_LINEAR)
	}
}

func TestApiDisconnect(t *testing.T) {
	api := getApi()
	err := api.Spot.Subscribe("orderbook.1.BTCUSDT")
	if err != nil {
		t.Fatalf("Colud not connect to ws %v", err)
	}
	go api.Disconnect()
	for i := 0; i < 1000; i++ {
		data, ok := <-api.Spot.DataCh
		if !ok {
			return
		}
		fmt.Println(data)
	}
}
