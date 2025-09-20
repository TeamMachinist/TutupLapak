#!/bin/bash

# Configuration
MAIN_BASE_URL="${MAIN_BASE_URL:-http://localhost:8002/api/v1}"

# Load credentials from auth test (choose email or phone)
if [ -f "test-data-auth-email.txt" ]; then
    source test-data-auth-email.txt
    echo "Loaded email credentials"
elif [ -f "test-data-auth-phone.txt" ]; then
    source test-data-auth-phone.txt
    echo "Loaded phone credentials"
else
    echo "ERROR: No auth credentials found. Run test-data-auth-endpoints.sh first!"
    exit 1
fi

# Test data
TIMESTAMP=$(date +%s%N)
SHORT_TS=${TIMESTAMP:12:16}

echo "========================================="
echo "CORE/PRODUCT SERVICE TEST"
echo "Base URL: $MAIN_BASE_URL"
echo "Using token: ${token:0:30}..."
echo "========================================="
echo

# GET all products (public, no auth needed)
echo ">>> Test 1: Get all products (public)"
echo "Endpoint: GET $MAIN_BASE_URL/product"
echo -n "Response: "
curl -X GET "$MAIN_BASE_URL/product" \
  -H "Content-Type: application/json" \
  -w "\nHTTP Status: %{http_code}\n" \
  -s
echo
echo

# POST create first product
echo ">>> Test 2: Create first product"
echo "Endpoint: POST $MAIN_BASE_URL/product"
PRODUCT1_DATA="{\"name\":\"Laptop $SHORT_TS\",\"category\":\"electronic\",\"qty\":3,\"price\":15000000,\"sku\":\"ELECT$SHORT_TS\",\"fileId\":\"0199604f-78e9-759b-9d3e-7bab2a68c469\"}"
echo "Body: $PRODUCT1_DATA"
echo -n "Response: "
PRODUCT1_RESPONSE=$(curl -X POST "$MAIN_BASE_URL/product" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $token" \
  -d "$PRODUCT1_DATA" \
  -w "\nHTTP_STATUS:%{http_code}" \
  -s)

echo "$PRODUCT1_RESPONSE" | sed 's/HTTP_STATUS:/\nHTTP Status: /'
PRODUCT1_ID=$(echo "$PRODUCT1_RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
echo "Product ID: $PRODUCT1_ID"
echo
echo

# POST create second product
echo ">>> Test 3: Create second product"
echo "Endpoint: POST $MAIN_BASE_URL/product"
PRODUCT2_DATA="{\"name\":\"Kaos $SHORT_TS\",\"category\":\"clothing\",\"qty\":3,\"price\":80000000,\"sku\":\"CLOTH$SHORT_TS\",\"fileId\":\"0199604f-78e9-759b-9d3e-7bab2a68c469\"}"
echo "Body: $PRODUCT2_DATA"
echo -n "Response: "
PRODUCT2_RESPONSE=$(curl -X POST "$MAIN_BASE_URL/product" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $token" \
  -d "$PRODUCT2_DATA" \
  -w "\nHTTP_STATUS:%{http_code}" \
  -s)

echo "$PRODUCT2_RESPONSE" | sed 's/HTTP_STATUS:/\nHTTP Status: /'
PRODUCT2_ID=$(echo "$PRODUCT2_RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
echo "Product ID: $PRODUCT2_ID"
echo
echo

# GET products with filters
echo ">>> Test 4: Get products with filters"
echo "Endpoint: GET $MAIN_BASE_URL/product?category=electronic&limit=5"
echo -n "Response: "
curl -X GET "$MAIN_BASE_URL/product?category=electronic&limit=5" \
  -H "Content-Type: application/json" \
  -w "\nHTTP Status: %{http_code}\n" \
  -s
echo
echo

# PUT update product
if [ -n "$PRODUCT1_ID" ]; then
    echo ">>> Test 5: Update product"
    echo "Endpoint: PUT $MAIN_BASE_URL/product/$PRODUCT1_ID"
    UPDATE_DATA="{\"name\":\"Updated Laptop Gaming\",\"price\":14000000}"
    echo "Body: $UPDATE_DATA"
    echo -n "Response: "
    curl -X PUT "$MAIN_BASE_URL/product/$PRODUCT1_ID" \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $token" \
      -d "$UPDATE_DATA" \
      -w "\nHTTP Status: %{http_code}\n" \
      -s
    echo
    echo
fi

# DELETE product
if [ -n "$PRODUCT2_ID" ]; then
    echo ">>> Test 6: Delete product"
    echo "Endpoint: DELETE $MAIN_BASE_URL/product/$PRODUCT2_ID"
    echo -n "Response: "
    curl -X DELETE "$MAIN_BASE_URL/product/$PRODUCT2_ID" \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $token" \
      -w "\nHTTP Status: %{http_code}\n" \
      -s
    echo
    echo
fi

# GET all products again to see changes
echo ">>> Test 7: Get all products (verify changes)"
echo "Endpoint: GET $MAIN_BASE_URL/product"
echo -n "Response: "
curl -X GET "$MAIN_BASE_URL/product" \
  -H "Content-Type: application/json" \
  -w "\nHTTP Status: %{http_code}\n" \
  -s
echo