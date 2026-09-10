package tunbridge

import "testing"

func TestMarkTrafficReadyOnce(t *testing.T) {
	resetTrafficReady()
	var lines []string
	SetLogger(func(msg string) { lines = append(lines, msg) })
	defer SetLogger(nil)

	markTrafficReady("RELAY")
	markTrafficReady("DIRECT")
	if len(lines) != 2 {
		t.Fatalf("want 2 logs (traffic_ready + TRAFFIC_READY), got %d: %v", len(lines), lines)
	}
	if lines[0] != "[vpn] traffic_ready via=RELAY" {
		t.Fatalf("log0=%q", lines[0])
	}
	if lines[1] != "[lifecycle] TRAFFIC_READY via=RELAY" {
		t.Fatalf("log1=%q", lines[1])
	}

	resetTrafficReady()
	markTrafficReady("DIRECT")
	if len(lines) != 4 ||
		lines[2] != "[vpn] traffic_ready via=DIRECT" ||
		lines[3] != "[lifecycle] TRAFFIC_READY via=DIRECT" {
		t.Fatalf("after reset got %v", lines)
	}
}
