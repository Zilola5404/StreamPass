package hyconfig

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"syscall"
	"time"

	"streampass/go_core/internal/protect"
)

// tcpPacketConn frames UDP datagrams over a single TCP stream so QUIC can
// run when raw UDP is blocked (ТЗ §10 TCP 443 / TCP 8443 underlay).
// Wire format: [uint16 BE length][payload], max 65535 bytes per datagram.
type tcpPacketConn struct {
	conn   net.Conn
	remote net.Addr
	rmu    sync.Mutex
	wmu    sync.Mutex
}

func dialTCPUnderlay(serverHostPort string, remoteUDP net.Addr) (net.PacketConn, error) {
	d := net.Dialer{
		Timeout: 12 * time.Second,
		Control: protect.Control,
	}
	conn, err := d.Dial("tcp", serverHostPort)
	if err != nil {
		return nil, err
	}
	if sc, ok := conn.(syscall.Conn); ok {
		if err := protect.Conn(sc); err != nil {
			_ = conn.Close()
			return nil, fmt.Errorf("protect underlay tcp: %w", err)
		}
	}
	return &tcpPacketConn{conn: conn, remote: remoteUDP}, nil
}

func (c *tcpPacketConn) ReadFrom(p []byte) (int, net.Addr, error) {
	c.rmu.Lock()
	defer c.rmu.Unlock()
	var hdr [2]byte
	if _, err := io.ReadFull(c.conn, hdr[:]); err != nil {
		return 0, nil, err
	}
	n := int(binary.BigEndian.Uint16(hdr[:]))
	if n == 0 {
		return 0, c.remote, nil
	}
	if n > len(p) {
		buf := make([]byte, n)
		if _, err := io.ReadFull(c.conn, buf); err != nil {
			return 0, nil, err
		}
		copy(p, buf[:len(p)])
		return len(p), c.remote, fmt.Errorf("tcp underlay: datagram truncated")
	}
	if _, err := io.ReadFull(c.conn, p[:n]); err != nil {
		return 0, nil, err
	}
	return n, c.remote, nil
}

func (c *tcpPacketConn) WriteTo(p []byte, _ net.Addr) (int, error) {
	if len(p) > 65535 {
		return 0, fmt.Errorf("tcp underlay: datagram too large (%d)", len(p))
	}
	c.wmu.Lock()
	defer c.wmu.Unlock()
	var hdr [2]byte
	binary.BigEndian.PutUint16(hdr[:], uint16(len(p)))
	if _, err := c.conn.Write(hdr[:]); err != nil {
		return 0, err
	}
	if _, err := c.conn.Write(p); err != nil {
		return 0, err
	}
	return len(p), nil
}

func (c *tcpPacketConn) Close() error                       { return c.conn.Close() }
func (c *tcpPacketConn) LocalAddr() net.Addr                { return c.conn.LocalAddr() }
func (c *tcpPacketConn) SetDeadline(t time.Time) error      { return c.conn.SetDeadline(t) }
func (c *tcpPacketConn) SetReadDeadline(t time.Time) error  { return c.conn.SetReadDeadline(t) }
func (c *tcpPacketConn) SetWriteDeadline(t time.Time) error { return c.conn.SetWriteDeadline(t) }

// tcpUnderlayConnFactory dials a framed TCP underlay toward the relay.
type tcpUnderlayConnFactory struct {
	tcpAddr  string
	udpRemote net.Addr
	obfsType string
	obfsPass string
}

func (f *tcpUnderlayConnFactory) New(net.Addr) (net.PacketConn, error) {
	pc, err := dialTCPUnderlay(f.tcpAddr, f.udpRemote)
	if err != nil {
		return nil, err
	}
	if f.obfsType == "" {
		return pc, nil
	}
	return wrapObfs(pc, f.obfsType, f.obfsPass)
}
