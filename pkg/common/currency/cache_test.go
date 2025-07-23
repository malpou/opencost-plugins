package currency

import (
	"testing"
	"time"
)

func TestMemoryCache_SetAndGet(t *testing.T) {
	cache := NewMemoryCache(1 * time.Hour)
	defer cache.Stop()
	
	// Test setting and getting rates
	rates := &CachedRates{
		Rates: map[string]float64{
			"EUR": 0.85,
			"GBP": 0.73,
		},
		BaseCode:  "USD",
		FetchedAt: time.Now(),
	}
	
	cache.Set("USD", rates)
	
	// Test successful get
	retrieved, found := cache.Get("USD")
	if !found {
		t.Error("expected to find cached rates")
	}
	
	if retrieved.BaseCode != "USD" {
		t.Errorf("expected base code USD, got %s", retrieved.BaseCode)
	}
	
	if len(retrieved.Rates) != 2 {
		t.Errorf("expected 2 rates, got %d", len(retrieved.Rates))
	}
	
	// Test non-existent key
	_, found = cache.Get("EUR")
	if found {
		t.Error("expected not to find rates for EUR")
	}
}

func TestMemoryCache_Expiration(t *testing.T) {
	// Use short TTL for testing
	cache := NewMemoryCache(100 * time.Millisecond)
	defer cache.Stop()
	
	rates := &CachedRates{
		Rates: map[string]float64{
			"EUR": 0.85,
		},
		BaseCode:  "USD",
		FetchedAt: time.Now(),
	}
	
	cache.Set("USD", rates)
	
	// Should find it immediately
	_, found := cache.Get("USD")
	if !found {
		t.Error("expected to find cached rates immediately")
	}
	
	// Wait for expiration
	time.Sleep(150 * time.Millisecond)
	
	// Should not find it after expiration
	_, found = cache.Get("USD")
	if found {
		t.Error("expected rates to be expired")
	}
}

func TestMemoryCache_Clear(t *testing.T) {
	cache := NewMemoryCache(1 * time.Hour)
	defer cache.Stop()
	
	// Add multiple entries
	for _, base := range []string{"USD", "EUR", "GBP"} {
		rates := &CachedRates{
			Rates:     map[string]float64{"TEST": 1.0},
			BaseCode:  base,
			FetchedAt: time.Now(),
		}
		cache.Set(base, rates)
	}
	
	// Verify all entries exist
	for _, base := range []string{"USD", "EUR", "GBP"} {
		_, found := cache.Get(base)
		if !found {
			t.Errorf("expected to find rates for %s", base)
		}
	}
	
	// Clear cache
	cache.Clear()
	
	// Verify all entries are gone
	for _, base := range []string{"USD", "EUR", "GBP"} {
		_, found := cache.Get(base)
		if found {
			t.Errorf("expected not to find rates for %s after clear", base)
		}
	}
}

func TestMemoryCache_Stats(t *testing.T) {
	cache := NewMemoryCache(1 * time.Hour)
	defer cache.Stop()
	
	// Initially empty
	entries, _ := cache.Stats()
	if entries != 0 {
		t.Errorf("expected 0 entries, got %d", entries)
	}
	
	// Add entries
	now := time.Now()
	for i, base := range []string{"USD", "EUR", "GBP"} {
		rates := &CachedRates{
			Rates:     map[string]float64{"TEST": 1.0},
			BaseCode:  base,
			FetchedAt: now.Add(time.Duration(i) * time.Minute),
		}
		cache.Set(base, rates)
	}
	
	entries, oldest := cache.Stats()
	if entries != 3 {
		t.Errorf("expected 3 entries, got %d", entries)
	}
	
	// The oldest should be the first one we added (USD)
	if !oldest.Equal(now) {
		t.Errorf("expected oldest entry to be %v, got %v", now, oldest)
	}
}

func TestMemoryCache_Cleanup(t *testing.T) {
	// Use very short TTL for testing
	cache := NewMemoryCache(50 * time.Millisecond)
	defer cache.Stop()
	
	// Add entry
	rates := &CachedRates{
		Rates:     map[string]float64{"EUR": 0.85},
		BaseCode:  "USD",
		FetchedAt: time.Now(),
	}
	cache.Set("USD", rates)
	
	// Verify it exists
	entries, _ := cache.Stats()
	if entries != 1 {
		t.Errorf("expected 1 entry, got %d", entries)
	}
	
	// Wait for cleanup cycle (janitor runs every TTL/2 = 25ms)
	// Wait a bit longer to ensure cleanup has run
	time.Sleep(100 * time.Millisecond)
	
	// Verify it's been cleaned up
	entries, _ = cache.Stats()
	if entries != 0 {
		t.Errorf("expected 0 entries after cleanup, got %d", entries)
	}
}