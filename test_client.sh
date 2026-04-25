#!/bin/bash

# Configuration
API_URL="http://localhost:8080"
# Generate valid UUIDs for testing
VOUCHER_CODE="RACE_TEST_$(date +%s)"
USER_ID=$(uuidgen | tr '[:upper:]' '[:lower:]')

function verify() {
    local label=$1
    local expected=$2
    local actual=$3
    if [[ "$actual" == *"$expected"* ]]; then
        echo "[PASS] $label"
    else
        echo "[FAIL] $label (Expected: $expected, Actual: $actual)"
        exit 1
    fi
}

echo "--- 1. Creating a Voucher (limit: 5 uses) ---"
resp=$(curl -s -X POST "$API_URL/vouchers" \
  -H "Content-Type: application/json" \
  -d "{
    \"code\": \"$VOUCHER_CODE\",
    \"discount_type\": \"flat\",
    \"discount_value\": 10000,
    \"max_uses\": 5,
    \"expires_at\": \"2026-12-31T23:59:59Z\",
    \"min_transaction_amount\": 50000
  }")
echo "$resp"
verify "Voucher creation code" "$VOUCHER_CODE" "$(echo "$resp" | jq -r .code)"

echo  "--- 2. Getting Voucher Details ---"
resp=$(curl -s -X GET "$API_URL/vouchers/$VOUCHER_CODE")
echo "$resp"
verify "Get voucher remaining_claims" "5" "$(echo "$resp" | jq -r .remaining_claims)"

echo  "--- 3. Single Claim Test ---"
resp=$(curl -s -X POST "$API_URL/vouchers/$VOUCHER_CODE/claim" \
  -H "Content-Type: application/json" \
  -d "{\"user_id\": \"$USER_ID\"}")
echo "$resp"
verify "Claim user_id" "$USER_ID" "$(echo "$resp" | jq -r .user_id)"

echo  "--- 4. Single Redeem Test ---"
resp=$(curl -s -X POST "$API_URL/vouchers/$VOUCHER_CODE/redeem" \
  -H "Content-Type: application/json" \
  -d "{
    \"user_id\": \"$USER_ID\",
    \"transaction_amount\": 100000
  }")
echo "$resp"
verify "Redeem final_amount" "90000" "$(echo "$resp" | jq -r .final_amount)"

echo  "--- 5. Concurrency Test with 'ab' ---"
BENCH_USER=$(uuidgen | tr '[:upper:]' '[:lower:]')
echo "{\"user_id\": \"$BENCH_USER\"}" > /tmp/claim_payload.json
ab -n 100 -c 10 -p /tmp/claim_payload.json -T "application/json" "$API_URL/vouchers/$VOUCHER_CODE/claim" > /dev/null
echo "Concurrency test finished."

echo  "--- 6. Verifying Final State ---"
resp=$(curl -s -X GET "$API_URL/vouchers/$VOUCHER_CODE")
echo "$resp"
verify "Voucher remaining_claims" "3" "$(echo "$resp" | jq -r .remaining_claims)"

echo  "--- 7. Exhaustion Test (Concurrent unique users) ---"
for i in {1..10}; do
  RAND_USER=$(uuidgen | tr '[:upper:]' '[:lower:]')
  curl -s -X POST "$API_URL/vouchers/$VOUCHER_CODE/claim" \
    -H "Content-Type: application/json" \
    -d "{\"user_id\": \"$RAND_USER\"}" > /dev/null &
done
wait
echo "Exhaustion burst finished."

echo  "--- 8. Verifying Limit is Enforced ---"
resp=$(curl -s -X GET "$API_URL/vouchers/$VOUCHER_CODE")
echo "$resp"
verify "Voucher remaining_claims" "0" "$(echo "$resp" | jq -r .remaining_claims)"

echo  "All tests passed successfully!"
rm /tmp/claim_payload.json
