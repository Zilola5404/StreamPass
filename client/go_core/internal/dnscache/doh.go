package dnscache

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/dns/dnsmessage"

	"streampass/go_core/internal/protect"
)

// DoH talks to Cloudflare by IP so we never need to resolve a hostname
// through the VPN DNS path (that would recurse into HandleQuery and hang/crash).
// Russian TLDs (.ru/.рф/…) use Yandex plain DNS for correct geo answers (ТЗ DIRECT).
const (
	doHEndpoint   = "https://1.1.1.1/dns-query"
	doHServerName = "cloudflare-dns.com"
	plainDNSAddr  = "1.1.1.1:53"
	ruDNSAddr     = "77.88.8.8:53"
)

// LogFunc receives diagnostic lines (optional).
type LogFunc func(message string)

// Resolver answers DNS queries via DoH with a local TTL cache (ТЗ §7).
type Resolver struct {
	cache  *Cache
	client *http.Client
	logf   LogFunc
}

var (
	defaultOnce sync.Once
	defaultRes  *Resolver
	logMu       sync.RWMutex
	globalLog   LogFunc
	dohOKOnce   sync.Once
)

// SetLogger installs a process-wide DNS diagnostic logger.
func SetLogger(fn LogFunc) {
	logMu.Lock()
	globalLog = fn
	logMu.Unlock()
	if r := defaultRes; r != nil {
		r.logf = fn
	}
}

func logLine(fn LogFunc, msg string) {
	if fn == nil {
		logMu.RLock()
		fn = globalLog
		logMu.RUnlock()
	}
	if fn != nil {
		fn(msg)
	}
}

// Default returns the process-wide DNS resolver used by the TUN bridge.
func Default() *Resolver {
	defaultOnce.Do(func() {
		defaultRes = NewResolver()
		logMu.RLock()
		defaultRes.logf = globalLog
		logMu.RUnlock()
	})
	return defaultRes
}

// NewResolver builds a DoH resolver. HTTP dials use VpnService.protect when set
// and always connect to 1.1.1.1 (no hostname lookup).
func NewResolver() *Resolver {
	dialer := &net.Dialer{
		Timeout:   8 * time.Second,
		KeepAlive: 30 * time.Second,
		Control:   protect.Control,
	}
	transport := &http.Transport{
		Proxy: nil, // never use system HTTP proxy for DoH
		DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			// Ignore resolved address — always hit Cloudflare anycast IP.
			return dialer.DialContext(ctx, "tcp", "1.1.1.1:443")
		},
		TLSClientConfig: &tls.Config{
			ServerName:         doHServerName,
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: false,
		},
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          4,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   8 * time.Second,
		ResponseHeaderTimeout: 8 * time.Second,
	}
	r := &Resolver{
		cache: New(512),
		client: &http.Client{
			Timeout:   10 * time.Second,
			Transport: transport,
		},
	}
	logLine(r.logf, "[dns] bootstrap=1.1.1.1 protected=true")
	return r
}

