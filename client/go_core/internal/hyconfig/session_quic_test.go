package hyconfig

import "testing"

func TestSessionQUICIdleIsLongLived(t *testing.T) {
	if SessionMaxIdleTimeout < SessionKeepAlivePeriod*2 {
		t.Fatalf("MaxIdleTimeout=%v must allow KeepAlive=%v", SessionMaxIdleTimeout, SessionKeepAlivePeriod)
	}
	// Handshake candidate budget must stay short; session idle must not equal it.
	if SessionMaxIdleTimeout <= handshakePerCandidate {
		t.Fatalf("session idle %v must exceed handshake candidate %v (issue #2)", SessionMaxIdleTimeout, handshakePerCandidate)
	}
	if SessionMaxIdleTimeout < 60e9 { // 60s
		t.Fatalf("session MaxIdleTimeout too short for idle/sleep: %v", SessionMaxIdleTimeout)
	}
}
