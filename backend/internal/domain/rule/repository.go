package rule

import (
	"context"
	"time"
)

// Repository is the port the Rule Service application layer depends on.
type Repository interface {
	// Latest returns the most recently published rule Set.
	Latest(ctx context.Context) (*Set, error)
	// Publish stores rules as a new version (version = previous max + 1,
	// assigned by the repository since "what's the next version number"
	// is a storage concern).
	Publish(ctx context.Context, rules []Rule, createdAt time.Time) (*Set, error)
}
