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
)

type tunnelRuntime struct {
	cancel context.CancelFunc
	bridge *tunbridge.Session
	hy     client.Client
}

// StartTunnel starts the Hysteria2 tunnel on the supplied Android TUN fd.
// connectionConfig is a hysteria2:// URI from GET /servers; relayHost and
// relayPort are kept as fallback when the URI is empty.
func StartTunnel(fd int, relayHost string, relayPort int, connectionConfig string, cb StatusCallback) {
	go runTunnel(fd, relayHost, relayPort, connectionConfig, cb)
}

// StopTunnel stops the active tunnel session, if any.
func StopTunnel() {
	tunnelMu.Lock()
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

	StopTunnel()

	cfg, parsed, err := hyconfig.BuildClientConfig(connectionConfig, relayHost, relayPort)
	if err != nil {
		emitError(cb, err)
		return
	}

	start := time.Now()
	hyClient, _, err := client.NewClient(cfg)
	if err != nil {
		emitError(cb, fmt.Errorf("hysteria connect: %w", err))
		return
	}
	pingMs := int(time.Since(start).Milliseconds())

	ctx, cancel := context.WithCancel(context.Background())
	bridge, err := tunbridge.Start(ctx, fd, hyClient, parsed.MTU)
	if err != nil {
		cancel()
		_ = hyClient.Close()
		emitError(cb, fmt.Errorf("tun bridge: %w", err))
		return
	}

	runtime := &tunnelRuntime{
		cancel: cancel,
		bridge: bridge,
		hy:     hyClient,
	}
	tunnelMu.Lock()
	active = runtime
	tunnelMu.Unlock()

	relayLabel := relayHost
	if relayLabel == "" {
		relayLabel = parsed.ServerHost
	}
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
