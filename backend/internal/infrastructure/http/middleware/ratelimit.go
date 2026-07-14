package middleware

import (
	"encoding/json"
	"net"
	"net/http"
	"sync"
	"time"

	apperrors "streampass/shared/errors"
)

// RateLimiter is a simple per-IP fixed-window limiter. A fixed window
// (rather than a token bucket or sliding log) was chosen deliberately:
// StreamPass's MVP traffic doesn't need burst smoothing, and a fixed
// window is a single counter + timestamp per IP — the simplest thing that
// satisfies the spec's rate-limiting requirement (KISS).
type RateLimiter struct {
	limit  int
	window time.Duration

	mu      sync.Mutex
	buckets map[string]*bucket
}

type bucket struct {
	count      int
	windowFrom time.Time
}

// NewRateLimiter builds a limiter allowing `limit` requests per `window`
// per client IP. It starts a background goroutine that periodically
// evicts stale per-IP buckets so long-running processes don't accumulate
// an unbounded map of one-off client IPs.
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{limit: limit, window: window, buckets: make(map[string]*bucket)}
	go rl.evictLoop()
	return rl
}

func (rl *RateLimiter) evictLoop() {
	ticker := time.NewTicker(rl.window)
	defer ticker.Stop()
	for range ticker.C {
		rl.evictStale()
	}
}

func (rl *RateLimiter) evictStale() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	for key, b := range rl.buckets {
		if now.Sub(b.windowFrom) >= rl.window {
			delete(rl.buckets, key)
		}
	}
}

// Middleware returns the http middleware enforcing this limiter.
func (rl *RateLimiter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r)
			if !rl.allow(ip) {
				writeRateLimited(w)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (rl *RateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b, ok := rl.buckets[key]
	if !ok || now.Sub(b.windowFrom) >= rl.window {
		rl.buckets[key] = &bucket{count: 1, windowFrom: now}
		return true
	}
	if b.count >= rl.limit {
		return false
	}
	b.count++
	return true
}

func clientIP(r *http.Request) string {
	// StreamPass runs behind a reverse proxy in production (deploy/README);
	// X-Forwarded-For is trusted here because the proxy is expected to set
	// it and strip any client-supplied value before forwarding — the same
	// assumption every request makes about reaching this process at all.
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		return fwd
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func writeRateLimited(w http.ResponseWriter) {
	body, err := json.Marshal(struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}{
		Error: struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}{Code: string(apperrors.CodeRateLimited), Message: "too many requests"},
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusTooManyRequests)
	if err == nil {
		_, _ = w.Write(body)
	}
}