// HandleQuery resolves a DNS wire query via cache, DoH, or plain UDP fallback.
func (r *Resolver) HandleQuery(ctx context.Context, query []byte) (resp []byte, err error) {
	defer func() {
		if rec := recover(); rec != nil {
			err = fmt.Errorf("dns panic: %v", rec)
			resp = nil
		}
	}()

	var parser dnsmessage.Parser
	header, err := parser.Start(query)
	if err != nil {
		return nil, fmt.Errorf("dns parse: %w", err)
	}
	q, err := parser.Question()
	if err != nil {
		return nil, fmt.Errorf("dns question: %w", err)
	}
	name := q.Name.String()
	qtype := uint16(q.Type)

	// VPN/relay path is IPv4-only (VPS has no working IPv6 egress). Answering
	// AAAA makes apps dial [IPv6]:443 via Hysteria → "no IPv4 address available".
	if q.Type == dnsmessage.TypeAAAA {
		hostTrim := trimDot(name)
		logLine(r.logf, fmt.Sprintf("[dns] query %s via=aaaa-suppress (IPv4-only VPN)", hostTrim))
		emitDNSDiag(hostTrim, "aaaa-suppress", 0, "")
		emitDNSRoute(hostTrim)
		// Android often keeps A in its own cache and never asks us for TypeA, so
		// TCP arrives IP-only (host=). Prefetch A + pin BEFORE answering AAAA so
		// 2ip.ru stays DIRECT under DefaultMode=RELAY.
		r.prefetchAndPin(ctx, hostTrim)
		return buildEmptyAnswer(header, q)
	}

	if raw, ok := r.cache.GetRaw(qtype, name); ok {
		logLine(r.logf, fmt.Sprintf("[dns] query %s via=cache", trimDot(name)))
		IndexAnswers(name, raw)
		pinRouteFromAnswer(trimDot(name), raw)
		emitDNSDiag(trimDot(name), "cache", 0, "")
		emitDNSRoute(trimDot(name))
		return rewriteID(raw, header.ID)
	}

	start := time.Now()
	var raw []byte
	var ttl time.Duration
	via := "doh"
	hostTrim := trimDot(name)

	// Android NetworkMonitor / captive portal must resolve fast (<~3s) or the
	// OS leaves the VPN without IS_VALIDATED → apps show "no internet".
	if isNetworkMonitorHost(hostTrim) {
		raw, ttl, err = r.fetchPlainUDPTo(ctx, query, ruDNSAddr)
		via = "yandex-nm"
		if err != nil {
			raw, ttl, err = r.fetchPlainUDPTo(ctx, query, plainDNSAddr)
			via = "udp-nm"
		}
	} else if IsRussianDomain(name) {
		raw, ttl, err = r.fetchPlainUDPTo(ctx, query, ruDNSAddr)
		via = "yandex"
		if err != nil {
			raw, ttl, err = r.fetchPlainUDPTo(ctx, query, plainDNSAddr)
			via = "udp-fallback"
		}
	} else {
		// Foreign: DoH to Cloudflare often blocked/slow in RU (8s+ → Android
		// NetworkMonitor fails → "VPN connected, no internet"). Cap DoH, then
		// fall back to Yandex (reachable) before 1.1.1.1 UDP.
		dohCtx, cancel := context.WithTimeout(ctx, 1500*time.Millisecond)
		raw, ttl, err = r.fetchDoH(dohCtx, query)
		cancel()
		if err != nil {
			raw, ttl, err = r.fetchPlainUDPTo(ctx, query, ruDNSAddr)
			via = "yandex"
			if err != nil {
				raw, ttl, err = r.fetchPlainUDPTo(ctx, query, plainDNSAddr)
				via = "udp"
			}
		} else {
			dohOKOnce.Do(func() {
				logLine(r.logf, "[dns] doh connected ip=1.1.1.1 sni=cloudflare-dns.com")
			})
		}
	}
	if err != nil {
		logLine(r.logf, fmt.Sprintf("[dns] query %s via=fail err=%v", trimDot(name), err))
		emitDNSDiag(trimDot(name), "fail", 0, err.Error())
		return buildServFail(header, q)
	}
	r.cache.PutRaw(qtype, name, raw, ttl)
	IndexAnswers(name, raw)
	pinRouteFromAnswer(trimDot(name), raw)
	rtt := time.Since(start).Milliseconds()
	RememberResolveMS(trimDot(name), rtt)
	logLine(r.logf, fmt.Sprintf("[dns] query %s via=%s rtt=%dms", trimDot(name), via, rtt))
	emitDNSDiag(trimDot(name), via, rtt, "")
	emitDNSRoute(trimDot(name))
	return rewriteID(raw, header.ID)
}

// pinRouteFromAnswer pins answer A-IPs as DIRECT or RELAY from DNS-time
// classification so IP-only TCP keeps the right mode (2ip.ru vs ifconfig.me).
func pinRouteFromAnswer(host string, raw []byte) {
	if host == "" || len(raw) == 0 {
		return
	}
	_, route, _ := ClassifyDNSRoute(host)
	ips := ExtractAIPs(raw)
	if len(ips) == 0 {
		logLine(nil, fmt.Sprintf("[dns] pin-skip host=%s route=%s (no A in answer/additional)", host, route))
		return
	}
	switch strings.ToUpper(route) {
	case "RELAY":
		for _, ip := range ips {
			PinRelayIP(host, ip)
		}
		logLine(nil, fmt.Sprintf("[dns] pin-relay host=%s ips=%s", host, strings.Join(ips, ",")))
	case "DIRECT":
		for _, ip := range ips {
			PinDirectIP(host, ip)
		}
		logLine(nil, fmt.Sprintf("[dns] pin-direct host=%s ips=%s", host, strings.Join(ips, ",")))
	}
}

