package bybit_api

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	json "github.com/goccy/go-json"
	"github.com/gorilla/schema"
)

type RESTManager struct {
	api     *BybitApi
	logger  Logger
	timeout time.Duration

	encoder *schema.Encoder
}

func NewRESTManager(api *BybitApi) *RESTManager {
	rm := &RESTManager{
		api:     api,
		encoder: schema.NewEncoder(),
		logger:  api.Logger,
		timeout: api.timeout,
	}

	return rm
}

func (r *RESTManager) getHTTPClient() *http.Client {
	return &http.Client{Timeout: 10 * time.Second}
}

func (r *RESTManager) encodeToQuery(params interface{}) (string, error) {
	values := url.Values{}
	if err := r.encoder.Encode(params, values); err != nil {
		return "", fmt.Errorf("failed to encode params: %w", err)
	}
	return values.Encode(), nil
}

func (r *RESTManager) sendRequest(req *http.Request, result interface{}) error {
	client := r.getHTTPClient()

	ctx, cancel := context.WithTimeout(r.api.context, r.timeout)
	defer cancel()

	req = req.WithContext(ctx)

	r.logger.Debug("Sending request to: %s", req.URL.String())
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	r.logger.Debug("Response status: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return json.Unmarshal(body, result)
}

func (r *RESTManager) setAuthHeaders(req *http.Request, signature string, timestamp int64) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(apiRequestKey, r.api.ApiKey)
	req.Header.Set(signatureKey, signature)
	req.Header.Set(timestampKey, strconv.FormatInt(timestamp, 10))
	req.Header.Set(signTypeKey, "2")
	req.Header.Set(recvWindowKey, RECV_WINDOW)
}

