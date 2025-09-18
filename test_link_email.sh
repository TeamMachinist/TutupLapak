#!/bin/bash

# Test script for POST /v1/user/link/email endpoint

BASE_URL="http://localhost:8002"
ENDPOINT="/api/v1/user/link/email"

echo "Testing POST $BASE_URL$ENDPOINT"
echo "=================================="

# Test 1: Valid email
echo "Test 1: Valid email"
curl -X POST "$BASE_URL$ENDPOINT" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN_HERE" \
  -d '{"email": "test@example.com"}' \
  -w "\nHTTP Status: %{http_code}\n" \
  -s

echo -e "\n"

# Test 2: Invalid email format
echo "Test 2: Invalid email format"
curl -X POST "$BASE_URL$ENDPOINT" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN_HERE" \
  -d '{"email": "invalid-email"}' \
  -w "\nHTTP Status: %{http_code}\n" \
  -s

echo -e "\n"

# Test 3: Missing email
echo "Test 3: Missing email"
curl -X POST "$BASE_URL$ENDPOINT" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN_HERE" \
  -d '{}' \
  -w "\nHTTP Status: %{http_code}\n" \
  -s

echo -e "\n"

# Test 4: Empty email
echo "Test 4: Empty email"
curl -X POST "$BASE_URL$ENDPOINT" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN_HERE" \
  -d '{"email": ""}' \
  -w "\nHTTP Status: %{http_code}\n" \
  -s

echo -e "\n"

# Test 5: No authorization header
echo "Test 5: No authorization header"
curl -X POST "$BASE_URL$ENDPOINT" \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com"}' \
  -w "\nHTTP Status: %{http_code}\n" \
  -s

echo -e "\n"
echo "Test completed!"
