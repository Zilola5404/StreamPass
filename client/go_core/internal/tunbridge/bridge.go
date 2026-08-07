package tunbridge

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/apernet/hysteria/core/v2/client"
	"github.com/sagernet/sing/common/buf"
	"github.com/sagernet/sing/common/bufio"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
	tun "github.com/sagernet/sing-tun"

	"streampass/go_core/internal/decision"
	"streampass/go_core/internal/dnscache"
	"streampass/go_core/internal/protect"
)

// LogFunc receives diagnostic lines from the TUN bridge.
type LogFunc func(message string)

var (
	logMu        sync.RWMutex
	logf         LogFunc
	trafficReady atomic.Bool
)

// resetTrafficReady clears the session traffic-ready latch (handshake ≠ data path).
func resetTrafficReady() {
	trafficReady.Store(false)
}

// markTrafficReady emits a one-shot [vpn] traffic_ready when the first user-plane
// byte/packet returns via DIRECT or RELAY (Architect Issue #1 / BL-001).
func markTrafficReady(via string) {
	if trafficReady.CompareAndSwap(false, true) {
		logLine(fmt.Sprintf("[vpn] traffic_ready via=%s", via))
	}
}

// SetLogger installs an optional diagnostic logger for dial failures.
func SetLogger(fn LogFunc) {
	logMu.Lock()
	logf = fn
	logMu.Unlock()
}

func logLine(msg string) {
	logMu.RLock()
	fn := logf
	logMu.RUnlock()
	if fn != nil {
		fn(msg)
	}
}

type Session struct {
	cancel context.CancelFunc
	stop   func()
	engine *decision.AtomicEngine
}

// UpdateEngine hot-reloads routing rules on an active tunnel (BL-006).
func (s *Session) UpdateEngine(rulesJSON, exclusionsJSON string) error {
	if s == nil || s.engine == nil {
		return fmt.Errorf("no active tunnel session")
	}
	return s.engine.Update(rulesJSON, exclusionsJSON)
}

// RulesVersion returns the currently loaded rule set version.
func (s *Session) RulesVersion() int {
	if s == nil || s.engine == nil {
		return 0
	}
	return s.engine.Version()
}

func (s *Session) Close() {
	if s == nil {
		return
	}
	if s.cancel != nil {
		s.cancel()
	}
	if s.stop != nil {
		s.stop()
	}
}

// TunIPv4Prefix is the Android VpnService address + sing-tun system-stack prefix.
// Must be the first host of a /30 (…1), so Addr().Next() is a unicast host (…2),
// not the subnet broadcast (…3). Using 10.10.0.2/30 made Next()=10.10.0.3 and
// broke TCP NAT → connected with no internet.
func TunIPv4Prefix() netip.Prefix {
	return netip.PrefixFrom(netip.MustParseAddr("10.10.0.1"), 30)
}

// TunIPv4Host is the address assigned on the Android TUN interface.
func TunIPv4Host() string {
	return TunIPv4Prefix().Addr().String()
}

// Options configures diagnostic network behaviour on the TUN bridge.
type Options struct {
	// BlockUDP443 drops UDP/443 so apps fall back to TCP/443 (QUIC off).
	BlockUDP443 bool
}

// Start brings up the TUN stack and routes each flow via the Decision Engine.
// relayID is an operator label (host/id) attached to RELAY diag events.
func Start(ctx context.Context, fd int, hyClient client.Client, mtu uint32, engine *decision.AtomicEngine, relayID string) (*Session, error) {
	return StartWithOptions(ctx, fd, hyClient, mtu, engine, relayID, Options{})
}

