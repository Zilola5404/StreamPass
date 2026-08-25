//go:build windows

// StreamPass Windows traffic core: Wintun + Decision + Hysteria.
// Flutter talks JSON-lines over 127.0.0.1 (see --port-file).
package main

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	portFile := flag.String("port-file", "", "write {port,pid,token} JSON here after bind")
	flag.Parse()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "listen: %v\n", err)
		os.Exit(1)
	}
	defer ln.Close()

	token := randomToken()
	addr := ln.Addr().(*net.TCPAddr)
	if *portFile != "" {
		body, _ := json.Marshal(map[string]any{
			"port":  addr.Port,
			"pid":   os.Getpid(),
			"token": token,
		})
		if err := os.WriteFile(*portFile, body, 0600); err != nil {
			fmt.Fprintf(os.Stderr, "port-file: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Printf("READY %d %s\n", addr.Port, token)
	}

	rt := &runtime{}
	defer rt.stop()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig
		_ = ln.Close()
		rt.stop()
		os.Exit(0)
	}()

	conn, err := ln.Accept()
	if err != nil {
		return
	}
	defer conn.Close()
	_ = ln.Close()

	serve(conn, token, rt)
}

func serve(conn net.Conn, token string, rt *runtime) {
	var writeMu sync.Mutex
	emit := func(ev Event) {
		writeMu.Lock()
		defer writeMu.Unlock()
		_, _ = conn.Write(marshalLine(ev))
	}

	scanner := bufio.NewScanner(conn)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 8*1024*1024)

	authed := false
	for scanner.Scan() {
		var req Request
		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			emitReply(conn, &writeMu, Reply{Type: "reply", Error: "bad json"})
			continue
		}
		if req.Token != token {
			emitReply(conn, &writeMu, Reply{ID: req.ID, Type: "reply", Error: "bad token"})
			continue
		}
		authed = true
		switch req.Cmd {
		case "start":
			err := rt.start(req, emit)
			if err != nil {
				emit(Event{Type: "status", Event: "error", Error: err.Error()})
				emitReply(conn, &writeMu, Reply{ID: req.ID, Type: "reply", Error: err.Error()})
			} else {
				emitReply(conn, &writeMu, Reply{ID: req.ID, Type: "reply", OK: true})
			}
		case "stop":
			rt.stop()
			emit(Event{Type: "status", Event: "disconnected"})
			emitReply(conn, &writeMu, Reply{ID: req.ID, Type: "reply", OK: true})
		case "recover":
			err := rt.recoverRelay(emit)
			if err != nil {
				emitReply(conn, &writeMu, Reply{ID: req.ID, Type: "reply", Error: err.Error()})
			} else {
				emitReply(conn, &writeMu, Reply{ID: req.ID, Type: "reply", OK: true})
			}
		case "update_rules":
			err := rt.updateRules(req.RulesJSON, req.ExclusionsJSON)
			if err != nil {
				emitReply(conn, &writeMu, Reply{ID: req.ID, Type: "reply", Error: err.Error()})
			} else {
				emitReply(conn, &writeMu, Reply{ID: req.ID, Type: "reply", OK: true})
			}
		case "ping":
			emitReply(conn, &writeMu, Reply{ID: req.ID, Type: "reply", OK: true})
		default:
			emitReply(conn, &writeMu, Reply{ID: req.ID, Type: "reply", Error: "unknown cmd"})
		}
	}
	rt.stop()
	if authed {
		emit(Event{Type: "status", Event: "disconnected"})
	}
}

func emitReply(conn net.Conn, mu *sync.Mutex, r Reply) {
	if r.Type == "" {
		r.Type = "reply"
	}
	mu.Lock()
	defer mu.Unlock()
	_, _ = conn.Write(marshalLine(r))
}

func randomToken() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "fallback-token"
	}
	return hex.EncodeToString(b[:])
}
