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
	return e.DecideDetailed(t).Mode
}

// Decision is a routing choice plus why it was made (TASK-01 diagnostics).
type Decision struct {
	Mode   Mode
	Rule   string // matched pattern (or "default")
	Source string // exclusion | rule | default
	Reason string // stable machine reason code
}

// DecideDetailed returns mode and the matching rule/reason for diagnostics.
func (e *Engine) DecideDetailed(t Target) Decision {
	host := normalizeHost(t.Host)

	if mode, rule, ok := e.matchRulesDetail(e.exclusions, host, t.IP); ok {
		return Decision{
			Mode:   normalizeMode(mode),
			Rule:   rule,
			Source: "exclusion",
			Reason: "user_exclusion_direct",
		}
	}
	if mode, rule, ok := e.matchRulesDetail(e.rules, host, t.IP); ok {
		m := normalizeMode(mode)
		return Decision{
			Mode:   m,
			Rule:   rule,
			Source: "rule",
			Reason: reasonForRule(m, rule),
		}
	}
	return Decision{
		Mode:   e.defaultMode,
		Rule:   "default",
		Source: "default",
		Reason: reasonForDefault(e.defaultMode),
	}
}

func reasonForRule(mode Mode, pattern string) string {
	p := strings.ToLower(pattern)
	switch mode {
	case ModeDirect:
		if p == "*.ru" || strings.HasSuffix(p, ".ru") || strings.Contains(p, "рф") ||
			p == "*.su" || p == "*.xn--p1ai" {
			return "ru_domain_bypass"
		}
		if strings.Contains(p, "bank") || strings.Contains(p, "gos") {
			return "critical_ru_app_direct"
		}
		return "rule_direct"
	case ModeFallback:
		return "rule_fallback"
	default:
		if strings.Contains(p, "youtube") || strings.Contains(p, "google") ||
			strings.Contains(p, "instagram") || strings.Contains(p, "telegram") {
			return "global_service_relay"
		}
		return "international_traffic_relay"
	}
}

func reasonForDefault(mode Mode) string {
	if mode == ModeDirect {
		return "default_direct"
	}
	return "default_relay_foreign"
}

func (e *Engine) matchRules(rules []Rule, host string, ip netip.Addr) (Mode, bool) {
	mode, _, ok := e.matchRulesDetail(rules, host, ip)
	return mode, ok
}

func (e *Engine) matchRulesDetail(rules []Rule, host string, ip netip.Addr) (Mode, string, bool) {
	for _, r := range rules {
		switch r.Kind {
		case KindDomain:
			if host != "" && domainMatches(r.Pattern, host) {
				return r.Mode, r.Pattern, true
			}
		case KindCIDR:
			if ip.IsValid() && cidrContains(r.Pattern, ip) {
				return r.Mode, r.Pattern, true
			}
		}
	}
	return "", "", false
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