// prefetchAndPin resolves TypeA for host (side effect) and pins DIRECT/RELAY.
// Called on AAAA-suppress because apps often dial a cached A without querying us.
func (r *Resolver) prefetchAndPin(ctx context.Context, host string) {
	host = trimDot(host)
	if host == "" || r == nil {
		return
	}
	// Reuse cached A if present.
	if raw, ok := r.cache.GetRaw(uint16(dnsmessage.TypeA), host+"."); ok {
		IndexAnswers(host, raw)
		pinRouteFromAnswer(host, raw)
		return
	}
	qwire, err := buildAQuery(host)
	if err != nil {
		return
	}
	pctx, cancel := context.WithTimeout(ctx, 450*time.Millisecond)
	defer cancel()
	var raw []byte
	if IsRussianDomain(host) || isNetworkMonitorHost(host) {
		raw, _, err = r.fetchPlainUDPTo(pctx, qwire, ruDNSAddr)
		if err != nil {
			raw, _, err = r.fetchPlainUDPTo(pctx, qwire, plainDNSAddr)
		}
	} else {
		raw, _, err = r.fetchDoH(pctx, qwire)
		if err != nil {
			raw, _, err = r.fetchPlainUDPTo(pctx, qwire, plainDNSAddr)
		}
	}
	if err != nil || len(raw) == 0 {
		return
	}
	r.cache.PutRaw(uint16(dnsmessage.TypeA), host+".", raw, 30*time.Second)
	IndexAnswers(host, raw)
	pinRouteFromAnswer(host, raw)
}

func buildAQuery(host string) ([]byte, error) {
	host = trimDot(host)
	if host == "" {
		return nil, fmt.Errorf("empty host")
	}
	name, err := dnsmessage.NewName(host + ".")
	if err != nil {
		return nil, err
	}
	builder := dnsmessage.NewBuilder(nil, dnsmessage.Header{
		ID:               uint16(time.Now().UnixNano()),
		RecursionDesired: true,
	})
	if err := builder.StartQuestions(); err != nil {
		return nil, err
	}
	if err := builder.Question(dnsmessage.Question{
		Name:  name,
		Type:  dnsmessage.TypeA,
		Class: dnsmessage.ClassINET,
	}); err != nil {
		return nil, err
	}
	return builder.Finish()
}

func emitDNSRoute(host string) {
	if host == "" {
		return
	}
	rule, route, reason := ClassifyDNSRoute(host)
	logLine(nil, fmt.Sprintf(
		"[dns-route] host=%s rule=%s route=%s reason=%s",
		host, rule, route, reason,
	))
}

// ClassifyDNSRoute returns diagnostic rule/route/reason for a hostname.
// Uses SetRouteHint (Decision Engine / forceMode) when installed; otherwise
// mirrors DefaultMode=RELAY for unknown foreign (RU still DIRECT via IsRussianDomain).
func ClassifyDNSRoute(host string) (rule, route, reason string) {
	if hint := routeHint(); hint != nil {
		return hint(host)
	}
	if IsRussianDomain(host) {
		return "*.ru", "DIRECT", "ru_domain_bypass"
	}
	if isNetworkMonitorHost(host) {
		return "network_monitor", "DIRECT", "network_monitor_direct"
	}
	return "foreign", "RELAY", "default_relay_foreign"
}

func isNetworkMonitorHost(host string) bool {
	h := strings.ToLower(strings.TrimSuffix(host, "."))
	switch {
	case h == "connectivitycheck.gstatic.com",
		strings.HasSuffix(h, ".connectivitycheck.gstatic.com"),
		h == "connectivitycheck.android.com",
		h == "clients3.google.com",
		h == "clients1.google.com",
		h == "clients2.google.com",
		h == "clients4.google.com",
		h == "android.clients.google.com":
		return true
	default:
		return false
	}
}

// RouteHint classifies a hostname the same way Decision Engine will (incl. forceMode).
type RouteHint func(host string) (rule, route, reason string)

var (
	routeHintMu sync.RWMutex
	routeHintFn RouteHint
)

// SetRouteHint installs Decision-aware DNS route logging (cleared with nil).
func SetRouteHint(fn RouteHint) {
	routeHintMu.Lock()
	routeHintFn = fn
	routeHintMu.Unlock()
}

func routeHint() RouteHint {
	routeHintMu.RLock()
	defer routeHintMu.RUnlock()
	return routeHintFn
}

func emitDNSDiag(host, via string, rttMS int64, errMsg string) {
	result := "ok"
	reason := "dns_" + via
	if via == "fail" || errMsg != "" {
		result = "fail"
		reason = "dns_fail"
		if errMsg != "" {
			reason = "dns_fail:" + sanitizeDNSErr(errMsg)
		}
	}
	site := ""
	if host != "" {
		site = "https://" + host
	}
	logLine(nil, fmt.Sprintf(
		"[diag] proto=dns site=%s host=%s dest_ip= dest_port=53 mode=DNS via=%s rule=dns decision=%s relay_id= result=%s latency_ms=%d slow=0 speed_kbps=0 reason=%s error=%s",
		site, host, strings.ToUpper(via), reason, result, rttMS, reason, sanitizeDNSErr(errMsg),
	))
}