// StartWithOptions is Start plus diagnostic toggles (UDP/443 block, …).
func StartWithOptions(ctx context.Context, fd int, hyClient client.Client, mtu uint32, engine *decision.AtomicEngine, relayID string, opts Options) (*Session, error) {
	resetTrafficReady()
	if mtu == 0 {
		mtu = 1400
	}
	if engine == nil {
		engine = decision.NewAtomicEngine(decision.NewEngine(nil, nil, decision.DefaultMode), 0)
	}

	tunOptions := tun.Options{
		FileDescriptor: fd,
		MTU:            mtu,
		AutoRoute:      false,
		Inet4Address:   []netip.Prefix{TunIPv4Prefix()},
	}

	tunDev, err := tun.New(tunOptions)
	if err != nil {
		return nil, err
	}

	runCtx, cancel := context.WithCancel(ctx)
	handler := &routingHandler{
		client:      hyClient,
		engine:      engine,
		relayID:     relayID,
		blockUDP443: opts.BlockUDP443,
	}
	if opts.BlockUDP443 {
		logLine("[vpn] udp443=blocked (force TCP/443 fallback)")
	}
	stack, err := tun.NewStack("system", tun.StackOptions{
		Context:    runCtx,
		Tun:        tunDev,
		TunOptions: tunOptions,
		Handler:    handler,
		UDPTimeout: 5 * time.Minute,
	})
	if err != nil {
		cancel()
		_ = tunDev.Close()
		return nil, err
	}

	if err := tunDev.Start(); err != nil {
		cancel()
		_ = stack.Close()
		_ = tunDev.Close()
		return nil, err
	}
	if err := stack.Start(); err != nil {
		cancel()
		_ = stack.Close()
		_ = tunDev.Close()
		return nil, err
	}

	stop := func() {
		_ = stack.Close()
		_ = tunDev.Close()
	}
	return &Session{cancel: cancel, stop: stop, engine: engine}, nil
}

type routingHandler struct {
	client      client.Client
	engine      *decision.AtomicEngine
	relayID     string
	blockUDP443 bool
}

func (h *routingHandler) PrepareConnection(
	network string,
	source M.Socksaddr,
	destination M.Socksaddr,
	routeContext tun.DirectRouteContext,
	timeout time.Duration,
) (tun.DirectRouteDestination, error) {
	return nil, nil
}

func (h *routingHandler) targetFrom(destination M.Socksaddr) decision.Target {
	t := decision.Target{}
	if destination.IsFqdn() {
		t.Host = destination.Fqdn
	}
	if destination.Addr.IsValid() {
		t.IP = destination.Addr
		if t.Host == "" {
			// OS often delivers IP-only flows after system DNS; recover hostname
			// so *.ru / domain rules still match (TASK network diagnostics).
			t.Host = dnscache.HostForIP(destination.Addr.String())
		}
	}
	return t
}

func (h *routingHandler) NewConnectionEx(
	ctx context.Context,
	conn net.Conn,
	source M.Socksaddr,
	destination M.Socksaddr,
	onClose N.CloseHandlerFunc,
) {
	defer func() {
		if r := recover(); r != nil {
			_ = conn.Close()
		}
	}()
	defer conn.Close()
	if onClose != nil {
		defer onClose(nil)
	}

	dec := h.engine.DecideDetailed(h.targetFrom(destination))
	mode := dec.Mode
	host, destIP, destPort := splitDest(destination)
	logLine(fmt.Sprintf("[tun] tcp mode=%s rule=%s reason=%s dest=%s", mode, dec.Rule, dec.Reason, destination.String()))
	logLine(fmt.Sprintf("[decision] host=%s ip=%s rule=%s action=%s reason=%s", host, destIP, dec.Rule, mode, dec.Reason))
	switch mode {
	case decision.ModeDirect:
		h.pipeTCP(ctx, conn, destination, host, destIP, destPort, dec, true, false)
	case decision.ModeFallback:
		// Decision policy: RELAY-first + allow_direct_fallback (docs/07.4 §6).
		if !h.pipeTCP(ctx, conn, destination, host, destIP, destPort, dec, false, true) {
			dec.Reason = "fallback_after_relay_fail"
			logLine(fmt.Sprintf("[tun] FALLBACK after relay fail dest=%s", destination.String()))
			h.pipeTCP(ctx, conn, destination, host, destIP, destPort, dec, true, false)
		}
	default:
		// Must-relay: no silent DIRECT (docs/07.4 §6.2).
		if !h.pipeTCP(ctx, conn, destination, host, destIP, destPort, dec, false, true) {
			logLine(fmt.Sprintf("[tun] must-relay fail dest=%s (no DIRECT fallback)", destination.String()))
		}
	}
}

