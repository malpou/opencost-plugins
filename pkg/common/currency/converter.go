package currency

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type CurrencyConverter struct {
	client Client
	cache  Cache
	config Config
	mu     sync.RWMutex
}

func NewConverter(config Config) (*CurrencyConverter, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}
	
	if config.CacheTTL == 0 {
		config.CacheTTL = 24 * time.Hour
	}
	
	if config.APITimeout == 0 {
		config.APITimeout = 10 * time.Second
	}
	
	client := NewExchangeRateClient(config.APIKey, config.APITimeout)
	cache := NewMemoryCache(config.CacheTTL)
	
	return &CurrencyConverter{
		client: client,
		cache:  cache,
		config: config,
	}, nil
}

func (c *CurrencyConverter) Convert(amount float64, from, to string) (float64, error) {
	from = strings.ToUpper(strings.TrimSpace(from))
	to = strings.ToUpper(strings.TrimSpace(to))
	
	if from == to {
		return amount, nil
	}
	
	rate, err := c.GetRate(from, to)
	if err != nil {
		return 0, fmt.Errorf("failed to get exchange rate from %s to %s: %w", from, to, err)
	}
	
	return amount * rate, nil
}

func (c *CurrencyConverter) GetRate(from, to string) (float64, error) {
	from = strings.ToUpper(strings.TrimSpace(from))
	to = strings.ToUpper(strings.TrimSpace(to))
	
	if from == to {
		return 1.0, nil
	}
	
	cachedRates, found := c.cache.Get(from)
	if found && cachedRates.Rates != nil {
		if rate, exists := cachedRates.Rates[to]; exists {
			return rate, nil
		}
	}
	
	rates, err := c.fetchAndCacheRates(from)
	if err != nil {
		return 0, err
	}
	
	rate, exists := rates[to]
	if !exists {
		return 0, fmt.Errorf("currency %s not supported or not found in exchange rates", to)
	}
	
	return rate, nil
}

func (c *CurrencyConverter) GetSupportedCurrencies() ([]string, error) {
	rates, err := c.fetchAndCacheRates("USD")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch supported currencies: %w", err)
	}
	
	currencies := make([]string, 0, len(rates))
	for currency := range rates {
		currencies = append(currencies, currency)
	}
	
	return currencies, nil
}

func (c *CurrencyConverter) fetchAndCacheRates(baseCurrency string) (map[string]float64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if cachedRates, found := c.cache.Get(baseCurrency); found {
		return cachedRates.Rates, nil
	}
	
	response, err := c.client.FetchRates(baseCurrency)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch rates from API: %w", err)
	}
	
	cachedRates := &CachedRates{
		Rates:     response.ConversionRates,
		BaseCode:  response.BaseCode,
		FetchedAt: time.Now(),
	}
	c.cache.Set(baseCurrency, cachedRates)
	
	return response.ConversionRates, nil
}

func (c *CurrencyConverter) SetClient(client Client) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.client = client
}

func (c *CurrencyConverter) SetCache(cache Cache) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache = cache
}

func (c *CurrencyConverter) ClearCache() {
	c.cache.Clear()
}