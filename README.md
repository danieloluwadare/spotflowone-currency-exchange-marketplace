# Currency Exchange Marketplace API (Go)

API for vendor allocations and customer orders that aggregates liquidity using best-rate-first logic with precise decimal math.

## Overview
- **Allocations**: vendors publish available inventory for a currency pair.
- **Orders**: customers place buy/sell orders in base currency; system fills across vendors to get best rates.
- **Precision**: uses `shopspring/decimal` to avoid floating-point errors.
- **Rates**: market rate fetched from Frankfurter with cached fallback.

## Architecture
- `cmd/api`: entrypoint and wiring.
- `internal/domain`: core types and enums.
- `internal/service`: allocation and order services (matching logic).
- `internal/repo`: in-memory allocation repository.
- `internal/external/rates`: Frankfurter client + cached provider.
- `internal/transport/http`: Gorilla Mux handlers and schemas.

## Setup
```bash
go mod tidy
go run ./cmd/api
```

Server runs on `:8080` by default or `PORT` env var.

## API
### POST /allocations
Create a vendor allocation.

Request
```json
{
  "vendorId": "VENDOR_F",
  "baseCurrency": "USD",
  "quoteCurrency": "EUR",
  "rate": "1.23",
  "available": "4000"
}
```

Response `201`
```json
{
  "vendorId": "VENDOR_F",
  "baseCurrency": "USD",
  "quoteCurrency": "EUR",
  "rate": "1.23",
  "available": "4000"
}
```

### POST /orders
Place a buy/sell order.

Request
```json
{
  "direction": "BUY",
  "baseCurrency": "USD",
  "quoteCurrency": "EUR",
  "amount": "7000"
}
```

Response `200` (filled) or `206` (partial)
```json
{
  "status": "FILLED",
  "requestedAmount": "7000",
  "filledAmount": "7000",
  "fills": [
    { "vendorId": "VENDOR_B", "rate": "1.22", "amount": "3000" },
    { "vendorId": "VENDOR_A", "rate": "1.25", "amount": "4000" }
  ],
  "blendedRate": "1.237142857142857142",
  "marketRate": "1.24",
  "marketRateSource": "LIVE"
}
```

### Status Codes
- `400` invalid input.
- `422` no liquidity for pair.
- `200` filled, `206` partial fill.

## Order Matching Logic
- **BUY**: lowest rate first.
- **SELL**: highest rate first.
- Blended rate = sum(fillAmount * rate) / sum(fillAmount).

## Market Rate Fallback
If Frankfurter is unavailable, the service returns the **last known rate** from cache. If no cached rate exists, `marketRate` is `null`.

## Tests

### Unit Tests
```bash
go test ./...
```

### API Integration Tests
See [API_TESTS.md](API_TESTS.md) for comprehensive API endpoint testing guide with curl examples, test scenarios, and expected responses.

## Assumptions
- Orders are executed immediately (no persistence beyond response).
- Vendor allocations are mutable in memory and reduced after fills.
- Currency validation is syntactic (3-letter ISO codes).

## What I’d Improve With More Time
- Persistent storage (Postgres) and audit logging.
- Idempotency keys for orders.
- Concurrency-safe reservation system to avoid race conditions at scale.

# spotflowone-currency-exchange-marketplace
