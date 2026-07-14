// Package relay is the Relay Manager application layer: selecting which
// relay servers to advertise to clients (spec section 9) and recording
// health-check results.
package relay

import (
	"context"
	"sort"
	"time"

	"streampass/backend/internal/domain/relay"
	apperrors "streampass/shared/errors"
	"streampass/shared/logger"
)

// Service implements the Relay Manager use cases.
type Service struct {
	repo relay.Repository
	log  *logger.Logger
}

// NewService wires the Relay Manager via constructor injection.
func NewService(repo relay.Repository, log *logger.Logger) *Service {
	return &Service{repo: repo, log: log.With("relay_service")}
}

// ListAvailable implements "GET /servers": returns healthy relay servers,
// best first (lowest load, then lowest RTT), so a client that just wants
// the top entry gets a reasonable default without its own scoring logic
// (spec: "Клиент автоматически выбирает лучший relay" — the client still
// makes the final call, this only orders the candidates).
func (s *Service) ListAvailable(ctx context.Context) ([]relay.Server, error) {
	all, err := s.repo.List(ctx)
	if err != nil {
		s.log.Error(ctx, err)
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to list relay servers", err)
	}

	available := make([]relay.Server, 0, len(all))
	for _, srv := range all {
		if srv.IsSelectable() {
			available = append(available, srv)
		}
	}

	sort.Slice(available, func(i, j int) bool {
		if available[i].LoadRatio != available[j].LoadRatio {
			return available[i].LoadRatio < available[j].LoadRatio
		}
		return available[i].RTTMillis < available[j].RTTMillis
	})

	return available, nil
}

// RecordHealthCheck implements the Health Monitor's write path.
func (s *Service) RecordHealthCheck(ctx context.Context, id relay.ID, healthy bool, loadRatio float64, rttMillis int, checkedAt time.Time) error {
	if err := s.repo.UpdateHealth(ctx, id, healthy, loadRatio, rttMillis, checkedAt); err != nil {
		s.log.Error(ctx, err)
		return apperrors.Wrap(apperrors.CodeInternal, "failed to record relay health check", err)
	}
	return nil
}
