package integrationtest

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type tokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func postJSON(t *testing.T, srv *httptest.Server, path string, body any, headers map[string]string) (int, map[string]any, []byte) {
	t.Helper()
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequest(http.MethodPost, srv.URL+path, rdr)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	var decoded map[string]any
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &decoded)
	}
	return res.StatusCode, decoded, raw
}

func getJSON(t *testing.T, srv *httptest.Server, path string, headers map[string]string) (int, []byte) {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, srv.URL+path, nil)
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	return res.StatusCode, raw
}

func putJSON(t *testing.T, srv *httptest.Server, path string, body any, headers map[string]string) (int, []byte) {
	t.Helper()
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodPut, srv.URL+path, bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	return res.StatusCode, raw
}

func decodeTokens(t *testing.T, body map[string]any) tokenPair {
	t.Helper()
	access, _ := body["access_token"].(string)
	refresh, _ := body["refresh_token"].(string)
	if access == "" || refresh == "" {
		t.Fatalf("missing tokens in %#v", body)
	}
	return tokenPair{AccessToken: access, RefreshToken: refresh}
}

func TestAuth_registerLoginRefreshLogout(t *testing.T) {
	db := StartPostgres(t)
	h, _ := NewTestHandler(t, db)
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	email := "user1@example.com"
	pass := "password123"

	status, body, _ := postJSON(t, srv, "/api/v1/register", map[string]string{
		"email": email, "password": pass,
	}, nil)
	if status != http.StatusCreated {
		t.Fatalf("register status=%d body=%v", status, body)
	}
	tokens := decodeTokens(t, body)

	status, raw := getJSON(t, srv, "/api/v1/servers", nil)
	if status != http.StatusUnauthorized {
		t.Fatalf("servers without auth: %d %s", status, raw)
	}

	status, raw = getJSON(t, srv, "/api/v1/servers", map[string]string{
		"Authorization": "Bearer " + tokens.AccessToken,
	})
	if status != http.StatusOK {
		t.Fatalf("servers with auth: %d %s", status, raw)
	}

	status, refreshBody, _ := postJSON(t, srv, "/api/v1/refresh", map[string]string{
		"refresh_token": tokens.RefreshToken,
	}, nil)
	if status != http.StatusOK {
		t.Fatalf("refresh status=%d body=%v", status, refreshBody)
	}
	if refreshBody["access_token"] == "" {
		t.Fatalf("refresh missing access_token: %#v", refreshBody)
	}

	status, _, _ = postJSON(t, srv, "/api/v1/logout", map[string]string{
		"refresh_token": tokens.RefreshToken,
	}, nil)
	if status != http.StatusNoContent && status != http.StatusOK {
		t.Fatalf("logout status=%d", status)
	}

	status, refreshBody, _ = postJSON(t, srv, "/api/v1/refresh", map[string]string{
		"refresh_token": tokens.RefreshToken,
	}, nil)
	if status == http.StatusOK {
		t.Fatalf("refresh after logout should fail, got %#v", refreshBody)
	}
}

func TestRelay_adminRegisterHealthAndList(t *testing.T) {
	db := StartPostgres(t)
	h, _ := NewTestHandler(t, db)
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	status, body, _ := postJSON(t, srv, "/api/v1/register", map[string]string{
		"email": "relay-user@example.com", "password": "password123",
	}, nil)
	if status != http.StatusCreated {
		t.Fatalf("register: %d %#v", status, body)
	}
	tokens := decodeTokens(t, body)

	admin := map[string]string{"X-Admin-Key": testAdminKey}
	status, regBody, raw := postJSON(t, srv, "/api/v1/servers", map[string]any{
		"id":                "nl-test-1",
		"region":            "nl",
		"host":              "203.0.113.10",
		"port":              443,
		"connection_config": "hysteria2://secret@203.0.113.10:443/",
	}, admin)
	if status != http.StatusCreated {
		t.Fatalf("register server: %d %s %#v", status, raw, regBody)
	}

	status, _, raw = postJSON(t, srv, "/api/v1/servers/health", map[string]any{
		"id": "nl-test-1", "healthy": true, "load_ratio": 0.1, "rtt_ms": 42,
	}, admin)
	if status != http.StatusNoContent && status != http.StatusOK {
		t.Fatalf("health: %d %s", status, raw)
	}

	status, listRaw := getJSON(t, srv, "/api/v1/servers", map[string]string{
		"Authorization": "Bearer " + tokens.AccessToken,
	})
	if status != http.StatusOK {
		t.Fatalf("list servers: %d %s", status, listRaw)
	}
	var servers []map[string]any
	if err := json.Unmarshal(listRaw, &servers); err != nil {
		t.Fatal(err)
	}
	if len(servers) != 1 || servers[0]["id"] != "nl-test-1" {
		t.Fatalf("servers=%#v", servers)
	}
	if servers[0]["healthy"] != true {
		t.Fatalf("expected healthy server: %#v", servers[0])
	}
}

func TestExclusions_putAndGet(t *testing.T) {
	db := StartPostgres(t)
	h, _ := NewTestHandler(t, db)
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	status, body, _ := postJSON(t, srv, "/api/v1/register", map[string]string{
		"email": "excl@example.com", "password": "password123",
	}, nil)
	if status != http.StatusCreated {
		t.Fatalf("register: %d %#v", status, body)
	}
	tokens := decodeTokens(t, body)
	auth := map[string]string{"Authorization": "Bearer " + tokens.AccessToken}

	status, raw := putJSON(t, srv, "/api/v1/exclusions", map[string]any{
		"domains": []string{"*.bank.ru", "mail.ru"},
	}, auth)
	if status != http.StatusOK {
		t.Fatalf("put exclusions: %d %s", status, raw)
	}

	status, raw = getJSON(t, srv, "/api/v1/exclusions", auth)
	if status != http.StatusOK {
		t.Fatalf("get exclusions: %d %s", status, raw)
	}
	var got struct {
		Domains []string `json:"domains"`
	}
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Domains) != 2 {
		t.Fatalf("domains=%#v", got.Domains)
	}
}

func TestBilling_createPayment(t *testing.T) {
	db := StartPostgres(t)
	h, _ := NewTestHandler(t, db)
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	status, body, _ := postJSON(t, srv, "/api/v1/register", map[string]string{
		"email": "bill@example.com", "password": "password123",
	}, nil)
	if status != http.StatusCreated {
		t.Fatalf("register: %d %#v", status, body)
	}
	tokens := decodeTokens(t, body)

	status, payBody, raw := postJSON(t, srv, "/api/v1/payments", map[string]any{}, map[string]string{
		"Authorization": "Bearer " + tokens.AccessToken,
	})
	if status != http.StatusCreated {
		t.Fatalf("create payment: %d %s %#v", status, raw, payBody)
	}
	url, _ := payBody["confirmation_url"].(string)
	if url == "" {
		t.Fatalf("missing confirmation_url: %#v", payBody)
	}
}
