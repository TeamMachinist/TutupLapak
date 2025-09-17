#!/bin/bash

echo "🧪 Testing POST /v1/user/link/phone endpoint"
echo "============================================="

# Get JWT token from auth service
echo "1. Getting JWT token..."
TOKEN=$(curl -s -X POST http://localhost:8001/api/v1/token \
  -H "Content-Type: application/json" \
  -d '{"user_id": "11111111-1111-1111-1111-111111111111"}' | jq -r '.token')

if [ "$TOKEN" = "null" ] || [ -z "$TOKEN" ]; then
    echo "❌ Failed to get JWT token"
    exit 1
fi

echo "✅ JWT token obtained: ${TOKEN:0:20}..."

# Test 1: Valid phone number
echo ""
echo "2. Testing with valid phone number..."
curl -X POST http://localhost:8002/api/v1/user/link/phone \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"phone": "+1234567890"}' \
  -w "\nStatus: %{http_code}\n"

# Test 2: Invalid phone number (no + prefix)
echo ""
echo "3. Testing with invalid phone number (no + prefix)..."
curl -X POST http://localhost:8002/api/v1/user/link/phone \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"phone": "1234567890"}' \
  -w "\nStatus: %{http_code}\n"

# Test 3: Missing Authorization header
echo ""
echo "4. Testing without Authorization header..."
curl -X POST http://localhost:8002/api/v1/user/link/phone \
  -H "Content-Type: application/json" \
  -d '{"phone": "+1234567890"}' \
  -w "\nStatus: %{http_code}\n"

# Test 4: Invalid token
echo ""
echo "5. Testing with invalid token..."
curl -X POST http://localhost:8002/api/v1/user/link/phone \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer invalid-token" \
  -d '{"phone": "+1234567890"}' \
  -w "\nStatus: %{http_code}\n"

echo ""
echo "6. Testing POST /v1/user/link/email endpoint..."
echo "============================================="

# Test 1: Valid email
echo ""
echo "6.1 Testing with valid email..."
curl -X POST http://localhost:8002/api/v1/user/link/email \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"email": "test@example.com"}' \
  -w "\nStatus: %{http_code}\n"

# Test 2: Invalid email format
echo ""
echo "6.2 Testing with invalid email format..."
curl -X POST http://localhost:8002/api/v1/user/link/email \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"email": "invalid-email"}' \
  -w "\nStatus: %{http_code}\n"

# Test 3: Missing email
echo ""
echo "6.3 Testing with missing email..."
curl -X POST http://localhost:8002/api/v1/user/link/email \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{}' \
  -w "\nStatus: %{http_code}\n"

# Test 4: Empty email
echo ""
echo "6.4 Testing with empty email..."
curl -X POST http://localhost:8002/api/v1/user/link/email \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"email": ""}' \
  -w "\nStatus: %{http_code}\n"

echo ""
echo "🎯 Test completed!"


