package mobile

import (
	"context"
	"fmt"
	"net/netip"
	"sync"
	"time"

	"github.com/apernet/hysteria/core/v2/client"
	"streampass/go_core/internal/decision"
	"streampass/go_core/internal/hyconfig"
	"streampass/go_core/internal/protect"
	"streampass/go_core/internal/tunbridge"
)

// StatusCallback is the Kotlin-side callback interface expected by gomobile.
type StatusCallback interface {
	OnConnecting()
	OnConnected(relay string, pingMs int)
	OnDisconnected()
	OnError(message string)
}

// SocketProtector is implemented by Android VpnService.protect(fd).
// Must be set before PrepareRelay so the Hysteria QUIC underlay bypasses TUN.
type SocketProtector interface {
	Protect(fd int) bool
}

type socketProtectorAdapter struct {
	p SocketProtector
}

func (a socketProtectorAdapter) Protect(fd int) bool {
	if a.p == nil {
		return false
	}
	return a.p.Protect(fd)
}

// SetSocketProtector installs or clears the platform socket protector.
func SetSocketProtector(p SocketProtector) {
	if p == nil {
		protect.Clear()
		return
	}
	protect.Set(socketProtectorAdapter{p: p})
}

var (
	tunnelMu   sync.Mutex
	active     *tunnelRuntime
	prepared   *tunnelRuntime
	runTunnelWg sync.WaitGroup
)

type tunnelRuntime struct {
	cancel     context.CancelFunc
	bridge     *tunbridge.Session
	hy         client.Client
	mtu        uint32
	relayLabel string
	pingMs     int
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
	pingMs := int(time.Since(start).Milliseconds())

	relayLabel := relayHost
	if relayLabel == "" {
		relayLabel = parsed.ServerHost
	}

	tunnelMu.Lock()
	prepared = &tunnelRuntime{
		hy:         hyClient,
		mtu:        parsed.MTU,
		relayLabel: relayLabel,
		pingMs:     pingMs,
	}
	tunnelMu.Unlock()
	return ""
}

func takePreparedSession() *tunnelRuntime {
	tunnelMu.Lock()
	defer tunnelMu.Unlock()
	s := prepared
	prepared = nil
	return s
}

// StartTunnel attaches the Android TUN fd to an active session. Call
// PrepareRelay first on Android so the relay handshake completes before
// VpnService routes all traffic into TUN.
// rulesJSON / exclusionsJSON are optional payloads from GET /api/v1/rules
// and local user exclusions (BL-005 Decision Engine).
func StartTunnel(fd int, relayHost string, relayPort int, connectionConfig string, rulesJSON string, exclusionsJSON string, cb StatusCallback) {
	runTunnelWg.Add(1)
	go func() {
		defer runTunnelWg.Done()
		runTunnel(fd, relayHost, relayPort, connectionConfig, rulesJSON, exclusionsJSON, cb)
	}()
}

// DecideRoute evaluates routing for diagnostics (host may be empty when only IP known).
func DecideRoute(rulesJSON, exclusionsJSON, host, ip string) string {
	engine, err := decision.NewEngineFromJSON(rulesJSON, exclusionsJSON)
	if err != nil {
		return string(decision.DefaultMode)
	}
	target := decision.Target{Host: host}
	if addr, parseErr := parseIP(ip); parseErr == nil {
		target.IP = addr
	}
	return string(engine.Decide(target))
}

func parseIP(raw string) (netip.Addr, error) {
	return netip.ParseAddr(raw)
}

// UpdateRules hot-reloads the rule set on the active tunnel (BL-006).
// Returns empty string on success, or an error message.
func UpdateRules(rulesJSON, exclusionsJSON string) string {
	tunnelMu.Lock()
	rt := active
	tunnelMu.Unlock()
	if rt == nil || rt.bridge == nil {
		return "no active tunnel"
	}
	if err := rt.bridge.UpdateEngine(rulesJSON, exclusionsJSON); err != nil {
		return err.Error()
	}
	return ""
}

// ActiveRulesVersion returns the rule set version applied to the active tunnel.
func ActiveRulesVersion() int {
	tunnelMu.Lock()
	rt := active
	tunnelMu.Unlock()
	if rt == nil || rt.bridge == nil {
		return 0
	}
	return rt.bridge.RulesVersion()
}

// StopTunnel stops the active tunnel session, if any.
func StopTunnel() {
	stopTunnelSessions()
	runTunnelWg.Wait()
}

func stopTunnelSessions() {
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

func runTunnel(fd int, relayHost string, relayPort int, connectionConfig string, rulesJSON, exclusionsJSON string, cb StatusCallback) {
	defer func() {
		if r := recover(); r != nil {
			emitError(cb, fmt.Errorf("tunnel panic: %v", r))
		}
	}()

	relaySession := takePreparedSession()

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
		pingMs = relaySession.pingMs
	} else {
		stopTunnelSessions()

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
	engine, err := decision.NewAtomicEngineFromJSON(rulesJSON, exclusionsJSON)
	if err != nil {
		cancel()
		_ = hyClient.Close()
		emitError(cb, fmt.Errorf("decision engine: %w", err))
		return
	}
	bridge, err := tunbridge.Start(ctx, fd, hyClient, mtu, engine)
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
