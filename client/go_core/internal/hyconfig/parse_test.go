package hyconfig

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
)

func TestLoopbackRelayUsesCertificatePin(t *testing.T) {
	leaf := []byte("local relay certificate")
	fingerprint := sha256.Sum256(leaf)
	pin := hex.EncodeToString(fingerprint[:])
	local, _, err := BuildClientConfig("hysteria2://test@127.0.0.1:8443/?pinSHA256="+pin, "", 0)
	if err != nil {
		t.Fatal(err)
	}
	verify := local.TLSConfig.VerifyPeerCertificate
	if !local.TLSConfig.InsecureSkipVerify || verify == nil {
		t.Fatal("loopback must use the supplied certificate pin")
	}
	if err := verify([][]byte{leaf}, nil); err != nil {
		t.Fatal(err)
	}
	if err := verify([][]byte{[]byte("different certificate")}, nil); err == nil {
		t.Fatal("accepted a different certificate")
	}
	remote, _, err := BuildClientConfig("hysteria2://test@192.0.2.1:8443/?pinSHA256="+pin, "", 0)
	if err != nil {
		t.Fatal(err)
	}
	if remote.TLSConfig.InsecureSkipVerify {
		t.Fatal("remote relay must keep system certificate verification")
	}
}

func TestRequireVerifiedTLS(t *testing.T) {
	const endpoint = "hysteria2://test@relay.example:443/"
	for _, tc := range []struct {
		name, uri string
		valid     bool
	}{
		{"missing", "", false},
		{"implicit insecure", endpoint, false},
		{"explicit insecure", endpoint + "?insecure=true", false},
		{"public certificate", endpoint + "?insecure=false", true},
		{"bad pin", endpoint + "?pinSHA256=invalid", false},
		{"valid pin format", endpoint + "?pinSHA256=" + strings.Repeat("ab", 32), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := RequireVerifiedTLS(tc.uri, "", 0); (err == nil) != tc.valid {
				t.Fatalf("valid=%v, err=%v", tc.valid, err)
			}
		})
	}
}

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

func TestParsePinSHA256ForcesSecure(t *testing.T) {
	parsed, err := Parse(
		"hysteria2://auth@198.51.100.1:443/?sni=relay.example&pinSHA256=aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899",
		"",
		0,
	)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if parsed.Insecure {
		t.Fatal("pinSHA256 must force Insecure=false")
	}
	if parsed.SNI != "relay.example" {
		t.Fatalf("sni = %q", parsed.SNI)
	}
	if parsed.PinSHA256 == "" {
		t.Fatal("expected pinSHA256")
	}
}

func TestParseUnsupportedScheme(t *testing.T) {
	_, err := Parse("https://example.com/secret-token", "", 0)
	if err == nil {
		t.Fatal("expected unsupported scheme error")
	}
	if got := err.Error(); got != `unsupported scheme "https"` {
		t.Fatalf("error must not include URI body, got %q", got)
	}
}

func TestNormalizeCertHash(t *testing.T) {
	hexForm := normalizeCertHash("AA:BB:cc:DD")
	if hexForm != "aabbccdd" {
		t.Fatalf("hex normalize = %q", hexForm)
	}
}
