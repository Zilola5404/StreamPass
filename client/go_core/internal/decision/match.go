package decision

import (
	"net/netip"
	"strings"
)

// Target describes one connection the engine evaluates.
type Target struct {
	// Host is a domain name when known (empty for raw IP flows).
	Host string
	// IP is the destination address when available.
	IP netip.Addr
}

// Engine applies ordered rules and user exclusions to pick DIRECT / RELAY / FALLBACK.
type Engine struct {
	rules       []Rule
	exclusions  []Rule
	defaultMode Mode
}

// NewEngine builds an engine. User exclusions are checked before published rules.
func NewEngine(rules []Rule, exclusions []Rule, defaultMode Mode) *Engine {
	if defaultMode == "" {
		defaultMode = DefaultMode
	}
	cp := func(in []Rule) []Rule {
		if len(in) == 0 {
			return nil
		}
		out := make([]Rule, len(in))
		copy(out, in)
		return out
	}
	return &Engine{
		rules:       cp(rules),
		exclusions:  cp(exclusions),
		defaultMode: defaultMode,
	}
}

// Decide returns the routing mode for a connection target.
func (e *Engine) Decide(t Target) Mode {
	host := normalizeHost(t.Host)

	if mode, ok := e.matchRules(e.exclusions, host, t.IP); ok {
		return normalizeMode(mode)
	}
	if mode, ok := e.matchRules(e.rules, host, t.IP); ok {
		return normalizeMode(mode)
	}
	return e.defaultMode
}

func (e *Engine) matchRules(rules []Rule, host string, ip netip.Addr) (Mode, bool) {
	for _, r := range rules {
		switch r.Kind {
		case KindDomain:
			if host != "" && domainMatches(r.Pattern, host) {
				return r.Mode, true
			}
		case KindCIDR:
			if ip.IsValid() && cidrContains(r.Pattern, ip) {
				return r.Mode, true
			}
		}
	}
	return "", false
}

func normalizeHost(host string) string {
	host = strings.TrimSpace(strings.ToLower(host))
	host = strings.TrimSuffix(host, ".")
	return host
}

func normalizeMode(m Mode) Mode {
	switch m {
	case ModeDirect, ModeRelay, ModeFallback:
		return m
	default:
		return DefaultMode
	}
}

// domainMatches supports patterns like "*.ru", "youtube.com", "_bank".
func domainMatches(pattern, host string) bool {
	pattern = normalizeHost(pattern)
	if pattern == "" || host == "" {
		return false
	}
	if strings.HasPrefix(pattern, "*.") {
		suffix := pattern[1:] // ".ru"
		if host == suffix[1:] {
			return true
		}
		return strings.HasSuffix(host, suffix)
	}
	if host == pattern {
		return true
	}
	return strings.HasSuffix(host, "."+pattern)
}

func cidrContains(pattern string, ip netip.Addr) bool {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" || !ip.IsValid() {
		return false
	}
	if !strings.Contains(pattern, "/") {
		addr, err := netip.ParseAddr(pattern)
		if err != nil {
			return false
		}
		return addr == ip
	}
	prefix, err := netip.ParsePrefix(pattern)
	if err != nil {
		return false
	}
	return prefix.Contains(ip)
}