// pipeTCP dials (direct or relay), emits [diag] with dial RTT + decision reason,
// then copies until close (emitting throughput when enough bytes moved).
// detectBlackhole: for RELAY path, dial-ok without first_byte within 3s is a fail
// (docs/07.4 §6) so Decision can apply FALLBACK — transport itself never switches Mode.
func (h *routingHandler) pipeTCP(
	ctx context.Context,
	conn net.Conn,
	destination M.Socksaddr,
	host, destIP string,
	destPort int,
	dec decision.Decision,
	direct bool,
	detectBlackhole bool,
) bool {
	modeLabel := string(dec.Mode)
	if !direct && dec.Mode == decision.ModeFallback {
		modeLabel = "FALLBACK"
	} else if direct && dec.Mode == decision.ModeFallback {
		modeLabel = "FALLBACK"
	} else if direct {
		modeLabel = "DIRECT"
	} else {
		modeLabel = "RELAY"
	}
	start := time.Now()
	var remote net.Conn
	var err error
	if direct {
		dialer := net.Dialer{Timeout: 15 * time.Second, Control: protect.Control}
		remote, err = dialer.DialContext(ctx, "tcp", destination.String())
		if err != nil {
			logLine(fmt.Sprintf("[tun] direct-tcp fail dest=%s err=%v", destination.String(), err))
			emitDiag("tcp", host, destIP, destPort, modeLabel, dec.Rule, dec.Reason, h.relayID, false, time.Since(start), err.Error(), 0)
			emitConnLog(host, destIP, "TCP", destPort, modeLabel, dec.Rule, dec.Reason, dnscache.LastResolveMS(host), time.Since(start), 0, 0, err.Error())
			return false
		}
	} else {
		remote, err = h.client.TCP(destination.String())
		if err != nil {
			logLine(fmt.Sprintf("[tun] relay-tcp fail dest=%s err=%v", destination.String(), err))
			emitDiag("tcp", host, destIP, destPort, modeLabel, dec.Rule, dec.Reason, h.relayID, false, time.Since(start), err.Error(), 0)
			emitConnLog(host, destIP, "TCP", destPort, modeLabel, dec.Rule, dec.Reason, dnscache.LastResolveMS(host), time.Since(start), 0, 0, err.Error())
			return false
		}
	}
	dialDur := time.Since(start)
	emitDiag("tcp", host, destIP, destPort, modeLabel, dec.Rule, dec.Reason, h.relayID, true, dialDur, "", 0)

	var bytes atomic.Int64
	var firstByte atomic.Int64 // nanoseconds; 0 = none yet
	xferStart := time.Now()
	counted := &countingConn{
		Conn: remote,
		onByte: func(n int, fromRemote bool) {
			if n <= 0 {
				return
			}
			bytes.Add(int64(n))
			if fromRemote && firstByte.Load() == 0 {
				firstByte.Store(time.Since(xferStart).Nanoseconds())
				markTrafficReady(modeLabel)
			}
		},
	}

	if detectBlackhole && !direct {
		xferCtx, cancel := context.WithCancel(ctx)
		done := make(chan struct{})
		go func() {
			defer close(done)
			_ = bufio.CopyConn(xferCtx, conn, counted)
		}()
		deadline := time.NewTimer(3 * time.Second)
		defer deadline.Stop()
		select {
		case <-done:
			_ = remote.Close()
		case <-deadline.C:
			if firstByte.Load() == 0 {
				cancel()
				_ = remote.Close()
				<-done
				logLine(fmt.Sprintf("[tun] relay blackhole (no first_byte in 3s) dest=%s", destination.String()))
				emitDiag("tcp", host, destIP, destPort, modeLabel, dec.Rule, "relay_blackhole", h.relayID, false, 3*time.Second, "relay_blackhole", 0)
				emitConnLog(host, destIP, "TCP", destPort, modeLabel, dec.Rule, "relay_blackhole", dnscache.LastResolveMS(host), dialDur, 0, 0, "relay_blackhole")
				return false
			}
			<-done
			_ = remote.Close()
		case <-ctx.Done():
			cancel()
			_ = remote.Close()
			<-done
			return false
		}
	} else {
		_ = bufio.CopyConn(ctx, conn, counted)
		_ = remote.Close()
	}

	xferDur := time.Since(xferStart)
	total := bytes.Load()
	fb := time.Duration(firstByte.Load())
	kbps := 0
	if total >= 32*1024 && xferDur > 500*time.Millisecond {
		kbps = int(float64(total*8/1024) / xferDur.Seconds())
		emitDiag("tcp", host, destIP, destPort, modeLabel, dec.Rule, "transfer_done", h.relayID, true, xferDur, "", kbps)
	}
	dnsMS := dnscache.LastResolveMS(host)
	emitConnLog(host, destIP, "TCP", destPort, modeLabel, dec.Rule, dec.Reason, dnsMS, dialDur, fb, kbps, "")
	return true
}

