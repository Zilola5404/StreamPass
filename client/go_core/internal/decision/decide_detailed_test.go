package decision_test

import (
	"testing"

	"streampass/go_core/internal/decision"
)

func TestDecideDetailed_russianDirect(t *testing.T) {
	e := decision.NewEngine(decision.MergeWithDefaults(nil), nil, decision.ModeRelay)
	d := e.DecideDetailed(decision.Target{Host: "ya.ru"})
	if d.Mode != decision.ModeDirect {
		t.Fatalf("mode=%s want DIRECT", d.Mode)
	}
	if d.Reason == "" || d.Rule == "" {
		t.Fatalf("missing rule/reason: %+v", d)
	}
}

func TestDecideDetailed_foreignRelay(t *testing.T) {
	e := decision.NewEngine(decision.MergeWithDefaults(nil), nil, decision.ModeRelay)
	d := e.DecideDetailed(decision.Target{Host: "example.com"})
	if d.Mode != decision.ModeRelay {
		t.Fatalf("mode=%s want RELAY", d.Mode)
	}
}
