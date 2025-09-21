#!/bin/bash

# Configuration
AUTH_BASE_URL="${AUTH_BASE_URL:-http://localhost:8001/api/v1}"
TIMESTAMP=$(date +%s%N)

# Test data
TEST_EMAIL="test${TIMESTAMP}@example.com"
TEST_PHONE="+628${TIMESTAMP:0:10}"
TEST_PASSWORD="TestPass123"

# Output files
EMAIL_CREDS_FILE="test-data-auth-email.txt"
PHONE_CREDS_FILE="test-data-auth-phone.txt"

echo "========================================="
echo "AUTH SERVICE TEST"
echo "Base URL: $AUTH_BASE_URL"
echo "========================================="
echo

echo "========================================="
echo "1. EMAIL AUTHENTICATION"
echo "========================================="
echo

# Register with email
echo ">>> Test 1: Register with email"
echo "Endpoint: POST $AUTH_BASE_URL/register/email"
echo "Body: {\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}"
echo -n "Response: "
curl -X POST "$AUTH_BASE_URL/register/email" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}" \
  -w "\nHTTP Status: %{http_code}\n" \
  -s
echo
echo

# Login with email  
echo ">>> Test 2: Login with email"
echo "Endpoint: POST $AUTH_BASE_URL/login/email"
echo "Body: {\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}"
echo -n "Response: "
LOGIN_RESPONSE=$(curl -X POST "$AUTH_BASE_URL/login/email" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}" \
  -w "\nHTTP_STATUS:%{http_code}" \
  -s)

# Print response with status
echo "$LOGIN_RESPONSE" | sed 's/HTTP_STATUS:/\nHTTP Status: /'
echo

# Extract token
EMAIL_TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
if [ -z "$EMAIL_TOKEN" ]; then
    EMAIL_TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"accessToken":"[^"]*"' | cut -d'"' -f4)
fi

# Save email credentials
echo "email=$TEST_EMAIL" > "$EMAIL_CREDS_FILE"
echo "password=$TEST_PASSWORD" >> "$EMAIL_CREDS_FILE"
echo "token=$EMAIL_TOKEN" >> "$EMAIL_CREDS_FILE"

echo "Saved to $EMAIL_CREDS_FILE"
echo "Token: ${EMAIL_TOKEN:0:30}..."
echo

echo "========================================="
echo "2. PHONE AUTHENTICATION"
echo "========================================="
echo

# Register with phone
echo ">>> Test 3: Register with phone"
echo "Endpoint: POST $AUTH_BASE_URL/register/phone"
echo "Body: {\"phone\":\"$TEST_PHONE\",\"password\":\"$TEST_PASSWORD\"}"
echo -n "Response: "
curl -X POST "$AUTH_BASE_URL/register/phone" \
  -H "Content-Type: application/json" \
  -d "{\"phone\":\"$TEST_PHONE\",\"password\":\"$TEST_PASSWORD\"}" \
  -w "\nHTTP Status: %{http_code}\n" \
  -s
echo
echo

# Login with phone
echo ">>> Test 4: Login with phone"
echo "Endpoint: POST $AUTH_BASE_URL/login/phone"
echo "Body: {\"phone\":\"$TEST_PHONE\",\"password\":\"$TEST_PASSWORD\"}"
echo -n "Response: "
LOGIN_RESPONSE=$(curl -X POST "$AUTH_BASE_URL/login/phone" \
  -H "Content-Type: application/json" \
  -d "{\"phone\":\"$TEST_PHONE\",\"password\":\"$TEST_PASSWORD\"}" \
  -w "\nHTTP_STATUS:%{http_code}" \
  -s)

# Print response with status
echo "$LOGIN_RESPONSE" | sed 's/HTTP_STATUS:/\nHTTP Status: /'
echo

# Save phone credentials
PHONE_TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

echo "phone=$TEST_PHONE" > "$PHONE_CREDS_FILE"
echo "password=$TEST_PASSWORD" >> "$PHONE_CREDS_FILE"
echo "token=$PHONE_TOKEN" >> "$PHONE_CREDS_FILE"

echo "Saved to $PHONE_CREDS_FILE"
echo "Token: ${PHONE_TOKEN:0:30}..."
echo