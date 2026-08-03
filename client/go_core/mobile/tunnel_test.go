package mobile_test

import (
	"os"
	"testing"

	"streampass/go_core/mobile"
)

func TestPrepareRelayIntegration(t *testing.T) {
	uri := os.Getenv("STREAMPASS_RELAY_URI")
	if uri == "" {
		t.Skip("STREAMPASS_RELAY_URI not set")
	}

	mobile.StopTunnel()
	errMsg := mobile.PrepareRelay("", 0, uri)
	if errMsg != "" {
		t.Fatalf("PrepareRelay: %s", errMsg)
	}
	mobile.StopTunnel()
}
