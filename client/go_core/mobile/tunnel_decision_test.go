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
	got := mobile.DecideRoute(rules, "", "cloudflare.com", "")
	if got != "DIRECT" {
		t.Fatalf("got %q want DIRECT (DefaultMode FS §6)", got)
	}
}

func TestDecideRoute_ipOnlyTelegramDC(t *testing.T) {
	got := mobile.DecideRoute(`{"version":1,"rules":[]}`, "", "", "149.154.167.50")
	if got != "RELAY" {
		t.Fatalf("got %q want RELAY (Telegram DC CIDR)", got)
	}
}

func TestDecideRoute_builtinRelayFallback(t *testing.T) {
	rules := `{"version":1,"rules":[]}`
	got := mobile.DecideRoute(rules, "", "gemini.google.com", "")
	if got != "RELAY" {
		t.Fatalf("got %q want RELAY (DefaultRelayRules)", got)
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
