// Package metrics exposes a minimal Prometheus text exposition endpoint
// without pulling the full client_golang dependency (KISS for BL-021).
package metrics

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Collector holds process-local counters/gauges for /metrics.
type Collector struct {
	httpTotal   sync.Map // key: method|path|status -> *atomic.Uint64
	httpSeconds sync.Map // key: method|path -> *atomic.Uint64 (milliseconds sum)
	httpCount   sync.Map // key: method|path -> *atomic.Uint64
	started     time.Time
}

// Default is the process-wide collector used by middleware and Handler.
var Default = New()

func New() *Collector {
	return &Collector{started: time.Now().UTC()}
}

func (c *Collector) ObserveHTTP(method, path string, status int, d time.Duration) {
	path = normalizePath(path)
	statusKey := method + "|" + path + "|" + strconv.Itoa(status)
	incMap(&c.httpTotal, statusKey)

	latencyKey := method + "|" + path
	addMap(&c.httpSeconds, latencyKey, uint64(d.Milliseconds()))
	incMap(&c.httpCount, latencyKey)
}

func normalizePath(path string) string {
	if path == "" {
		return "/"
	}
	// Avoid high-cardinality path labels from IDs.
	parts := strings.Split(path, "/")
	for i, p := range parts {
		if p == "" {
			continue
		}
		if looksLikeID(p) {
			parts[i] = ":id"
		}
	}
	out := strings.Join(parts, "/")
	if !strings.HasPrefix(out, "/") {
		out = "/" + out
	}
	return out
}

func looksLikeID(s string) bool {
	if strings.HasPrefix(s, "usr_") || strings.HasPrefix(s, "rel_") {
		return true
	}
	if len(s) >= 16 {
		hexish := true
		for _, r := range s {
			if (r < '0' || r > '9') && (r < 'a' || r > 'f') && (r < 'A' || r > 'F') && r != '-' {
				hexish = false
				break
			}
		}
		if hexish {
			return true
		}
	}
	return false
}

func incMap(m *sync.Map, key string) {
	v, _ := m.LoadOrStore(key, &atomic.Uint64{})
	v.(*atomic.Uint64).Add(1)
}

func addMap(m *sync.Map, key string, delta uint64) {
	v, _ := m.LoadOrStore(key, &atomic.Uint64{})
	v.(*atomic.Uint64).Add(delta)
}

// Handler serves Prometheus text format at GET /metrics.
func Handler(w http.ResponseWriter, r *http.Request) {
	Default.WritePrometheus(w)
}

func (c *Collector) WritePrometheus(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	var b strings.Builder
	b.WriteString("# HELP streampass_up 1 if the process is serving metrics.\n")
	b.WriteString("# TYPE streampass_up gauge\n")
	b.WriteString("streampass_up 1\n")
	b.WriteString("# HELP streampass_uptime_seconds Process uptime in seconds.\n")
	b.WriteString("# TYPE streampass_uptime_seconds gauge\n")
	fmt.Fprintf(&b, "streampass_uptime_seconds %.0f\n", time.Since(c.started).Seconds())

	b.WriteString("# HELP streampass_http_requests_total HTTP requests by method, path, status.\n")
	b.WriteString("# TYPE streampass_http_requests_total counter\n")
	c.httpTotal.Range(func(key, value any) bool {
		parts := strings.Split(key.(string), "|")
		if len(parts) != 3 {
			return true
		}
		fmt.Fprintf(&b, "streampass_http_requests_total{method=%q,path=%q,status=%q} %d\n",
			parts[0], parts[1], parts[2], value.(*atomic.Uint64).Load())
		return true
	})

	b.WriteString("# HELP streampass_http_request_duration_milliseconds_sum Sum of request durations.\n")
	b.WriteString("# TYPE streampass_http_request_duration_milliseconds_sum counter\n")
	c.httpSeconds.Range(func(key, value any) bool {
		parts := strings.Split(key.(string), "|")
		if len(parts) != 2 {
			return true
		}
		fmt.Fprintf(&b, "streampass_http_request_duration_milliseconds_sum{method=%q,path=%q} %d\n",
			parts[0], parts[1], value.(*atomic.Uint64).Load())
		return true
	})

	b.WriteString("# HELP streampass_http_request_duration_milliseconds_count Request count for latency sum.\n")
	b.WriteString("# TYPE streampass_http_request_duration_milliseconds_count counter\n")
	c.httpCount.Range(func(key, value any) bool {
		parts := strings.Split(key.(string), "|")
		if len(parts) != 2 {
			return true
		}
		fmt.Fprintf(&b, "streampass_http_request_duration_milliseconds_count{method=%q,path=%q} %d\n",
			parts[0], parts[1], value.(*atomic.Uint64).Load())
		return true
	})

	_, _ = w.Write([]byte(b.String()))
}

// Middleware records HTTP metrics around the handler chain.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/metrics" || strings.HasSuffix(r.URL.Path, "/metrics") {
			next.ServeHTTP(w, r)
			return
		}
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		Default.ObserveHTTP(r.Method, r.URL.Path, rec.status, time.Since(start))
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}
