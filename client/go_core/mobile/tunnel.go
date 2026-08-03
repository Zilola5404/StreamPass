package mobile

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/apernet/hysteria/core/v2/client"
	"streampass/go_core/internal/hyconfig"
	"streampass/go_core/internal/tunbridge"
)

// StatusCallback is the Kotlin-side callback interface expected by gomobile.
type StatusCallback interface {
	OnConnecting()
	OnConnected(relay string, pingMs int)
	OnDisconnected()
	OnError(message string)
}

var (
	tunnelMu sync.Mutex
	active   *tunnelRuntime
	prepared *tunnelRuntime
)

type tunnelRuntime struct {
	cancel     context.CancelFunc
	bridge     *tunbridge.Session
	hy         client.Client
	mtu        uint32
	relayLabel string
}

// PrepareRelay dials the Hysteria relay before Android brings up the TUN
// interface. Must be called while the relay is still reachable on the
// underlying network — if TUN default route is already active, QUIC
// handshake packets loop into the empty tunnel and time out.
// Returns an empty string on success, or an error message.
func PrepareRelay(relayHost string, relayPort int, connectionConfig string) string {
	StopTunnel()

	cfg, parsed, err := hyconfig.BuildClientConfig(connectionConfig, relayHost, relayPort)
	if err != nil {
		return err.Error()
	}

	start := time.Now()
	hyClient, _, err := client.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("hysteria connect: %w", err).Error()
	}
	_ = time.Since(start)

	relayLabel := relayHost
	if relayLabel == "" {
		relayLabel = parsed.ServerHost
	}

	tunnelMu.Lock()
	prepared = &tunnelRuntime{
		hy:         hyClient,
		mtu:        parsed.MTU,
		relayLabel: relayLabel,
	}
	tunnelMu.Unlock()
	return ""
}

// StartTunnel attaches the Android TUN fd to an active session. Call
// PrepareRelay first on Android so the relay handshake completes before
// VpnService routes all traffic into TUN.
func StartTunnel(fd int, relayHost string, relayPort int, connectionConfig string, cb StatusCallback) {
	go runTunnel(fd, relayHost, relayPort, connectionConfig, cb)
}

// StopTunnel stops the active tunnel session, if any.
func StopTunnel() {
	tunnelMu.Lock()
	if prepared != nil {
		prepared.close()
		prepared = nil
	}
	current := active
	active = nil
	tunnelMu.Unlock()
	if current != nil {
		current.close()
	}
}

func runTunnel(fd int, relayHost string, relayPort int, connectionConfig string, cb StatusCallback) {
	if cb != nil {
		cb.OnConnecting()
	}

	tunnelMu.Lock()
	relaySession := prepared
	prepared = nil
	tunnelMu.Unlock()

	var hyClient client.Client
	var mtu uint32 = hyconfig.DefaultMTU()
	var relayLabel string
	var pingMs int

	if relaySession != nil && relaySession.hy != nil {
		hyClient = relaySession.hy
		if relaySession.mtu > 0 {
			mtu = relaySession.mtu
		}
		relayLabel = relaySession.relayLabel
	} else {
		StopTunnel()

		cfg, parsed, err := hyconfig.BuildClientConfig(connectionConfig, relayHost, relayPort)
		if err != nil {
			emitError(cb, err)
			return
		}

		start := time.Now()
		var err2 error
		hyClient, _, err2 = client.NewClient(cfg)
		if err2 != nil {
			emitError(cb, fmt.Errorf("hysteria connect: %w", err2))
			return
		}
		pingMs = int(time.Since(start).Milliseconds())
		if parsed.MTU > 0 {
			mtu = parsed.MTU
		}
		relayLabel = relayHost
		if relayLabel == "" {
			relayLabel = parsed.ServerHost
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	bridge, err := tunbridge.Start(ctx, fd, hyClient, mtu)
	if err != nil {
		cancel()
		_ = hyClient.Close()
		emitError(cb, fmt.Errorf("tun bridge: %w", err))
		return
	}

	runtime := &tunnelRuntime{
		cancel:     cancel,
		bridge:     bridge,
		hy:         hyClient,
		mtu:        mtu,
		relayLabel: relayLabel,
	}
	tunnelMu.Lock()
	active = runtime
	tunnelMu.Unlock()

	if cb != nil {
		cb.OnConnected(relayLabel, pingMs)
	}

	<-ctx.Done()
}

func (r *tunnelRuntime) close() {
	if r.cancel != nil {
		r.cancel()
	}
	if r.bridge != nil {
		r.bridge.Close()
	}
	if r.hy != nil {
		_ = r.hy.Close()
	}
}

func emitError(cb StatusCallback, err error) {
	if cb != nil {
		cb.OnError(err.Error())
	}
}
