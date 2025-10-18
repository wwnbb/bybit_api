package bybit_api

import (
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	pp "github.com/wwnbb/pprint"
	"github.com/wwnbb/ptr"
)

func TestWsPlaceOrder(t *testing.T) {
	api := getApi()
	go func() {
		for {
			fmt.Println(<-api.Trade.DataCh)
		}
	}()

	now := time.Now()

	nanosInCurrentSecond := now.Nanosecond()
	targetNanos := 940 * time.Millisecond

	var waitDuration time.Duration
	if nanosInCurrentSecond < int(targetNanos) {
		// We haven't reached 900ms yet in this second
		waitDuration = targetNanos - time.Duration(nanosInCurrentSecond)
	} else {
		// We've passed 900ms, wait for 900ms of next second
		waitDuration = time.Second - time.Duration(nanosInCurrentSecond) + targetNanos
	}

	fmt.Printf("Current time: %v\n", now.Format("15:04:05.000"))
	fmt.Printf("Waiting %v milliseconds...\n", waitDuration.Milliseconds())

	// Wait until 100ms before next second
	time.Sleep(waitDuration)

	// Execute order (should now be at ~900ms mark, order hits exchange at 1-10ms into next second)
	executeTime := time.Now()
	fmt.Printf("Executing order at: %v\n", executeTime.Format("15:04:05.000"))

	params := PlaceOrderParams{
		Symbol:    "ZBTUSDT",
		Side:      "Sell",
		OrderType: "Market",
		Qty:       "15",
		Category:  "linear",
	}
	api.Trade.PlaceOrder(params)

	time.Sleep(5 * time.Second)
}

