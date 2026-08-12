package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const apiBase = "https://api.telegram.org"

// Client talks to Telegram Bot API (Stars invoices + messaging).
type Client struct {
	token  string
	http   *http.Client
	baseURL string
}

// New builds a Bot API client. Empty token disables outbound calls.
func New(token string) *Client {
	return &Client{
		token:   token,
		baseURL: apiBase,
		http:    &http.Client{Timeout: 12 * time.Second},
	}
}

// Enabled reports whether a bot token is configured.
func (c *Client) Enabled() bool { return c != nil && c.token != "" }

type apiResponse struct {
	OK          bool            `json:"ok"`
	Result      json.RawMessage `json:"result"`
	Description string          `json:"description"`
}

// CreateInvoiceLink creates a Stars (XTR) invoice link.
func (c *Client) CreateInvoiceLink(ctx context.Context, title, description, payload string, starsAmount int) (string, error) {
	if !c.Enabled() {
		return "", fmt.Errorf("telegram bot token not configured")
	}
	body := map[string]any{
		"title":          title,
		"description":    description,
		"payload":        payload,
		"currency":       "XTR",
		"prices":         []map[string]any{{"label": title, "amount": starsAmount}},
		"provider_token": "", // Stars: empty provider token
	}
	var link string
	if err := c.call(ctx, "createInvoiceLink", body, &link); err != nil {
		return "", err
	}
	return link, nil
}

// AnswerPreCheckoutQuery confirms or rejects a Stars checkout.
func (c *Client) AnswerPreCheckoutQuery(ctx context.Context, queryID string, ok bool, errMsg string) error {
	body := map[string]any{
		"pre_checkout_query_id": queryID,
		"ok":                    ok,
	}
	if !ok && errMsg != "" {
		body["error_message"] = errMsg
	}
	return c.call(ctx, "answerPreCheckoutQuery", body, nil)
}

// SendMessage sends a plain text message to a Telegram user.
func (c *Client) SendMessage(ctx context.Context, chatID int64, text string) error {
	body := map[string]any{
		"chat_id": chatID,
		"text":    text,
	}
	return c.call(ctx, "sendMessage", body, nil)
}

// SetWebhook registers the bot webhook URL with Telegram.
func (c *Client) SetWebhook(ctx context.Context, url, secret string) error {
	body := map[string]any{"url": url}
	if secret != "" {
		body["secret_token"] = secret
	}
	return c.call(ctx, "setWebhook", body, nil)
}

func (c *Client) call(ctx context.Context, method string, body any, out any) error {
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/bot%s/%s", c.baseURL, c.token, method), bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return err
	}
	var parsed apiResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return err
	}
	if !parsed.OK {
		return fmt.Errorf("telegram api %s: %s", method, parsed.Description)
	}
	if out == nil || len(parsed.Result) == 0 || string(parsed.Result) == "true" {
		return nil
	}
	return json.Unmarshal(parsed.Result, out)
}
