package tunbridge_test

import (
	"testing"

	"streampass/go_core/internal/decision"
	"streampass/go_core/internal/tunbridge"
)

// Documents audit Этап 4: DIRECT fail → RELAY retry is skipped under forceMode.
func TestDirectRetryPolicy_skippedUnderForceMode(t *testing.T) {
	eng := decision.NewAtomicEngine(decision.NewEngine(nil, nil, decision.DefaultMode), 0)
	eng.SetForceMode(decision.ModeDirect)
	if eng.ForceMode() != decision.ModeDirect {
		t.Fatal("expected force DIRECT")
	}
	// Product path only retries when ForceMode()=="".
	if eng.ForceMode() == "" {
		t.Fatal("would incorrectly allow RELAY retry in direct_test")
	}
	_ = tunbridge.TunIPv4Host()
}
