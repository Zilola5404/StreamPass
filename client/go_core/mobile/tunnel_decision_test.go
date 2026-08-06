package mobile_test

import (
	"testing"

	"streampass/go_core/mobile"
)

func TestDecideRoute_domainDirect(t *testing.T) {
	rules := `{"version":1,"rules":[{"kind":"DOMAIN","pattern":"*.ru","mode":"DIRECT"}]}`
	got := mobile.DecideRoute(rules, "", "yandex.ru", "")
	if got != "DIRECT" {
		t.Fatalf("got %q want DIRECT", got)
	}
}

func TestDecideRoute_defaultDirect(t *testing.T) {
	rules := `{"version":1,"rules":[{"kind":"DOMAIN","pattern":"*.ru","mode":"DIRECT"}]}`
	got := mobile.DecideRoute(rules, "", "google.com", "")
	if got != "DIRECT" {
		t.Fatalf("got %q want DIRECT (decision.DefaultMode)", got)
	}
}

func TestUpdateRules_noActiveTunnel(t *testing.T) {
	mobile.StopTunnel()
	err := mobile.UpdateRules(`{"version":1,"rules":[]}`, "[]")
	if err != "no active tunnel" {
		t.Fatalf("got %q want no active tunnel", err)
	}
}

func TestActiveRulesVersion_noTunnel(t *testing.T) {
	mobile.StopTunnel()
	if v := mobile.ActiveRulesVersion(); v != 0 {
		t.Fatalf("got %d want 0", v)
	}
}
