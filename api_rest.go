/*
 * Copyright (c) 2024 QuantHill <info@quanthill.ae>
 */
package qant_api_bybit

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	gjson "github.com/goccy/go-json"
)

// REST API
func (api *BybitApi) GetServerTime() (*GetServerTimeResponse, error) {
	const path = "/v5/market/time"
	url := api.BASE_REST_URL + path
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var r GetServerTimeResponse
	err = gjson.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (api *BybitApi) PlaceOrder(params PlaceOrderParams) (*PlaceOrderResponse, error) {
	const path = "/v5/order/create"
	url := BASE_URL + path
	paramsJson, err := gjson.Marshal(params)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(paramsJson))
	if err != nil {
		return nil, err
	}

	signature, ts, err := GenSignature(params, api.ApiKey, api.ApiSecret, RECV_WINDOW)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-BAPI-API-KEY", api.ApiKey)
	req.Header.Set("X-BAPI-SIGN", signature)
	req.Header.Set("X-BAPI-TIMESTAMP", strconv.FormatInt(ts, 10))
	req.Header.Set("X-BAPI-SIGN-TYPE", "2")
	req.Header.Set("X-BAPI-RECV-WINDOW", RECV_WINDOW)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var r PlaceOrderResponse

	err = gjson.NewDecoder(resp.Body).Decode(&r)
	return &r, nil
}

func (api *BybitApi) GetOrders(params OpenOrderRequest) (*GetOrdersResponse, error) {
	const path = "/v5/order/realtime"

	baseUrl, err := url.Parse(BASE_URL + path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base URL: %w", err)
	}

	values := url.Values{}
	baseUrl.RawQuery = values.Encode()

	req, err := http.NewRequest("GET", baseUrl.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	signature, ts, err := GenSignature(values.Encode(), api.ApiKey, api.ApiSecret, RECV_WINDOW)
	if err != nil {
		return nil, fmt.Errorf("failed to generate signature: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-BAPI-API-KEY", api.ApiKey)
	req.Header.Set("X-BAPI-SIGN", signature)
	req.Header.Set("X-BAPI-TIMESTAMP", strconv.FormatInt(ts, 10))
	req.Header.Set("X-BAPI-SIGN-TYPE", "2")
	req.Header.Set("X-BAPI-RECV-WINDOW", RECV_WINDOW)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var result GetOrdersResponse
	err = gjson.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}
