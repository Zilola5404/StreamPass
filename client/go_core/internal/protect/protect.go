// Package protect bridges Android VpnService.protect(fd) into Go dials.
// Without it, after TUN installs 0.0.0.0/0 every underlay socket loops into TUN.
package protect

import (
	"fmt"
	"sync"
	"syscall"
)

// Protector marks a socket fd so it bypasses the VPN routing table.
type Protector interface {
	Protect(fd int) bool
}

var (
	mu        sync.RWMutex
	protector Protector
)

// Set installs the platform protector (Android VpnService). Nil clears it.
func Set(p Protector) {
	mu.Lock()
	protector = p
	mu.Unlock()
}

// Clear removes the active protector.
func Clear() {
	Set(nil)
}

// FD protects a raw file descriptor. No-op when no protector is installed
// (desktop tests / non-VPN builds).
func FD(fd int) error {
	mu.RLock()
	p := protector
	mu.RUnlock()
	if p == nil {
		return nil
	}
	if !p.Protect(fd) {
		return fmt.Errorf("protect(%d) failed", fd)
	}
	return nil
}

// Conn protects an existing connection via SyscallConn.
func Conn(c syscall.Conn) error {
	if c == nil {
		return nil
	}
	raw, err := c.SyscallConn()
	if err != nil {
		return err
	}
	var perr error
	if err := raw.Control(func(fd uintptr) {
		perr = FD(int(fd))
	}); err != nil {
		return err
	}
	return perr
}

// Control returns a net.Dialer / ListenConfig Control func that protects
// newly created sockets.
func Control(network, address string, c syscall.RawConn) error {
	var perr error
	if err := c.Control(func(fd uintptr) {
		perr = FD(int(fd))
	}); err != nil {
		return err
	}
	return perr
}
