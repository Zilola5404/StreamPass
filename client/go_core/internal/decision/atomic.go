package decision

import "sync"

// AtomicEngine wraps Engine for hot-reload from the Rule Engine (BL-006).
type AtomicEngine struct {
	mu        sync.RWMutex
	engine    *Engine
	version   int
	forceMode Mode // empty = normal; DIRECT/RELAY for network-mode tests
}

// NewAtomicEngine builds a hot-swappable engine.
func NewAtomicEngine(engine *Engine, version int) *AtomicEngine {
	if engine == nil {
		engine = NewEngine(nil, nil, DefaultMode)
	}
	return &AtomicEngine{engine: engine, version: version}
}

// Decide evaluates the current rule set.
func (a *AtomicEngine) Decide(t Target) Mode {
	return a.DecideDetailed(t).Mode
}

// DecideDetailed evaluates mode + rule/reason for diagnostics.
func (a *AtomicEngine) DecideDetailed(t Target) Decision {
	a.mu.RLock()
	force := a.forceMode
	eng := a.engine
	a.mu.RUnlock()
	if force != "" {
		return Decision{
			Mode:   force,
			Rule:   "network_mode",
			Source: "force",
			Reason: "network_mode_" + string(force),
		}
	}
	return eng.DecideDetailed(t)
}

// SetForceMode overrides all decisions (direct_test / full_relay diagnostics).
// Empty string clears the override.
func (a *AtomicEngine) SetForceMode(mode Mode) {
	a.mu.Lock()
	a.forceMode = mode
	a.mu.Unlock()
}

// Version returns the loaded rule set version (0 if unknown).
func (a *AtomicEngine) Version() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.version
}

// Update replaces rules and exclusions without restarting the tunnel.
func (a *AtomicEngine) Update(rulesJSON, exclusionsJSON string) error {
	set, err := ParseRuleSetJSON(rulesJSON)
	if err != nil {
		return err
	}
	exclusions, err := ParseExclusionsJSON(exclusionsJSON)
	if err != nil {
		return err
	}
	engine := NewEngine(MergeWithDefaults(set.Rules), exclusions, DefaultMode)
	a.mu.Lock()
	a.engine = engine
	a.version = set.Version
	a.mu.Unlock()
	return nil
}

// NewAtomicEngineFromJSON parses backend payloads and builds an AtomicEngine.
func NewAtomicEngineFromJSON(rulesJSON, exclusionsJSON string) (*AtomicEngine, error) {
	set, err := ParseRuleSetJSON(rulesJSON)
	if err != nil {
		return nil, err
	}
	exclusions, err := ParseExclusionsJSON(exclusionsJSON)
	if err != nil {
		return nil, err
	}
	return NewAtomicEngine(NewEngine(MergeWithDefaults(set.Rules), exclusions, DefaultMode), set.Version), nil
}
