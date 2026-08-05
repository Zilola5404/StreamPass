// TCP ↔ UDP datagram bridge for Hysteria2 TCP underlay (ТЗ §10).
//
// Listens on TCP ports and forwards length-prefixed frames to a local
// Hysteria UDP listener (usually 127.0.0.1:443).
//
// Wire format (same as client hyconfig.tcpPacketConn):
//
//	[uint16 BE length][payload]
//
// Usage:
//
//	go run . -listen :8443,:24443 -udp 127.0.0.1:443
package main

import (
	"encoding/binary"
	"errors"
	"flag"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

func main() {
	listen := flag.String("listen", ":8443,:24443", "TCP listen addresses (comma-separated)")
	udpTarget := flag.String("udp", "127.0.0.1:443", "local Hysteria UDP address")
	flag.Parse()

	udpAddr, err := net.ResolveUDPAddr("udp", *udpTarget)
	if err != nil {
		log.Fatalf("resolve udp: %v", err)
	}

	var listeners []net.Listener
	for _, addr := range strings.Split(*listen, ",") {
		addr = strings.TrimSpace(addr)
		if addr == "" {
			continue
		}
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			log.Fatalf("listen %s: %v", addr, err)
		}
		listeners = append(listeners, ln)
		log.Printf("listening on %s → UDP %s", ln.Addr(), udpAddr)
		go acceptLoop(ln, udpAddr)
	}
	if len(listeners) == 0 {
		log.Fatal("no listen addresses")
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	for _, ln := range listeners {
		_ = ln.Close()
	}
}

func acceptLoop(ln net.Listener, udpAddr *net.UDPAddr) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			log.Printf("accept: %v", err)
			continue
		}
		go handleConn(conn, udpAddr)
	}
}

func handleConn(tcp net.Conn, udpAddr *net.UDPAddr) {
	defer tcp.Close()
	udp, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Printf("udp dial %s: %v", udpAddr, err)
		return
	}
	defer udp.Close()

	log.Printf("session %s ↔ %s", tcp.RemoteAddr(), udpAddr)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		_ = tcpToUDP(tcp, udp)
		_ = tcp.Close()
		_ = udp.Close()
	}()
	go func() {
		defer wg.Done()
		_ = udpToTCP(udp, tcp)
		_ = tcp.Close()
		_ = udp.Close()
	}()
	wg.Wait()
}

func tcpToUDP(tcp net.Conn, udp *net.UDPConn) error {
	for {
		var hdr [2]byte
		if _, err := io.ReadFull(tcp, hdr[:]); err != nil {
			return err
		}
		n := int(binary.BigEndian.Uint16(hdr[:]))
		if n == 0 {
			continue
		}
		buf := make([]byte, n)
		if _, err := io.ReadFull(tcp, buf); err != nil {
			return err
		}
		if _, err := udp.Write(buf); err != nil {
			return err
		}
	}
}

func udpToTCP(udp *net.UDPConn, tcp net.Conn) error {
	buf := make([]byte, 65535)
	for {
		n, err := udp.Read(buf)
		if err != nil {
			return err
		}
		if n == 0 {
			continue
		}
		var hdr [2]byte
		binary.BigEndian.PutUint16(hdr[:], uint16(n))
		if _, err := tcp.Write(hdr[:]); err != nil {
			return err
		}
		if _, err := tcp.Write(buf[:n]); err != nil {
			return err
		}
	}
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lmsgprefix)
	log.SetPrefix("[tcpunderlay] ")
}
