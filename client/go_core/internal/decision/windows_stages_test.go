package decision_test

import (
	"testing"

	"streampass/go_core/internal/decision"
)

// Stage 8 — split Decision matrix used by Windows sidecar (same engine as Android).
func TestWindowsSplitMatrix(t *testing.T) {
	e := decision.NewEngine(decision.MergeWithDefaults(nil), nil, decision.DefaultMode)
	cases := []struct {
		host string
		want decision.Mode
	}{
		{"ya.ru", decision.ModeDirect},
		{"mail.ru", decision.ModeDirect},
		{"vk.com", decision.ModeDirect},
		{"youtube.com", decision.ModeRelay},
		{"instagram.com", decision.ModeRelay},
		{"github.com", decision.ModeRelay},
		{"cloudflare.com", decision.ModeDirect}, // DefaultMode=DIRECT (unknown)
	}
	for _, tc := range cases {
		got := e.Decide(decision.Target{Host: tc.host})
		if got != tc.want {
			t.Fatalf("%s: got %s want %s", tc.host, got, tc.want)
		}
	}
}

// Stage 11 — Adaptive fallback is Decision ModeFallback (classified), not silent leak.
func TestWindowsAdaptiveFallbackMode(t *testing.T) {
	rules := []decision.Rule{
		{Kind: decision.KindDomain, Pattern: "slow.example", Mode: decision.ModeFallback},
	}
	e := decision.NewEngine(decision.MergeWithDefaults(rules), nil, decision.DefaultMode)
	if got := e.Decide(decision.Target{Host: "slow.example"}); got != decision.ModeFallback {
		t.Fatalf("got %s want FALLBACK", got)
	}
	// must-relay still RELAY
	if got := e.Decide(decision.Target{Host: "youtube.com"}); got != decision.ModeRelay {
		t.Fatalf("youtube must stay RELAY, got %s", got)
	}
}
