package bitflyer

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

const baseURL = "https://api.bitflyer.com/v1/"

type APIClient struct {
	key        string
	secret     string
	httpClient *http.Client
}

func New(key, secret string) *APIClient {
	apiClient := &APIClient{key, secret, &http.Client{}}
	return apiClient
}

func (api APIClient) header(method, endpoint string, body []byte) map[string]string {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	log.Println(timestamp)
	message := timestamp + method + endpoint + string(body)

	mac := hmac.New(sha256.New, []byte(api.secret))
	mac.Write([]byte(message))
	sign := hex.EncodeToString(mac.Sum(nil))
	return map[string]string{
		"ACCESS-KEY":       api.key,
		"ACCESS-TIMESTAMP": timestamp,
		"ACCESS-SIGN":      sign,
		"Content-Type":     "application/json",
	}
}

func (api *APIClient) doRequest(method, urlPath string, query map[string]string, data []byte) ([]byte, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base URL: %w", err)
	}
	apiURL, err := url.Parse(urlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse API URL: %w", err)
	}
	endpoint := base.ResolveReference(apiURL).String()
	log.Printf("action=doRequest=%s", endpoint)
	req, err := http.NewRequest(method, endpoint, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	q := req.URL.Query()
	for key, value := range query {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()

	for key, value := range api.header(method, req.URL.RequestURI(), data) {
		req.Header.Add(key, value)
	}
	resp, err := api.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}

type Balance struct {
	CurrencyCode string  `json:"currency_code"` // "currency_code" に対応
	Amount       float64 `json:"amount"`        // "amount" に対応
	Available    float64 `json:"available"`     // "available" に対応
}

func (api *APIClient) GetBalance() ([]Balance, error) {
	url := "me/getbalance"
	resp, err := api.doRequest("GET", url, map[string]string{}, nil)
	log.Printf("url=%s resp=%s", url, string(resp))
	if err != nil {
		log.Printf("action=GetBalance err=%s", err.Error())
		return nil, err
	}

	log.Printf("Raw Response: %s", string(resp))

	var balance []Balance
	err = json.Unmarshal(resp, &balance)
	if err != nil {
		log.Printf("action=GetBalance err=%s", err.Error())
		return nil, err
	}

	log.Printf("Decoded Balance: %+v", balance)
	return balance, nil
}

type Ticker struct {
	ProductCode     string  `json:"product_code"`
	State           string  `json:"state"`
	Timestamp       string  `json:"timestamp"`
	TickID          float64 `json:"tick_id"`
	BestBid         float64 `json:"best_bid"`
	BestAsk         float64 `json:"best_ask"`
	BestBidSize     float64 `json:"best_bid_size"`
	BestAskSize     float64 `json:"best_ask_size"`
	TotalBidDepth   float64 `json:"total_bid_depth"`
	TotalAskDepth   float64 `json:"total_ask_depth"`
	MarketBidSize   float64 `json:"market_bid_size"`
	MarketAskSize   float64 `json:"market_ask_size"`
	Ltp             float64 `json:"ltp"`
	Volume          float64 `json:"volume"`
	VolumeByProduct float64 `json:"volume_by_product"`
}

func (t *Ticker) GetMidPrice() float64 {
	return (t.BestBid + t.BestAsk) / 2
}

func (t *Ticker) DateTime() time.Time {
	dateTime, err := time.Parse(time.RFC3339, t.Timestamp)
	if err != nil {
		log.Printf("acrion=Datetime, err=%s", err.Error())
	}
	return dateTime
}

func (t *Ticker) TruncateDateTime(duration time.Duration) time.Time {
	return t.DateTime().Truncate(duration)
}

func (api *APIClient) GetTicker(productCode string) (*Ticker, error) {
	// product_code をクエリに含める
	url := fmt.Sprintf("ticker?product_code=%s", productCode)
	resp, err := api.doRequest("GET", url, map[string]string{}, nil)
	if err != nil {
		log.Printf("Failed to fetch ticker data: %v", err)
		return nil, fmt.Errorf("failed to fetch ticker data: %w", err)
	}

	var ticker Ticker
	if err := json.Unmarshal(resp, &ticker); err != nil {
		log.Printf("Failed to decode ticker response: %v", err)
		return nil, fmt.Errorf("failed to decode ticker response: %w", err)
	}

	// レスポンスに ProductCode が含まれない場合、手動で設定
	if ticker.ProductCode == "" {
		ticker.ProductCode = productCode
	}

	return &ticker, nil
}

func InitializeData(apiClient *APIClient, productCode string) (*Ticker, error) {
	ticker, err := apiClient.GetTicker(productCode)
	if err != nil {
		log.Printf("Failed to fetch initial data: %v", err)
		return nil, fmt.Errorf("failed to fetch initial ticker data: %w", err)
	}
	log.Printf("Initial Ticker Data: %+v", ticker)
	return ticker, nil
}

// websocketAPIを使用した処理

type JsonRPC2 struct {
	Version string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	Result  interface{} `json:"result,omitempty"`
	Id      *int        `json:"id,omitempty"`
}

type SubscribeParams struct {
	Channel string `json:"channel"`
}

func (api *APIClient) GetRealTimeTicker(ctx context.Context, symbol string, ch chan<- Ticker) {
	u := url.URL{Scheme: "wss", Host: "ws.lightstream.bitflyer.com", Path: "/json-rpc"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Printf("WebSocket dial error: %v", err)
	} else {
		log.Printf("WebSocket successfully connected to: %s", u.String())
	}

	defer func() {
		log.Println("Closing WebSocket connection")
		c.Close()
	}()

	channel := fmt.Sprintf("lightning_ticker_%s", symbol)
	log.Printf("Subscribing to channel: %s", channel)
	if err := c.WriteJSON(&JsonRPC2{
		Version: "2.0",
		Method:  "subscribe",
		Params:  &SubscribeParams{Channel: channel},
	}); err != nil {
		log.Printf("subscribe error: %v", err)
		return
	}

	//受信ループ
	for {
		select {
		case <-ctx.Done(): // 外部からの終了シグナル
			log.Println("context canceled, exiting WebSocket loop")
			return
		default:
			var message JsonRPC2
			if err := c.ReadJSON(&message); err != nil {
				log.Printf("WebSocket read error: %v", err)
				return
			}

			log.Printf("Raw Websocket message: %+v", message)

			if message.Method == "channelMessage" {
				if params, ok := message.Params.(map[string]interface{}); ok {
					if binary, found := params["message"]; found {
						log.Printf("Raw Params[\"message\"]: %+v", binary) // ここで確認
						jsonData, err := json.Marshal(binary)
						if err != nil {
							log.Printf("Failed to marshal binary: %v", err)
							return
						}
						var ticker Ticker
						if err := json.Unmarshal(jsonData, &ticker); err != nil {
							log.Printf("Failed to unmarshal ticker JSON: %v", err)
							return
						}
						ch <- ticker
					} else {
						log.Println("message field not found in params")
					}
				} else {
					log.Println("Invalid params type in WebSocket message")
				}
			} else {
				log.Printf("Unexpected message method: %s", message.Method)
			}
		}
	}
}
