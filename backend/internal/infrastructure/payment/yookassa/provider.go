// Package yookassa implements billing.PaymentProvider against the ЮKassa
// REST API (https://yookassa.ru/developers/api). This is a real HTTP
// client, not a mock — but it has not been exercised against live ЮKassa
// credentials in this environment (the sandbox has no network access to
// api.yookassa.ru), so treat it as needing a smoke test against a real
// ЮKassa shop before production use.
package yookassa

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"streampass/backend/internal/application/billing"
	apperrors "streampass/shared/errors"
)

const apiBaseURL = "https://api.yookassa.ru/v3"

// Provider implements billing.PaymentProvider against ЮKassa.
type Provider struct {
	shopID     string
	secretKey  string
	returnURL  string
	httpClient *http.Client
}

// Config holds the ЮKassa shop credentials, loaded from config/env — never
// hardcoded (spec: "Запрещено использовать hardcode").
type Config struct {
	ShopID    string
	SecretKey string
	// ReturnURL is where ЮKassa redirects the user's browser after they
	// confirm payment (spec section 15: payment confirmation flow).
	ReturnURL string
}

// New builds a ЮKassa Provider.
func New(cfg Config) *Provider {
	return &Provider{
		shopID:     cfg.ShopID,
		secretKey:  cfg.SecretKey,
		returnURL:  cfg.ReturnURL,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

type createPaymentRequest struct {
	Amount struct {
		Value    string `json:"value"`
		Currency string `json:"currency"`
	} `json:"amount"`
	Confirmation struct {
		Type      string `json:"type"`
		ReturnURL string `json:"return_url"`
	} `json:"confirmation"`
	Description string `json:"description"`
	Capture     bool   `json:"capture"`
}

type createPaymentResponse struct {
	ID           string `json:"id"`
	Status       string `json:"status"`
	Confirmation struct {
		ConfirmationURL string `json:"confirmation_url"`
	} `json:"confirmation"`
}

// CreatePayment implements billing.PaymentProvider.
func (p *Provider) CreatePayment(ctx context.Context, userID string, amountRUB int64, description string) (string, string, error) {
	body := createPaymentRequest{Description: description, Capture: true}
	body.Amount.Value = fmt.Sprintf("%d.00", amountRUB)
	body.Amount.Currency = "RUB"
	body.Confirmation.Type = "redirect"
	body.Confirmation.ReturnURL = p.returnURL

	var resp createPaymentResponse
	if err := p.doRequest(ctx, http.MethodPost, "/payments", userID /* idempotence key */, body, &resp); err != nil {
		return "", "", err
	}
	return resp.ID, resp.Confirmation.ConfirmationURL, nil
}

type paymentStatusResponse struct {
	Status string `json:"status"`
}

// FetchPaymentStatus implements billing.PaymentProvider.
func (p *Provider) FetchPaymentStatus(ctx context.Context, providerPaymentID string) (billing.PaymentStatus, error) {
	var resp paymentStatusResponse
	if err := p.doRequest(ctx, http.MethodGet, "/payments/"+providerPaymentID, "", nil, &resp); err != nil {
		return "", err
	}
	return mapStatus(resp.Status), nil
}

// mapStatus translates ЮKassa's status vocabulary to the provider-agnostic
// one billing.PaymentStatus defines, per the abstraction spec section 15
// requires so a second provider can be added later without touching
// use-case logic.
func mapStatus(yookassaStatus string) billing.PaymentStatus {
	switch yookassaStatus {
	case "succeeded":
		return billing.PaymentStatusSucceeded
	case "canceled":
		return billing.PaymentStatusCanceled
	default:
		return billing.PaymentStatusPending
	}
}

func (p *Provider) doRequest(ctx context.Context, method, path, idempotenceKey string, reqBody any, out any) error {
	var bodyReader io.Reader
	if reqBody != nil {
		b, err := json.Marshal(reqBody)
		if err != nil {
			return apperrors.Wrap(apperrors.CodeInternal, "failed to marshal yookassa request", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, apiBaseURL+path, bodyReader)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to build yookassa request", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+basicAuth(p.shopID, p.secretKey))
	if idempotenceKey != "" {
		req.Header.Set("Idempotence-Key", idempotenceKey)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeUnavailable, "yookassa request failed", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return apperrors.New(apperrors.CodePaymentFailed, "yookassa returned an error").
			WithDetails(map[string]any{"status": resp.StatusCode, "body": string(respBody)})
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return apperrors.Wrap(apperrors.CodeInternal, "failed to decode yookassa response", err)
		}
	}
	return nil
}

func basicAuth(user, pass string) string {
	return base64.StdEncoding.EncodeToString([]byte(user + ":" + pass))
}
