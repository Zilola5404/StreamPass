//go:build windows

package protect

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

const (
	ipUnicastIF   = 31
	ipv6UnicastIF = 31
)

var (
	underlayMu     sync.Mutex
	underlayCached bool
	underlayIdx    int
	underlayName   string
)

type interfaceProtector struct {
	index int
}

func (p interfaceProtector) Protect(fd int) bool {
	handle := syscall.Handle(fd)
	if err := bind4(handle, p.index); err != nil {
		return false
	}
	// IPv6 may be disabled on the NIC; ignore bind6 failure.
	_ = bind6(handle, p.index)
	return true
}

func bind4(handle syscall.Handle, ifaceIdx int) error {
	var bytes [4]byte
	binary.BigEndian.PutUint32(bytes[:], uint32(ifaceIdx))
	idx := *(*uint32)(unsafe.Pointer(&bytes[0]))
	return syscall.SetsockoptInt(handle, syscall.IPPROTO_IP, ipUnicastIF, int(idx))
}

func bind6(handle syscall.Handle, ifaceIdx int) error {
	return syscall.SetsockoptInt(handle, syscall.IPPROTO_IPV6, ipv6UnicastIF, ifaceIdx)
}

// BindInterface pins underlay/DIRECT sockets to a physical NIC so they do not
// follow the TUN default route (Windows equivalent of VpnService.protect).
// Call before installing 0.0.0.0/0 via TUN. Index 0 clears the protector.
func BindInterface(ifIndex int) {
	if ifIndex <= 0 {
		Clear()
		return
	}
	Set(interfaceProtector{index: ifIndex})
}

// IsTunnelInterface reports VPN/TAP/TUN adapters that must never host underlay.
func IsTunnelInterface(name string) bool {
	n := strings.ToLower(strings.TrimSpace(name))
	if n == "" {
		return false
	}
	keys := []string{
		"streampass",
		"wintun",
		"wireguard",
		"outline",
		"tap-windows",
		"tap0901",
		"tun",
		"utun",
		"wsl",
		"vethernet",
		"hyper-v",
		"loopback",
	}
	for _, k := range keys {
		if strings.Contains(n, k) {
			return true
		}
	}
	return false
}

// BindPhysicalUnderlay finds a real NIC (never StreamPass/Wintun/TAP) and
// installs the protector. Must run BEFORE Hysteria handshake and before AutoRoute.
// Result is cached for the session (StartDesktop must not re-probe — PowerShell
// Get-NetRoute can hang on repeated calls).
func BindPhysicalUnderlay() (index int, name string, err error) {
	underlayMu.Lock()
	if underlayCached {
		idx, n := underlayIdx, underlayName
		underlayMu.Unlock()
		BindInterface(idx)
		return idx, n, nil
	}
	underlayMu.Unlock()

	index, name, err = PhysicalInterfaceIndex()
	if err != nil {
		return 0, "", err
	}
	BindInterface(index)
	underlayMu.Lock()
	underlayIdx, underlayName = index, name
	underlayCached = true
	underlayMu.Unlock()
	return index, name, nil
}

// HasUnderlay reports whether BindPhysicalUnderlay succeeded this session.
func HasUnderlay() bool {
	underlayMu.Lock()
	defer underlayMu.Unlock()
	return underlayCached
}

func resetUnderlaySession() {
	underlayMu.Lock()
	underlayCached = false
	underlayIdx = 0
	underlayName = ""
	underlayMu.Unlock()
}

// ClearStaleTunnelDefaultRoute removes leftover 0.0.0.0/0 via StreamPass from a
// previous crash so the OS is not stuck on a dead TUN. See route_windows.go.
func ClearStaleTunnelDefaultRoute() error {
	return clearStaleTunnelDefaultRouteImpl()
}

// PhysicalInterfaceIndex is the IPv4 NIC used for internet *before* TUN routes.
// Never returns StreamPass / Wintun / TAP — otherwise underlay loops into TUN.
func PhysicalInterfaceIndex() (index int, name string, err error) {
	// Prefer the OS default-route interface (audit P1), then dial-owner, then probe.
	if idx, n, ok := defaultRouteInterfaceIndex(); ok {
		return idx, n, nil
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		return 0, "", err
	}

	type cand struct {
		idx  int
		name string
	}
	var cands []cand
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		if IsTunnelInterface(iface.Name) {
			continue
		}
		addrs, aerr := iface.Addrs()
		if aerr != nil {
			continue
		}
		hasV4 := false
		for _, a := range addrs {
			ipnet, ok := a.(*net.IPNet)
			if !ok || ipnet.IP == nil || ipnet.IP.To4() == nil || ipnet.IP.IsLoopback() {
				continue
			}
			// Skip link-local / our TUN subnet leftovers mis-assigned
			ip4 := ipnet.IP.To4()
			if ip4[0] == 169 && ip4[1] == 254 {
				continue
			}
			if ip4[0] == 10 && ip4[1] == 10 && ip4[2] == 0 {
				continue
			}
			hasV4 = true
			break
		}
		if hasV4 {
			cands = append(cands, cand{idx: iface.Index, name: iface.Name})
		}
	}
	if len(cands) == 0 {
		return 0, "", fmt.Errorf("no physical IPv4 interface (all up ifaces look like TUN/TAP)")
	}

	// Prefer the interface that owns the UDP dial source — but only if not a tunnel.
	if idx, name, ok := dialOwnerInterface(ifaces); ok && !IsTunnelInterface(name) {
		return idx, name, nil
	}

	// Default route is stuck on StreamPass: probe each physical candidate with IP_UNICAST_IF.
	for _, c := range cands {
		if probeDialOnInterface(c.idx) {
			return c.idx, c.name, nil
		}
	}
	// Last resort: first physical candidate (bind may still help once routes are fixed).
	return cands[0].idx, cands[0].name, nil
}

func dialOwnerInterface(ifaces []net.Interface) (int, string, bool) {
	c, err := net.Dial("udp4", "1.1.1.1:53")
	if err != nil {
		c, err = net.Dial("udp4", "8.8.8.8:53")
		if err != nil {
			return 0, "", false
		}
	}
	defer c.Close()
	local, ok := c.LocalAddr().(*net.UDPAddr)
	if !ok || local.IP == nil {
		return 0, "", false
	}
	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			ipnet, ok := a.(*net.IPNet)
			if !ok || ipnet.IP == nil {
				continue
			}
			if sameIPv4(ipnet.IP, local.IP) {
				return iface.Index, iface.Name, true
			}
		}
	}
	return 0, "", false
}

func probeDialOnInterface(ifIndex int) bool {
	d := net.Dialer{
		Timeout: 2 * time.Second,
		Control: func(network, address string, c syscall.RawConn) error {
			var perr error
			if err := c.Control(func(fd uintptr) {
				perr = bind4(syscall.Handle(fd), ifIndex)
			}); err != nil {
				return err
			}
			return perr
		},
	}
	c, err := d.Dial("udp4", "1.1.1.1:53")
	if err != nil {
		return false
	}
	_ = c.Close()
	return true
}

func sameIPv4(a, b net.IP) bool {
	a4, b4 := a.To4(), b.To4()
	if a4 == nil || b4 == nil {
		return a.Equal(b)
	}
	return a4.Equal(b4)
}
