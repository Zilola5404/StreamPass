package diagsvc

import (
	"context"
	"log/slog"
	"strings"
	"time"
	"unicode"

	"streampass/backend/internal/domain/diag"
	apperrors "streampass/shared/errors"
	"streampass/shared/logger"
)

const (
	maxBatchSize   = 100
	maxHostLen     = 253
	retentionDays  = 7
	defaultListLim = 100
)

// Clock is injected for tests.
type Clock interface {
	Now() time.Time
}

// SystemClock is the production clock.
type SystemClock struct{}

// Now returns UTC now.
func (SystemClock) Now() time.Time { return time.Now().UTC() }

// Service ingests and lists diagnostic events.
type Service struct {
	repo  diag.Repository
	clock Clock
	log   *logger.Logger
}

// NewService wires the diag service.
func NewService(repo diag.Repository, clock Clock, log *logger.Logger) *Service {
	return &Service{repo: repo, clock: clock, log: log.With("diag_service")}
}

// RecordBatch validates and stores client diagnostic events.
func (s *Service) RecordBatch(ctx context.Context, userID string, events []diag.Event) error {
	if userID == "" {
		return apperrors.New(apperrors.CodeInvalidInput, "diag event missing user id")
	}
	if len(events) == 0 {
		return nil
	}
	if len(events) > maxBatchSize {
		return apperrors.New(apperrors.CodeInvalidInput, "diag batch too large").
			WithDetails(map[string]any{"max": maxBatchSize})
	}

	now := s.clock.Now()
	cleaned := make([]diag.Event, 0, len(events))
	for _, e := range events {
		e.UserID = userID
		e.Host = sanitizeHost(e.Host)
		e.Site = sanitizeSite(e.Site, e.Host, e.DestIP)
		e.DestIP = sanitizeToken(e.DestIP, 64)
		e.Mode = strings.ToUpper(sanitizeToken(e.Mode, 16))
		e.Result = strings.ToLower(sanitizeToken(e.Result, 16))
		e.Proto = strings.ToLower(sanitizeToken(e.Proto, 8))
		e.Reason = sanitizeToken(e.Reason, 96)
		e.Rule = sanitizeToken(e.Rule, 128)
		e.DecisionReason = sanitizeToken(e.DecisionReason, 96)
		e.ErrorCode = sanitizeToken(e.ErrorCode, 64)
		e.RelayID = sanitizeToken(e.RelayID, 64)
		e.ClientVersion = sanitizeToken(e.ClientVersion, 32)
		if e.LatencyMS < 0 {
			e.LatencyMS = 0
		}
		if e.SpeedKbps < 0 {
			e.SpeedKbps = 0
		}
		if e.DestPort < 0 || e.DestPort > 65535 {
			e.DestPort = 0
		}
		if e.Proto == "" {
			e.Proto = "tcp"
		}
		if e.Result == "" {
			e.Result = "unknown"
		}
		if e.Result == "slow" {
			e.Slow = true
		}
		if e.RecordedAt.IsZero() {
			e.RecordedAt = now
		}
		cleaned = append(cleaned, e)
		s.log.Info(ctx, "diag_event",
			slog.String("site", e.Site),
			slog.String("host", e.Host),
			slog.String("dest_ip", e.DestIP),
			slog.Int("dest_port", e.DestPort),
			slog.String("mode", e.Mode),
			slog.String("result", e.Result),
			slog.Bool("slow", e.Slow),
			slog.Int("speed_kbps", e.SpeedKbps),
			slog.String("rule", e.Rule),
			slog.String("decision", e.DecisionReason),
			slog.String("reason", e.Reason),
			slog.Int("latency_ms", e.LatencyMS),
			slog.String("error_code", e.ErrorCode),
			slog.String("relay_id", e.RelayID),
			slog.String("user_id", e.UserID),
		)
	}

	if err := s.repo.RecordBatch(ctx, cleaned); err != nil {
		s.log.Error(ctx, err)
		return err
	}
	_, _ = s.repo.PurgeOlderThan(ctx, now.Add(-retentionDays*24*time.Hour))
	return nil
}

// List returns recent events for admin.
func (s *Service) List(ctx context.Context, userID string, limit int) ([]diag.Event, error) {
	if limit <= 0 {
		limit = defaultListLim
	}
	return s.repo.List(ctx, userID, limit)
}

func sanitizeHost(h string) string {
	h = strings.TrimSpace(strings.ToLower(h))
	if i := strings.Index(h, "://"); i >= 0 {
		h = h[i+3:]
	}
	if i := strings.IndexAny(h, "/?#"); i >= 0 {
		h = h[:i]
	}
	if i := strings.Index(h, ":"); i >= 0 && !strings.HasPrefix(h, "[") {
		h = h[:i]
	}
	if len(h) > maxHostLen {
		h = h[:maxHostLen]
	}
	return h
}

// sanitizeSite keeps only https://host or ip://addr (no path/query).
func sanitizeSite(site, host, destIP string) string {
	site = strings.TrimSpace(site)
	if site != "" {
		lower := strings.ToLower(site)
		switch {
		case strings.HasPrefix(lower, "https://"):
			rest := site[len("https://"):]
			if i := strings.IndexAny(rest, "/?#"); i >= 0 {
				rest = rest[:i]
			}
			rest = sanitizeHost(rest)
			if rest != "" {
				return "https://" + rest
			}
		case strings.HasPrefix(lower, "ip://"):
			ip := sanitizeToken(site[len("ip://"):], 64)
			if ip != "" {
				return "ip://" + ip
			}
		}
	}
	if host != "" {
		return "https://" + sanitizeHost(host)
	}
	if destIP != "" {
		return "ip://" + sanitizeToken(destIP, 64)
	}
	return ""
}

func sanitizeToken(s string, max int) string {
	s = strings.TrimSpace(s)
	var b strings.Builder
	for _, r := range s {
		if unicode.IsControl(r) {
			continue
		}
		b.WriteRune(r)
		if b.Len() >= max {
			break
		}
	}
	return b.String()
}
