// Package rule contains the backend Rule Service's domain model: the
// versioned RuleSet served to clients over GET /rules (spec section 6).
// This is distinct from client-core/domain/rules, which is the on-device
// matching engine; the backend only owns authoring, storage and
// versioning of the rules clients download.
package rule

import (
	"time"

	apperrors "streampass/shared/errors"
)

// Kind is the type of match a Rule performs.
type Kind string

const (
	KindDomain Kind = "DOMAIN"
	KindCIDR   Kind = "CIDR"
)

// Mode is the routing decision a Rule maps to (spec section 5: DIRECT /
// RELAY / FALLBACK).
type Mode string

const (
	ModeDirect   Mode = "DIRECT"
	ModeRelay    Mode = "RELAY"
	ModeFallback Mode = "FALLBACK"
)

// Rule is a single routing rule (e.g. "*.ru -> DIRECT").
type Rule struct {
	Kind    Kind
	Pattern string
	Mode    Mode
}

// Set is an immutable, versioned collection of rules. Versioning closes
// the gap flagged in memory ("lack of rule list versioning/diffing"):
// clients can report the version they hold, and the backend can compute a
// diff or simply tell them "unchanged" instead of re-sending the full set
// every poll.
type Set struct {
	Version   int
	Rules     []Rule
	CreatedAt time.Time
}

// NewSet constructs a new immutable rule set version.
func NewSet(version int, rules []Rule, createdAt time.Time) *Set {
	// Defensive copy so callers can't mutate the slice they passed in
	// after construction (immutability of a published version matters:
	// clients cache by version number).
	owned := make([]Rule, len(rules))
	copy(owned, rules)
	return &Set{Version: version, Rules: owned, CreatedAt: createdAt}
}

// ErrNoRuleSet is returned by Repository.Latest when no rule set has ever
// been published yet (fresh deployment before the operator seeds rules).
var ErrNoRuleSet = apperrors.New(apperrors.CodeNotFound, "no rule set has been published yet")
