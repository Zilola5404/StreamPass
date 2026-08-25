package dnscache

import (
	"fmt"
	"net"
	"strings"
	"sync"

	"golang.org/x/net/dns/dnsmessage"
)

// Reverse IP→hostname map filled from successful DNS answers so TUN flows
// that only carry a destination IP can still show a readable site name in
// operator diagnostics (hostname only — no URL paths).
//
// Cloudflare/Akamai anycast: many hostnames share one IP. We keep a *set* of
// hostnames per IP so a later DIRECT resolve cannot erase a prior RELAY host.
var (
	revMu       sync.RWMutex
	revByIP     = map[string]map[string]struct{}{}
	lastByIP    = map[string]string{}
	relayPinMu  sync.RWMutex
	relayByIP   = map[string]string{}
	directPinMu sync.RWMutex
	directByIP  = map[string]string{}
	rttMu       sync.RWMutex
	rttByHost   = map[string]int64{}
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
		for k := range revByIP {
			delete(revByIP, k)
			delete(lastByIP, k)
			break
		}
	}
	set := revByIP[ip]
	if set == nil {
		set = map[string]struct{}{}
		revByIP[ip] = set
	}
	set[host] = struct{}{}
	lastByIP[ip] = host
}

// HostForIP returns the most recent hostname that resolved to ip, if known.
func HostForIP(ip string) string {
	ip = strings.TrimSpace(ip)
	if parsed := net.ParseIP(ip); parsed != nil {
		ip = parsed.String()
	}
	revMu.RLock()
	defer revMu.RUnlock()
	if h := lastByIP[ip]; h != "" {
		return h
	}
	if set := revByIP[ip]; len(set) > 0 {
		for h := range set {
			return h
		}
	}
	return ""
}

// HostsForIP returns every hostname remembered for ip (anycast-safe).
func HostsForIP(ip string) []string {
	ip = strings.TrimSpace(ip)
	if parsed := net.ParseIP(ip); parsed != nil {
		ip = parsed.String()
	}
	revMu.RLock()
	defer revMu.RUnlock()
	set := revByIP[ip]
	if len(set) == 0 {
		return nil
	}
	out := make([]string, 0, len(set))
	for h := range set {
		out = append(out, h)
	}
	return out
}

// PinRelayIP marks ip as belonging to a RELAY hostname (DNS-time decision).
func PinRelayIP(host, ip string) {
	host = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(host), "."))
	ip = strings.TrimSpace(ip)
	if host == "" || ip == "" {
		return
	}
	if parsed := net.ParseIP(ip); parsed != nil {
		ip = parsed.String()
	}
	RememberIP(host, ip)
	relayPinMu.Lock()
	defer relayPinMu.Unlock()
	if len(relayByIP) >= maxReverseEntries {
		for k := range relayByIP {
			delete(relayByIP, k)
			break
		}
	}
	relayByIP[ip] = host
}

// PinDirectIP marks ip as belonging to a DIRECT hostname (*.ru, NetworkMonitor).
func PinDirectIP(host, ip string) {
	host = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(host), "."))
	ip = strings.TrimSpace(ip)
	if host == "" || ip == "" {
		return
	}
	if parsed := net.ParseIP(ip); parsed != nil {
		ip = parsed.String()
	}
	RememberIP(host, ip)
	directPinMu.Lock()
	defer directPinMu.Unlock()
	if len(directByIP) >= maxReverseEntries {
		for k := range directByIP {
			delete(directByIP, k)
			break
		}
	}
	directByIP[ip] = host
}

// DirectHostForIP returns a hostname pinned as DIRECT for this IP, if any.
func DirectHostForIP(ip string) string {
	ip = strings.TrimSpace(ip)
	if parsed := net.ParseIP(ip); parsed != nil {
		ip = parsed.String()
	}
	directPinMu.RLock()
	defer directPinMu.RUnlock()
	return directByIP[ip]
}

// DirectHostNearIP returns a DIRECT pin for ip or a sibling in the same IPv4 /24.
func DirectHostNearIP(ip string) string {
	if h := DirectHostForIP(ip); h != "" {
		return h
	}
	parsed := net.ParseIP(strings.TrimSpace(ip))
	if parsed == nil {
		return ""
	}
	v4 := parsed.To4()
	if v4 == nil {
		return ""
	}
	prefix := fmt.Sprintf("%d.%d.%d.", v4[0], v4[1], v4[2])
	directPinMu.RLock()
	defer directPinMu.RUnlock()
	for pip, host := range directByIP {
		if strings.HasPrefix(pip, prefix) {
			return host
		}
	}
	return ""
}

