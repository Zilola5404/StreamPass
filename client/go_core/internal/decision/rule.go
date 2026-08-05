package decision

// Mode is the per-connection routing decision (TZ section 5).
type Mode string

const (
	ModeDirect   Mode = "DIRECT"
	ModeRelay    Mode = "RELAY"
	ModeFallback Mode = "FALLBACK"
)

// Kind matches backend rule kinds (GET /api/v1/rules).
type Kind string

const (
	KindDomain Kind = "DOMAIN"
	KindCIDR   Kind = "CIDR"
)

// Rule is one routing rule from the published rule set.
type Rule struct {
	Kind    Kind
	Pattern string
	Mode    Mode
}

// RuleSet is the versioned rule list downloaded from the backend.
type RuleSet struct {
	Version int
	Rules   []Rule
}

// DefaultMode applies when nothing matches.
// StreamPass is a traffic accelerator (split tunnel), not a full-tunnel VPN:
// unknown destinations stay DIRECT unless explicitly forced to RELAY/FALLBACK.
const DefaultMode = ModeDirect
