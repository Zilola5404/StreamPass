// Package exclusion models per-user DIRECT domain overrides (ТЗ §6).
package exclusion

import (
	"context"

	"streampass/backend/internal/domain/user"
)

const MaxDomains = 100

// Repository persists a user's exclusion domain list.
type Repository interface {
	Get(ctx context.Context, userID user.ID) ([]string, error)
	Replace(ctx context.Context, userID user.ID, domains []string) error
}
