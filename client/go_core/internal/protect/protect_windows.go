//go:build windows

package protect

import (
	"encoding/binary"
	"fmt"
	"net"
	"syscall"
	"unsafe"
)

const (
	ipUnicastIF   = 31
	ipv6UnicastIF = 31
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

// PhysicalInterfaceIndex is the IPv4 NIC used for internet *before* TUN routes.
func PhysicalInterfaceIndex() (index int, name string, err error) {
	c, err := net.Dial("udp4", "1.1.1.1:53")
	if err != nil {
		c, err = net.Dial("udp4", "8.8.8.8:53")
		if err != nil {
			return 0, "", fmt.Errorf("probe default interface: %w", err)
		}
	}
	defer c.Close()
	local, ok := c.LocalAddr().(*net.UDPAddr)
	if !ok || local.IP == nil {
		return 0, "", fmt.Errorf("probe default interface: no local address")
	}
	ifaces, err := net.Interfaces()
	if err != nil {
		return 0, "", err
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
				return iface.Index, iface.Name, nil
			}
		}
	}
	return 0, "", fmt.Errorf("no interface owns %s", local.IP)
}

func sameIPv4(a, b net.IP) bool {
	a4, b4 := a.To4(), b.To4()
	if a4 == nil || b4 == nil {
		return a.Equal(b)
	}
	return a4.Equal(b4)
}
