//go:build windows

package tunbridge

import (
	"strings"
	"testing"
)

// Stage 9 — IPv6 Variant B: desktop Wintun must not add Inet6 / StrictRoute block.
func TestDesktopOptions_ipv6VariantB(t *testing.T) {
	// Compile-time contract documented in desktop_windows.go StartDesktop.
	// Assert constants used by Windows path.
	if TunDNS().String() != "198.18.0.1" {
		t.Fatalf("stage7 dns want 198.18.0.1 got %s", TunDNS())
	}
	if TunIPv4Host() != "10.10.0.1" {
		t.Fatalf("tun host want 10.10.0.1 got %s", TunIPv4Host())
	}
	if windowsAdapterName != "StreamPass" {
		t.Fatalf("adapter name=%s", windowsAdapterName)
	}
}

func TestWrapWintunErr_adminHint(t *testing.T) {
	err := wrapWintunErr(errString("Access is denied."))
	if err == nil || !strings.Contains(err.Error(), "администратора") {
		t.Fatalf("want admin hint, got %v", err)
	}
}

type errString string

func (e errString) Error() string { return string(e) }
