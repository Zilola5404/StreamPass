package dnscache_test

import (
	"testing"

	"streampass/go_core/internal/dnscache"
)

func TestClassifyDNSRoute_defaultRelayForForeign(t *testing.T) {
	dnscache.SetRouteHint(nil)
	rule, route, reason := dnscache.ClassifyDNSRoute("example.com")
	if route != "RELAY" || reason != "default_relay_foreign" {
		t.Fatalf("got rule=%s route=%s reason=%s want RELAY/default_relay_foreign", rule, route, reason)
	}
}

func TestClassifyDNSRoute_russian(t *testing.T) {
	dnscache.SetRouteHint(nil)
	_, route, reason := dnscache.ClassifyDNSRoute("ya.ru")
	if route != "DIRECT" || reason != "ru_domain_bypass" {
		t.Fatalf("got route=%s reason=%s", route, reason)
	}
}

func TestClassifyDNSRoute_forceModeHint(t *testing.T) {
	dnscache.SetRouteHint(func(host string) (string, string, string) {
		return "network_mode", "RELAY", "network_mode_RELAY"
	})
	defer dnscache.SetRouteHint(nil)
	rule, route, reason := dnscache.ClassifyDNSRoute("ya.ru")
	if rule != "network_mode" || route != "RELAY" || reason != "network_mode_RELAY" {
		t.Fatalf("got %s/%s/%s", rule, route, reason)
	}
}

func TestPinDirectIP_survivesDefaultRelay(t *testing.T) {
	dnscache.PinDirectIP("2ip.ru", "185.178.208.137")
	if got := dnscache.DirectHostNearIP("185.178.208.140"); got != "2ip.ru" {
		t.Fatalf("DirectHostNearIP=%q want 2ip.ru", got)
	}
}
