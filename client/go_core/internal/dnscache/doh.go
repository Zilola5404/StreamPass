package dnscache

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/net/dns/dnsmessage"

	"streampass/go_core/internal/protect"
)

const defaultDoHURL = "https://cloudflare-dns.com/dns-query"

// Resolver answers DNS queries via DoH with a local TTL cache (ТЗ §7).
type Resolver struct {
	cache  *Cache
	client *http.Client
	dohURL string
}

var (
	defaultOnce sync.Once
	defaultRes  *Resolver
)

// Default returns the process-wide DNS resolver used by the TUN bridge.
func Default() *Resolver {
	defaultOnce.Do(func() {
		defaultRes = NewResolver(defaultDoHURL)
	})
	return defaultRes
}

// NewResolver builds a DoH resolver. HTTP dials use VpnService.protect when set.
func NewResolver(dohURL string) *Resolver {
	if dohURL == "" {
		dohURL = defaultDoHURL
	}
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
			Control:   protect.Control,
		}).DialContext,
		ForceAttemptHTTP2: true,
		IdleConnTimeout:   90 * time.Second,
	}
	return &Resolver{
		cache:  New(512),
		dohURL: dohURL,
		client: &http.Client{Timeout: 12 * time.Second, Transport: transport},
	}
}

// HandleQuery resolves a DNS wire query via cache or DoH and returns a response packet.
func (r *Resolver) HandleQuery(ctx context.Context, query []byte) ([]byte, error) {
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
		return rewriteID(raw, header.ID)
	}

	raw, ttl, err := r.fetchDoH(ctx, query)
	if err != nil {
		return buildServFail(header, q)
	}
	r.cache.PutRaw(qtype, name, raw, ttl)
	return rewriteID(raw, header.ID)
}

func (r *Resolver) fetchDoH(ctx context.Context, query []byte) ([]byte, time.Duration, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.dohURL, bytes.NewReader(query))
	if err != nil {
		return nil, 0, err
	}
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

	ttl := 60 * time.Second
	var parser dnsmessage.Parser
	if _, err := parser.Start(raw); err == nil {
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
	}
	return raw, ttl, nil
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
