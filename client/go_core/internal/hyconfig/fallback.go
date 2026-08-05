package hyconfig

import (
	"fmt"
	"net"
	"strconv"
)

// FallbackUDPPorts is the UDP dial order from ТЗ §10 after the primary port.
var FallbackUDPPorts = []int{8443, 24443}

// FallbackTCPPorts is tried after all UDP candidates fail (ТЗ §10).
// TCP/443 is skipped on the StreamPass VPS because Caddy owns it; the
// underlay bridge listens on TCP/8443 and TCP/24443 and forwards framed
// datagrams to local Hysteria UDP/443.
var FallbackTCPPorts = []int{8443, 24443}

// DialCandidate is one endpoint to try for the relay handshake.
type DialCandidate struct {
	Host    string // host:port
	Port    int
	Network string // "udp" or "tcp"
}

// FallbackCandidates builds the dial list: primary UDP, then UDP alternates,
// then TCP underlay ports (ТЗ §10), deduped per network+port.
func FallbackCandidates(host string, primaryPort int) []DialCandidate {
	hostOnly := host
	if h, p, err := net.SplitHostPort(host); err == nil {
		hostOnly = h
		if primaryPort <= 0 {
			if parsed, err := strconv.Atoi(p); err == nil {
				primaryPort = parsed
			}
		}
	}
	if primaryPort <= 0 {
		primaryPort = 443
	}

	seen := map[string]struct{}{}
	var out []DialCandidate
	add := func(network string, port int) {
		if port <= 0 {
			return
		}
		key := network + "/" + strconv.Itoa(port)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		out = append(out, DialCandidate{
			Host:    net.JoinHostPort(hostOnly, strconv.Itoa(port)),
			Port:    port,
			Network: network,
		})
	}
	add("udp", primaryPort)
	for _, p := range FallbackUDPPorts {
		add("udp", p)
	}
	for _, p := range FallbackTCPPorts {
		add("tcp", p)
	}
	return out
}

// FormatCandidate returns a short label for logs.
func (c DialCandidate) String() string {
	if c.Network == "" {
		return fmt.Sprintf("udp/%d", c.Port)
	}
	return fmt.Sprintf("%s/%d", c.Network, c.Port)
}
