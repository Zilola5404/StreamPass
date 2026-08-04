package relay

import (
	"context"
	"sort"
	"strings"
	"time"

	"streampass/backend/internal/domain/relay"
	apperrors "streampass/shared/errors"
	"streampass/shared/logger"
	"streampass/shared/region"
)

type Service struct {
	repo relay.Repository
	log  *logger.Logger
}

func NewService(repo relay.Repository, log *logger.Logger) *Service {
	return &Service{repo: repo, log: log.With("relay_service")}
}

// ListAvailable returns healthy relays ranked by load then RTT.
// When regionFilter is non-empty it is normalized and used to keep only
// matching servers (ТЗ Этап 6: client can prefer a region such as Warsaw).
func (s *Service) ListAvailable(ctx context.Context, regionFilter string) ([]relay.Server, error) {
	all, err := s.repo.List(ctx)
	if err != nil {
		s.log.Error(ctx, err)
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to list relay servers", err)
	}

	wantRegion := region.Normalize(regionFilter)

	available := make([]relay.Server, 0, len(all))
	for _, srv := range all {
		if !srv.IsSelectable() {
			continue
		}
		srv.Region = relay.Region(region.Normalize(string(srv.Region)))
		if wantRegion != "" && string(srv.Region) != wantRegion {
			continue
		}
		available = append(available, srv)
	}

	sort.Slice(available, func(i, j int) bool {
		if available[i].LoadRatio != available[j].LoadRatio {
			return available[i].LoadRatio < available[j].LoadRatio
		}
		return available[i].RTTMillis < available[j].RTTMillis
	})

	return available, nil
}

// ListAll returns every registered relay server, healthy or not — unlike
// ListAvailable, nothing is filtered out. Used by admin tooling and by the
// Health Monitor, which must be able to see currently-unhealthy servers in
// order to detect when they recover.
func (s *Service) ListAll(ctx context.Context) ([]relay.Server, error) {
	all, err := s.repo.List(ctx)
	if err != nil {
		s.log.Error(ctx, err)
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to list relay servers", err)
	}
	for i := range all {
		all[i].Region = relay.Region(region.Normalize(string(all[i].Region)))
	}
	return all, nil
}

func (s *Service) Register(ctx context.Context, id relay.ID, regionCode relay.Region, host string, port int, connectionConfig string, registeredAt time.Time) (*relay.Server, error) {
	normalized := relay.Region(region.Normalize(string(regionCode)))
	if err := validateRegistration(id, normalized, host, port); err != nil {
		return nil, err
	}

	srv, err := s.repo.Register(ctx, relay.Server{
		ID:               id,
		Region:           normalized,
		Host:             host,
		Port:             port,
		ConnectionConfig: connectionConfig,
		UpdatedAt:        registeredAt,
	})
	if err != nil {
		s.log.Error(ctx, err)
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to register relay server", err)
	}
	return srv, nil
}

func validateRegistration(id relay.ID, regionCode relay.Region, host string, port int) error {
	if id == "" {
		return apperrors.New(apperrors.CodeInvalidInput, "id must not be empty").WithDetails(map[string]any{"field": "id"})
	}
	if strings.TrimSpace(string(regionCode)) == "" {
		return apperrors.New(apperrors.CodeInvalidInput, "region must not be empty").WithDetails(map[string]any{"field": "region"})
	}
	if !region.IsKnown(string(regionCode)) {
		return apperrors.New(apperrors.CodeInvalidInput, "region must be one of: de, nl, pl, fi").WithDetails(map[string]any{
			"field":   "region",
			"allowed": []string{"de", "nl", "pl", "fi"},
		})
	}
	if host == "" {
		return apperrors.New(apperrors.CodeInvalidInput, "host must not be empty").WithDetails(map[string]any{"field": "host"})
	}
	if port <= 0 || port > 65535 {
		return apperrors.New(apperrors.CodeInvalidInput, "port must be between 1 and 65535").WithDetails(map[string]any{"field": "port"})
	}
	return nil
}

func (s *Service) Delete(ctx context.Context, id relay.ID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		s.log.Error(ctx, err)
		return err
	}
	return nil
}

func (s *Service) RecordHealthCheck(ctx context.Context, id relay.ID, healthy bool, loadRatio float64, rttMillis int, checkedAt time.Time) error {
	if err := s.repo.UpdateHealth(ctx, id, healthy, loadRatio, rttMillis, checkedAt); err != nil {
		s.log.Error(ctx, err)
		return apperrors.Wrap(apperrors.CodeInternal, "failed to record relay health check", err)
	}
	return nil
}
