package tunbridge_test

import (
	"net/netip"
	"testing"

	"streampass/go_core/internal/decision"
	"streampass/go_core/internal/dnscache"
)

// TrafficPathDiagnosis documents why foreign sites show geo-block / no traffic.
// See docs/07.4_RoutingPolicy.md and scripts/DiagnoseTrafficBlock.ps1.
func TestTrafficPathDiagnosis_ipOnlyMetaUsesCidrRelay(t *testing.T) {
	engine, err := decision.NewEngineFromJSON(`{"version":1,"rules":[]}`, "")
	if err != nil {
		t.Fatal(err)
	}
	// IP-only Meta CDN without HostForIP must still RELAY via builtin CIDR safety net.
	ip := netip.MustParseAddr("157.240.241.174")
	got := engine.Decide(decision.Target{IP: ip})
	if got != decision.ModeRelay {
		t.Fatalf("IP-only Meta without HostForIP=%s want RELAY (CIDR safety net)", got)
	}
}

func TestTrafficPathDiagnosis_hostForIPEnablesRelayRules(t *testing.T) {
	dnscache.RememberIP("www.instagram.com", "157.240.241.174")
	engine, err := decision.NewEngineFromJSON(`{"version":1,"rules":[]}`, "")
	if err != nil {
		t.Fatal(err)
	}
	host := dnscache.HostForIP("157.240.241.174")
	if host == "" {
		t.Fatal("RememberIP not stored")
	}
	got := engine.Decide(decision.Target{IP: netip.MustParseAddr("157.240.241.174"), Host: host})
	if got != decision.ModeRelay {
		t.Fatalf("with HostForIP=%q got %s want RELAY", host, got)
	}
}

func TestTrafficPathDiagnosis_youtubeHostMustRelay(t *testing.T) {
	engine, err := decision.NewEngineFromJSON(`{"version":1,"rules":[]}`, "")
	if err != nil {
		t.Fatal(err)
	}
	d := engine.DecideDetailed(decision.Target{Host: "www.youtube.com"})
	if d.Mode != decision.ModeRelay {
		t.Fatalf("youtube mode=%s want RELAY", d.Mode)
	}
}

func TestTrafficPathDiagnosis_ruHostDirect(t *testing.T) {
	engine, err := decision.NewEngineFromJSON(`{"version":1,"rules":[]}`, "")
	if err != nil {
		t.Fatal(err)
	}
	d := engine.DecideDetailed(decision.Target{Host: "ya.ru"})
	if d.Mode != decision.ModeDirect {
		t.Fatalf("ya.ru mode=%s want DIRECT", d.Mode)
	}
	if d.Reason != "ru_domain_bypass" {
		t.Fatalf("reason=%q want ru_domain_bypass", d.Reason)
	}
}
