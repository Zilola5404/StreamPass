package hyconfig

import "testing"

func TestParseHysteria2URI(t *testing.T) {
	parsed, err := Parse(
		"hysteria2://test-auth@198.51.100.1:443/?obfs=salamander&obfs-password=test-obfs-secret&insecure=1#StreamPass",
		"",
		0,
	)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if parsed.ServerHost != "198.51.100.1:443" {
		t.Fatalf("server host = %q", parsed.ServerHost)
	}
	if parsed.Auth != "test-auth" {
		t.Fatalf("auth = %q", parsed.Auth)
	}
	if parsed.ObfsType != "salamander" {
		t.Fatalf("obfs = %q", parsed.ObfsType)
	}
	if parsed.ObfsPass != "test-obfs-secret" {
		t.Fatalf("obfs pass = %q", parsed.ObfsPass)
	}
	if !parsed.Insecure {
		t.Fatal("expected insecure=true for self-signed relay")
	}
}

func TestParseFallbackHostPort(t *testing.T) {
	parsed, err := Parse("", "relay.example.com", 8443)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if parsed.ServerHost != "relay.example.com:8443" {
		t.Fatalf("server host = %q", parsed.ServerHost)
	}
}
