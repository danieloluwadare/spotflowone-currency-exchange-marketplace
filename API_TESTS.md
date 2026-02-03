# API Testing Guide

This document provides comprehensive test cases and examples for all API endpoints.

## Prerequisites

1. Start the server:

   **Option A: Run directly with Go**
   ```bash
   go run ./cmd/api
   ```

   **Option B: Run with Docker Compose (Recommended)**
   ```bash
   # Start the service
   docker-compose up -d

   # View logs
   docker-compose logs -f

   # Stop the service
   docker-compose down
   ```

   **Option C: Run with Docker (Manual)**
   ```bash
   # Build the Docker image
   docker build -t spotflowone-api .

   # Run the container
   docker run -d -p 8080:8080 --name spotflowone-api spotflowone-api

   # Stop and remove the container
   docker stop spotflowone-api
   docker rm spotflowone-api
   ```

2. Server runs on `http://localhost:8080` by default (or `PORT` environment variable).

3. All requests use `Content-Type: application/json`.

---

## 1. Allocation Endpoint Tests

### 1.1 Create Allocation - Success

**Request:**
```bash
curl -X POST http://localhost:8080/allocations \
  -H "Content-Type: application/json" \
  -d '{
    "vendorId": "VENDOR_F",
    "baseCurrency": "USD",
    "quoteCurrency": "EUR",
    "rate": "1.23",
    "available": "4000"
  }'
```

**Expected Response:** `201 Created`
```json
{
  "vendorId": "VENDOR_F",
  "baseCurrency": "USD",
  "quoteCurrency": "EUR",
  "rate": "1.23",
  "available": "4000"
}
```

### 1.2 Create Allocation - Invalid Currency Code

**Request:**
```bash
curl -X POST http://localhost:8080/allocations \
  -H "Content-Type: application/json" \
  -d '{
    "vendorId": "VENDOR_G",
    "baseCurrency": "INVALID",
    "quoteCurrency": "EUR",
    "rate": "1.20",
    "available": "1000"
  }'
```

**Expected Response:** `400 Bad Request`
```json
{
  "error": "invalid input"
}
```

### 1.3 Create Allocation - Same Base and Quote Currency

**Request:**
```bash
curl -X POST http://localhost:8080/allocations \
  -H "Content-Type: application/json" \
  -d '{
    "vendorId": "VENDOR_H",
    "baseCurrency": "USD",
    "quoteCurrency": "USD",
    "rate": "1.00",
    "available": "1000"
  }'
```

**Expected Response:** `400 Bad Request`
```json
{
  "error": "invalid input"
}
```

### 1.4 Create Allocation - Negative Rate

**Request:**
```bash
curl -X POST http://localhost:8080/allocations \
  -H "Content-Type: application/json" \
  -d '{
    "vendorId": "VENDOR_I",
    "baseCurrency": "USD",
    "quoteCurrency": "EUR",
    "rate": "-1.20",
    "available": "1000"
  }'
```

**Expected Response:** `400 Bad Request`
```json
{
  "error": "invalid input"
}
```

### 1.5 Create Allocation - Invalid JSON

**Request:**
```bash
curl -X POST http://localhost:8080/allocations \
  -H "Content-Type: application/json" \
  -d 'invalid json'
```

**Expected Response:** `400 Bad Request`
```json
{
  "error": "invalid JSON body"
}
```

---

## 2. Order Endpoint Tests

### 2.1 BUY Order - Fully Filled

**Request:**
```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{
    "direction": "BUY",
    "baseCurrency": "USD",
    "quoteCurrency": "EUR",
    "amount": "7000"
  }'
```

**Expected Response:** `200 OK`
```json
{
  "status": "FILLED",
  "requestedAmount": "7000",
  "filledAmount": "7000",
  "fills": [
    {
      "vendorId": "VENDOR_B",
      "rate": "1.22",
      "amount": "3000"
    },
    {
      "vendorId": "VENDOR_F",
      "rate": "1.23",
      "amount": "4000"
    }
  ],
  "blendedRate": "1.2257142857142857",
  "marketRate": "0.84459",
  "marketRateSource": "LIVE"
}
```

**Notes:**
- For BUY orders, allocations are sorted by **lowest rate first** (best for customer)
- VENDOR_B (1.22) is used before VENDOR_F (1.23)
- Blended rate is weighted average: (3000×1.22 + 4000×1.23) / 7000

