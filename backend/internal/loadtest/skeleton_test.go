package loadtest_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadtestSourcesPresent(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	for _, rel := range []string{
		"scripts/loadtest/main.go",
		"scripts/loadtest/k6-public.js",
		"scripts/LoadTest.ps1",
	} {
		data, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			t.Fatalf("missing %s: %v", rel, err)
		}
		if len(strings.TrimSpace(string(data))) == 0 {
			t.Fatalf("%s empty", rel)
		}
	}
	main := string(mustRead(t, filepath.Join(root, "scripts/loadtest/main.go")))
	for _, needle := range []string{"/api/v1/rules", "/api/v1/regions", "p99", "BL-032"} {
		if !strings.Contains(main, needle) {
			t.Fatalf("loadtest main missing %q", needle)
		}
	}
}

func TestLoadtestAgainstPublicAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("skip network loadtest in -short")
	}
	if os.Getenv("STREAMPASS_LOADTEST") == "0" {
		t.Skip("STREAMPASS_LOADTEST=0")
	}

	root := filepath.Join("..", "..", "..")
	base := os.Getenv("STREAMPASS_BASE_URL")
	if base == "" {
		base = "https://212-43-156-33.nip.io"
	}

	ctxTimeout := 45 * time.Second
	cmd := exec.Command("go", "run", "./scripts/loadtest",
		"-base", base,
		"-duration", "8s",
		"-rps", "20",
	)
	cmd.Dir = root
	cmd.Env = os.Environ()

	done := make(chan error, 1)
	go func() { done <- cmd.Run() }()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("loadtest failed: %v", err)
		}
	case <-time.After(ctxTimeout):
		_ = cmd.Process.Kill()
		t.Fatal("loadtest timed out")
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
