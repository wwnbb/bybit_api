package bybit_api

import (
	"fmt"
	"github.com/wwnbb/ptr"
	"testing"
	"time"
)

func TestWsPlaceOrder(t *testing.T) {
	api := getApi()
	go func() {
		for {
			fmt.Println(<-api.Trade.DataCh)
		}
	}()
	// {
	// 	"symbol": "ETHUSDT",
	// 	"side": "Buy",
	// 	"orderType": "Limit",
	// 	"qty": "0.2",
	// 	"price": "2800",
	// 	"category": "linear",
	// 	"timeInForce": "PostOnly"
	// 	}
	params := PlaceOrderParams{
		Symbol:      "ETHUSDT",
		Side:        "Buy",
		OrderType:   "Limit",
		Qty:         "0.2",
		Price:       ptr.Ptr("2500"),
		Category:    "linear",
		TimeInForce: ptr.Ptr("PostOnly"),
	}
	api.Trade.PlaceOrder(params)
	time.Sleep(5 * time.Second)
}

func TestWsCancelOrder(t *testing.T) {
	api := getApi()
	idChan := make(chan string, 1)
	done := make(chan bool)
	errChan := make(chan error, 1)

	// Goroutine to listen for order creation responses
	go func() {
		defer close(idChan)
		timeout := time.After(10 * time.Second)

		for {
			select {
			case data := <-api.Trade.DataCh:
				if data.Op == "order.create" {
					v, ok := data.Data.(OrderWebsocketCreateResponse)
					if !ok {
						errChan <- fmt.Errorf("failed to cast data to OrderWebsocketCreateResponse")
						return
					}

					if v.RetCode != 0 {
						errChan <- fmt.Errorf("order creation failed: %s", v.RetMsg)
						return
					}

					println("Order created:", v.Data.OrderId)
					idChan <- v.Data.OrderId
					return
				}
			case <-timeout:
				errChan <- fmt.Errorf("timeout waiting for order creation")
				return
			case <-done:
				return
			}
		}
	}()

	// Goroutine to cancel the order once created
	go func() {
		timeout := time.After(15 * time.Second)

		select {
		case id := <-idChan:
			println("Canceling order:", id)
			err := api.Trade.CancelOrder(CancelOrderParams{
				Category: "linear",
				Symbol:   "ETHUSDT",
				OrderId:  ptr.Ptr(id),
			})
			if err != nil {
				errChan <- fmt.Errorf("cancel order failed: %w", err)
				return
			}

			// Wait for cancel confirmation
			cancelTimeout := time.After(5 * time.Second)
			for {
				select {
				case data := <-api.Trade.DataCh:
					if data.Op == "order.cancel" {
						v, ok := data.Data.(OrderWebsocketCancelResponse)
						if !ok {
							errChan <- fmt.Errorf("failed to cast cancel response")
							return
						}

						if v.RetCode == 0 {
							println("Order cancelled successfully:", v.Data.OrderId)
							done <- true
							return
						} else {
							errChan <- fmt.Errorf("cancel failed: %s", v.RetMsg)
							return
						}
					}
				case <-cancelTimeout:
					errChan <- fmt.Errorf("timeout waiting for cancel confirmation")
					return
				}
			}
		case <-timeout:
			errChan <- fmt.Errorf("timeout waiting for order ID")
			return
		case <-done:
			return
		}
	}()

	// Place the order
	params := PlaceOrderParams{
		Symbol:      "ETHUSDT",
		Side:        "Buy",
		OrderType:   "Limit",
		Qty:         "0.2",
		Price:       ptr.Ptr("2500"),
		Category:    "linear",
		TimeInForce: ptr.Ptr("PostOnly"),
	}

	err := api.Trade.PlaceOrder(params)
	if err != nil {
		t.Fatalf("Failed to place order: %v", err)
	}

	// Wait for completion or error
	select {
	case err := <-errChan:
		t.Fatalf("Test failed: %v", err)
	case <-done:
		t.Log("Order created and cancelled successfully")
	case <-time.After(20 * time.Second):
		t.Fatal("Test timeout")
	}
}