// GetServerTime retrieves the current server time
func (r *RESTManager) GetServerTime() (*GetServerTimeResponse, error) {
	const path = "/v5/market/time"
	url := r.api.BASE_REST_URL + path

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	var result GetServerTimeResponse
	if err := r.sendRequest(req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *RESTManager) PlaceOrder(params PlaceOrderParams) (*PlaceOrderResponse, error) {
	const path = "/v5/order/create"
	url := r.api.BASE_REST_URL + path

	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal params: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(paramsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	signature, timestamp, err := genSignature(
		params,
		r.api.ApiKey,
		r.api.ApiSecret,
		RECV_WINDOW,
		&TimeProvider{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate signature: %w", err)
	}

	r.setAuthHeaders(req, signature, timestamp)

	var result PlaceOrderResponse
	if err := r.sendRequest(req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *RESTManager) CancelOrder(params CancelOrderParams) (*CancelOrderResponse, error) {
	const path = "/v5/order/cancel"
	url := r.api.BASE_REST_URL + path

	if params.Category == "" || params.Symbol == "" {
		return nil, fmt.Errorf("category and symbol are required fields")
	}
	if params.OrderId == nil && params.OrderLinkId == nil {
		return nil, fmt.Errorf("either orderId or orderLinkId must be provided")
	}

	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal params: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(paramsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	signature, timestamp, err := genSignature(
		params,
		r.api.ApiKey,
		r.api.ApiSecret,
		RECV_WINDOW,
		&TimeProvider{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate signature: %w", err)
	}

	r.setAuthHeaders(req, signature, timestamp)

	var result CancelOrderResponse
	if err := r.sendRequest(req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *RESTManager) GetOrders(params OpenOrderRequest) (*GetOrdersResponse, error) {
	const path = "/v5/order/realtime"

	queryStr, err := r.encodeToQuery(params)
	if err != nil {
		return nil, err
	}

	reqURL := fmt.Sprintf("%s%s?%s", r.api.BASE_REST_URL, path, queryStr)

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	signature, timestamp, err := genSignature(
		queryStr,
		r.api.ApiKey,
		r.api.ApiSecret,
		RECV_WINDOW,
		&TimeProvider{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate signature: %w", err)
	}

	r.setAuthHeaders(req, signature, timestamp)

	var result GetOrdersResponse
	if err := r.sendRequest(req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *RESTManager) GetKline(params GetKlineParams) (*GetKlineResponse, error) {
	const path = "/v5/market/kline"

	queryStr, err := r.encodeToQuery(params)
	if err != nil {
		return nil, fmt.Errorf("failed to encode query parameters: %w", err)
	}

	reqURL := fmt.Sprintf("%s%s?%s", r.api.BASE_REST_URL, path, queryStr)

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var result GetKlineResponse
	if err := r.sendRequest(req, &result); err != nil {
		return nil, fmt.Errorf("failed to get kline data: %w", err)
	}

	return &result, nil
}

func (r *RESTManager) GetMarkPriceKline(params GetKlineParams) (*GetKlineResponse, error) {
	const path = "/v5/market/mark-price-kline"

	queryStr, err := r.encodeToQuery(params)
	if err != nil {
		return nil, fmt.Errorf("failed to encode query parameters: %w", err)
	}

	reqURL := fmt.Sprintf("%s%s?%s", r.api.BASE_REST_URL, path, queryStr)

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var result GetKlineResponse
	if err := r.sendRequest(req, &result); err != nil {
		return nil, fmt.Errorf("failed to get mark price kline data: %w", err)
	}

	return &result, nil
}

func (r *RESTManager) GetIndexPriceKline(params GetKlineParams) (*GetKlineResponse, error) {
	const path = "/v5/market/index-price-kline"

	queryStr, err := r.encodeToQuery(params)
	if err != nil {
		return nil, fmt.Errorf("failed to encode query parameters: %w", err)
	}

	reqURL := fmt.Sprintf("%s%s?%s", r.api.BASE_REST_URL, path, queryStr)

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var result GetKlineResponse
	if err := r.sendRequest(req, &result); err != nil {
		return nil, fmt.Errorf("failed to get index price kline data: %w", err)
	}

	return &result, nil
}

func (r *RESTManager) GetOrderbook(params GetOrderbookParams) (*GetOrderbookResponse, error) {
	const path = "/v5/market/orderbook"

	queryStr, err := r.encodeToQuery(params)
	if err != nil {
		return nil, fmt.Errorf("failed to encode query parameters: %w", err)
	}

	reqURL := fmt.Sprintf("%s%s?%s", r.api.BASE_REST_URL, path, queryStr)

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var result GetOrderbookResponse
	if err := r.sendRequest(req, &result); err != nil {
		return nil, fmt.Errorf("failed to get orderbook data: %w", err)
	}

	return &result, nil
}

func (r *RESTManager) GetInstrumentsInfo(params GetInstrumentsInfoParams) (*GetInstrumentsInfoResponse, error) {
	const path = "/v5/market/instruments-info"

	buildRequest := func(params GetInstrumentsInfoParams) (*http.Request, error) {
		queryStr, err := r.encodeToQuery(params)
		if err != nil {
			return nil, fmt.Errorf("failed to encode query parameters: %w", err)
		}

		reqURL := fmt.Sprintf("%s%s?%s", r.api.BASE_REST_URL, path, queryStr)

		req, err := http.NewRequest(http.MethodGet, reqURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		return req, nil
	}
	req, err := buildRequest(params)
	if err != nil {
		return nil, err
	}

	var result GetInstrumentsInfoResponse
	if err := r.sendRequest(req, &result); err != nil {
		return nil, fmt.Errorf("failed to get instruments info: %w", err)
	}
	cursor := result.Result.NextPageCursor
	for cursor != "" {
		params.Cursor = cursor
		req, err := buildRequest(params)
		if err != nil {
			return nil, err
		}

		var cursorResult GetInstrumentsInfoResponse
		if err := r.sendRequest(req, &cursorResult); err != nil {
			return nil, fmt.Errorf("failed to get instruments info: %w", err)
		}
		result.Result.List = append(result.Result.List, cursorResult.Result.List...)
		cursor = cursorResult.Result.NextPageCursor
		if cursor == "" {
			result.Result.NextPageCursor = cursor
		}
	}

	return &result, nil
}

func (r *RESTManager) GetTicker(params GetTickerParams) (*GetTickerResponse, error) {
	const path = "/v5/market/tickers"
	queryStr, err := r.encodeToQuery(params)
	if err != nil {
		return nil, fmt.Errorf("failed to encode query parameters: %w", err)
	}

	reqURL := fmt.Sprintf("%s%s?%s", r.api.BASE_REST_URL, path, queryStr)
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var result GetTickerResponse
	if err := r.sendRequest(req, &result); err != nil {
		return nil, fmt.Errorf("failed to get recent public trades: %w", err)
	}

	return &result, nil
}

func (r *RESTManager) GetAccountInfo() (*GetAccountInfoResponse, error) {
	const path = "/v5/account/info"

	reqURL := fmt.Sprintf("%s%s", r.api.BASE_REST_URL, path)
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	timestamp := time.Now().UnixMilli()
	signature, timestamp, err := genSignature(
		"",
		r.api.ApiKey,
		r.api.ApiSecret,
		RECV_WINDOW,
		&TimeProvider{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate signature: %w", err)
	}

	r.setAuthHeaders(req, signature, timestamp)

	var result GetAccountInfoResponse
	if err := r.sendRequest(req, &result); err != nil {
		return nil, fmt.Errorf("failed to get wallet balance: %w", err)
	}

	return &result, nil
}

func (r *RESTManager) GetRecentPublicTrades(params GetRecentPublicTradesParams) (*GetRecentPublicTradesResponse, error) {

	const path = "/v5/market/recent-trade"

	queryStr, err := r.encodeToQuery(params)
	if err != nil {
		return nil, fmt.Errorf("failed to encode query parameters: %w", err)
	}

	reqURL := fmt.Sprintf("%s%s?%s", r.api.BASE_REST_URL, path, queryStr)
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var result GetRecentPublicTradesResponse
	if err := r.sendRequest(req, &result); err != nil {
		return nil, fmt.Errorf("failed to get recent public trades: %w", err)
	}

	return &result, nil

}

func (r *RESTManager) GetWalletBalance(params GetWalletBalanceParams) (*GetWalletBalanceResponse, error) {
	const path = "/v5/account/wallet-balance"

	queryStr, err := r.encodeToQuery(params)
	if err != nil {
		return nil, fmt.Errorf("failed to encode query parameters: %w", err)
	}

	reqURL := fmt.Sprintf("%s%s?%s", r.api.BASE_REST_URL, path, queryStr)
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	timestamp := time.Now().UnixMilli()
	signature, timestamp, err := genSignature(
		queryStr,
		r.api.ApiKey,
		r.api.ApiSecret,
		RECV_WINDOW,
		&TimeProvider{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate signature: %w", err)
	}

	r.setAuthHeaders(req, signature, timestamp)

	var result GetWalletBalanceResponse
	if err := r.sendRequest(req, &result); err != nil {
		return nil, fmt.Errorf("failed to get wallet balance: %w", err)
	}

	return &result, nil
}

func (r *RESTManager) GetPositions(params GetPositionParams) (*GetPositionResponse, error) {
	const path = "/v5/position/list"

	queryStr, err := r.encodeToQuery(params)
	if err != nil {
		return nil, fmt.Errorf("failed to encode query parameters: %w", err)
	}

	reqURL := fmt.Sprintf("%s%s?%s", r.api.BASE_REST_URL, path, queryStr)

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	signature, timestamp, err := genSignature(
		queryStr,
		r.api.ApiKey,
		r.api.ApiSecret,
		RECV_WINDOW,
		&TimeProvider{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate signature: %w", err)
	}

	r.setAuthHeaders(req, signature, timestamp)

	var result GetPositionResponse
	if err := r.sendRequest(req, &result); err != nil {
		return nil, fmt.Errorf("failed to get position data: %w", err)
	}

	return &result, nil
}
