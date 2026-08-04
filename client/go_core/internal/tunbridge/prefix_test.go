package tunbridge

import (
	"net/netip"
	"testing"

	tun "github.com/sagernet/sing-tun"
)

func TestTunIPv4PrefixHasUsableNextHost(t *testing.T) {
	prefix := TunIPv4Prefix()
	if !tun.HasNextAddress(prefix, 1) {
		t.Fatalf("prefix %s must contain Addr().Next() for system stack", prefix)
	}
	next := prefix.Addr().Next()
	broadcast := tun.BroadcastAddr([]netip.Prefix{prefix})
	if next == broadcast {
		t.Fatalf("Addr().Next()=%s is the broadcast of %s — TCP NAT will break", next, prefix)
	}
	if prefix.Addr() != netip.MustParseAddr("10.10.0.1") {
		t.Fatalf("expected TUN host 10.10.0.1, got %s", prefix.Addr())
	}
}

func TestLegacyTunPrefixWasBroadcastNext(t *testing.T) {
	// Documents the +13 bug: 10.10.0.2/30 → Next() == broadcast.
	legacy := netip.PrefixFrom(netip.MustParseAddr("10.10.0.2"), 30)
	next := legacy.Addr().Next()
	broadcast := tun.BroadcastAddr([]netip.Prefix{legacy})
	if next != broadcast {
		t.Fatalf("expected legacy next==broadcast, got next=%s broadcast=%s", next, broadcast)
	}
}
