package bybit_api

import (
	"context"
	pp "github.com/wwnbb/pprint"
	"os"
	"testing"
	"time"
)

func TestGetAccountInfo(t *testing.T) {
	apiKey := os.Getenv("BYBIT_API_KEY")
	secretKey := os.Getenv("BYBIT_API_SECRET")

	if apiKey == "" || secretKey == "" {
		t.Skip("Skipping test: BYBIT_API_KEY and BYBIT_SECRET_KEY environment variables not set")
	}

	ctx := context.Background()
	api := NewBybitApi(apiKey, secretKey, ctx)
	api.ConfigureMainNetUrls()

	resp, err := api.REST.GetAccountInfo()
	if err != nil {
		t.Fatalf("GetAccountInfo failed: %v", err)
	}

	if resp.RetMsg != "OK" {
		t.Errorf("unexpected response message: got %s, want OK", resp.RetMsg)
	}

	if resp.RetCode != 0 {
		t.Errorf("unexpected return code: got %d, want 0", resp.RetCode)
	}

	if resp.Result.MarginMode == "" {
		t.Error("margin mode is empty")
	}
}

func TestAuthPrivateChannel(t *testing.T) {
	api := GetApi()
	api.Trade.Connect()
	go func() {
		for {
			pp.PrettyPrint(<-api.Trade.DataCh)
		}
	}()
	time.Sleep(20 * time.Second)
}
