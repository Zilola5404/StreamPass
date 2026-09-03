//go:build windows

package protect

import "testing"

func TestFindStaleStreamPassGateways_fallback(t *testing.T) {
	gws := findStaleStreamPassGateways()
	if len(gws) == 0 {
		t.Fatal("expected at least one gateway candidate")
	}
	for _, gw := range gws {
		if gw[:7] != "10.10.0" {
			t.Fatalf("unexpected gateway %q", gw)
		}
	}
}

func TestRegisterAndClearSessionGateway(t *testing.T) {
	RegisterTunnelSessionGateway("10.10.0.2")
	RegisterTunnelSessionGateway("10.10.0.2")
	// idempotent register; clear is best-effort (route may not exist)
	_ = ClearSessionTunnelRoutes()
}
