// Command healthmonitor is a small standalone worker: on a timer, it asks
// the StreamPass backend for every registered relay server (GET
// /servers/all, including currently-unhealthy ones so recoveries are
// detected), TCP-probes each one directly, and reports the result back via
// POST /servers/health.
//
// It deliberately does not speak the relay's actual VPN protocol
// (Reality/Hysteria2/etc) — a successful TCP connect within the timeout is
// treated as "reachable", which is the right level of check for this
// worker: it answers "is the network path to this relay alive", not "is
// the VPN handshake correct". Protocol-level correctness is caught faster
// by users' own connection attempts than by a generic prober, and keeping
// this worker protocol-agnostic means it doesn't need to change every time
// a new relay protocol is added (KISS/YAGNI).
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type serverDTO struct {
	ID   string `json:"id"`
	Host string `json:"host"`
	Port int    `json:"port"`
}

type healthCheckRequest struct {
	ID        string  `json:"id"`
	Healthy   bool    `json:"healthy"`
	LoadRatio float64 `json:"load_ratio"`
	RTTMillis int     `json:"rtt_ms"`
}

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With(slog.String("module", "healthmonitor"))

	cfg, err := loadConfig()
	if err != nil {
		log.Error("invalid configuration", slog.String("error", err.Error()))
		os.Exit(1)
	}

	client := &http.Client{Timeout: cfg.checkTimeout + 5*time.Second}

	log.Info("health monitor starting",
		slog.String("backend_url", cfg.backendURL),
		slog.Duration("check_interval", cfg.checkInterval),
		slog.Duration("check_timeout", cfg.checkTimeout),
	)

	ticker := time.NewTicker(cfg.checkInterval)
	defer ticker.Stop()

	runOnce(client, cfg, log)
	for range ticker.C {
		runOnce(client, cfg, log)
	}
}

type config struct {
	backendURL    string
	adminKey      string
	checkInterval time.Duration
	checkTimeout  time.Duration
	// udpOnlyPorts are Hysteria listeners without TCP; a failed TCP probe
	// must not flip them to unhealthy (region DE/PL/FI ports on shared VPS).
	udpOnlyPorts map[int]bool
}

func loadConfig() (config, error) {
	backendURL := os.Getenv("BACKEND_URL")
	if backendURL == "" {
		return config{}, fmt.Errorf("BACKEND_URL must be set")
	}
	adminKey := os.Getenv("ADMIN_API_KEY")
	if adminKey == "" {
		return config{}, fmt.Errorf("ADMIN_API_KEY must be set")
	}

	interval := envDurationOr("CHECK_INTERVAL", 60*time.Second)
	timeout := envDurationOr("CHECK_TIMEOUT", 5*time.Second)

	return config{
		backendURL:    backendURL,
		adminKey:      adminKey,
		checkInterval: interval,
		checkTimeout:  timeout,
		udpOnlyPorts:  parseUDPOnlyPorts(os.Getenv("HEALTH_UDP_ONLY_PORTS")),
	}, nil
}

func parseUDPOnlyPorts(raw string) map[int]bool {
	out := map[int]bool{}
	if strings.TrimSpace(raw) == "" {
		// Defaults for StreamPass region listeners on the API VPS.
		raw = "8443,24443,34443"
	}
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		n, err := strconv.Atoi(part)
		if err != nil {
			continue
		}
		out[n] = true
	}
	return out
}

func envDurationOr(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}

func runOnce(client *http.Client, cfg config, log *slog.Logger) {
	servers, err := fetchServers(client, cfg)
	if err != nil {
		log.Error("failed to fetch server list", slog.String("error", err.Error()))
		return
	}

	log.Info("checking servers", slog.Int("count", len(servers)))

	for _, srv := range servers {
		healthy, rttMillis := probeTCP(srv.Host, srv.Port, cfg.checkTimeout)
		if !healthy && cfg.udpOnlyPorts[srv.Port] {
			log.Info("skip unhealthy flip for udp-only port",
				slog.String("server_id", srv.ID),
				slog.Int("port", srv.Port),
			)
			continue
		}

		if err := reportHealth(client, cfg, srv.ID, healthy, rttMillis); err != nil {
			log.Error("failed to report health",
				slog.String("server_id", srv.ID), slog.String("error", err.Error()))
			continue
		}

		log.Info("checked server",
			slog.String("server_id", srv.ID),
			slog.Bool("healthy", healthy),
			slog.Int("rtt_ms", rttMillis),
		)
	}
}

func fetchServers(client *http.Client, cfg config) ([]serverDTO, error) {
	req, err := http.NewRequest(http.MethodGet, cfg.backendURL+"/api/v1/servers/all", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Admin-Key", cfg.adminKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from GET /servers/all", resp.StatusCode)
	}

	var servers []serverDTO
	if err := json.NewDecoder(resp.Body).Decode(&servers); err != nil {
		return nil, fmt.Errorf("failed to decode server list: %w", err)
	}
	return servers, nil
}

func probeTCP(host string, port int, timeout time.Duration) (healthy bool, rttMillis int) {
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	start := time.Now()

	conn, err := net.DialTimeout("tcp", addr, timeout)
	elapsed := time.Since(start)

	if err != nil {
		return false, int(elapsed.Milliseconds())
	}
	conn.Close()
	return true, int(elapsed.Milliseconds())
}

func reportHealth(client *http.Client, cfg config, id string, healthy bool, rttMillis int) error {
	body, err := json.Marshal(healthCheckRequest{
		ID:        id,
		Healthy:   healthy,
		LoadRatio: 0,
		RTTMillis: rttMillis,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, cfg.backendURL+"/api/v1/servers/health", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Admin-Key", cfg.adminKey)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status %d from POST /servers/health", resp.StatusCode)
	}
	return nil
}
