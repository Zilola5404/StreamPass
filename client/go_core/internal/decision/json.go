package decision

import (
	"encoding/json"
	"fmt"
	"strings"
)

type ruleSetDTO struct {
	Version int       `json:"version"`
	Rules   []ruleDTO `json:"rules"`
}

type ruleDTO struct {
	Kind    string `json:"kind"`
	Pattern string `json:"pattern"`
	Mode    string `json:"mode"`
}

// ParseRuleSetJSON decodes GET /api/v1/rules payload into a RuleSet.
func ParseRuleSetJSON(raw string) (RuleSet, error) {
	if raw == "" {
		return RuleSet{}, nil
	}
	var dto ruleSetDTO
	if err := json.Unmarshal([]byte(raw), &dto); err != nil {
		return RuleSet{}, fmt.Errorf("decode rule set: %w", err)
	}
	out := RuleSet{Version: dto.Version, Rules: make([]Rule, 0, len(dto.Rules))}
	for i, item := range dto.Rules {
		kind := Kind(strings.TrimSpace(item.Kind))
		mode := Mode(strings.TrimSpace(item.Mode))
		pattern := strings.TrimSpace(item.Pattern)
		if pattern == "" {
			return RuleSet{}, fmt.Errorf("rule %d: empty pattern", i)
		}
		switch kind {
		case KindDomain, KindCIDR:
		default:
			return RuleSet{}, fmt.Errorf("rule %d: unsupported kind %q", i, item.Kind)
		}
		switch mode {
		case ModeDirect, ModeRelay, ModeFallback:
		default:
			return RuleSet{}, fmt.Errorf("rule %d: unsupported mode %q", i, item.Mode)
		}
		out.Rules = append(out.Rules, Rule{Kind: kind, Pattern: pattern, Mode: mode})
	}
	return out, nil
}

// ParseExclusionsJSON decodes a JSON string array of domain patterns.
// Each exclusion is treated as a DIRECT domain rule (TZ user exclusions).
func ParseExclusionsJSON(raw string) ([]Rule, error) {
	if raw == "" {
		return nil, nil
	}
	var patterns []string
	if err := json.Unmarshal([]byte(raw), &patterns); err != nil {
		return nil, fmt.Errorf("decode exclusions: %w", err)
	}
	out := make([]Rule, 0, len(patterns))
	for _, p := range patterns {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, Rule{Kind: KindDomain, Pattern: p, Mode: ModeDirect})
	}
	return out, nil
}

// NewEngineFromJSON builds an engine from backend rules JSON and exclusions JSON.
func NewEngineFromJSON(rulesJSON, exclusionsJSON string) (*Engine, error) {
	set, err := ParseRuleSetJSON(rulesJSON)
	if err != nil {
		return nil, err
	}
	exclusions, err := ParseExclusionsJSON(exclusionsJSON)
	if err != nil {
		return nil, err
	}
	return NewEngine(set.Rules, exclusions, DefaultMode), nil
}
