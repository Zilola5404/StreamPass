package decision

import "sync"

// AtomicEngine wraps Engine for hot-reload from the Rule Engine (BL-006).
type AtomicEngine struct {
	mu      sync.RWMutex
	engine  *Engine
	version int
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
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.engine.Decide(t)
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
	engine := NewEngine(set.Rules, exclusions, DefaultMode)
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
	return NewAtomicEngine(NewEngine(set.Rules, exclusions, DefaultMode), set.Version), nil
}
