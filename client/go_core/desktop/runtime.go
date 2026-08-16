//go:build windows

package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/apernet/hysteria/core/v2/client"

	"streampass/go_core/internal/decision"
	"streampass/go_core/internal/dnscache"
	"streampass/go_core/internal/hyconfig"
	"streampass/go_core/internal/protect"
	"streampass/go_core/internal/tunbridge"
)

// startEngineBudget: cmd=start must reply with ENGINE_STARTED well under Flutter's 45s RPC timeout.
// Hysteria candidate dial runs AFTER the RPC reply (separate lifecycle).
const startEngineBudget = 20 * time.Second

type runtime struct {
	mu          sync.Mutex
	cancel      context.CancelFunc
	relayCancel context.CancelFunc
	bridge      *tunbridge.Session
	hy          client.Client
}

func (r *runtime) start(req Request, emit func(Event)) error {
	startAt := time.Now()
	stage := "init"
	defer func() {
		if rec := recover(); rec != nil {
			emit(Event{Type: "log", Message: fmt.Sprintf(
				"[START_TIMEOUT] last_stage=%s elapsed_ms=%d panic=%v",
				stage, time.Since(startAt).Milliseconds(), rec,
			)})
			panic(rec)
		}
	}()

	r.stop()

	logFn := func(message string) {
		emit(Event{Type: "log", Message: message})
	}
	tunbridge.SetLogger(logFn)
	dnscache.SetLogger(logFn)

	emit(Event{Type: "log", Message: "[CONNECT] platform=windows"})
	emit(Event{Type: "status", Event: "connecting", Relay: req.RelayHost})
	emit(Event{Type: "log", Message: fmt.Sprintf("[START] command_received mtu=%d networkMode=%s", req.MTU, req.NetworkMode)})

	if req.ConnectionConfig != "" && strings.Contains(req.ConnectionConfig, "insecure=1") {
		emit(Event{Type: "log", Message: "[vpn] WARN insecure=1 in connection_config — forbidden for Windows production (TLS pin required before ship)"})
	}

	deadline := time.Now().Add(startEngineBudget)
	stage = "stale_route"
	if err := protect.ClearStaleTunnelDefaultRoute(); err != nil {
		emit(Event{Type: "log", Message: fmt.Sprintf("[vpn] stale route cleanup: %v (need Admin if sites still broken)", err)})
	} else {
		emit(Event{Type: "log", Message: "[vpn] stale StreamPass default route cleared"})
	}
	if time.Now().After(deadline) {
		return fmt.Errorf("ENGINE_START timeout at %s (%s)", stage, startEngineBudget)
	}

	stage = "underlay_bind"
	ifIdx, ifName, err := protect.BindPhysicalUnderlay()
	if err != nil {
		return fmt.Errorf("underlay interface: %w", err)
	}
	emit(Event{Type: "log", Message: fmt.Sprintf("[vpn] UNDERLAY_IF index=%d name=%s (before engine)", ifIdx, ifName)})

	mtu := uint32(1400)
	if req.MTU >= 1200 && req.MTU <= 1500 {
		mtu = uint32(req.MTU)
	}

	stage = "decision_engine"
	engine, err := decision.NewAtomicEngineFromJSON(req.RulesJSON, req.ExclusionsJSON)
	if err != nil {
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

	relayLabel := req.RelayHost
	ctx, cancel := context.WithCancel(context.Background())
	blockUDP443 := req.BlockUDP443 || req.NetworkMode == "tcp_only"

	// ENGINE_STARTED first — Hysteria handshake is NOT part of cmd=start.
	stage = "wintun"
	bridge, err := tunbridge.StartDesktop(ctx, nil, mtu, engine, relayLabel, tunbridge.Options{
		BlockUDP443: blockUDP443,
	})
	if err != nil {
		cancel()
		dnscache.SetRouteHint(nil)
		emit(Event{Type: "log", Message: fmt.Sprintf(
			"[START_TIMEOUT] last_stage=%s elapsed_ms=%d error=%v",
			stage, time.Since(startAt).Milliseconds(), err,
		)})
		return fmt.Errorf("wintun: %w", err)
	}

	r.mu.Lock()
	r.cancel = cancel
	r.bridge = bridge
	r.hy = nil
	r.mu.Unlock()

	stage = "engine_started"
	emit(Event{Type: "log", Message: fmt.Sprintf(
		"[ENGINE] started elapsed_ms=%d underlay=%s",
		time.Since(startAt).Milliseconds(), ifName,
	)})
	emit(Event{Type: "status", Event: "connected", Relay: relayLabel, PingMs: 0})
	emit(Event{Type: "log", Message: "[RPC] start_response_sent result=ENGINE_STARTED"})

	wantRelay := req.ConnectionConfig != "" && req.NetworkMode != "direct_test"
	if wantRelay {
		relayCtx, relayCancel := context.WithCancel(ctx)
		r.mu.Lock()
		r.relayCancel = relayCancel
		r.mu.Unlock()
		go r.attachRelayAsync(relayCtx, req, emit, bridge, relayLabel)
	} else {
		emit(Event{Type: "log", Message: "[RELAY] skipped (direct_test or empty config)"})
	}
	return nil
}

func (r *runtime) attachRelayAsync(ctx context.Context, req Request, emit func(Event), bridge *tunbridge.Session, relayLabel string) {
	emit(Event{Type: "log", Message: fmt.Sprintf("[RELAY] handshake_started host=%s port=%d", req.RelayHost, req.RelayPort)})
	type result struct {
		out *hyconfig.ConnectResult
		err error
	}
	ch := make(chan result, 1)
	go func() {
		out, err := hyconfig.ConnectWithFallback(req.ConnectionConfig, req.RelayHost, req.RelayPort)
		ch <- result{out: out, err: err}
	}()

	select {
	case <-ctx.Done():
		emit(Event{Type: "log", Message: "[RELAY] handshake_cancelled"})
		return
	case res := <-ch:
		if res.err != nil {
			emit(Event{Type: "log", Message: fmt.Sprintf("[TRAFFIC_FAILED] reason=relay_unavailable error=%v", res.err)})
			emit(Event{Type: "log", Message: "[RELAY] unavailable — DIRECT path remains active"})
			return
		}
		if ctx.Err() != nil {
			_ = res.out.Client.Close()
			return
		}
		label := relayLabel
		if label == "" && res.out.Parsed != nil {
			label = res.out.Parsed.ServerHost
		}
		bridge.SetHysteriaClient(res.out.Client, label)
		r.mu.Lock()
		old := r.hy
		r.hy = res.out.Client
		r.mu.Unlock()
		if old != nil {
			_ = old.Close()
		}
		emit(Event{Type: "log", Message: fmt.Sprintf(
			"[RELAY] connected via=%s/%s pingMs=%d",
			res.out.Candidate.Network, res.out.Candidate.Host, res.out.PingMs,
		)})
		emit(Event{Type: "status", Event: "connected", Relay: label, PingMs: res.out.PingMs})
	}
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
	relayCancel := r.relayCancel
	cancel := r.cancel
	bridge := r.bridge
	hy := r.hy
	r.relayCancel = nil
	r.cancel = nil
	r.bridge = nil
	r.hy = nil
	r.mu.Unlock()

	dnscache.SetRouteHint(nil)
	if relayCancel != nil {
		relayCancel()
	}
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
