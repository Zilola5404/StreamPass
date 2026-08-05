package hyconfig

import (
	"bytes"
	"net"
	"testing"
	"time"
)

func TestTCPPacketConn_roundTrip(t *testing.T) {
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()

	remote := &net.UDPAddr{IP: net.IPv4(1, 2, 3, 4), Port: 443}
	a := &tcpPacketConn{conn: c1, remote: remote}
	b := &tcpPacketConn{conn: c2, remote: remote}

	payload := []byte("quic-datagram-payload")
	done := make(chan error, 1)
	go func() {
		buf := make([]byte, 2048)
		n, addr, err := b.ReadFrom(buf)
		if err != nil {
			done <- err
			return
		}
		if addr.String() != remote.String() {
			done <- errString("bad addr " + addr.String())
			return
		}
		if !bytes.Equal(buf[:n], payload) {
			done <- errString("payload mismatch")
			return
		}
		done <- nil
	}()

	if _, err := a.WriteTo(payload, remote); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	}
}

type errString string

func (e errString) Error() string { return string(e) }
