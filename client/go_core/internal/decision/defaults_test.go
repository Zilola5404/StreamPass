package decision_test

import (
	"testing"

	"streampass/go_core/internal/decision"
)

func TestDefaultDirectRules_mergedWithBackend(t *testing.T) {
	rules := []decision.Rule{
		{Kind: decision.KindDomain, Pattern: "youtube.com", Mode: decision.ModeRelay},
	}
	e := decision.NewEngine(decision.MergeWithDefaults(rules), nil, decision.DefaultMode)

	if got := e.Decide(decision.Target{Host: "yandex.ru"}); got != decision.ModeDirect {
		t.Fatalf("yandex.ru = %s, want DIRECT", got)
	}
	if got := e.Decide(decision.Target{Host: "www.youtube.com"}); got != decision.ModeRelay {
		t.Fatalf("youtube = %s, want RELAY", got)
	}
	if got := e.Decide(decision.Target{Host: "google.com"}); got != decision.ModeRelay {
		t.Fatalf("google.com default = %s, want RELAY (built-in accelerator fallback)", got)
	}
	if got := e.Decide(decision.Target{Host: "cdninstagram.com"}); got != decision.ModeRelay {
		t.Fatalf("cdninstagram.com = %s, want RELAY", got)
	}
	if got := e.Decide(decision.Target{Host: "gemini.google.com"}); got != decision.ModeRelay {
		t.Fatalf("gemini.google.com = %s, want RELAY", got)
	}
}

func TestNewEngineFromJSON_includesDefaults(t *testing.T) {
	raw := `{"version":1,"rules":[{"kind":"DOMAIN","pattern":"instagram.com","mode":"RELAY"}]}`
	engine, err := decision.NewEngineFromJSON(raw, "")
	if err != nil {
		t.Fatal(err)
	}
	if got := engine.Decide(decision.Target{Host: "gosuslugi.ru"}); got != decision.ModeDirect {
		t.Fatalf("gosuslugi.ru = %s, want DIRECT", got)
	}
}