// RelayHostForIP returns a hostname pinned as RELAY for this IP, if any.
func RelayHostForIP(ip string) string {
	ip = strings.TrimSpace(ip)
	if parsed := net.ParseIP(ip); parsed != nil {
		ip = parsed.String()
	}
	relayPinMu.RLock()
	defer relayPinMu.RUnlock()
	return relayByIP[ip]
}

// RelayHostNearIP returns a RELAY pin for ip or a sibling in the same IPv4 /24.
func RelayHostNearIP(ip string) string {
	if h := RelayHostForIP(ip); h != "" {
		return h
	}
	parsed := net.ParseIP(strings.TrimSpace(ip))
	if parsed == nil {
		return ""
	}
	v4 := parsed.To4()
	if v4 == nil {
		return ""
	}
	prefix := fmt.Sprintf("%d.%d.%d.", v4[0], v4[1], v4[2])
	relayPinMu.RLock()
	defer relayPinMu.RUnlock()
	for pip, host := range relayByIP {
		if strings.HasPrefix(pip, prefix) {
			return host
		}
	}
	return ""
}

// ExtractAIPs returns A-record addresses from Answer + Additional sections
// (CNAME targets are often placed in Additional by Yandex/Cloudflare).
func ExtractAIPs(raw []byte) []string {
	if len(raw) == 0 {
		return nil
	}
	var msg dnsmessage.Message
	if err := msg.Unpack(raw); err != nil {
		return extractAIPsParser(raw)
	}
	seen := map[string]struct{}{}
	var out []string
	add := func(ip string) {
		if ip == "" {
			return
		}
		if _, ok := seen[ip]; ok {
			return
		}
		seen[ip] = struct{}{}
		out = append(out, ip)
	}
	collect := func(rrs []dnsmessage.Resource) {
		for _, rr := range rrs {
			if ar, ok := rr.Body.(*dnsmessage.AResource); ok {
				add(net.IP(ar.A[:]).String())
			}
		}
	}
	collect(msg.Answers)
	collect(msg.Additionals)
	return out
}

// extractAIPsParser is a fallback using the streaming Parser API.
func extractAIPsParser(raw []byte) []string {
	var parser dnsmessage.Parser
	if _, err := parser.Start(raw); err != nil {
		return nil
	}
	for {
		_, err := parser.Question()
		if err == dnsmessage.ErrSectionDone {
			break
		}
		if err != nil {
			return nil
		}
	}
	seen := map[string]struct{}{}
	var out []string
	add := func(ip string) {
		if ip == "" {
			return
		}
		if _, ok := seen[ip]; ok {
			return
		}
		seen[ip] = struct{}{}
		out = append(out, ip)
	}
	for {
		ah, err := parser.AnswerHeader()
		if err == dnsmessage.ErrSectionDone {
			break
		}
		if err != nil {
			return out
		}
		switch ah.Type {
		case dnsmessage.TypeA:
			r, err := parser.AResource()
			if err != nil {
				return out
			}
			add(net.IP(r.A[:]).String())
		case dnsmessage.TypeCNAME:
			if _, err := parser.CNAMEResource(); err != nil {
				return out
			}
		case dnsmessage.TypeAAAA:
			if _, err := parser.AAAAResource(); err != nil {
				_ = parser.SkipAnswer()
			}
		default:
			_ = parser.SkipAnswer()
		}
	}
	_ = parser.SkipAllAuthorities()
	for {
		ah, err := parser.AdditionalHeader()
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
				break
			}
			add(net.IP(r.A[:]).String())
		default:
			_ = parser.SkipAdditional()
		}
	}
	return out
}

// IndexAnswers extracts A addresses from a DNS response and remembers IP→name.
func IndexAnswers(name string, raw []byte) {
	name = trimDot(name)
	if name == "" || len(raw) == 0 {
		return
	}
	for _, ip := range ExtractAIPs(raw) {
		RememberIP(name, ip)
	}
}

// ClearSessionMaps drops reverse/pin/RTT maps after idle or network change (Issue #2).
func ClearSessionMaps() {
	revMu.Lock()
	revByIP = map[string]map[string]struct{}{}
	lastByIP = map[string]string{}
	revMu.Unlock()

	relayPinMu.Lock()
	relayByIP = map[string]string{}
	relayPinMu.Unlock()

	directPinMu.Lock()
	directByIP = map[string]string{}
	directPinMu.Unlock()

	rttMu.Lock()
	rttByHost = map[string]int64{}
	rttMu.Unlock()
}

// InvalidateAfterIdle clears DNS cache + hostname associations so post-idle
// traffic re-resolves and re-pins (Issue #2 §12).
func InvalidateAfterIdle() {
	Default().cache.Clear()
	ClearSessionMaps()
	logLine(nil, "[dns] invalidated after idle/network change")
}
