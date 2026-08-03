package mobile

import "testing"

func TestTakePreparedSession_clearsPrepared(t *testing.T) {
	StopTunnel()
	prepared = &tunnelRuntime{relayLabel: "nl-native-1", pingMs: 68}

	got := takePreparedSession()
	if got == nil || got.pingMs != 68 {
		t.Fatalf("takePreparedSession() = %+v, want pingMs=68", got)
	}

	tunnelMu.Lock()
	left := prepared
	tunnelMu.Unlock()
	if left != nil {
		t.Fatal("prepared should be nil after takePreparedSession")
	}
}

func TestStopTunnel_clearsPrepared(t *testing.T) {
	StopTunnel()
	prepared = &tunnelRuntime{relayLabel: "test", pingMs: 42}
	StopTunnel()

	tunnelMu.Lock()
	left := prepared
	tunnelMu.Unlock()
	if left != nil {
		t.Fatal("StopTunnel should clear prepared session")
	}
}
