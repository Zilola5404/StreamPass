package decision_test

import (
	"testing"

	"streampass/go_core/internal/decision"
)

func TestAtomicEngine_hotReload(t *testing.T) {
	v1 := `{"version":1,"rules":[{"kind":"DOMAIN","pattern":"*.ru","mode":"DIRECT"}]}`
	engine, err := decision.NewAtomicEngineFromJSON(v1, "")
	if err != nil {
		t.Fatal(err)
	}
	if got := engine.Decide(decision.Target{Host: "yandex.ru"}); got != decision.ModeDirect {
		t.Fatalf("v1 = %s", got)
	}

	v2 := `{"version":2,"rules":[{"kind":"DOMAIN","pattern":"*.ru","mode":"RELAY"}]}`
	if err := engine.Update(v2, ""); err != nil {
		t.Fatal(err)
	}
	if engine.Version() != 2 {
		t.Fatalf("version = %d", engine.Version())
	}
	// Built-in Russian DIRECT defaults take precedence over backend RELAY for *.ru.
	if got := engine.Decide(decision.Target{Host: "yandex.ru"}); got != decision.ModeDirect {
		t.Fatalf("v2 = %s, want DIRECT (default *.ru rule wins)", got)
	}
}

func TestAtomicEngine_updateExclusions(t *testing.T) {
	rules := `{"version":1,"rules":[{"kind":"DOMAIN","pattern":"*.google.com","mode":"RELAY"}]}`
	engine, err := decision.NewAtomicEngineFromJSON(rules, `[]`)
	if err != nil {
		t.Fatal(err)
	}
	if err := engine.Update(rules, `["docs.google.com"]`); err != nil {
		t.Fatal(err)
	}
	if got := engine.Decide(decision.Target{Host: "docs.google.com"}); got != decision.ModeDirect {
		t.Fatalf("exclusion = %s", got)
	}
}