### 2.2 SELL Order - Fully Filled

**Request:**
```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{
    "direction": "SELL",
    "baseCurrency": "USD",
    "quoteCurrency": "EUR",
    "amount": "5000"
  }'
```

**Expected Response:** `200 OK`
```json
{
  "status": "FILLED",
  "requestedAmount": "5000",
  "filledAmount": "5000",
  "fills": [
    {
      "vendorId": "VENDOR_C",
      "rate": "1.28",
      "amount": "5000"
    }
  ],
  "blendedRate": "1.28",
  "marketRate": "0.84459",
  "marketRateSource": "LIVE"
}
```

**Notes:**
- For SELL orders, allocations are sorted by **highest rate first** (best for customer)
- VENDOR_C (1.28) is selected as it offers the highest rate

### 2.3 BUY Order - Partial Fill

**Request:**
```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{
    "direction": "BUY",
    "baseCurrency": "USD",
    "quoteCurrency": "EUR",
    "amount": "20000"
  }'
```

**Expected Response:** `206 Partial Content`
```json
{
  "status": "PARTIALLY_FILLED",
  "requestedAmount": "20000",
  "filledAmount": "10000",
  "fills": [
    {
      "vendorId": "VENDOR_A",
      "rate": "1.25",
      "amount": "5000"
    },
    {
      "vendorId": "VENDOR_C",
      "rate": "1.28",
      "amount": "5000"
    }
  ],
  "blendedRate": "1.265",
  "marketRate": "0.84459",
  "marketRateSource": "LIVE"
}
```

**Notes:**
- Order requested 20000 but only 10000 was available
- Status is `PARTIALLY_FILLED`
- HTTP status code is `206 Partial Content`

### 2.4 BUY Order - GBP Currency Pair

**Request:**
```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{
    "direction": "BUY",
    "baseCurrency": "USD",
    "quoteCurrency": "GBP",
    "amount": "5000"
  }'
```

**Expected Response:** `200 OK`
```json
{
  "status": "FILLED",
  "requestedAmount": "5000",
  "filledAmount": "5000",
  "fills": [
    {
      "vendorId": "VENDOR_E",
      "rate": "0.84",
      "amount": "5000"
    }
  ],
  "blendedRate": "0.84",
  "marketRate": "0.73125",
  "marketRateSource": "LIVE"
}
```

**Notes:**
- Tests different currency pair (USD/GBP)
- VENDOR_E (0.84) selected as lowest rate for BUY
- Market rate fetched successfully for GBP

### 2.5 Order - No Liquidity

**Request:**
```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{
    "direction": "BUY",
    "baseCurrency": "USD",
    "quoteCurrency": "JPY",
    "amount": "1000"
  }'
```

**Expected Response:** `422 Unprocessable Entity`
```json
{
  "error": "no liquidity"
}
```

**Notes:**
- No allocations exist for USD/JPY pair
- Returns appropriate error status code

### 2.6 Order - Invalid Direction

**Request:**
```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{
    "direction": "INVALID",
    "baseCurrency": "USD",
    "quoteCurrency": "EUR",
    "amount": "1000"
  }'
```

**Expected Response:** `400 Bad Request`
```json
{
  "error": "invalid input"
}
```

### 2.7 Order - Zero Amount

**Request:**
```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{
    "direction": "BUY",
    "baseCurrency": "USD",
    "quoteCurrency": "EUR",
    "amount": "0"
  }'
```

**Expected Response:** `400 Bad Request`
```json
{
  "error": "invalid input"
}
```

### 2.8 Order - Negative Amount

**Request:**
```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{
    "direction": "BUY",
    "baseCurrency": "USD",
    "quoteCurrency": "EUR",
    "amount": "-1000"
  }'
```

**Expected Response:** `400 Bad Request`
```json
{
  "error": "invalid input"
}
```

### 2.9 Order - Invalid JSON

**Request:**
```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d 'invalid json'
```

**Expected Response:** `400 Bad Request`
```json
{
  "error": "invalid JSON body"
}
```

### 2.10 Order - Missing Required Fields

**Request:**
```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{
    "direction": "BUY",
    "baseCurrency": "USD"
  }'
```

**Expected Response:** `400 Bad Request`
```json
{
  "error": "invalid amount"
}
```

---

## 3. Test Scenarios Summary

