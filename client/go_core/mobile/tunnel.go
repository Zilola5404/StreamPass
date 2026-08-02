package mobile

// StatusCallback is the Kotlin-side callback interface expected by gomobile.
type StatusCallback interface {
	OnConnecting()
	OnConnected(relay string, pingMs int)
	OnDisconnected()
	OnError(message string)
}

// StartTunnel starts the tunnel using the supplied file descriptor and relay data.
// This is a minimal stub intentionally kept compile-safe so the Android bridge can
// be wired and tested without a full transport implementation yet.
func StartTunnel(fd int, relayHost string, relayPort int, authPassword string, cb StatusCallback) {
	if cb != nil {
		cb.OnConnecting()
	}
	if cb != nil {
		cb.OnError("Go core tunnel binding is present but the transport implementation is still stubbed")
	}
}

// StopTunnel stops the active tunnel.
func StopTunnel() {}
