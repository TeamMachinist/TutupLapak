#!/bin/bash

# Configuration
FILE_BASE_URL="${FILE_BASE_URL:-http://localhost:8003/api/v1}"
FILE_PATH="${FILE_PATH:-/Your/absolute/path/to/the-file.png}"

# Load credentials from auth test
if [ -f "test-data-auth-email.txt" ]; then
    source test-data-auth-email.txt
    echo "Loaded email credentials"
else
    echo "ERROR: No auth credentials found. Run test-data-auth-endpoints.sh first!"
    exit 1
fi

# Test data
TIMESTAMP=$(date +%s%N)
SHORT_TS=${TIMESTAMP:12:16}

# Output files
UPLODADS_FILE="test-data-file.txt"

echo "========================================="
echo "FILE SERVICE TEST"
echo "Base URL: $FILE_BASE_URL"
echo "Using token: ${token:0:30}..."
echo "========================================="
echo

# POST upload first file
echo ">>> Test 1: Upload file"
echo "Endpoint: POST $FILE_BASE_URL/file"
echo "File: $FILE_PATH"
echo -n "Response: "
FILE_UPLOAD_RESPONSE=$(curl -X POST "$FILE_BASE_URL/file" \
  -H "Authorization: Bearer $token" \
  -F "file=@$FILE_PATH" \
  -w "\nHTTP_STATUS:%{http_code}" \
  -s)
echo "$FILE_UPLOAD_RESPONSE" | sed 's/HTTP_STATUS:/\nHTTP Status: /'

# Save first file data
FILE_ID_ONE=$(echo "$FILE_UPLOAD_RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
echo "file_id=$FILE_ID_ONE" > "$UPLODADS_FILE"
echo "Saved to $UPLODADS_FILE"
echo "File ID: $FILE_ID_ONE"
echo

# POST upload second file
echo ">>> Test 2: Upload file"
echo "Endpoint: POST $FILE_BASE_URL/file"
echo "File: $FILE_PATH"
echo -n "Response: "
FILE_UPLOAD_RESPONSE=$(curl -X POST "$FILE_BASE_URL/file" \
  -H "Authorization: Bearer $token" \
  -F "file=@$FILE_PATH" \
  -w "\nHTTP_STATUS:%{http_code}" \
  -s)
echo "$FILE_UPLOAD_RESPONSE" | sed 's/HTTP_STATUS:/\nHTTP Status: /'

# Save second file data
FILE_ID_TWO=$(echo "$FILE_UPLOAD_RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
echo "file_ids=$FILE_ID_ONE,$FILE_ID_TWO" >> "$UPLODADS_FILE"
echo "Saved to $UPLODADS_FILE"
echo "File ID: $FILE_ID_TWO"
echo

# GET one file
if [ -n "$FILE_ID_ONE" ]; then
    echo ">>> Test 3: Get one file by ID"
    echo "Endpoint: GET $FILE_BASE_URL/file/$FILE_ID_ONE"
    echo -n "Response: "
    curl -X GET "$FILE_BASE_URL/file/$FILE_ID_ONE" \
      -H "Content-Type: application/json" \
      -w "\nHTTP Status: %{http_code}\n" \
      -s
    echo
    echo
fi

# GET multiple files
echo ">>> Test 4: Get multiple files by IDs"
echo "Endpoint: GET $FILE_BASE_URL/file?id=$FILE_ID_ONE,$FILE_ID_TWO"
echo -n "Response: "
curl -X GET "$FILE_BASE_URL/file?id=$FILE_ID_ONE,$FILE_ID_TWO" \
  -H "Content-Type: application/json" \
  -w "\nHTTP Status: %{http_code}\n" \
  -s
echo
echo

# GET with empty IDs param
echo ">>> Test 5: Get multiple files with empty param"
echo "Endpoint: GET $FILE_BASE_URL/file?id="
echo -n "Response: "
curl -X GET "$FILE_BASE_URL/file?id=" \
  -H "Content-Type: application/json" \
  -w "\nHTTP Status: %{http_code}\n" \
  -s
echo
