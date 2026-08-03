package hyconfig

import (
	"bufio"
	"context"
	"io"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/apernet/hysteria/core/v2/client"
)

const defaultTestConnectionConfig = "hysteria2://streampass-secure-auth@212.43.159.198:443/?obfs=salamander&obfs-password=streampass-relay-2024#StreamPass"

func integrationConnectionConfig() string {
	if v := os.Getenv("STREAMPASS_RELAY_URI"); v != "" {
		return v
	}
	return defaultTestConnectionConfig
}

// TestIntegrationHysteriaConnect verifies the client can authenticate to the live relay.
// Run without -short: go test -v -run TestIntegrationHysteriaConnect ./internal/hyconfig/ -timeout 2m
func TestIntegrationHysteriaConnect(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test skipped in short mode")
	}
	if os.Getenv("STREAMPASS_RELAY_TEST") == "0" {
		t.Skip("STREAMPASS_RELAY_TEST=0")
	}

	cfg, _, err := BuildClientConfig(integrationConnectionConfig(), "", 0)
	if err != nil {
		t.Fatalf("build config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	type result struct {
		hy client.Client
		err error
	}
	ch := make(chan result, 1)
	go func() {
		hy, _, err := client.NewClient(cfg)
		ch <- result{hy: hy, err: err}
	}()

	var hy client.Client
	select {
	case <-ctx.Done():
		t.Fatal("hysteria handshake timed out — relay unreachable or misconfigured")
	case r := <-ch:
		if r.err != nil {
			t.Fatalf("hysteria connect: %v", r.err)
		}
		hy = r.hy
	}
	defer hy.Close()

	t.Log("hysteria handshake OK")
}

// TestIntegrationHysteriaForeignIP dials ifconfig.me through the relay and checks the response body.
func TestIntegrationHysteriaForeignIP(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test skipped in short mode")
	}
	if os.Getenv("STREAMPASS_RELAY_TEST") == "0" {
		t.Skip("STREAMPASS_RELAY_TEST=0")
	}

	cfg, _, err := BuildClientConfig(integrationConnectionConfig(), "", 0)
	if err != nil {
		t.Fatalf("build config: %v", err)
	}

	hy, _, err := client.NewClient(cfg)
	if err != nil {
		t.Fatalf("hysteria connect: %v", err)
	}
	defer hy.Close()

	conn, err := hy.TCP("ifconfig.me:80")
	if err != nil {
		t.Fatalf("tcp dial via relay: %v", err)
	}
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(20 * time.Second))
	req := "GET /ip HTTP/1.1\r\nHost: ifconfig.me\r\nConnection: close\r\n\r\n"
	if _, err := io.WriteString(conn, req); err != nil {
		t.Fatalf("write request: %v", err)
	}

	scanner := bufio.NewScanner(conn)
	var body string
	inBody := false
	for scanner.Scan() {
		line := scanner.Text()
		if inBody {
			body = strings.TrimSpace(line)
			break
		}
		if line == "" {
			inBody = true
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("read response: %v", err)
	}
	if body == "" {
		t.Fatal("empty IP in response from ifconfig.me via relay")
	}

	t.Logf("foreign IP via relay: %s", body)

	// Relay is in DE/EU — expect non-empty public IP (not RFC1918).
	if strings.HasPrefix(body, "10.") || strings.HasPrefix(body, "192.168.") || strings.HasPrefix(body, "127.") {
		t.Fatalf("unexpected private/loopback IP: %s", body)
	}

	// Sanity: response looks like an IP address.
	if !strings.Contains(body, ".") {
		t.Fatalf("response does not look like IPv4: %q", body)
	}
}

// TestIntegrationHysteriaHTTPHead is a lighter check using HTTP through TCP tunnel.
func TestIntegrationHysteriaHTTPHead(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test skipped in short mode")
	}
	if os.Getenv("STREAMPASS_RELAY_TEST") == "0" {
		t.Skip("STREAMPASS_RELAY_TEST=0")
	}

	cfg, _, err := BuildClientConfig(integrationConnectionConfig(), "", 0)
	if err != nil {
		t.Fatalf("build config: %v", err)
	}

	hy, _, err := client.NewClient(cfg)
	if err != nil {
		t.Fatalf("hysteria connect: %v", err)
	}
	defer hy.Close()

	conn, err := hy.TCP("example.com:80")
	if err != nil {
		t.Fatalf("tcp dial: %v", err)
	}
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(15 * time.Second))
	if _, err := io.WriteString(conn, "HEAD / HTTP/1.1\r\nHost: example.com\r\nConnection: close\r\n\r\n"); err != nil {
		t.Fatalf("write: %v", err)
	}

	buf := make([]byte, 512)
	n, err := conn.Read(buf)
	if err != nil && err != io.EOF {
		t.Fatalf("read: %v", err)
	}
	resp := string(buf[:n])
	if !strings.Contains(resp, "HTTP/") {
		t.Fatalf("unexpected response: %q", resp)
	}
	if !strings.Contains(resp, "200") && !strings.Contains(resp, "301") && !strings.Contains(resp, "302") {
		t.Logf("warning: non-200 HEAD response: %s", strings.Split(resp, "\n")[0])
	}
	_ = net.ParseIP("127.0.0.1") // ensure net imported
}
