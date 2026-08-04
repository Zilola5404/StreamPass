package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestWritePrometheusContainsCoreSeries(t *testing.T) {
	c := New()
	c.ObserveHTTP(http.MethodGet, "/api/v1/health", 200, 12*time.Millisecond)
	c.ObserveHTTP(http.MethodGet, "/api/v1/users/usr_abc123def4567890", 401, 5*time.Millisecond)

	rec := httptest.NewRecorder()
	c.WritePrometheus(rec)
	body := rec.Body.String()
	for _, needle := range []string{
		"streampass_up 1",
		"streampass_http_requests_total",
		`path="/api/v1/health"`,
		`path="/api/v1/users/:id"`,
		`status="200"`,
	} {
		if !strings.Contains(body, needle) {
			t.Fatalf("metrics body missing %q\n%s", needle, body)
		}
	}
}

func TestHandlerOK(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	Handler(rec, req)
	if rec.Code != 200 {
		t.Fatalf("status=%d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "text/plain") {
		t.Fatalf("content-type=%q", ct)
	}
}
