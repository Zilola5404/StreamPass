package hyconfig

import (
	"fmt"
	"net"
	"strconv"
)

// FallbackUDPPorts is the UDP dial order from ТЗ §10 after the primary port.
// TCP 443/8443 are listed in the TZ but hysteria core v2 is QUIC/UDP-only;
// those are skipped until a TCP underlay exists.
var FallbackUDPPorts = []int{8443, 24443}

// DialCandidate is one UDP endpoint to try for the relay handshake.
type DialCandidate struct {
	Host string // host:port
	Port int
}

// FallbackCandidates builds the UDP endpoint list: primary port first, then
// ТЗ §10 alternates (deduped).
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

	seen := map[int]struct{}{}
	var out []DialCandidate
	add := func(port int) {
		if port <= 0 {
			return
		}
		if _, ok := seen[port]; ok {
			return
		}
		seen[port] = struct{}{}
		out = append(out, DialCandidate{
			Host: net.JoinHostPort(hostOnly, strconv.Itoa(port)),
			Port: port,
		})
	}
	add(primaryPort)
	for _, p := range FallbackUDPPorts {
		add(p)
	}
	return out
}

// FormatCandidate returns a short label for logs.
func (c DialCandidate) String() string {
	return fmt.Sprintf("udp/%d", c.Port)
}
