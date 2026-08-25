// Package platega implements billing.PaymentProvider against Platega.io
// (https://docs.platega.io). Live credentials are required for production;
// without them CreatePayment returns a clear configuration error.
package platega

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"streampass/backend/internal/application/billing"
	apperrors "streampass/shared/errors"
)

const apiBaseURL = "https://app.platega.io"

// PaymentMethodSBPQR is Platega paymentMethod for СБП QR (recommended MVP).
const PaymentMethodSBPQR = 2

// PaymentMethodCardRUB is Platega paymentMethod for bank cards.
const PaymentMethodCardRUB = 10

// Provider implements billing.PaymentProvider.
type Provider struct {
	merchantID    string
	secret        string
	returnURL     string
	failedURL     string
	paymentMethod int
	httpClient    *http.Client
}

// Config holds Platega credentials from env/config (never hardcode secrets).
type Config struct {
	MerchantID    string
	Secret        string
	ReturnURL     string
	FailedURL     string
	PaymentMethod int // default SBP QR = 2
}

// New builds a Platega Provider.
func New(cfg Config) *Provider {
	pm := cfg.PaymentMethod
	if pm == 0 {
		pm = PaymentMethodSBPQR
	}
	failed := cfg.FailedURL
	if failed == "" {
		failed = cfg.ReturnURL
	}
	return &Provider{
		merchantID:    cfg.MerchantID,
		secret:        cfg.Secret,
		returnURL:     cfg.ReturnURL,
		failedURL:     failed,
		paymentMethod: pm,
		httpClient:    &http.Client{Timeout: 15 * time.Second},
	}
}

// Name identifies this provider in payments.provider.
func (p *Provider) Name() string { return "platega" }

// Enabled reports whether credentials are configured.
func (p *Provider) Enabled() bool {
	return p != nil && p.merchantID != "" && p.secret != "" && p.returnURL != ""
}

type createBody struct {
	PaymentMethod  int `json:"paymentMethod"`
	PaymentDetails struct {
		Amount   int    `json:"amount"`
		Currency string `json:"currency"`
	} `json:"paymentDetails"`
	Description string `json:"description"`
	ReturnURL   string `json:"returnUrl"`
	FailedURL   string `json:"failedUrl"`
	Payload     string `json:"payload,omitempty"`
}

type createResp struct {
	TransactionID string `json:"transactionId"`
	Redirect      string `json:"redirect"`
	Status        string `json:"status"`
}

// CreatePayment implements billing.PaymentProvider.
func (p *Provider) CreatePayment(ctx context.Context, userID string, amountRUB int64, description string) (string, string, error) {
	if !p.Enabled() {
		return "", "", apperrors.New(apperrors.CodePaymentFailed, "platega is not configured")
	}
	var body createBody
	body.PaymentMethod = p.paymentMethod
	body.PaymentDetails.Amount = int(amountRUB)
	body.PaymentDetails.Currency = "RUB"
	body.Description = description
	body.ReturnURL = p.returnURL
	body.FailedURL = p.failedURL
	body.Payload = userID

	raw, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiBaseURL+"/transaction/process", bytes.NewReader(raw))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-MerchantId", p.merchantID)
	req.Header.Set("X-Secret", p.secret)

	res, err := p.httpClient.Do(req)
	if err != nil {
		return "", "", apperrors.Wrap(apperrors.CodePaymentFailed, "platega create failed", err)
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return "", "", apperrors.New(apperrors.CodePaymentFailed, "platega create rejected").
			WithDetails(map[string]any{"status": res.StatusCode, "body": truncate(string(b), 200)})
	}
	var out createResp
	if err := json.Unmarshal(b, &out); err != nil {
		return "", "", apperrors.Wrap(apperrors.CodePaymentFailed, "platega create decode failed", err)
	}
	if out.TransactionID == "" || out.Redirect == "" {
		return "", "", apperrors.New(apperrors.CodePaymentFailed, "platega create missing redirect")
	}
	return out.TransactionID, out.Redirect, nil
}

type statusResp struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// FetchPaymentStatus implements billing.PaymentProvider.
func (p *Provider) FetchPaymentStatus(ctx context.Context, providerPaymentID string) (billing.PaymentStatus, error) {
	if !p.Enabled() {
		return "", apperrors.New(apperrors.CodePaymentFailed, "platega is not configured")
	}
	url := fmt.Sprintf("%s/transaction/%s", apiBaseURL, providerPaymentID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("X-MerchantId", p.merchantID)
	req.Header.Set("X-Secret", p.secret)
	res, err := p.httpClient.Do(req)
	if err != nil {
		return "", apperrors.Wrap(apperrors.CodePaymentFailed, "platega status failed", err)
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return "", apperrors.New(apperrors.CodePaymentFailed, "platega status rejected").
			WithDetails(map[string]any{"status": res.StatusCode})
	}
	var out statusResp
	if err := json.Unmarshal(b, &out); err != nil {
		return "", err
	}
	switch strings.ToUpper(out.Status) {
	case "CONFIRMED", "SUCCESS", "SUCCEEDED", "PAID":
		return billing.PaymentStatusSucceeded, nil
	case "CANCELED", "CANCELLED", "FAILED", "CHARGEBACKED":
		return billing.PaymentStatusCanceled, nil
	default:
		return billing.PaymentStatusPending, nil
	}
}

// VerifyCallbackHeaders checks X-MerchantId / X-Secret from Platega callback.
func (p *Provider) VerifyCallbackHeaders(merchantID, secret string) bool {
	if !p.Enabled() {
		return false
	}
	return merchantID == p.merchantID && secret == p.secret
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
