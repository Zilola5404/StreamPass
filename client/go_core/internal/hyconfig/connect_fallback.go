package hyconfig

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/apernet/hysteria/core/v2/client"
)

// ConnectResult is a successful Hysteria handshake after optional port fallback.
type ConnectResult struct {
	Client    client.Client
	Parsed    *Parsed
	PingMs    int
	Candidate DialCandidate
}

// ConnectWithFallback dials the relay trying UDP ports in ТЗ §10 order.
func ConnectWithFallback(connectionConfig, relayHost string, relayPort int) (*ConnectResult, error) {
	baseCfg, parsed, err := BuildClientConfig(connectionConfig, relayHost, relayPort)
	if err != nil {
		return nil, err
	}

	hostOnly := relayHost
	if hostOnly == "" {
		hostOnly, _, _ = net.SplitHostPort(parsed.ServerHost)
	} else if h, _, err := net.SplitHostPort(hostOnly); err == nil {
		hostOnly = h
	}
	if hostOnly == "" {
		if h, _, err := net.SplitHostPort(parsed.ServerHost); err == nil {
			hostOnly = h
		} else {
			hostOnly = parsed.ServerHost
		}
	}

	primary := relayPort
	if primary <= 0 {
		if _, p, err := net.SplitHostPort(parsed.ServerHost); err == nil {
			primary, _ = strconv.Atoi(p)
		}
	}

	var failures []string
	for _, c := range FallbackCandidates(hostOnly, primary) {
		addr, err := net.ResolveUDPAddr("udp", c.Host)
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s: resolve: %v", c, err))
			continue
		}
		cfg := cloneClientConfig(baseCfg)
		cfg.ServerAddr = addr
		// Fresh ConnFactory so each attempt gets its own protected UDP socket.
		cfg.ConnFactory = &protectedConnFactory{
			obfsType: parsed.ObfsType,
			obfsPass: parsed.ObfsPass,
		}

		start := time.Now()
		hy, _, err := client.NewClient(cfg)
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", c, err))
			continue
		}

		out := *parsed
		out.ServerHost = c.Host
		return &ConnectResult{
			Client:    hy,
			Parsed:    &out,
			PingMs:    int(time.Since(start).Milliseconds()),
			Candidate: c,
		}, nil
	}

	if len(failures) == 0 {
		return nil, fmt.Errorf("no fallback candidates")
	}
	return nil, fmt.Errorf("all UDP fallback endpoints failed: %s", strings.Join(failures, "; "))
}

func cloneClientConfig(in *client.Config) *client.Config {
	if in == nil {
		return &client.Config{}
	}
	out := *in
	return &out
}