type countingConn struct {
	net.Conn
	onByte func(n int, fromRemote bool)
}

func (c *countingConn) Read(p []byte) (int, error) {
	n, err := c.Conn.Read(p)
	if n > 0 && c.onByte != nil {
		c.onByte(n, true)
	}
	return n, err
}

func (c *countingConn) Write(p []byte) (int, error) {
	n, err := c.Conn.Write(p)
	if n > 0 && c.onByte != nil {
		c.onByte(n, false)
	}
	return n, err
}

func (h *routingHandler) NewPacketConnectionEx(
	ctx context.Context,
	conn N.PacketConn,
	source M.Socksaddr,
	destination M.Socksaddr,
	onClose N.CloseHandlerFunc,
) {
	defer func() {
		if r := recover(); r != nil {
			_ = conn.Close()
		}
	}()
	defer conn.Close()
	if onClose != nil {
		defer onClose(nil)
	}

	// Local DNS Cache + DoH (ТЗ §7 / BL-016): answer UDP/53 without hairpinning.
	if destination.Port == 53 {
		h.handleDNS(ctx, conn, destination)
		return
	}

	// Diagnostic: force TCP/443 by dropping QUIC (UDP/443).
	if h.blockUDP443 && destination.Port == 443 {
		host, destIP, destPort := splitDest(destination)
		logLine(fmt.Sprintf("[vpn] drop udp/443 host=%s ip=%s (tcp_only)", host, destIP))
		emitDiag("udp", host, destIP, destPort, "DROP", "udp443", "udp443_blocked", h.relayID, false, 0, "udp443_blocked", 0)
		return
	}

	dec := h.engine.DecideDetailed(h.targetFrom(destination))
	mode := dec.Mode
	host, destIP, destPort := splitDest(destination)
	logLine(fmt.Sprintf("[decision] host=%s ip=%s rule=%s action=%s reason=%s proto=udp", host, destIP, dec.Rule, mode, dec.Reason))

	switch mode {
	case decision.ModeDirect:
		h.relayUDP(ctx, conn, destination, true, host, destIP, destPort, dec)
	case decision.ModeFallback:
		if !h.relayUDP(ctx, conn, destination, false, host, destIP, destPort, dec) {
			dec.Reason = "fallback_after_relay_fail"
			h.relayUDP(ctx, conn, destination, true, host, destIP, destPort, dec)
		}
	default:
		// Must-relay UDP: no silent DIRECT / no quic_direct_bypass (docs/07.4 §2).
		if !h.relayUDP(ctx, conn, destination, false, host, destIP, destPort, dec) {
			logLine(fmt.Sprintf("[tun] must-relay-udp fail dest=%s (no DIRECT fallback)", destination.String()))
		}
	}
}

