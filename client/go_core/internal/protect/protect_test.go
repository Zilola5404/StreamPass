package protect

import (
	"net"
	"syscall"
	"testing"
)

type fakeProtector struct {
	fds []int
	ok  bool
}

func (f *fakeProtector) Protect(fd int) bool {
	f.fds = append(f.fds, fd)
	return f.ok
}

func TestFD_noopWithoutProtector(t *testing.T) {
	Clear()
	if err := FD(3); err != nil {
		t.Fatalf("FD without protector: %v", err)
	}
}

func TestFD_callsProtector(t *testing.T) {
	Clear()
	t.Cleanup(Clear)
	fp := &fakeProtector{ok: true}
	Set(fp)
	if err := FD(42); err != nil {
		t.Fatalf("FD: %v", err)
	}
	if len(fp.fds) != 1 || fp.fds[0] != 42 {
		t.Fatalf("fds=%v want [42]", fp.fds)
	}
}

func TestFD_failure(t *testing.T) {
	Clear()
	t.Cleanup(Clear)
	Set(&fakeProtector{ok: false})
	if err := FD(7); err == nil {
		t.Fatal("expected protect failure")
	}
}

func TestConn_udp(t *testing.T) {
	Clear()
	t.Cleanup(Clear)
	fp := &fakeProtector{ok: true}
	Set(fp)

	c, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		t.Fatalf("ListenUDP: %v", err)
	}
	defer c.Close()

	if err := Conn(c); err != nil {
		t.Fatalf("Conn: %v", err)
	}
	if len(fp.fds) != 1 || fp.fds[0] <= 0 {
		t.Fatalf("expected one positive fd, got %v", fp.fds)
	}

	// Control path used by Dialer
	raw, err := c.SyscallConn()
	if err != nil {
		t.Fatal(err)
	}
	if err := Control("udp", "", raw); err != nil {
		t.Fatalf("Control: %v", err)
	}
	if len(fp.fds) != 2 {
		t.Fatalf("expected 2 protect calls, got %v", fp.fds)
	}
}

var _ syscall.Conn = (*net.UDPConn)(nil)
