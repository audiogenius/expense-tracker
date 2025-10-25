#!/bin/bash

# Test script for authentication fixes
# Usage: ./test-auth.sh [API_URL]

API_URL=${1:-"http://localhost:8080"}

echo "🧪 Testing authentication fixes..."
echo "API URL: $API_URL"
echo ""

# Test 1: Health check
echo "1. Testing API health..."
if curl -s "$API_URL/health" | grep -q "ok"; then
    echo "✅ API is healthy"
else
    echo "❌ API health check failed"
    exit 1
fi

# Test 2: CORS headers
echo ""
echo "2. Testing CORS headers..."
CORS_HEADERS=$(curl -s -I -X OPTIONS "$API_URL/api/login" | grep -i "access-control")
if [ -n "$CORS_HEADERS" ]; then
    echo "✅ CORS headers present"
    echo "$CORS_HEADERS"
else
    echo "❌ CORS headers missing"
fi

# Test 3: Login endpoint accessibility
echo ""
echo "3. Testing login endpoint..."
LOGIN_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST \
  -H "Content-Type: application/json" \
  -d '{"id": "test", "username": "test"}' \
  "$API_URL/api/login")

HTTP_CODE=$(echo "$LOGIN_RESPONSE" | tail -n1)
BODY=$(echo "$LOGIN_RESPONSE" | head -n -1)

echo "Response code: $HTTP_CODE"
echo "Response body: $BODY"

if [ "$HTTP_CODE" = "403" ]; then
    echo "✅ Login endpoint working (403 = user not whitelisted, which is expected)"
elif [ "$HTTP_CODE" = "400" ]; then
    echo "✅ Login endpoint working (400 = bad request, which is expected)"
elif [ "$HTTP_CODE" = "200" ]; then
    echo "✅ Login successful"
else
    echo "❌ Unexpected response: $HTTP_CODE"
fi

echo ""
echo "4. Testing with whitelisted user (if TELEGRAM_WHITELIST=*)..."
if [ "$TELEGRAM_WHITELIST" = "*" ]; then
    echo "Testing with test user..."
    TEST_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST \
      -H "Content-Type: application/json" \
      -d '{
        "id": "123456789",
        "username": "testuser",
        "first_name": "Test",
        "last_name": "User"
      }' \
      "$API_URL/api/login")
    
    TEST_HTTP_CODE=$(echo "$TEST_RESPONSE" | tail -n1)
    TEST_BODY=$(echo "$TEST_RESPONSE" | head -n -1)
    
    echo "Response code: $TEST_HTTP_CODE"
    if [ "$TEST_HTTP_CODE" = "200" ]; then
        echo "✅ Login successful with test user"
        echo "Token received: $(echo "$TEST_BODY" | grep -o '"token":"[^"]*"' | cut -d'"' -f4 | head -c 20)..."
    else
        echo "❌ Login failed with test user: $TEST_HTTP_CODE"
        echo "Response: $TEST_BODY"
    fi
else
    echo "⚠️  TELEGRAM_WHITELIST is not '*', skipping test user login"
fi

echo ""
echo "🎉 Authentication test completed!"
echo ""
echo "📋 Summary:"
echo "- API Health: $(curl -s "$API_URL/health" | grep -q "ok" && echo "✅ OK" || echo "❌ FAILED")"
echo "- CORS Headers: $([ -n "$CORS_HEADERS" ] && echo "✅ OK" || echo "❌ MISSING")"
echo "- Login Endpoint: $([ "$HTTP_CODE" = "403" ] || [ "$HTTP_CODE" = "400" ] || [ "$HTTP_CODE" = "200" ] && echo "✅ OK" || echo "❌ FAILED")"
