package decision_test

import (
	"testing"

	"streampass/go_core/internal/decision"
)

func TestDecideDetailed_russianDirect(t *testing.T) {
	e := decision.NewEngine(decision.MergeWithDefaults(nil), nil, decision.DefaultMode)
	d := e.DecideDetailed(decision.Target{Host: "ya.ru"})
	if d.Mode != decision.ModeDirect {
		t.Fatalf("mode=%s want DIRECT", d.Mode)
	}
	if d.Reason != "ru_domain_bypass" {
		t.Fatalf("reason=%q want ru_domain_bypass", d.Reason)
	}
	if d.Rule != "*.ru" {
		t.Fatalf("rule=%q want *.ru", d.Rule)
	}
}

func TestDecideDetailed_foreignDefaultDirect(t *testing.T) {
	e := decision.NewEngine(decision.MergeWithDefaults(nil), nil, decision.DefaultMode)
	d := e.DecideDetailed(decision.Target{Host: "example.com"})
	if d.Mode != decision.ModeDirect {
		t.Fatalf("mode=%s want DIRECT (DefaultMode)", d.Mode)
	}
}

func TestDecideDetailed_youtubeMustRelay(t *testing.T) {
	e := decision.NewEngine(decision.MergeWithDefaults(nil), nil, decision.DefaultMode)
	d := e.DecideDetailed(decision.Target{Host: "www.youtube.com"})
	if d.Mode != decision.ModeRelay {
		t.Fatalf("mode=%s want RELAY", d.Mode)
	}
}
