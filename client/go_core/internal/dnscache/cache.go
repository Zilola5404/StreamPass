package dnscache

import (
	"fmt"
	"sync"
	"time"
)

type cacheEntry struct {
	raw     []byte
	expires time.Time
}

// Cache stores raw DNS response packets keyed by qtype+name.
type Cache struct {
	mu      sync.RWMutex
	entries map[string]cacheEntry
	maxSize int
}

// New returns a cache with a soft upper bound on entries.
func New(maxSize int) *Cache {
	if maxSize <= 0 {
		maxSize = 512
	}
	return &Cache{entries: make(map[string]cacheEntry), maxSize: maxSize}
}

func cacheKey(qtype uint16, name string) string {
	return fmt.Sprintf("%d|%s", qtype, name)
}

// GetRaw returns a cached DNS response wire packet if still fresh.
func (c *Cache) GetRaw(qtype uint16, name string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.entries[cacheKey(qtype, name)]
	if !ok || time.Now().After(e.expires) {
		return nil, false
	}
	return append([]byte(nil), e.raw...), true
}

// PutRaw stores a DNS response with the given TTL (clamped).
func (c *Cache) PutRaw(qtype uint16, name string, raw []byte, ttl time.Duration) {
	if ttl < 30*time.Second {
		ttl = 30 * time.Second
	}
	if ttl > 1*time.Hour {
		ttl = 1 * time.Hour
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.entries) >= c.maxSize {
		now := time.Now()
		for k, e := range c.entries {
			if now.After(e.expires) {
				delete(c.entries, k)
				break
			}
		}
		if len(c.entries) >= c.maxSize {
			for k := range c.entries {
				delete(c.entries, k)
				break
			}
		}
	}
	c.entries[cacheKey(qtype, name)] = cacheEntry{
		raw:     append([]byte(nil), raw...),
		expires: time.Now().Add(ttl),
	}
}