func sanitizeDNSErr(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "\n", "")
	if len(s) > 64 {
		s = s[:64]
	}
	return s
}

func trimDot(name string) string {
	if len(name) > 0 && name[len(name)-1] == '.' {
		return name[:len(name)-1]
	}
	return name
}

func (r *Resolver) fetchDoH(ctx context.Context, query []byte) ([]byte, time.Duration, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, doHEndpoint, bytes.NewReader(query))
	if err != nil {
		return nil, 0, err
	}
	req.Host = doHServerName
	req.Header.Set("Host", doHServerName)
	req.Header.Set("Content-Type", "application/dns-message")
	req.Header.Set("Accept", "application/dns-message")

	res, err := r.client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 256))
		return nil, 0, fmt.Errorf("doh status %d: %s", res.StatusCode, body)
	}
	raw, err := io.ReadAll(io.LimitReader(res.Body, 64*1024))
	if err != nil {
		return nil, 0, err
	}
	return raw, extractTTL(raw), nil
}

func (r *Resolver) fetchPlainUDP(ctx context.Context, query []byte) ([]byte, time.Duration, error) {
	return r.fetchPlainUDPTo(ctx, query, plainDNSAddr)
}

func (r *Resolver) fetchPlainUDPTo(ctx context.Context, query []byte, addr string) ([]byte, time.Duration, error) {
	var lc net.ListenConfig
	lc.Control = protect.Control
	pc, err := lc.ListenPacket(ctx, "udp", ":0")
	if err != nil {
		return nil, 0, err
	}
	defer pc.Close()

	raddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, 0, err
	}
	_ = pc.SetDeadline(time.Now().Add(2 * time.Second))
	if _, err := pc.WriteTo(query, raddr); err != nil {
		return nil, 0, err
	}
	buf := make([]byte, 4096)
	n, _, err := pc.ReadFrom(buf)
	if err != nil {
		return nil, 0, err
	}
	raw := append([]byte(nil), buf[:n]...)
	return raw, extractTTL(raw), nil
}

func extractTTL(raw []byte) time.Duration {
	ttl := 60 * time.Second
	var parser dnsmessage.Parser
	if _, err := parser.Start(raw); err != nil {
		return ttl
	}
	_, _ = parser.Question()
	for {
		ah, err := parser.AnswerHeader()
		if err == dnsmessage.ErrSectionDone {
			break
		}
		if err != nil {
			break
		}
		if ah.TTL > 0 {
			ttl = time.Duration(ah.TTL) * time.Second
		}
		_ = parser.SkipAnswer()
	}
	return ttl
}

func rewriteID(packet []byte, id uint16) ([]byte, error) {
	if len(packet) < 2 {
		return nil, fmt.Errorf("dns packet too short")
	}
	out := append([]byte(nil), packet...)
	out[0] = byte(id >> 8)
	out[1] = byte(id)
	return out, nil
}

func buildServFail(req dnsmessage.Header, q dnsmessage.Question) ([]byte, error) {
	builder := dnsmessage.NewBuilder(nil, dnsmessage.Header{
		ID:               req.ID,
		Response:         true,
		OpCode:           req.OpCode,
		RCode:            dnsmessage.RCodeServerFailure,
		RecursionDesired: req.RecursionDesired,
	})
	if err := builder.StartQuestions(); err != nil {
		return nil, err
	}
	if err := builder.Question(q); err != nil {
		return nil, err
	}
	return builder.Finish()
}

// buildEmptyAnswer returns a successful NOERROR response with zero answers
// (used to suppress AAAA on IPv4-only relay deployments).
func buildEmptyAnswer(req dnsmessage.Header, q dnsmessage.Question) ([]byte, error) {
	builder := dnsmessage.NewBuilder(nil, dnsmessage.Header{
		ID:                 req.ID,
		Response:           true,
		OpCode:             req.OpCode,
		RCode:              dnsmessage.RCodeSuccess,
		RecursionDesired:   req.RecursionDesired,
		RecursionAvailable: true,
	})
	if err := builder.StartQuestions(); err != nil {
		return nil, err
	}
	if err := builder.Question(q); err != nil {
		return nil, err
	}
	return builder.Finish()
}
