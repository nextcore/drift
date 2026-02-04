package svg

import (
	"errors"
	"sync"
)

// IconCache caches loaded SVG icons by key.
//
// This is an opt-in helper for static assets to avoid reparsing SVG data and
// reduce paint churn when widgets rebuild.
type IconCache struct {
	mu    sync.Mutex
	items map[string]*Icon
}

// NewIconCache creates an empty icon cache.
func NewIconCache() *IconCache {
	return &IconCache{items: make(map[string]*Icon)}
}

// Get returns a cached icon or loads and caches it using the loader.
//
// If the cache is nil, the loader is invoked directly.
//
// Note: To avoid holding the lock during I/O, concurrent requests for the same
// uncached key may invoke the loader multiple times (thundering herd). Only one
// result is stored; duplicates are discarded. This is acceptable for typical
// icon loading at startup.
func (c *IconCache) Get(key string, loader func() (*Icon, error)) (*Icon, error) {
	if loader == nil {
		return nil, errors.New("svg: loader is nil")
	}
	if c == nil {
		return loader()
	}

	c.mu.Lock()
	if icon := c.items[key]; icon != nil {
		c.mu.Unlock()
		return icon, nil
	}
	c.mu.Unlock()

	icon, err := loader()
	if err != nil || icon == nil {
		return icon, err
	}

	c.mu.Lock()
	if existing := c.items[key]; existing != nil {
		c.mu.Unlock()
		return existing, nil
	}
	c.items[key] = icon
	c.mu.Unlock()

	return icon, nil
}
