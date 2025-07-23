# Currency Package

Convert costs between currencies in OpenCost plugins using live exchange rates.

## Quick Start

```go
import "github.com/opencost/opencost-plugins/pkg/common/currency"

config := currency.Config{
    APIKey:   "your-api-key",
    CacheTTL: 24 * time.Hour,
}

converter, err := currency.NewConverter(config)
if err != nil {
    log.Fatal(err)
}

// Convert 100 USD to EUR
amount, err := converter.Convert(100.0, "USD", "EUR")
```

## Setup

Get a free API key from [exchangerate-api.com](https://www.exchangerate-api.com/) (1,500 requests/month).

## How it Works

The package fetches exchange rates and caches them for 24 hours. This keeps API usage low - most plugins use under 50 requests per month.

Supports all ISO 4217 currencies (161 total). Thread-safe with automatic cache cleanup.

## MongoDB Atlas Example

```go
// Config
type AtlasConfig struct {
    TargetCurrency  string `json:"target_currency"`
    ExchangeAPIKey  string `json:"exchange_api_key"`
}

// Usage
if atlasConfig.ExchangeAPIKey != "" {
    converter, _ := currency.NewConverter(currency.Config{
        APIKey:   atlasConfig.ExchangeAPIKey,
        CacheTTL: 24 * time.Hour,
    })
}

// Convert costs
if converter != nil {
    cost, _ = converter.Convert(cost, "USD", targetCurrency)
}
```

## Testing

```bash
cd pkg/common/currency
go test -v
```

Tests use mocks - no API calls needed.