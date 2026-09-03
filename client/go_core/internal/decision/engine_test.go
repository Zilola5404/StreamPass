package decision_test

import (
	"net/netip"
	"testing"

	"streampass/go_core/internal/decision"
)

func TestDomainMatch_wildcardRu(t *testing.T) {
	rules := []decision.Rule{
		{Kind: decision.KindDomain, Pattern: "*.ru", Mode: decision.ModeDirect},
		{Kind: decision.KindDomain, Pattern: "youtube.com", Mode: decision.ModeRelay},
	}
	e := decision.NewEngine(rules, nil, decision.DefaultMode)

	cases := []struct {
		host string
		want decision.Mode
	}{
		{"yandex.ru", decision.ModeDirect},
		{"sub.mail.ru", decision.ModeDirect},
		{"www.youtube.com", decision.ModeRelay},
		{"google.com", decision.ModeDirect}, // DefaultMode=DIRECT when no rule
	}
	for _, tc := range cases {
		got := e.Decide(decision.Target{Host: tc.host})
		if got != tc.want {
			t.Errorf("Decide(%q) = %s, want %s", tc.host, got, tc.want)
		}
	}
}

func TestCIDRMatch_directRussianRange(t *testing.T) {
	rules := []decision.Rule{
		{Kind: decision.KindCIDR, Pattern: "178.248.232.0/21", Mode: decision.ModeDirect},
		{Kind: decision.KindDomain, Pattern: "google.com", Mode: decision.ModeRelay},
	}
	e := decision.NewEngine(rules, nil, decision.DefaultMode)

	ip := netip.MustParseAddr("178.248.233.10")
	got := e.Decide(decision.Target{IP: ip})
	if got != decision.ModeDirect {
		t.Fatalf("CIDR match = %s, want DIRECT", got)
	}

	outside := netip.MustParseAddr("8.8.8.8")
	got = e.Decide(decision.Target{IP: outside, Host: "dns.google"})
	if got != decision.ModeDirect {
		t.Fatalf("unmatched 8.8.8.8 = %s, want DIRECT (DefaultMode)", got)
	}
	got = e.Decide(decision.Target{Host: "google.com"})
	if got != decision.ModeRelay {
		t.Fatalf("google.com = %s, want RELAY", got)
	}
}

func TestUserExclusions_overrideRelayRules(t *testing.T) {
	rules := []decision.Rule{
		{Kind: decision.KindDomain, Pattern: "*.google.com", Mode: decision.ModeRelay},
	}
	exclusions := []decision.Rule{
		{Kind: decision.KindDomain, Pattern: "docs.google.com", Mode: decision.ModeDirect},
	}
	e := decision.NewEngine(rules, exclusions, decision.ModeRelay)

	if got := e.Decide(decision.Target{Host: "docs.google.com"}); got != decision.ModeDirect {
		t.Fatalf("exclusion = %s, want DIRECT", got)
	}
	if got := e.Decide(decision.Target{Host: "www.google.com"}); got != decision.ModeRelay {
		t.Fatalf("relay rule = %s, want RELAY", got)
	}
}

func TestFirstMatchingRuleWins(t *testing.T) {
	rules := []decision.Rule{
		{Kind: decision.KindDomain, Pattern: "*.ru", Mode: decision.ModeDirect},
		{Kind: decision.KindDomain, Pattern: "blocked.ru", Mode: decision.ModeRelay},
	}
	e := decision.NewEngine(rules, nil, decision.ModeRelay)

	if got := e.Decide(decision.Target{Host: "blocked.ru"}); got != decision.ModeDirect {
		t.Fatalf("first rule = %s, want DIRECT (first match wins)", got)
	}
}

func TestParseRuleSetJSON(t *testing.T) {
	raw := `{"version":2,"rules":[{"kind":"DOMAIN","pattern":"*.ru","mode":"DIRECT"}]}`
	set, err := decision.ParseRuleSetJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if set.Version != 2 || len(set.Rules) != 1 {
		t.Fatalf("set = %+v", set)
	}
	e := decision.NewEngine(set.Rules, nil, decision.ModeRelay)
	if got := e.Decide(decision.Target{Host: "vk.ru"}); got != decision.ModeDirect {
		t.Fatalf("got %s", got)
	}
}

func TestFallbackModePreserved(t *testing.T) {
	rules := []decision.Rule{
		{Kind: decision.KindDomain, Pattern: "slow.example", Mode: decision.ModeFallback},
	}
	e := decision.NewEngine(rules, nil, decision.ModeRelay)
	if got := e.Decide(decision.Target{Host: "slow.example"}); got != decision.ModeFallback {
		t.Fatalf("got %s, want FALLBACK", got)
	}
}

func TestIPTargetWithoutHost(t *testing.T) {
	rules := []decision.Rule{
		{Kind: decision.KindCIDR, Pattern: "142.250.0.0/15", Mode: decision.ModeRelay},
	}
	e := decision.NewEngine(rules, nil, decision.DefaultMode)

	ip := netip.MustParseAddr("142.250.1.2")
	if got := e.Decide(decision.Target{IP: ip}); got != decision.ModeRelay {
		t.Fatalf("got %s, want RELAY", got)
	}

	meta := netip.MustParseAddr("157.240.1.2")
	e2, err := decision.NewEngineFromJSON(`{"version":1,"rules":[]}`, "")
	if err != nil {
		t.Fatal(err)
	}
	// Meta CDN IP → builtin CIDR safety net RELAY (HostForIP miss path).
	if got := e2.Decide(decision.Target{IP: meta}); got != decision.ModeRelay {
		t.Fatalf("Meta CDN IP got %s, want RELAY", got)
	}
	tg := netip.MustParseAddr("149.154.167.50")
	if got := e2.Decide(decision.Target{IP: tg}); got != decision.ModeRelay {
		t.Fatalf("Telegram DC IP got %s, want RELAY", got)
	}
}
