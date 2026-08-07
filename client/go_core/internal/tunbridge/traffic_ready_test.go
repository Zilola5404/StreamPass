package tunbridge

import "testing"

func TestMarkTrafficReadyOnce(t *testing.T) {
	resetTrafficReady()
	var lines []string
	SetLogger(func(msg string) { lines = append(lines, msg) })
	defer SetLogger(nil)

	markTrafficReady("RELAY")
	markTrafficReady("DIRECT")
	if len(lines) != 1 {
		t.Fatalf("want 1 traffic_ready log, got %d: %v", len(lines), lines)
	}
	if lines[0] != "[vpn] traffic_ready via=RELAY" {
		t.Fatalf("log=%q", lines[0])
	}

	resetTrafficReady()
	markTrafficReady("DIRECT")
	if len(lines) != 2 || lines[1] != "[vpn] traffic_ready via=DIRECT" {
		t.Fatalf("after reset got %v", lines)
	}
}
