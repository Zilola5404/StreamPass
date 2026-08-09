package decision_test

import (
	"testing"

	"streampass/go_core/internal/decision"
)

func TestForceMode_overridesDecide(t *testing.T) {
	eng := decision.NewAtomicEngine(
		decision.NewEngine(decision.MergeWithDefaults(nil), nil, decision.DefaultMode),
		1,
	)
	if got := eng.ForceMode(); got != "" {
		t.Fatalf("ForceMode=%q want empty", got)
	}
	eng.SetForceMode(decision.ModeDirect)
	if eng.ForceMode() != decision.ModeDirect {
		t.Fatal("ForceMode not set")
	}
	d := eng.DecideDetailed(decision.Target{Host: "youtube.com"})
	if d.Mode != decision.ModeDirect || d.Reason != "network_mode_DIRECT" {
		t.Fatalf("forced decide=%+v", d)
	}
	eng.SetForceMode("")
	d = eng.DecideDetailed(decision.Target{Host: "youtube.com"})
	if d.Mode != decision.ModeRelay {
		t.Fatalf("youtube after clear force=%s want RELAY", d.Mode)
	}
}
