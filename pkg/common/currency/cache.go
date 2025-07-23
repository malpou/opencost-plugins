package currency

import (
	"sync"
	"time"
)

type MemoryCache struct {
	mu      sync.RWMutex
	data    map[string]*CachedRates
	ttl     time.Duration
	janitor *time.Ticker
}

func NewMemoryCache(ttl time.Duration) *MemoryCache {
	if ttl == 0 {
		ttl = 24 * time.Hour
	}
	
	cache := &MemoryCache{
		data:    make(map[string]*CachedRates),
		ttl:     ttl,
		janitor: time.NewTicker(ttl / 2),
	}
	
	go cache.cleanup()
	
	return cache
}

func (c *MemoryCache) Get(baseCurrency string) (*CachedRates, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	rates, exists := c.data[baseCurrency]
	if !exists {
		return nil, false
	}
	
	if time.Now().After(rates.ValidUntil) {
		return nil, false
	}
	
	return rates, true
}

func (c *MemoryCache) Set(baseCurrency string, rates *CachedRates) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	rates.ValidUntil = rates.FetchedAt.Add(c.ttl)
	c.data[baseCurrency] = rates
}

func (c *MemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.data = make(map[string]*CachedRates)
}

func (c *MemoryCache) cleanup() {
	for range c.janitor.C {
		c.removeExpired()
	}
}

func (c *MemoryCache) removeExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	now := time.Now()
	for key, rates := range c.data {
		if now.After(rates.ValidUntil) {
			delete(c.data, key)
		}
	}
}

func (c *MemoryCache) Stop() {
	if c.janitor != nil {
		c.janitor.Stop()
	}
}

func (c *MemoryCache) Stats() (entries int, oldestEntry time.Time) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	entries = len(c.data)
	
	for _, rates := range c.data {
		if oldestEntry.IsZero() || rates.FetchedAt.Before(oldestEntry) {
			oldestEntry = rates.FetchedAt
		}
	}
	
	return entries, oldestEntry
}