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
