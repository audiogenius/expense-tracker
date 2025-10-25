#!/bin/bash

# Debug script for authentication issues
# Usage: ./debug-auth.sh [API_URL]

API_URL=${1:-"http://localhost:8080"}

echo "🔍 Debugging authentication issues..."
echo "API URL: $API_URL"
echo ""

# Check if API is running
echo "1. Checking API health..."
if curl -s "$API_URL/health" > /dev/null; then
    echo "✅ API is running"
else
    echo "❌ API is not responding"
    exit 1
fi

# Check environment variables
echo ""
echo "2. Checking environment variables..."
echo "JWT_SECRET: ${JWT_SECRET:+SET}"
echo "TELEGRAM_BOT_TOKEN: ${TELEGRAM_BOT_TOKEN:+SET}"
echo "TELEGRAM_WHITELIST: ${TELEGRAM_WHITELIST:-NOT_SET}"

# Test login endpoint
echo ""
echo "3. Testing login endpoint..."
echo "Testing with test data..."

# Create test payload
TEST_PAYLOAD='{
  "id": "123456789",
  "username": "testuser",
  "first_name": "Test",
  "last_name": "User"
}'

echo "Payload: $TEST_PAYLOAD"

# Test login
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST \
  -H "Content-Type: application/json" \
  -d "$TEST_PAYLOAD" \
  "$API_URL/api/login")

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n -1)

echo "Response code: $HTTP_CODE"
echo "Response body: $BODY"

if [ "$HTTP_CODE" = "200" ]; then
    echo "✅ Login successful"
elif [ "$HTTP_CODE" = "403" ]; then
    echo "❌ User not in whitelist"
    echo "💡 Add your Telegram ID to TELEGRAM_WHITELIST"
elif [ "$HTTP_CODE" = "401" ]; then
    echo "❌ Authentication failed"
    echo "💡 Check JWT_SECRET and TELEGRAM_BOT_TOKEN"
else
    echo "❌ Unexpected response: $HTTP_CODE"
fi

echo ""
echo "4. Checking Docker containers..."
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep -E "(expense_|api|db)"

echo ""
echo "5. Checking logs..."
echo "Recent API logs:"
docker logs expense_api --tail 10 2>/dev/null || echo "No API container found"

echo ""
echo "🔧 Troubleshooting tips:"
echo "1. Make sure TELEGRAM_WHITELIST contains your Telegram ID or '*'"
echo "2. Verify JWT_SECRET is set and not empty"
echo "3. Check TELEGRAM_BOT_TOKEN is valid"
echo "4. Ensure database is running and accessible"
echo "5. Check CORS settings if accessing from different domain"