func (h *routingHandler) handleDNS(ctx context.Context, conn N.PacketConn, destination M.Socksaddr) {
	defer func() {
		if r := recover(); r != nil {
			// Never let DNS path take down the VPN process.
		}
	}()

	readBuffer := buf.NewPacket()
	defer readBuffer.Release()
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		_, err := conn.ReadPacket(readBuffer)
		if err != nil {
			return
		}
		query := append([]byte(nil), readBuffer.Bytes()...)
		readBuffer.Reset()

		resp, err := dnscache.Default().HandleQuery(ctx, query)
		if err != nil || len(resp) == 0 {
			continue
		}
		writeBuffer := buf.As(resp)
		_ = conn.WritePacket(writeBuffer, destination)
		writeBuffer.Release()
	}
}

func (h *routingHandler) relayUDP(
	ctx context.Context,
	conn N.PacketConn,
	destination M.Socksaddr,
	direct bool,
	host, destIP string,
	destPort int,
	dec decision.Decision,
) bool {
	destAddr := destination.String()
	start := time.Now()
	modeLabel := "RELAY"
	if direct {
		modeLabel = "DIRECT"
	} else if dec.Mode == decision.ModeFallback {
		modeLabel = "FALLBACK"
	}

	if direct {
		var lc net.ListenConfig
		lc.Control = protect.Control
		packetConn, err := lc.ListenPacket(ctx, "udp", "")
		if err != nil {
			emitDiag("udp", host, destIP, destPort, modeLabel, dec.Rule, dec.Reason, h.relayID, false, time.Since(start), err.Error(), 0)
			return false
		}
		defer packetConn.Close()
		udpRemote, ok := packetConn.(*net.UDPConn)
		if !ok {
			emitDiag("udp", host, destIP, destPort, modeLabel, dec.Rule, dec.Reason, h.relayID, false, time.Since(start), "not_udp", 0)
			return false
		}
		raddr, err := net.ResolveUDPAddr("udp", destAddr)
		if err != nil {
			emitDiag("udp", host, destIP, destPort, modeLabel, dec.Rule, dec.Reason, h.relayID, false, time.Since(start), err.Error(), 0)
			return false
		}
		emitDiag("udp", host, destIP, destPort, modeLabel, dec.Rule, dec.Reason, h.relayID, true, time.Since(start), "", 0)
		h.copyUDPToAddr(ctx, conn, udpRemote, destination, raddr)
		return true
	}

	hyUDP, err := h.client.UDP()
	if err != nil {
		emitDiag("udp", host, destIP, destPort, modeLabel, dec.Rule, dec.Reason, h.relayID, false, time.Since(start), err.Error(), 0)
		return false
	}
	defer hyUDP.Close()
	emitDiag("udp", host, destIP, destPort, modeLabel, dec.Rule, dec.Reason, h.relayID, true, time.Since(start), "", 0)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		readBuffer := buf.NewPacket()
		defer readBuffer.Release()
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			_, err := conn.ReadPacket(readBuffer)
			if err != nil {
				return
			}
			payload := append([]byte(nil), readBuffer.Bytes()...)
			readBuffer.Reset()
			if err := hyUDP.Send(payload, destAddr); err != nil {
				return
			}
		}
	}()

	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			data, addr, err := hyUDP.Receive()
			if err != nil {
				return
			}
			// Hysteria may return "1.1.1.1:53" vs "[::ffff:1.1.1.1]:53" etc.
			// Strict string equality dropped valid replies (AUDIT-003).
			if !sameUDPEndpoint(addr, destAddr) {
				continue
			}
			writeBuffer := buf.As(data)
			if err := conn.WritePacket(writeBuffer, destination); err != nil {
				writeBuffer.Release()
				return
			}
			writeBuffer.Release()
			markTrafficReady(modeLabel)
		}
	}()

	wg.Wait()
	return true
}

