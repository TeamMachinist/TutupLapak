#!/bin/bash

# Configuration
MAIN_BASE_URL="${MAIN_BASE_URL:-http://localhost:8002/api/v1}"

# Test data
TIMESTAMP=$(date +%s%N)
SHORT_TS=${TIMESTAMP:12:16}

echo "========================================="
echo "CORE/PURCHASE SERVICE TEST"
echo "Base URL: $MAIN_BASE_URL"
echo "========================================="
echo

# POST create purchase with single item
echo ">>> Test 1: Create purchase with single item"
echo "Endpoint: POST $MAIN_BASE_URL/purchase"
PURCHASE1_DATA="{\"purchasedItems\":[{\"productId\":\"00000000-0000-0000-0000-100000000000\",\"qty\":2}],\"senderName\":\"John Doe Test\",\"senderContactType\":\"email\",\"senderContactDetail\":\"john.test@example.com\"}"
echo "Body: $PURCHASE1_DATA"
echo -n "Response: "
PURCHASE1_RESPONSE=$(curl -X POST "$MAIN_BASE_URL/purchase" \
  -H "Content-Type: application/json" \
  -d "$PURCHASE1_DATA" \
  -w "\nHTTP_STATUS:%{http_code}" \
  -s)

echo "$PURCHASE1_RESPONSE" | sed 's/HTTP_STATUS:/\nHTTP Status: /'
PURCHASE1_ID=$(echo "$PURCHASE1_RESPONSE" | grep -o '"purchaseId":"[^"]*"' | head -1 | cut -d'"' -f4)
echo "Purchase ID: $PURCHASE1_ID"
echo
echo

# POST create purchase with multiple items
echo ">>> Test 2: Create purchase with multiple items"
echo "Endpoint: POST $MAIN_BASE_URL/purchase"
PURCHASE2_DATA="{\"purchasedItems\":[{\"productId\":\"00000000-0000-0000-0000-200000000000\",\"qty\":2},{\"productId\":\"00000000-0000-0000-0000-300000000000\",\"qty\":2}],\"senderName\":\"Jane Smith\",\"senderContactType\":\"phone\",\"senderContactDetail\":\"+628123456789\"}"
echo "Body: $PURCHASE2_DATA"
echo -n "Response: "
PURCHASE2_RESPONSE=$(curl -X POST "$MAIN_BASE_URL/purchase" \
  -H "Content-Type: application/json" \
  -d "$PURCHASE2_DATA" \
  -w "\nHTTP_STATUS:%{http_code}" \
  -s)

echo "$PURCHASE2_RESPONSE" | sed 's/HTTP_STATUS:/\nHTTP Status: /'
PURCHASE2_ID=$(echo "$PURCHASE2_RESPONSE" | grep -o '"purchaseId":"[^"]*"' | head -1 | cut -d'"' -f4)
echo "Purchase ID: $PURCHASE2_ID"
echo
echo

# POST payment confirmation with file IDs
if [ -n "$PURCHASE1_ID" ]; then
    echo ">>> Test 3: Confirm purchase payment"
    echo "Endpoint: POST $MAIN_BASE_URL/purchase/$PURCHASE1_ID"
    PAYMENT_DATA="{\"fileIds\":[\"00000000-0000-0000-0000-000000000100\"]}"
    echo "Body: $PAYMENT_DATA"
    echo -n "Response: "
    curl -X POST "$MAIN_BASE_URL/purchase/$PURCHASE1_ID" \
      -H "Content-Type: application/json" \
      -d "$PAYMENT_DATA" \
      -w "\nHTTP Status: %{http_code}\n" \
      -s
    echo
    echo
fi

# POST payment with multiple file IDs (simulating multiple bank accounts)
if [ -n "$PURCHASE2_ID" ]; then
    echo ">>> Test 4: Confirm purchase with multiple payment proofs"
    echo "Endpoint: POST $MAIN_BASE_URL/purchase/$PURCHASE2_ID"
    PAYMENT_MULTI_DATA="{\"fileIds\":[\"00000000-0000-0000-0000-000000000100\",\"00000000-0000-0000-0000-000000000200\"]}"
    echo "Body: $PAYMENT_MULTI_DATA"
    echo -n "Response: "
    curl -X POST "$MAIN_BASE_URL/purchase/$PURCHASE2_ID" \
      -H "Content-Type: application/json" \
      -d "$PAYMENT_MULTI_DATA" \
      -w "\nHTTP Status: %{http_code}\n" \
      -s
    echo
    echo
fi
