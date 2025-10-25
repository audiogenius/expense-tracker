#!/bin/bash

echo "=== Тестирование API Endpoints (без jq) ==="

BASE_URL="http://localhost:8080"
TELEGRAM_ID="260144148"

echo "1. Тест health endpoint:"
HEALTH_RESPONSE=$(curl -s "$BASE_URL/health")
echo "Ответ: $HEALTH_RESPONSE"
if [[ "$HEALTH_RESPONSE" == *"ok"* ]]; then
    echo "✅ Health endpoint работает"
else
    echo "❌ Health endpoint не работает"
fi
echo ""

echo "2. Тест categories endpoint:"
CATEGORIES_RESPONSE=$(curl -s "$BASE_URL/categories")
echo "Ответ: ${CATEGORIES_RESPONSE:0:100}..."
if [[ "$CATEGORIES_RESPONSE" == *"["* ]]; then
    echo "✅ Categories endpoint работает"
else
    echo "❌ Categories endpoint не работает"
fi
echo ""

echo "3. Тест логина через /api/login:"
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/login" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "'"$TELEGRAM_ID"'",
    "username": "testuser",
    "first_name": "Test",
    "last_name": "User"
  }')

echo "Ответ: $LOGIN_RESPONSE"
if [[ "$LOGIN_RESPONSE" == *"token"* ]]; then
    echo "✅ Login через /api/login работает"
    # Извлекаем токен простым способом
    TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    echo "Токен: ${TOKEN:0:30}..."
else
    echo "❌ Login через /api/login не работает"
    TOKEN=""
fi
echo ""

echo "4. Тест логина через /login:"
LOGIN_RESPONSE2=$(curl -s -X POST "$BASE_URL/login" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "'"$TELEGRAM_ID"'",
    "username": "testuser",
    "first_name": "Test",
    "last_name": "User"
  }')

echo "Ответ: $LOGIN_RESPONSE2"
if [[ "$LOGIN_RESPONSE2" == *"token"* ]]; then
    echo "✅ Login через /login работает"
    if [ -z "$TOKEN" ]; then
        TOKEN=$(echo "$LOGIN_RESPONSE2" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
        echo "Токен: ${TOKEN:0:30}..."
    fi
else
    echo "❌ Login через /login не работает"
fi
echo ""

echo "5. Тест защищенного endpoint с токеном:"
if [ -n "$TOKEN" ]; then
    echo "Используем токен: ${TOKEN:0:30}..."
    EXPENSES_RESPONSE=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/expenses")
    echo "Ответ: ${EXPENSES_RESPONSE:0:100}..."
    if [[ "$EXPENSES_RESPONSE" == *"["* ]]; then
        echo "✅ Защищенный endpoint /expenses работает"
    else
        echo "❌ Защищенный endpoint /expenses не работает"
    fi
else
    echo "❌ Нет токена для тестирования защищенного endpoint"
fi
echo ""

echo "6. Проверка статуса контейнеров:"
docker-compose ps | grep -E "(api|proxy|frontend)"
echo ""

echo "7. Проверка логов API сервиса (последние 5 строк):"
docker-compose logs --tail=5 api
echo ""

echo "=== Тестирование завершено ==="
