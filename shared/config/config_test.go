package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}
	return path
}

func TestFileLoader_ResolvesEnvPlaceholders(t *testing.T) {
	os.Setenv("SP_TEST_DB_HOST", "db.internal")
	defer os.Unsetenv("SP_TEST_DB_HOST")

	path := writeTempConfig(t, `
database:
  host: ${SP_TEST_DB_HOST}
  port: 5432
server:
  http_port: 8080
  debug: true
jwt:
  access_ttl: 15m
`)

	cfg, err := NewFileLoader(path).Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	host, err := cfg.String("database.host")
	if err != nil || host != "db.internal" {
		t.Errorf("database.host = %q, err = %v, want db.internal", host, err)
	}

	port, err := cfg.Int("database.port")
	if err != nil || port != 5432 {
		t.Errorf("database.port = %d, err = %v, want 5432", port, err)
	}

	debug, err := cfg.Bool("server.debug")
	if err != nil || !debug {
		t.Errorf("server.debug = %v, err = %v, want true", debug, err)
	}

	ttl, err := cfg.Duration("jwt.access_ttl")
	if err != nil || ttl != 15*time.Minute {
		t.Errorf("jwt.access_ttl = %v, err = %v, want 15m", ttl, err)
	}
}

func TestFileLoader_MissingKeyReturnsError(t *testing.T) {
	path := writeTempConfig(t, "server:\n  http_port: 8080\n")
	cfg, err := NewFileLoader(path).Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if _, err := cfg.String("nope.nope"); err == nil {
		t.Error("expected error for missing key, got nil")
	}
}

func TestFileLoader_DefaultsFallBack(t *testing.T) {
	path := writeTempConfig(t, "server:\n  http_port: 8080\n")
	cfg, err := NewFileLoader(path).Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got := cfg.IntOr("server.missing", 42); got != 42 {
		t.Errorf("IntOr default = %d, want 42", got)
	}
	if got := cfg.StringOr("server.missing", "fallback"); got != "fallback" {
		t.Errorf("StringOr default = %q, want fallback", got)
	}
}

func TestFileLoader_UnresolvedEnvVarResolvesToEmptyString(t *testing.T) {
	path := writeTempConfig(t, "secret:\n  value: ${SP_TEST_UNSET_VAR}\n")
	cfg, err := NewFileLoader(path).Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	v, err := cfg.String("secret.value")
	if err != nil {
		t.Fatalf("String() error = %v", err)
	}
	if v != "" {
		t.Errorf("unresolved placeholder = %q, want empty string (never the literal placeholder text)", v)
	}
}