### Success Cases
- ✅ Create allocation with valid data
- ✅ BUY order fully filled (best rate first)
- ✅ SELL order fully filled (best rate first)
- ✅ Partial fill scenario
- ✅ Different currency pairs (USD/EUR, USD/GBP)

### Error Cases
- ✅ Invalid currency codes
- ✅ Same base and quote currency
- ✅ Negative or zero amounts/rates
- ✅ No liquidity for currency pair
- ✅ Invalid direction
- ✅ Invalid JSON format
- ✅ Missing required fields

### Edge Cases
- ✅ Partial fills return 206 status
- ✅ Market rate fallback (when external API unavailable)
- ✅ Blended rate calculation accuracy
- ✅ Best-rate-first sorting (lowest for BUY, highest for SELL)

---

## 4. Automated Testing Script

Save the following as `test_api.sh`:

```bash
#!/bin/bash

BASE_URL="http://localhost:8080"

echo "=== Testing Allocation Endpoint ==="

# Test 1: Create allocation
echo "Test 1: Create allocation"
curl -s -X POST "$BASE_URL/allocations" \
  -H "Content-Type: application/json" \
  -d '{"vendorId":"VENDOR_F","baseCurrency":"USD","quoteCurrency":"EUR","rate":"1.23","available":"4000"}' \
  | jq .

# Test 2: Invalid currency
echo -e "\nTest 2: Invalid currency"
curl -s -X POST "$BASE_URL/allocations" \
  -H "Content-Type: application/json" \
  -d '{"vendorId":"VENDOR_G","baseCurrency":"INVALID","quoteCurrency":"EUR","rate":"1.20","available":"1000"}' \
  | jq .

echo -e "\n=== Testing Order Endpoint ==="

# Test 3: BUY order
echo "Test 3: BUY order"
curl -s -X POST "$BASE_URL/orders" \
  -H "Content-Type: application/json" \
  -d '{"direction":"BUY","baseCurrency":"USD","quoteCurrency":"EUR","amount":"7000"}' \
  | jq .

# Test 4: SELL order
echo -e "\nTest 4: SELL order"
curl -s -X POST "$BASE_URL/orders" \
  -H "Content-Type: application/json" \
  -d '{"direction":"SELL","baseCurrency":"USD","quoteCurrency":"EUR","amount":"5000"}' \
  | jq .

# Test 5: Partial fill
echo -e "\nTest 5: Partial fill"
curl -s -X POST "$BASE_URL/orders" \
  -H "Content-Type: application/json" \
  -d '{"direction":"BUY","baseCurrency":"USD","quoteCurrency":"EUR","amount":"20000"}' \
  | jq .

# Test 6: No liquidity
echo -e "\nTest 6: No liquidity"
curl -s -X POST "$BASE_URL/orders" \
  -H "Content-Type: application/json" \
  -d '{"direction":"BUY","baseCurrency":"USD","quoteCurrency":"JPY","amount":"1000"}' \
  | jq .

echo -e "\n=== All tests completed ==="
```

Make it executable and run:
```bash
chmod +x test_api.sh
./test_api.sh
```

---

## 5. Expected Behavior

### Best-Rate-First Logic
- **BUY orders**: Allocations sorted by **lowest rate first** (most favorable to customer)
- **SELL orders**: Allocations sorted by **highest rate first** (most favorable to customer)

### Blended Rate Calculation
```
blendedRate = sum(fillAmount × rate) / sum(fillAmount)
```

### Market Rate
- Fetched from Frankfurter API when available (Source: "LIVE")
- Falls back to cached rate if API unavailable (Source: "CACHED")
- Returns `null` if no cached rate exists

### Status Codes
- `200 OK`: Order fully filled
- `206 Partial Content`: Order partially filled
- `201 Created`: Allocation created successfully
- `400 Bad Request`: Invalid input
- `422 Unprocessable Entity`: No liquidity available
- `500 Internal Server Error`: Server error

---

## 6. Seed Data

The application is pre-seeded with the following allocations on startup:

| Vendor | Base | Quote | Rate | Available |
|--------|------|-------|------|-----------|
| VENDOR_A | USD | EUR | 1.25 | 5000 |
| VENDOR_B | USD | EUR | 1.22 | 3000 |
| VENDOR_C | USD | EUR | 1.28 | 10000 |
| VENDOR_D | USD | GBP | 0.86 | 8000 |
| VENDOR_E | USD | GBP | 0.84 | 6000 |

These allocations are available immediately after server startup for testing.
