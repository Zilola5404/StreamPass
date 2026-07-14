// Package rule is the Rule Service application layer (spec section 6):
// serving the current rule set to clients and letting an operator publish
// a new one. Business rules about what a "valid" rule set looks like live
// here, not in the HTTP handler or the Postgres repository.
package rule

import (
	"context"
	"time"

	"streampass/backend/internal/domain/rule"
	apperrors "streampass/shared/errors"
	"streampass/shared/logger"
)

// Clock is injected for deterministic tests, mirroring the auth package.
type Clock interface {
	Now() time.Time
}

// SystemClock is the production Clock implementation.
type SystemClock struct{}

// Now returns the current wall-clock time.
func (SystemClock) Now() time.Time { return time.Now().UTC() }

// Service implements the Rule Service use cases.
type Service struct {
	repo  rule.Repository
	clock Clock
	log   *logger.Logger
}

// NewService wires the Rule Service via constructor injection.
func NewService(repo rule.Repository, clock Clock, log *logger.Logger) *Service {
	return &Service{repo: repo, clock: clock, log: log.With("rule_service")}
}

// GetLatest implements "GET /rules": returns the current rule set.
func (s *Service) GetLatest(ctx context.Context) (*rule.Set, error) {
	set, err := s.repo.Latest(ctx)
	if err != nil {
		s.log.Error(ctx, err)
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to load rule set", err)
	}
	return set, nil
}

// Publish validates and stores a new rule set version. Used by the Admin
// Panel / operator tooling, not by end-user clients.
func (s *Service) Publish(ctx context.Context, rules []rule.Rule) (*rule.Set, error) {
	if err := validateRules(rules); err != nil {
		return nil, err
	}
	set, err := s.repo.Publish(ctx, rules, s.clock.Now())
	if err != nil {
		s.log.Error(ctx, err)
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to publish rule set", err)
	}
	return set, nil
}

func validateRules(rules []rule.Rule) error {
	if len(rules) == 0 {
		return apperrors.New(apperrors.CodeRuleSetInvalid, "rule set must not be empty")
	}
	for i, r := range rules {
		if r.Pattern == "" {
			return apperrors.New(apperrors.CodeRuleSetInvalid, "rule pattern must not be empty").
				WithDetails(map[string]any{"index": i})
		}
		switch r.Kind {
		case rule.KindDomain, rule.KindCIDR:
		default:
			return apperrors.New(apperrors.CodeRuleSetInvalid, "unsupported rule kind").
				WithDetails(map[string]any{"index": i, "kind": r.Kind})
		}
		switch r.Mode {
		case rule.ModeDirect, rule.ModeRelay, rule.ModeFallback:
		default:
			return apperrors.New(apperrors.CodeRuleSetInvalid, "unsupported rule mode").
				WithDetails(map[string]any{"index": i, "mode": r.Mode})
		}
	}
	return nil
}
