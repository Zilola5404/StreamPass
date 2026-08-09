package dnscache_test

import (
	"testing"

	"streampass/go_core/internal/dnscache"
)

func TestClassifyDNSRoute_defaultDirectForForeign(t *testing.T) {
	dnscache.SetRouteHint(nil)
	rule, route, reason := dnscache.ClassifyDNSRoute("example.com")
	if route != "DIRECT" || reason != "default_direct" {
		t.Fatalf("got rule=%s route=%s reason=%s want DIRECT/default_direct", rule, route, reason)
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
