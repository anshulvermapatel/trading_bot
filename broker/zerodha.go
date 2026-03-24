package broker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"trading_bot/models"
)

const (
	kiteBaseURL = "https://api.kite.trade"
)

type ZerodhaAuth struct {
	APIKey      string
	AccessToken string
}

type ZerodhaBroker struct {
	httpClient *http.Client
	auth       ZerodhaAuth
}

func NewZerodhaBroker(client *http.Client, auth ZerodhaAuth) *ZerodhaBroker {
	if client == nil {
		client = http.DefaultClient
	}
	return &ZerodhaBroker{httpClient: client, auth: auth}
}

func (z *ZerodhaBroker) PlaceOrder(order models.Order) error {
	if z.auth.APIKey == "" || z.auth.AccessToken == "" {
		return fmt.Errorf("zerodha auth missing API key or access token")
	}

	transactionType := "BUY"
	if order.Side == models.SellSignal {
		transactionType = "SELL"
	}

	payload := map[string]any{
		"tradingsymbol":    order.Symbol,
		"exchange":         "NSE",
		"transaction_type": transactionType,
		"order_type":       "MARKET",
		"quantity":         order.Qty,
		"product":          "MIS",
		"validity":         "DAY",
		"tag":              order.OrderTag,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal order payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, kiteBaseURL+"/orders/regular", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("build kite order request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Kite-Version", "3")
	req.Header.Set("Authorization", fmt.Sprintf("token %s:%s", z.auth.APIKey, z.auth.AccessToken))

	resp, err := z.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("kite order request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("kite order failed, status=%d", resp.StatusCode)
	}

	// TODO: Parse Kite response and return order_id for persistence/audit.
	return nil
}