func (h *routingHandler) copyUDPToAddr(ctx context.Context, conn N.PacketConn, udpRemote *net.UDPConn, destination M.Socksaddr, raddr *net.UDPAddr) {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		readBuffer := buf.NewPacket()
		defer readBuffer.Release()
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			_, err := conn.ReadPacket(readBuffer)
			if err != nil {
				return
			}
			payload := append([]byte(nil), readBuffer.Bytes()...)
			readBuffer.Reset()
			if _, err := udpRemote.WriteToUDP(payload, raddr); err != nil {
				return
			}
		}
	}()

	go func() {
		defer wg.Done()
		bufData := make([]byte, 2048)
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			n, _, err := udpRemote.ReadFromUDP(bufData)
			if err != nil {
				return
			}
			writeBuffer := buf.As(append([]byte(nil), bufData[:n]...))
			if err := conn.WritePacket(writeBuffer, destination); err != nil {
				writeBuffer.Release()
				return
			}
			writeBuffer.Release()
			markTrafficReady("DIRECT")
		}
	}()

	wg.Wait()
}

func (h *routingHandler) copyUDP(ctx context.Context, conn N.PacketConn, remote net.Conn, destination M.Socksaddr, destAddr string) {
	udpRemote, ok := remote.(*net.UDPConn)
	if !ok {
		return
	}
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		readBuffer := buf.NewPacket()
		defer readBuffer.Release()
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			_, err := conn.ReadPacket(readBuffer)
			if err != nil {
				return
			}
			payload := append([]byte(nil), readBuffer.Bytes()...)
			readBuffer.Reset()
			if _, err := udpRemote.Write(payload); err != nil {
				return
			}
		}
	}()

	go func() {
		defer wg.Done()
		bufData := make([]byte, 2048)
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			n, err := udpRemote.Read(bufData)
			if err != nil {
				return
			}
			writeBuffer := buf.As(append([]byte(nil), bufData[:n]...))
			if err := conn.WritePacket(writeBuffer, destination); err != nil {
				writeBuffer.Release()
				return
			}
			writeBuffer.Release()
		}
	}()

	wg.Wait()
}

// sameUDPEndpoint compares Hysteria UDP reply addresses without relying on
// brittle string equality (IPv4-mapped, bracket forms, etc.).
func sameUDPEndpoint(a, b string) bool {
	if a == b {
		return true
	}
	aa, err1 := net.ResolveUDPAddr("udp", a)
	bb, err2 := net.ResolveUDPAddr("udp", b)
	if err1 != nil || err2 != nil {
		return false
	}
	if aa.Port != bb.Port {
		return false
	}
	if aa.IP == nil || bb.IP == nil {
		return false
	}
	return aa.IP.Equal(bb.IP)
}

func splitDest(destination M.Socksaddr) (host, destIP string, destPort int) {
	destPort = int(destination.Port)
	if destination.IsFqdn() {
		host = strings.ToLower(strings.TrimSuffix(destination.Fqdn, "."))
	}
	if destination.Addr.IsValid() {
		destIP = destination.Addr.String()
	}
	if host == "" && destIP != "" {
		host = dnscache.HostForIP(destIP)
	}
	return host, destIP, destPort
}

// okDiagThrottle limits successful [diag] lines per host+mode (failures/slow always emit).
var okDiagThrottle sync.Map // key -> time.Time

const (
	okDiagMinInterval  = 30 * time.Second
	slowDialThreshold  = 1500 * time.Millisecond
)

