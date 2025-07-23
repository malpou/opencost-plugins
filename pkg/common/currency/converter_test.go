package currency

import (
	"fmt"
	"testing"
	"time"
)

type MockClient struct {
	rates map[string]map[string]float64
	err   error
}

func (m *MockClient) FetchRates(baseCurrency string) (*ExchangeRateResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	
	rates, exists := m.rates[baseCurrency]
	if !exists {
		return nil, fmt.Errorf("no rates for base currency %s", baseCurrency)
	}
	
	return &ExchangeRateResponse{
		Result:          "success",
		BaseCode:        baseCurrency,
		ConversionRates: rates,
	}, nil
}

type MockCache struct {
	data map[string]*CachedRates
}

func NewMockCache() *MockCache {
	return &MockCache{
		data: make(map[string]*CachedRates),
	}
}

func (m *MockCache) Get(baseCurrency string) (*CachedRates, bool) {
	rates, exists := m.data[baseCurrency]
	if !exists || time.Now().After(rates.ValidUntil) {
		return nil, false
	}
	return rates, true
}

func (m *MockCache) Set(baseCurrency string, rates *CachedRates) {
	m.data[baseCurrency] = rates
}

func (m *MockCache) Clear() {
	m.data = make(map[string]*CachedRates)
}

func TestCurrencyConverter_Convert(t *testing.T) {
	mockClient := &MockClient{
		rates: map[string]map[string]float64{
			"USD": {
				"USD": 1.0,
				"EUR": 0.85,
				"GBP": 0.73,
				"JPY": 110.0,
			},
			"EUR": {
				"EUR": 1.0,
				"USD": 1.18,
				"GBP": 0.86,
				"JPY": 129.53,
			},
		},
	}
	
	converter := &CurrencyConverter{
		client: mockClient,
		cache:  NewMockCache(),
		config: Config{APIKey: "test"},
	}
	
	tests := []struct {
		name        string
		amount      float64
		from        string
		to          string
		expected    float64
		expectError bool
	}{
		{
			name:     "USD to EUR",
			amount:   100,
			from:     "USD",
			to:       "EUR",
			expected: 85,
		},
		{
			name:     "USD to GBP",
			amount:   100,
			from:     "USD",
			to:       "GBP",
			expected: 73,
		},
		{
			name:     "EUR to USD",
			amount:   100,
			from:     "EUR",
			to:       "USD",
			expected: 118,
		},
		{
			name:     "Same currency",
			amount:   100,
			from:     "USD",
			to:       "USD",
			expected: 100,
		},
		{
			name:     "Case insensitive",
			amount:   100,
			from:     "usd",
			to:       "eur",
			expected: 85,
		},
		{
			name:        "Unsupported currency",
			amount:      100,
			from:        "USD",
			to:          "XYZ",
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.Convert(tt.amount, tt.from, tt.to)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			
			if result != tt.expected {
				t.Errorf("expected %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestCurrencyConverter_GetRate(t *testing.T) {
	mockClient := &MockClient{
		rates: map[string]map[string]float64{
			"USD": {
				"USD": 1.0,
				"EUR": 0.85,
				"GBP": 0.73,
			},
		},
	}
	
	converter := &CurrencyConverter{
		client: mockClient,
		cache:  NewMockCache(),
		config: Config{APIKey: "test"},
	}
	
	// Test getting rate
	rate, err := converter.GetRate("USD", "EUR")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if rate != 0.85 {
		t.Errorf("expected rate 0.85, got %f", rate)
	}
	
	// Test same currency
	rate, err = converter.GetRate("USD", "USD")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if rate != 1.0 {
		t.Errorf("expected rate 1.0, got %f", rate)
	}
	
	// Test cache hit
	rate, err = converter.GetRate("USD", "EUR")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if rate != 0.85 {
		t.Errorf("expected cached rate 0.85, got %f", rate)
	}
}

func TestCurrencyConverter_GetSupportedCurrencies(t *testing.T) {
	mockClient := &MockClient{
		rates: map[string]map[string]float64{
			"USD": {
				"USD": 1.0,
				"EUR": 0.85,
				"GBP": 0.73,
				"JPY": 110.0,
			},
		},
	}
	
	converter := &CurrencyConverter{
		client: mockClient,
		cache:  NewMockCache(),
		config: Config{APIKey: "test"},
	}
	
	currencies, err := converter.GetSupportedCurrencies()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	
	if len(currencies) != 4 {
		t.Errorf("expected 4 currencies, got %d", len(currencies))
	}
	
	// Check that all expected currencies are present
	expected := map[string]bool{
		"USD": false,
		"EUR": false,
		"GBP": false,
		"JPY": false,
	}
	
	for _, curr := range currencies {
		if _, exists := expected[curr]; exists {
			expected[curr] = true
		}
	}
	
	for curr, found := range expected {
		if !found {
			t.Errorf("expected currency %s not found", curr)
		}
	}
}

func TestNewConverter(t *testing.T) {
	// Test with empty API key
	_, err := NewConverter(Config{})
	if err == nil {
		t.Error("expected error for empty API key")
	}
	
	// Test with valid config
	converter, err := NewConverter(Config{APIKey: "test-key"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	
	if converter.config.CacheTTL != 24*time.Hour {
		t.Errorf("expected default cache TTL of 24h, got %v", converter.config.CacheTTL)
	}
	
	if converter.config.APITimeout != 10*time.Second {
		t.Errorf("expected default API timeout of 10s, got %v", converter.config.APITimeout)
	}
}