func TestGetOrderRealTime(t *testing.T) {
	api := getApi()
	val, err := api.REST.GetOrderRealTime(
		OrderRealtimeRequest{
			Category: "linear",
			OrderId:  "44395895-93de-4fac-852d-83e241852c56",
		},
	)
	if err != nil {
		t.Fatalf("GetOrderRealTime failed: %v", err)
	}
	pp.PrettyPrint(val)
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

	_, err := api.Trade.PlaceOrder(params)
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

// func TestWsPlaceOrder(t *testing.T) {
// 	api := getApi()
// 	go func() {
// 		for {
// 			fmt.Println(<-api.Trade.DataCh)
// 		}
// 	}()
//
// 	targetSecond := time.Now().Unix() + 5
// 	latency := 50 * time.Millisecond
//
// 	targetTime := time.Unix(targetSecond, 0)
// 	now := time.Now()
//
// 	fmt.Printf("Current time: %v (Unix: %d)\n", now.Format("15:04:05.000000"), now.Unix())
// 	fmt.Printf("Target second: %v (Unix: %d)\n", targetTime.Format("15:04:05.000000"), targetSecond)
//
// 	if now.Unix() >= targetSecond {
// 		t.Fatalf("Target second %d is in the past (current: %d)", targetSecond, now.Unix())
// 	}
//
// 	// Target: 10-20ms into the target second
// 	// We'll aim for 15ms as the midpoint, accounting for execution overhead
//
// 	exactTargetTime := time.Unix(targetSecond, 0)
//
// 	// Calculate wait duration
// 	waitDuration := exactTargetTime.Sub(now)
// 	waitDuration -= latency
//
// 	if waitDuration < 0 {
// 		t.Fatalf("Not enough time to reach target offset in second %d", targetSecond)
// 	}
//
// 	fmt.Printf("Waiting %v (%d milliseconds)...\n", waitDuration, waitDuration.Milliseconds())
//
// 	// Wait until target time
// 	time.Sleep(waitDuration)
//
// 	// Execute order (should now be at ~10-20ms into the target second)
// 	executeTime := time.Now()
// 	millisIntoSecond := (executeTime.Nanosecond() / 1_000_000)
// 	fmt.Printf("Executing order at: %v (offset: %dms into second, Unix: %d)\n",
// 		executeTime.Format("15:04:05.000000"), millisIntoSecond, executeTime.Unix())
//
// 	params := PlaceOrderParams{
// 		Symbol:      "COAIUSDT",
// 		Side:        "Buy",
// 		OrderType:   "limit",
// 		Qty:         "1",
// 		Price:       ptr.Ptr("14.25"),
// 		Category:    "linear",
// 		TimeInForce: ptr.Ptr("FOK"),
// 	}
// 	api.Trade.PlaceOrder(params)
//
// 	time.Sleep(5 * time.Second)
// }

func waitUntilTarget(target int64) {
	now := time.Now()
	fmt.Println("Current time:", time.Now().Format("15:04:05.000000"))
	targetTime := time.Unix(target, 0)
	latency := 950 * time.Millisecond
	waitTime := targetTime.Sub(now)
	waitTime -= latency
	time.Sleep(waitTime)
	fmt.Println("Woke up at:", time.Now().Format("15:04:05.000000"))
}

func TestWaitUntilTarget(t *testing.T) {
	timeUntil := int64(1760745600)

	api := getApi()

	params := PlaceOrderParams{
		Symbol:    "",
		Side:      "",
		OrderType: "",
		Qty:       "",
		Category:  "",
	}
	api.Trade.PlaceOrder(params)
	go func() {
		for {
			fmt.Println(<-api.Trade.DataCh)
		}
	}()
	params = PlaceOrderParams{
		Symbol:    "KUSDT",
		Side:      "Buy",
		OrderType: "Market",
		Qty:       "500",
		Category:  "linear",
	}
	waitUntilTarget(timeUntil)
	api.Trade.PlaceOrder(params)

	time.Sleep(5 * time.Second)
}

type ExecutionTimes struct {
	orderSend int64
	orderRecv int64
}

func TestCalculateLatency(t *testing.T) {
	OrderTimeMap := make(map[string]ExecutionTimes)
	var OrdMutex sync.Mutex // Use sync.Mutex instead of channel

	api := getApi()
	params := PlaceOrderParams{
		Symbol:    "",
		Side:      "",
		OrderType: "",
		Qty:       "",
		Category:  "",
	}
	api.Trade.PlaceOrder(params)
	time.Sleep(1)

	// Start goroutine before placing orders
	done := make(chan struct{}) // Add a way to stop the goroutine
	go func() {
		for {
			select {
			case data := <-api.Trade.DataCh:
				if data.Op == "order.create" {
					val := data.Data.(OrderWebsocketCreateResponse)
					recvTimeI := val.Header["Timenow"]
					recvTime, _ := strconv.ParseInt(recvTimeI, 10, 64)

					OrdMutex.Lock()
					if v, exists := OrderTimeMap[val.ReqId]; exists {
						v.orderRecv = recvTime
						OrderTimeMap[val.ReqId] = v // Update the map
					}
					OrdMutex.Unlock()
				}
			case <-done:
				return
			}
		}
	}()

	for i := range 10 {
		now := time.Now().UnixMilli()
		params = PlaceOrderParams{
			// Fill in actual parameters here
		}
		req_id, err := api.Trade.PlaceOrder(params)
		if err != nil {
			t.Errorf("Failed to place order %d: %v", i, err)
			continue
		}

		OrdMutex.Lock()
		OrderTimeMap[req_id] = ExecutionTimes{
			orderSend: now,
			orderRecv: 0,
		}
		OrdMutex.Unlock()

		time.Sleep(10 * time.Millisecond)
	}

	params = PlaceOrderParams{
		Category:  "linear",
		Symbol:    "ZBTUSDT",
		Side:      "Sell",
		OrderType: "Market",
		Qty:       "15",
	}
	req_id, _ := api.Trade.PlaceOrder(params)

	OrdMutex.Lock()
	OrderTimeMap[req_id] = ExecutionTimes{
		orderSend: time.Now().UnixMilli(),
		orderRecv: 0,
	}
	OrdMutex.Unlock()

	time.Sleep(5 * time.Second)
	close(done)

	// Print results
	OrdMutex.Lock()
	for k, v := range OrderTimeMap {
		latency := v.orderRecv - v.orderSend
		fmt.Printf("Order: %s, time send: %d, time recv: %d, latency: %d ms\n",
			k, v.orderSend, v.orderRecv, latency)

		if v.orderRecv == 0 {
			t.Errorf("Order %s never received a response", k)
		}
	}
	OrdMutex.Unlock()
}

func TestStructCreation(t *testing.T) {
	dataSTruct := ExecutionTimes{}
	fmt.Println(dataSTruct)
	dataSTruct.orderSend = 12345
	dataSTruct.orderRecv = 67890
}
