package dnscache

import (
	"net"
	"strings"
	"sync"

	"golang.org/x/net/dns/dnsmessage"
)

// Reverse IP→hostname map filled from successful DNS answers so TUN flows
// that only carry a destination IP can still show a readable site name in
// operator diagnostics (hostname only — no URL paths).
var (
	revMu   sync.RWMutex
	revByIP = map[string]string{}
	rttMu   sync.RWMutex
	rttByHost = map[string]int64{} // hostname → last resolve RTT ms
)

const maxReverseEntries = 2048

// RememberResolveMS stores the last DNS resolve latency for a hostname.
func RememberResolveMS(host string, ms int64) {
	host = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(host), "."))
	if host == "" || ms < 0 {
		return
	}
	rttMu.Lock()
	defer rttMu.Unlock()
	if len(rttByHost) >= maxReverseEntries {
		for k := range rttByHost {
			delete(rttByHost, k)
			break
		}
	}
	rttByHost[host] = ms
}

// LastResolveMS returns the last DNS RTT for host, or 0 if unknown.
func LastResolveMS(host string) int64 {
	host = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(host), "."))
	rttMu.RLock()
	defer rttMu.RUnlock()
	return rttByHost[host]
}

// RememberIP links an IPv4/IPv6 address to a hostname (FQDN without trailing dot).
func RememberIP(host, ip string) {
	host = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(host), "."))
	ip = strings.TrimSpace(ip)
	if host == "" || ip == "" {
		return
	}
	if parsed := net.ParseIP(ip); parsed != nil {
		ip = parsed.String()
	}
	revMu.Lock()
	defer revMu.Unlock()
	if len(revByIP) >= maxReverseEntries {
		// Drop an arbitrary entry to bound memory.
		for k := range revByIP {
			delete(revByIP, k)
			break
		}
	}
	revByIP[ip] = host
}

// HostForIP returns the last hostname that resolved to ip, if known.
func HostForIP(ip string) string {
	ip = strings.TrimSpace(ip)
	if parsed := net.ParseIP(ip); parsed != nil {
		ip = parsed.String()
	}
	revMu.RLock()
	defer revMu.RUnlock()
	return revByIP[ip]
}

// IndexAnswers extracts A/AAAA addresses from a DNS response wire packet
// and remembers IP→name for later diag enrichment.
func IndexAnswers(name string, raw []byte) {
	name = trimDot(name)
	if name == "" || len(raw) == 0 {
		return
	}
	var parser dnsmessage.Parser
	if _, err := parser.Start(raw); err != nil {
		return
	}
	if _, err := parser.Question(); err != nil {
		return
	}
	for {
		ah, err := parser.AnswerHeader()
		if err == dnsmessage.ErrSectionDone {
			break
		}
		if err != nil {
			break
		}
		switch ah.Type {
		case dnsmessage.TypeA:
			r, err := parser.AResource()
			if err != nil {
				_ = parser.SkipAnswer()
				continue
			}
			RememberIP(name, net.IP(r.A[:]).String())
		case dnsmessage.TypeAAAA:
			r, err := parser.AAAAResource()
			if err != nil {
				_ = parser.SkipAnswer()
				continue
			}
			RememberIP(name, net.IP(r.AAAA[:]).String())
		default:
			_ = parser.SkipAnswer()
		}
	}
}
