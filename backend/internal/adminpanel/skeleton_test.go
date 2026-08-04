package adminpanel_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAdminPanelFiles(t *testing.T) {
	root := filepath.Join("..", "..", "..", "admin")
	required := []string{"index.html", "app.js", "styles.css", "README.md"}
	for _, name := range required {
		path := filepath.Join(root, name)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("missing %s: %v", path, err)
		}
		if len(strings.TrimSpace(string(data))) == 0 {
			t.Fatalf("%s is empty", path)
		}
	}

	html := string(mustRead(t, filepath.Join(root, "index.html")))
	for _, needle := range []string{
		"view-login", "view-app", "login-submit", "health-check",
		"users-body", "relays-body", "relay-form",
		"rules-json", "rules-publish", "config-form",
		"latest_client_version", "client_download_url",
	} {
		if !strings.Contains(html, needle) {
			t.Fatalf("index.html missing %q", needle)
		}
	}

	js := string(mustRead(t, filepath.Join(root, "app.js")))
	for _, needle := range []string{
		"X-Admin-Key", "/servers/all", "/users", "DELETE",
		"loadUsers", "loadRelays", "loadRules", "loadConfig",
		"POST", "/rules", "/config", "publishRules",
	} {
		if !strings.Contains(js, needle) {
			t.Fatalf("app.js missing %q", needle)
		}
	}
}

func TestCaddyServesAdminPath(t *testing.T) {
	caddy := string(mustRead(t, filepath.Join("..", "..", "..", "Caddyfile")))
	if !strings.Contains(caddy, "handle_path /admin/*") {
		t.Fatal("Caddyfile must expose /admin/*")
	}
	if !strings.Contains(caddy, "/srv/admin") {
		t.Fatal("Caddyfile must root admin static files")
	}
}

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
