package exclusion

import (
	"context"
	"regexp"
	"strings"

	domainexcl "streampass/backend/internal/domain/exclusion"
	"streampass/backend/internal/domain/user"
	apperrors "streampass/shared/errors"
	"streampass/shared/logger"
)

var domainPattern = regexp.MustCompile(`^(\*\.)?[a-zA-Zа-яА-Я0-9-]+(\.[a-zA-Zа-яА-Я0-9-]+)+$`)

// Service implements GET/PUT exclusions use cases.
type Service struct {
	repo domainexcl.Repository
	log  *logger.Logger
}

// NewService wires the exclusions application service.
func NewService(repo domainexcl.Repository, log *logger.Logger) *Service {
	return &Service{repo: repo, log: log.With("exclusion_service")}
}

// List returns the user's exclusion domains (never nil).
func (s *Service) List(ctx context.Context, userID user.ID) ([]string, error) {
	domains, err := s.repo.Get(ctx, userID)
	if err != nil {
		s.log.Error(ctx, err)
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to load exclusions", err)
	}
	if domains == nil {
		return []string{}, nil
	}
	return domains, nil
}

// Replace validates and overwrites the user's exclusion list.
func (s *Service) Replace(ctx context.Context, userID user.ID, domains []string) ([]string, error) {
	normalized, err := Normalize(domains)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Replace(ctx, userID, normalized); err != nil {
		s.log.Error(ctx, err)
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to save exclusions", err)
	}
	return normalized, nil
}

// Normalize trims, lowercases, dedupes, and validates domains.
func Normalize(domains []string) ([]string, error) {
	if len(domains) > domainexcl.MaxDomains {
		return nil, apperrors.New(apperrors.CodeInvalidInput, "too many exclusions (max 100)")
	}
	seen := make(map[string]struct{}, len(domains))
	out := make([]string, 0, len(domains))
	for _, raw := range domains {
		d := strings.ToLower(strings.TrimSpace(raw))
		if d == "" {
			continue
		}
		if !domainPattern.MatchString(d) {
			return nil, apperrors.New(apperrors.CodeInvalidInput, "invalid exclusion domain: "+d)
		}
		if _, ok := seen[d]; ok {
			continue
		}
		seen[d] = struct{}{}
		out = append(out, d)
	}
	return out, nil
}
