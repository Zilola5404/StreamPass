//go:build windows

package protect

import (
	"net"
	"testing"
)

func TestPhysicalInterfaceIndex(t *testing.T) {
	idx, name, err := PhysicalInterfaceIndex()
	if err != nil {
		t.Skipf("no default IPv4 interface: %v", err)
	}
	if idx <= 0 {
		t.Fatalf("index=%d name=%s", idx, name)
	}
	t.Logf("physical if index=%d name=%s", idx, name)
}

func TestBindInterface_udpSocket(t *testing.T) {
	idx, _, err := PhysicalInterfaceIndex()
	if err != nil {
		t.Skip(err)
	}
	t.Cleanup(Clear)
	BindInterface(idx)

	c, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		t.Fatalf("ListenUDP: %v", err)
	}
	defer c.Close()
	if err := Conn(c); err != nil {
		t.Fatalf("protect Conn: %v", err)
	}
}

func TestBindInterface_zeroClears(t *testing.T) {
	BindInterface(1)
	BindInterface(0)
	if err := FD(3); err != nil {
		t.Fatalf("FD after clear: %v", err)
	}
}
