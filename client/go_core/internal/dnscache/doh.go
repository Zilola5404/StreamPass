package dnscache

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/net/dns/dnsmessage"

	"streampass/go_core/internal/protect"
)

// DoH talks to Cloudflare by IP so we never need to resolve a hostname
// through the VPN DNS path (that would recurse into HandleQuery and hang/crash).
const (
	doHEndpoint   = "https://1.1.1.1/dns-query"
	doHServerName = "cloudflare-dns.com"
	plainDNSAddr  = "1.1.1.1:53"
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

	if raw, ok := r.cache.GetRaw(qtype, name); ok {
		logLine(r.logf, fmt.Sprintf("[dns] query %s via=cache", trimDot(name)))
		return rewriteID(raw, header.ID)
	}

	start := time.Now()
	raw, ttl, err := r.fetchDoH(ctx, query)
	via := "doh"
	if err != nil {
		raw, ttl, err = r.fetchPlainUDP(ctx, query)
		via = "udp"
		if err != nil {
			logLine(r.logf, fmt.Sprintf("[dns] query %s via=fail err=%v", trimDot(name), err))
			return buildServFail(header, q)
		}
	} else {
		dohOKOnce.Do(func() {
			logLine(r.logf, "[dns] doh connected ip=1.1.1.1 sni=cloudflare-dns.com")
		})
	}
	r.cache.PutRaw(qtype, name, raw, ttl)
	logLine(r.logf, fmt.Sprintf("[dns] query %s via=%s rtt=%dms", trimDot(name), via, time.Since(start).Milliseconds()))
	return rewriteID(raw, header.ID)
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
	var lc net.ListenConfig
	lc.Control = protect.Control
	pc, err := lc.ListenPacket(ctx, "udp", ":0")
	if err != nil {
		return nil, 0, err
	}
	defer pc.Close()

	raddr, err := net.ResolveUDPAddr("udp", plainDNSAddr)
	if err != nil {
		return nil, 0, err
	}
	_ = pc.SetDeadline(time.Now().Add(5 * time.Second))
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
