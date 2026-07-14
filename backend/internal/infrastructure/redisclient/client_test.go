package redisclient

import (
	"bufio"
	"context"
	"net"
	"strings"
	"testing"
	"time"
)

// startMockRedis starts a tiny in-process TCP server that understands just
// enough RESP2 to drive the Client's unit tests without a real Redis
// instance. It keeps an in-memory map so SET/GET/DEL/EXISTS behave
// consistently across a test.
func startMockRedis(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start mock redis listener: %v", err)
	}
	t.Cleanup(func() { ln.Close() })

	store := map[string]string{}

	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go handleMockConn(c, store)
		}
	}()

	return ln.Addr().String()
}

func handleMockConn(c net.Conn, store map[string]string) {
	defer c.Close()
	r := bufio.NewReader(c)
	for {
		args, err := readMockCommand(r)
		if err != nil {
			return
		}
		if len(args) == 0 {
			continue
		}
		reply := dispatchMockCommand(args, store)
		if _, err := c.Write([]byte(reply)); err != nil {
			return
		}
	}
}

func readMockCommand(r *bufio.Reader) ([]string, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return nil, err
	}
	line = strings.TrimRight(line, "\r\n")
	if len(line) == 0 || line[0] != '*' {
		return nil, nil
	}
	var n int
	for _, c := range line[1:] {
		n = n*10 + int(c-'0')
	}
	args := make([]string, n)
	for i := 0; i < n; i++ {
		lenLine, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		lenLine = strings.TrimRight(lenLine, "\r\n")
		var l int
		for _, c := range lenLine[1:] {
			l = l*10 + int(c-'0')
		}
		buf := make([]byte, l+2)
		if _, err := r.Read(buf); err != nil {
			return nil, err
		}
		args[i] = string(buf[:l])
	}
	return args, nil
}

func dispatchMockCommand(args []string, store map[string]string) string {
	switch strings.ToUpper(args[0]) {
	case "SET":
		store[args[1]] = args[2]
		return "+OK\r\n"
	case "GET":
		v, ok := store[args[1]]
		if !ok {
			return "$-1\r\n"
		}
		return "$" + itoa(len(v)) + "\r\n" + v + "\r\n"
	case "DEL":
		delete(store, args[1])
		return ":1\r\n"
	case "EXISTS":
		if _, ok := store[args[1]]; ok {
			return ":1\r\n"
		}
		return ":0\r\n"
	default:
		return "-ERR unsupported command in mock\r\n"
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

func TestClient_SetGetDelExists(t *testing.T) {
	addr := startMockRedis(t)
	client := New(Config{Addr: addr, DialTimeout: time.Second})
	ctx := context.Background()

	if err := client.Set(ctx, "foo", "bar", time.Minute); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	got, err := client.Get(ctx, "foo")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got != "bar" {
		t.Errorf("Get() = %q, want bar", got)
	}

	exists, err := client.Exists(ctx, "foo")
	if err != nil || !exists {
		t.Errorf("Exists() = (%v, %v), want (true, nil)", exists, err)
	}

	if err := client.Del(ctx, "foo"); err != nil {
		t.Fatalf("Del() error = %v", err)
	}

	if _, err := client.Get(ctx, "foo"); err == nil {
		t.Error("Get() after Del() expected NotFound error, got nil")
	}
}
