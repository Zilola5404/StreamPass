// Command loadtest runs a lightweight HTTP load test against StreamPass API
// endpoints (BL-032). No external k6/vegeta binary required.
//
// Usage:
//   go run ./scripts/loadtest -base https://212-43-156-33.nip.io -duration 15s -rps 40
//   go run ./scripts/loadtest -base http://127.0.0.1:8080 -email u@x -password secret
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type scenario struct {
	name   string
	method string
	path   string
	auth   bool
	body   []byte
}

type sample struct {
	latency time.Duration
	status  int
	err     bool
}

func main() {
	base := flag.String("base", "https://212-43-156-33.nip.io", "API host (no /api/v1 suffix)")
	duration := flag.Duration("duration", 15*time.Second, "test duration")
	rps := flag.Int("rps", 30, "target aggregate requests per second")
	email := flag.String("email", "", "optional login for authenticated scenarios")
	password := flag.String("password", "", "optional login password")
	adminKey := flag.String("admin-key", "", "optional X-Admin-Key for /servers/all")
	flag.Parse()

	baseURL := strings.TrimRight(*base, "/")
	client := &http.Client{Timeout: 10 * time.Second}

	token := ""
	if *email != "" && *password != "" {
		var err error
		token, err = login(client, baseURL, *email, *password)
		if err != nil {
			fmt.Fprintf(os.Stderr, "login failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("authenticated: yes")
	}

	scenarios := []scenario{
		{name: "GET /health", method: http.MethodGet, path: "/health"},
		{name: "GET /api/v1/rules", method: http.MethodGet, path: "/api/v1/rules"},
		{name: "GET /api/v1/config", method: http.MethodGet, path: "/api/v1/config"},
		{name: "GET /api/v1/regions", method: http.MethodGet, path: "/api/v1/regions"},
	}
	if token != "" {
		scenarios = append(scenarios, scenario{
			name: "GET /api/v1/servers", method: http.MethodGet, path: "/api/v1/servers", auth: true,
		})
		scenarios = append(scenarios, scenario{
			name: "GET /api/v1/subscription", method: http.MethodGet, path: "/api/v1/subscription", auth: true,
		})
	}
	if *adminKey != "" {
		scenarios = append(scenarios, scenario{
			name: "GET /api/v1/servers/all", method: http.MethodGet, path: "/api/v1/servers/all",
		})
	}

	fmt.Printf("base=%s duration=%s target_rps=%d scenarios=%d\n", baseURL, duration, *rps, len(scenarios))

	ctx, cancel := context.WithTimeout(context.Background(), *duration)
	defer cancel()

	var (
		mu      sync.Mutex
		results = map[string][]sample{}
		total   atomic.Int64
		errors  atomic.Int64
	)

	interval := time.Second / time.Duration(*rps)
	if interval <= 0 {
		interval = time.Millisecond
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	var wg sync.WaitGroup
	i := 0
	for {
		select {
		case <-ctx.Done():
			wg.Wait()
			printReport(results, total.Load(), errors.Load(), *duration)
			if errors.Load() > total.Load()/10 {
				os.Exit(2)
			}
			return
		case <-ticker.C:
			sc := scenarios[i%len(scenarios)]
			i++
			wg.Add(1)
			go func(sc scenario) {
				defer wg.Done()
				s := runOnce(client, baseURL, sc, token, *adminKey)
				total.Add(1)
				if s.err || s.status >= 500 || s.status == 0 {
					errors.Add(1)
				}
				mu.Lock()
				results[sc.name] = append(results[sc.name], s)
				mu.Unlock()
			}(sc)
		}
	}
}

func login(client *http.Client, base, email, password string) (string, error) {
	body, _ := json.Marshal(map[string]string{"email": email, "password": password})
	req, err := http.NewRequest(http.MethodPost, base+"/api/v1/login", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	if res.StatusCode != 200 {
		return "", fmt.Errorf("status %d: %s", res.StatusCode, truncate(string(raw), 200))
	}
	var parsed struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", err
	}
	if parsed.AccessToken == "" {
		return "", fmt.Errorf("empty access_token")
	}
	return parsed.AccessToken, nil
}

func runOnce(client *http.Client, base string, sc scenario, token, adminKey string) sample {
	start := time.Now()
	req, err := http.NewRequest(sc.method, base+sc.path, bytes.NewReader(sc.body))
	if err != nil {
		return sample{latency: time.Since(start), err: true}
	}
	if len(sc.body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}
	if sc.auth && token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if strings.Contains(sc.path, "/servers/all") && adminKey != "" {
		req.Header.Set("X-Admin-Key", adminKey)
	}
	res, err := client.Do(req)
	lat := time.Since(start)
	if err != nil {
		return sample{latency: lat, err: true}
	}
	defer res.Body.Close()
	_, _ = io.Copy(io.Discard, res.Body)
	return sample{latency: lat, status: res.StatusCode, err: false}
}

func printReport(results map[string][]sample, total, errs int64, dur time.Duration) {
	fmt.Println()
	fmt.Println("=== Load test report (BL-032) ===")
	fmt.Printf("total=%d errors=%d effective_rps=%.1f\n", total, errs, float64(total)/dur.Seconds())

	names := make([]string, 0, len(results))
	for name := range results {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		samples := results[name]
		if len(samples) == 0 {
			continue
		}
		lats := make([]time.Duration, 0, len(samples))
		ok := 0
		statusCounts := map[int]int{}
		for _, s := range samples {
			lats = append(lats, s.latency)
			if !s.err && s.status > 0 && s.status < 500 {
				ok++
			}
			statusCounts[s.status]++
		}
		sort.Slice(lats, func(i, j int) bool { return lats[i] < lats[j] })
		p50 := lats[percentileIndex(len(lats), 50)]
		p95 := lats[percentileIndex(len(lats), 95)]
		p99 := lats[percentileIndex(len(lats), 99)]
		fmt.Printf("%-28s n=%-5d ok=%-5d p50=%-8s p95=%-8s p99=%-8s statuses=%v\n",
			name, len(samples), ok, p50.Round(time.Millisecond), p95.Round(time.Millisecond), p99.Round(time.Millisecond), statusCounts)
	}
}

func percentileIndex(n, p int) int {
	if n == 0 {
		return 0
	}
	idx := (p * n / 100) - 1
	if idx < 0 {
		return 0
	}
	if idx >= n {
		return n - 1
	}
	return idx
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
