//go:build windows

package protect

import (
	"net"
	"testing"
)

func TestIsTunnelInterface(t *testing.T) {
	cases := map[string]bool{
		"StreamPass":        true,
		"Wintun":            true,
		"outline-tap0":      true,
		"Ethernet":          false,
		"Wi-Fi":             false,
		"Беспроводная сеть": false,
		"vEthernet (WSL)":   true,
	}
	for name, want := range cases {
		if got := IsTunnelInterface(name); got != want {
			t.Fatalf("%q: got %v want %v", name, got, want)
		}
	}
}

func TestPhysicalInterfaceIndex_skipsStreamPass(t *testing.T) {
	idx, name, err := PhysicalInterfaceIndex()
	if err != nil {
		t.Skipf("no physical IF: %v", err)
	}
	if IsTunnelInterface(name) {
		t.Fatalf("picked tunnel iface index=%d name=%s", idx, name)
	}
	t.Logf("physical if index=%d name=%s", idx, name)
}

func TestBindPhysicalUnderlay(t *testing.T) {
	idx, name, err := BindPhysicalUnderlay()
	t.Cleanup(Clear)
	if err != nil {
		t.Skip(err)
	}
	if IsTunnelInterface(name) {
		t.Fatalf("underlay on tunnel %s", name)
	}
	c, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if err := Conn(c); err != nil {
		t.Fatalf("protect Conn: %v", err)
	}
	t.Logf("bound underlay if=%d %s", idx, name)
}
