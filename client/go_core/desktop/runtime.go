//go:build windows

package main

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/apernet/hysteria/core/v2/client"

	"streampass/go_core/internal/decision"
	"streampass/go_core/internal/dnscache"
	"streampass/go_core/internal/hyconfig"
	"streampass/go_core/internal/protect"
	"streampass/go_core/internal/tunbridge"
)

type runtime struct {
	mu     sync.Mutex
	cancel context.CancelFunc
	bridge *tunbridge.Session
	hy     client.Client
}

func (r *runtime) start(req Request, emit func(Event)) error {
	r.stop()

	logFn := func(message string) {
		emit(Event{Type: "log", Message: message})
	}
	tunbridge.SetLogger(logFn)
	dnscache.SetLogger(logFn)

	emit(Event{Type: "log", Message: "[vpn] CONNECT_REQUESTED platform=windows"})
	emit(Event{Type: "status", Event: "connecting", Relay: req.RelayHost})
	emit(Event{Type: "log", Message: fmt.Sprintf("[vpn] STAGE mtu=%d networkMode=%s", req.MTU, req.NetworkMode)})

	if req.ConnectionConfig != "" && strings.Contains(req.ConnectionConfig, "insecure=1") {
		emit(Event{Type: "log", Message: "[vpn] WARN insecure=1 in connection_config — forbidden for Windows production (TLS pin required before ship)"})
	}

	// Leftover StreamPass 0.0.0.0/0 from a crash blackholes everything — clear first.
	if err := protect.ClearStaleTunnelDefaultRoute(); err != nil {
		emit(Event{Type: "log", Message: fmt.Sprintf("[vpn] stale route cleanup: %v (need Admin if sites still broken)", err)})
	} else {
		emit(Event{Type: "log", Message: "[vpn] stale StreamPass default route cleared"})
	}

	// Bind underlay to Wi‑Fi/Ethernet BEFORE Hysteria dial (Android: protect before PrepareRelay).
	ifIdx, ifName, err := protect.BindPhysicalUnderlay()
	if err != nil {
		return fmt.Errorf("underlay interface: %w", err)
	}
	emit(Event{Type: "log", Message: fmt.Sprintf("[vpn] UNDERLAY_IF index=%d name=%s (before relay)", ifIdx, ifName)})

	var hyClient client.Client
	var pingMs int
	relayLabel := req.RelayHost

	if req.ConnectionConfig != "" && req.NetworkMode != "direct_test" {
		emit(Event{Type: "log", Message: fmt.Sprintf("[vpn] RELAY_CONNECTING host=%s port=%d", req.RelayHost, req.RelayPort)})
		result, err := hyconfig.ConnectWithFallback(req.ConnectionConfig, req.RelayHost, req.RelayPort)
		if err != nil {
			emit(Event{Type: "log", Message: fmt.Sprintf("[vpn] RELAY_CONNECT failed: %v — continuing DIRECT-only", err)})
		} else {
			hyClient = result.Client
			pingMs = result.PingMs
			if relayLabel == "" {
				relayLabel = result.Parsed.ServerHost
			}
			emit(Event{Type: "log", Message: fmt.Sprintf("[vpn] RELAY_CONNECTED via=%s/%s pingMs=%d", result.Candidate.Network, result.Candidate.Host, pingMs)})
		}
	} else {
		emit(Event{Type: "log", Message: "[vpn] RELAY skipped (direct_test or empty config)"})
	}

	mtu := uint32(1400)
	if req.MTU >= 1200 && req.MTU <= 1500 {
		mtu = uint32(req.MTU)
	}

	engine, err := decision.NewAtomicEngineFromJSON(req.RulesJSON, req.ExclusionsJSON)
	if err != nil {
		if hyClient != nil {
			_ = hyClient.Close()
		}
		return fmt.Errorf("decision engine: %w", err)
	}
	emit(Event{Type: "log", Message: fmt.Sprintf("[vpn] RULES_LOADED applied=true bytes=%d", len(req.RulesJSON))})
	switch req.NetworkMode {
	case "full_relay":
		engine.SetForceMode(decision.ModeRelay)
	case "direct_test":
		engine.SetForceMode(decision.ModeDirect)
	default:
		engine.SetForceMode("")
	}
	dnscache.SetRouteHint(func(host string) (rule, route, reason string) {
		d := engine.DecideDetailed(decision.Target{Host: host})
		return d.Rule, string(d.Mode), d.Reason
	})

	ctx, cancel := context.WithCancel(context.Background())
	blockUDP443 := req.BlockUDP443 || req.NetworkMode == "tcp_only"
	bridge, err := tunbridge.StartDesktop(ctx, hyClient, mtu, engine, relayLabel, tunbridge.Options{
		BlockUDP443: blockUDP443,
	})
	if err != nil {
		cancel()
		dnscache.SetRouteHint(nil)
		if hyClient != nil {
			_ = hyClient.Close()
		}
		return fmt.Errorf("wintun: %w", err)
	}

	r.mu.Lock()
	r.cancel = cancel
	r.bridge = bridge
	r.hy = hyClient
	r.mu.Unlock()

	emit(Event{Type: "status", Event: "connected", Relay: relayLabel, PingMs: pingMs})
	return nil
}

func (r *runtime) updateRules(rulesJSON, exclusionsJSON string) error {
	r.mu.Lock()
	bridge := r.bridge
	r.mu.Unlock()
	if bridge == nil {
		return fmt.Errorf("no active tunnel")
	}
	return bridge.UpdateEngine(rulesJSON, exclusionsJSON)
}

func (r *runtime) stop() {
	r.mu.Lock()
	cancel := r.cancel
	bridge := r.bridge
	hy := r.hy
	r.cancel = nil
	r.bridge = nil
	r.hy = nil
	r.mu.Unlock()

	dnscache.SetRouteHint(nil)
	if cancel != nil {
		cancel()
	}
	if bridge != nil {
		bridge.Close()
	}
	if hy != nil {
		_ = hy.Close()
	}
}
