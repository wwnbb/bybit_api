package qant_api_bybit

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"strconv"
	"testing"
	"time"
)

func httpClient() *http.Client {
	client := &http.Client{Timeout: 10 * time.Second}
	return client
}

var burl string = "https://api-testnet.bybit.com"
var api_key = "QU0G8RSs5aSsoGVir2"
var apiSecret = "IHmT3wcDaI7TBo0WlJaPlJj8JTMtdb5KQrZR"
var recv_window = "5000"
var signature = ""

func getRequest(client *http.Client, method string, params string, endPoint string) []byte {
	now := time.Now()
	unixNano := now.UnixNano()
	time_stamp := unixNano / 1000000
	hmac256 := hmac.New(sha256.New, []byte(apiSecret))
	hmac256.Write([]byte(strconv.FormatInt(time_stamp, 10) + api_key + recv_window + params))
	signature = hex.EncodeToString(hmac256.Sum(nil))
	request, error := http.NewRequest("GET", burl+endPoint+"?"+params, nil)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-BAPI-API-KEY", api_key)
	request.Header.Set("X-BAPI-SIGN", signature)
	request.Header.Set("X-BAPI-TIMESTAMP", strconv.FormatInt(time_stamp, 10))
	request.Header.Set("X-BAPI-SIGN-TYPE", "2")
	request.Header.Set("X-BAPI-RECV-WINDOW", recv_window)
	reqDump, err := httputil.DumpRequestOut(request, true)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Request Dump:\n%s", string(reqDump))
	response, error := client.Do(request)
	if error != nil {
		panic(error)
	}
	defer response.Body.Close()
	elapsed := time.Since(now).Seconds()
	fmt.Printf("\n%s took %v seconds \n", endPoint, elapsed)
	fmt.Println("response Status:", response.Status)
	fmt.Println("response Headers:", response.Header)
	body, _ := io.ReadAll(response.Body)
	fmt.Println("response Body:", string(body))
	return body
}

func postRequest(client *http.Client, method string, params interface{}, endPoint string) []byte {
	now := time.Now()
	unixNano := now.UnixNano()
	time_stamp := unixNano / 1000000
	jsonData, err := json.Marshal(params)
	if err != nil {
		log.Fatal(err)
	}
	hmac256 := hmac.New(sha256.New, []byte(apiSecret))
	hmac256.Write([]byte(strconv.FormatInt(time_stamp, 10) + api_key + recv_window + string(jsonData[:])))
	signature = hex.EncodeToString(hmac256.Sum(nil))
	request, error := http.NewRequest("POST", burl+endPoint, bytes.NewBuffer(jsonData))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-BAPI-API-KEY", api_key)
	request.Header.Set("X-BAPI-SIGN", signature)
	request.Header.Set("X-BAPI-TIMESTAMP", strconv.FormatInt(time_stamp, 10))
	request.Header.Set("X-BAPI-SIGN-TYPE", "2")
	request.Header.Set("X-BAPI-RECV-WINDOW", recv_window)
	reqDump, err := httputil.DumpRequestOut(request, true)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Request Dump:\n%s", string(reqDump))
	response, error := client.Do(request)
	if error != nil {
		panic(error)
	}
	defer response.Body.Close()
	elapsed := time.Since(now).Seconds()
	fmt.Printf("\n%s took %v seconds \n", endPoint, elapsed)
	fmt.Println("response Status:", response.Status)
	fmt.Println("response Headers:", response.Header)
	body, _ := io.ReadAll(response.Body)
	fmt.Println("response Body:", string(body))
	return body
}

func TestPostRequest(t *testing.T) {
	c := httpClient()

	postParams := map[string]interface{}{
		"category":  "linear",
		"symbol":    "BTCUSDT",
		"side":      "Buy",
		"orderType": "Market",
		"qty":       "0.001",
	}
	postEndPoint := "/v5/order/create"
	postRequest(c, http.MethodPost, postParams, postEndPoint)
}

func TestGetRequest(t *testing.T) {
	c := httpClient()

	getEndPoint := "/v5/order/realtime"
	getParams := "category=linear&settleCoin=USDT"
	getRequest(c, http.MethodGet, getParams, getEndPoint)
}
