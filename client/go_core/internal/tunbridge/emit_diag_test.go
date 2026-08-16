package tunbridge

import (
	"strings"
	"testing"
	"time"
)

func TestEmitDiagNeverUsesResultOK(t *testing.T) {
	var lines []string
	SetLogger(func(s string) { lines = append(lines, s) })
	t.Cleanup(func() { SetLogger(nil) })

	emitDiag("tcp", "api2.cursor.sh", "1.2.3.4", 443, "RELAY", "default", "stream_opened", "212.43.156.33", true, 100*time.Millisecond, "", 0, 0, 0)
	emitDiag("tcp", "api2.cursor.sh", "1.2.3.4", 443, "RELAY", "default", "transfer_done", "212.43.156.33", true, time.Second, "", 100, 1000, 2000)
	emitDiag("tcp", "api2.cursor.sh", "1.2.3.4", 443, "RELAY", "default", "stream_open_no_data", "212.43.156.33", false, 4*time.Second, "stream_open_no_data", 0, 12, 0)

	joined := strings.Join(lines, "\n")
	if strings.Contains(joined, "result=ok") {
		t.Fatalf("result=ok must not appear:\n%s", joined)
	}
	if !strings.Contains(joined, "result=stream_opened") {
		t.Fatalf("expected stream_opened:\n%s", joined)
	}
	if !strings.Contains(joined, "result=transfer_done") {
		t.Fatalf("expected transfer_done:\n%s", joined)
	}
	if !strings.Contains(joined, "result=stream_open_no_data") {
		t.Fatalf("expected stream_open_no_data:\n%s", joined)
	}
	if !strings.Contains(joined, "bytes_tx=12") || !strings.Contains(joined, "bytes_rx=0") {
		t.Fatalf("expected bytes counters:\n%s", joined)
	}
}
