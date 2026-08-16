package hyconfig

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/apernet/hysteria/core/v2/client"
)

// Handshake budgets keep Windows RPC start() under Flutter's 45s timeout.
const (
	handshakePerCandidate = 8 * time.Second
	handshakeOverall      = 32 * time.Second
)

// ConnectResult is a successful Hysteria handshake after optional port fallback.
type ConnectResult struct {
	Client    client.Client
	Parsed    *Parsed
	PingMs    int
	Candidate DialCandidate
}

// ConnectWithFallback dials the relay trying UDP then TCP underlay (ТЗ §10).
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

	deadline := time.Now().Add(handshakeOverall)
	var failures []string
	for _, c := range FallbackCandidates(hostOnly, primary) {
		remain := time.Until(deadline)
		if remain <= 0 {
			failures = append(failures, "overall handshake budget exhausted")
			break
		}
		budget := handshakePerCandidate
		if remain < budget {
			budget = remain
		}
		result, err := dialCandidate(baseCfg, parsed, hostOnly, c, budget)
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", c, err))
			continue
		}
		return result, nil
	}

	if len(failures) == 0 {
		return nil, fmt.Errorf("no fallback candidates")
	}
	return nil, fmt.Errorf("all fallback endpoints failed: %s", strings.Join(failures, "; "))
}

func dialCandidate(baseCfg *client.Config, parsed *Parsed, hostOnly string, c DialCandidate, budget time.Duration) (*ConnectResult, error) {
	// QUIC peer identity for TCP underlay is always the main Hysteria UDP/443:
	// the VPS bridge listens on TCP/8443+/24443 and forwards to 127.0.0.1:443.
	quicPort := c.Port
	if c.Network == "tcp" {
		quicPort = 443
	}
	udpRemote, err := net.ResolveUDPAddr("udp", net.JoinHostPort(hostOnly, strconv.Itoa(quicPort)))
	if err != nil {
		return nil, fmt.Errorf("resolve: %w", err)
	}

	cfg := cloneClientConfig(baseCfg)
	cfg.ServerAddr = udpRemote
	// Keep idle timeout within hysteria bounds but short enough for fallback.
	idle := budget
	if idle < 4*time.Second {
		idle = 4 * time.Second
	}
	if idle > 30*time.Second {
		idle = 30 * time.Second
	}
	cfg.QUICConfig.MaxIdleTimeout = idle
	cfg.QUICConfig.KeepAlivePeriod = 2 * time.Second

	switch c.Network {
	case "tcp":
		cfg.ConnFactory = &tcpUnderlayConnFactory{
			tcpAddr:   c.Host,
			udpRemote: udpRemote,
			obfsType:  parsed.ObfsType,
			obfsPass:  parsed.ObfsPass,
		}
	default:
		cfg.ConnFactory = &protectedConnFactory{
			obfsType: parsed.ObfsType,
			obfsPass: parsed.ObfsPass,
		}
		if addr, err := net.ResolveUDPAddr("udp", c.Host); err == nil {
			cfg.ServerAddr = addr
		}
	}

	type dialOut struct {
		hy  client.Client
		err error
	}
	ch := make(chan dialOut, 1)
	start := time.Now()
	go func() {
		hy, _, err := client.NewClient(cfg)
		ch <- dialOut{hy: hy, err: err}
	}()

	timer := time.NewTimer(budget)
	defer timer.Stop()
	select {
	case out := <-ch:
		if out.err != nil {
			return nil, out.err
		}
		copied := *parsed
		copied.ServerHost = c.Host
		return &ConnectResult{
			Client:    out.hy,
			Parsed:    &copied,
			PingMs:    int(time.Since(start).Milliseconds()),
			Candidate: c,
		}, nil
	case <-timer.C:
		// Late success must not leak a live client.
		go func() {
			if out := <-ch; out.hy != nil {
				_ = out.hy.Close()
			}
		}()
		return nil, fmt.Errorf("handshake timeout after %s", budget)
	}
}

func cloneClientConfig(in *client.Config) *client.Config {
	if in == nil {
		return &client.Config{}
	}
	out := *in
	return &out
}