func emitDiag(proto, host, destIP string, destPort int, mode, rule, decisionReason, relayID string, ok bool, d time.Duration, errMsg string, speedKbps int) {
	if host == "" && destIP != "" {
		host = dnscache.HostForIP(destIP)
	}
	site := siteLabel(host, destIP)
	via := strings.ToUpper(mode)
	result := "ok"
	reason := decisionReason
	if reason == "" {
		reason = reasonOK(via)
	}
	slow := 0
	if speedKbps > 0 && decisionReason == "transfer_done" {
		result = "xfer"
		reason = "transfer_done"
	} else if !ok {
		result, reason = classifyFail(via, errMsg)
	} else if d >= slowDialThreshold {
		result = "slow"
		slow = 1
		if reason == "" || strings.HasPrefix(reason, "ok_") || strings.HasPrefix(reason, "rule_") ||
			strings.HasPrefix(reason, "russian_") || strings.HasPrefix(reason, "global_") ||
			strings.HasPrefix(reason, "international_") || strings.HasPrefix(reason, "default_") ||
			strings.HasPrefix(reason, "user_") || strings.HasPrefix(reason, "critical_") {
			reason = "slow_dial_" + strings.ToLower(via)
		}
	} else if speedKbps == 0 {
		key := proto + "|" + via + "|" + host + "|" + destIP + "|" + fmt.Sprintf("%d", destPort) + "|" + rule
		if prev, loaded := okDiagThrottle.Load(key); loaded {
			if t, okT := prev.(time.Time); okT && time.Since(t) < okDiagMinInterval {
				return
			}
		}
		okDiagThrottle.Store(key, time.Now())
	}
	errCode := sanitizeErr(errMsg)
	rule = sanitizeErr(rule)
	decisionReason = sanitizeErr(decisionReason)
	relayID = sanitizeErr(relayID)
	if reason == "" {
		reason = decisionReason
	}
	reason = sanitizeErr(reason)
	logLine(fmt.Sprintf(
		"[diag] proto=%s site=%s host=%s dest_ip=%s dest_port=%d mode=%s via=%s rule=%s decision=%s relay_id=%s result=%s latency_ms=%d slow=%d speed_kbps=%d reason=%s error=%s",
		proto, site, host, destIP, destPort, via, via, rule, decisionReason, relayID, result, d.Milliseconds(), slow, speedKbps, reason, errCode,
	))
	logLine(fmt.Sprintf(
		"[route] %s ip=%s:%d via=%s rule=%s → %s (%dms, %dkbps) %s",
		siteOrDash(site), destIP, destPort, via, rule, result, d.Milliseconds(), speedKbps, reason,
	))
}

// emitConnLog writes the human-readable connection sample requested for network diagnostics.
func emitConnLog(host, destIP, protocol string, port int, route, rule, reason string, dnsMS int64, connect, firstByte time.Duration, speedKbps int, errMsg string) {
	result := "success"
	if errMsg != "" {
		result = "fail"
	}
	logLine(fmt.Sprintf(
		"[conn] app=- host=%s ip=%s protocol=%s port=%d route=%s rule=%s reason=%s dns=%dms connect=%dms tls=- first_byte=%dms speed=%dkbps result=%s error=%s",
		host, destIP, protocol, port, route, rule, reason, dnsMS, connect.Milliseconds(), firstByte.Milliseconds(), speedKbps, result, sanitizeErr(errMsg),
	))
}

func siteLabel(host, destIP string) string {
	if host != "" {
		// Origin only (scheme + host) — never path/query (ТЗ §14).
		return "https://" + host
	}
	if destIP != "" {
		return "ip://" + destIP
	}
	return ""
}

func siteOrDash(site string) string {
	if site == "" {
		return "-"
	}
	return site
}

func reasonOK(via string) string {
	switch via {
	case "DIRECT":
		return "ok_direct"
	case "RELAY":
		return "ok_relay"
	case "FALLBACK":
		return "ok_fallback"
	default:
		return "ok"
	}
}

func classifyFail(via, errMsg string) (result, reason string) {
	low := strings.ToLower(errMsg)
	switch {
	case strings.Contains(low, "timeout") || strings.Contains(low, "i/o timeout") || strings.Contains(low, "deadline"):
		return "timeout", "timeout_" + strings.ToLower(via)
	case strings.Contains(low, "refused"):
		return "fail", "connection_refused_" + strings.ToLower(via)
	case strings.Contains(low, "reset"):
		return "fail", "connection_reset_" + strings.ToLower(via)
	case strings.Contains(low, "no route") || strings.Contains(low, "network is unreachable"):
		return "fail", "network_unreachable"
	case strings.Contains(low, "drop"):
		return "drop", "traffic_cut_" + strings.ToLower(via)
	case via == "RELAY":
		return "fail", "relay_dial_fail"
	case via == "DIRECT":
		return "fail", "direct_dial_fail"
	default:
		return "fail", "dial_fail_" + strings.ToLower(via)
	}
}

func sanitizeErr(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\r", "")
	if len(s) > 64 {
		s = s[:64]
	}
	return s
}
