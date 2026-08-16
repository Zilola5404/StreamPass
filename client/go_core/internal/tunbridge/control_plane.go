package tunbridge

import (
	"net"
	"net/netip"
	"strings"
)

// isControlPlaneDest reports destinations that must never go via Hysteria RELAY:
// the API host (*.nip.io), the active relay IP/hostname, and loopback/private.
// Hairpin RELAY→same VPS blackholes control-plane TLS and breaks rules sync.
func isControlPlaneDest(host, destIP, relayID string) bool {
	h := strings.ToLower(strings.TrimSuffix(strings.TrimSpace(host), "."))
	ip := strings.TrimSpace(destIP)
	rid := strings.ToLower(strings.TrimSpace(relayID))

	if h == "localhost" || strings.HasSuffix(h, ".localhost") {
		return true
	}
	if ip != "" {
		if parsed, err := netip.ParseAddr(ip); err == nil {
			if parsed.IsLoopback() || parsed.IsPrivate() || parsed.IsLinkLocalUnicast() {
				return true
			}
		}
		if ip == "127.0.0.1" || ip == "::1" {
			return true
		}
	}

	// Staging/prod API on nip.io (same host as many relays).
	if h != "" && (h == "nip.io" || strings.HasSuffix(h, ".nip.io")) {
		return true
	}

	if rid == "" {
		return false
	}

	// relayID may be "212.43.156.33", "pl-warsaw-1", or a hostname.
	if ip != "" && (ip == rid || ip == stripHostPort(rid)) {
		return true
	}
	if h != "" && (h == rid || h == stripHostPort(rid)) {
		return true
	}
	if hostLooksLikeIP(rid) && ip == rid {
		return true
	}
	return false
}

func controlPlaneReason(host, destIP, relayID string) string {
	h := strings.ToLower(strings.TrimSuffix(strings.TrimSpace(host), "."))
	ip := strings.TrimSpace(destIP)
	if h == "localhost" || strings.HasSuffix(h, ".localhost") {
		return "private_network_bypass"
	}
	if ip != "" {
		if parsed, err := netip.ParseAddr(ip); err == nil {
			if parsed.IsLoopback() || parsed.IsPrivate() || parsed.IsLinkLocalUnicast() {
				return "private_network_bypass"
			}
		}
	}
	_ = relayID
	return "relay_endpoint_bypass"
}

func stripHostPort(s string) string {
	s = strings.TrimSpace(s)
	if h, _, err := net.SplitHostPort(s); err == nil {
		return h
	}
	return s
}

func hostLooksLikeIP(s string) bool {
	_, err := netip.ParseAddr(stripHostPort(s))
	return err == nil
}
