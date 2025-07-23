package currency

import (
	"time"
)

// ExchangeRateResponse represents the API response from exchangerate-api.com
type ExchangeRateResponse struct {
	Result            string             `json:"result"`
	Documentation     string             `json:"documentation"`
	TermsOfUse        string             `json:"terms_of_use"`
	TimeLastUpdateUnix int64             `json:"time_last_update_unix"`
	TimeLastUpdateUTC  string            `json:"time_last_update_utc"`
	TimeNextUpdateUnix int64             `json:"time_next_update_unix"`
	TimeNextUpdateUTC  string            `json:"time_next_update_utc"`
	BaseCode          string             `json:"base_code"`
	ConversionRates   map[string]float64 `json:"conversion_rates"`
}

// CachedRates stores exchange rates with metadata
type CachedRates struct {
	Rates      map[string]float64
	BaseCode   string
	FetchedAt  time.Time
	ValidUntil time.Time
}

// Converter interface defines currency conversion operations
type Converter interface {
	// Convert converts an amount from one currency to another
	Convert(amount float64, from, to string) (float64, error)
	
	// GetRate returns the exchange rate between two currencies
	GetRate(from, to string) (float64, error)
	
	// GetSupportedCurrencies returns a list of supported currency codes
	GetSupportedCurrencies() ([]string, error)
}

// Client interface for fetching exchange rates
type Client interface {
	// FetchRates fetches current exchange rates for a base currency
	FetchRates(baseCurrency string) (*ExchangeRateResponse, error)
}

// Cache interface for storing exchange rates
type Cache interface {
	// Get retrieves cached rates for a base currency
	Get(baseCurrency string) (*CachedRates, bool)
	
	// Set stores rates for a base currency with TTL
	Set(baseCurrency string, rates *CachedRates)
	
	// Clear removes all cached rates
	Clear()
}

// Config holds configuration for the currency converter
type Config struct {
	APIKey     string
	CacheTTL   time.Duration
	APITimeout time.Duration
}