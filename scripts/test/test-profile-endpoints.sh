#!/bin/bash

# Configuration
MAIN_BASE_URL="${MAIN_BASE_URL:-http://localhost:8002/api/v1}"

# Load both credential files:
if [ -f "test-data-auth-email.txt" ] && [ -f "test-data-auth-phone.txt" ]; then
    source test-data-auth-email.txt
    EMAIL_TOKEN=$token
    EMAIL_CREDS=$email
    
    source test-data-auth-phone.txt  
    PHONE_TOKEN=$token
    PHONE_CREDS=$phone
    
    # Test 1: Use email token to link the phone number from phone registration
    # Test 2: Use phone token to link the email from email registration
else
    echo "ERROR: Need both email and phone credentials"
    exit 1
fi

# Test data
TIMESTAMP=$(date +%s%N)
LINK_PHONE="+629${TIMESTAMP:0:10}"
LINK_EMAIL="link${TIMESTAMP}@example.com"

echo "========================================="
echo "CORE/USER SERVICE TEST"
echo "Base URL: $MAIN_BASE_URL"
echo "========================================="
echo

# GET user profile
echo ">>> Test 1: Get user profile"
echo "Endpoint: GET $MAIN_BASE_URL/user"
echo -n "Response: "
curl -X GET "$MAIN_BASE_URL/user" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $EMAIL_TOKEN" \
  -w "\nHTTP Status: %{http_code}\n" \
  -s
echo
echo

# PUT user profile
echo ">>> Test 2: Update user profile"
echo "Endpoint: PUT $MAIN_BASE_URL/user"
USER_UPDATE_DATA="{\"bankAccountName\":\"CIMB\",\"bankAccountHolder\":\"Test User\",\"bankAccountNumber\":\"1234567890\"}"
echo "Body: $USER_UPDATE_DATA"
echo -n "Response: "
curl -X PUT "$MAIN_BASE_URL/user" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $EMAIL_TOKEN" \
  -d "$USER_UPDATE_DATA" \
  -w "\nHTTP Status: %{http_code}\n" \
  -s
echo
echo

# POST link phone
echo ">>> Test 3: Link phone number"
echo "Endpoint: POST $MAIN_BASE_URL/user/link/phone"
LINK_PHONE_DATA="{\"phone\":\"$LINK_PHONE\"}"
echo "Body: $LINK_PHONE_DATA"
echo -n "Response: "
curl -X POST "$MAIN_BASE_URL/user/link/phone" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $EMAIL_TOKEN" \
  -d "$LINK_PHONE_DATA" \
  -w "\nHTTP Status: %{http_code}\n" \
  -s
echo
echo

# POST link email
echo ">>> Test 4: Link email address"
echo "Endpoint: POST $MAIN_BASE_URL/user/link/email"
LINK_EMAIL_DATA="{\"email\":\"$LINK_EMAIL\"}"
echo "Body: $LINK_EMAIL_DATA"
echo -n "Response: "
curl -X POST "$MAIN_BASE_URL/user/link/email" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $PHONE_TOKEN" \
  -d "$LINK_EMAIL_DATA" \
  -w "\nHTTP Status: %{http_code}\n" \
  -s
echo
echo

# GET user profile again to verify updates
echo ">>> Test 5: Get user profile (verify updates)"
echo "Endpoint: GET $MAIN_BASE_URL/user"
echo -n "Response: "
curl -X GET "$MAIN_BASE_URL/user" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $token" \
  -w "\nHTTP Status: %{http_code}\n" \
  -s
echo
