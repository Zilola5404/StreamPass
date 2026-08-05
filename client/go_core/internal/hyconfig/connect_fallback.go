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

	var failures []string
	for _, c := range FallbackCandidates(hostOnly, primary) {
		result, err := dialCandidate(baseCfg, parsed, hostOnly, c)
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

func dialCandidate(baseCfg *client.Config, parsed *Parsed, hostOnly string, c DialCandidate) (*ConnectResult, error) {
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

	start := time.Now()
	hy, _, err := client.NewClient(cfg)
	if err != nil {
		return nil, err
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

func cloneClientConfig(in *client.Config) *client.Config {
	if in == nil {
		return &client.Config{}
	}
	out := *in
	return &out
}
