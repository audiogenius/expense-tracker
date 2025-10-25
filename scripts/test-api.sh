#!/bin/bash

echo "=== Тестирование API Endpoints ==="

BASE_URL="http://localhost:8080"
TELEGRAM_ID="260144148"

echo "1. Тест health endpoint:"
curl -s "$BASE_URL/health" | jq . || echo "Health endpoint недоступен"
echo ""

echo "2. Тест categories endpoint:"
curl -s "$BASE_URL/categories" | jq '.[0:2]' || echo "Categories endpoint недоступен"
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

echo "$LOGIN_RESPONSE" | jq . || echo "Login через /api/login не работает"
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

echo "$LOGIN_RESPONSE2" | jq . || echo "Login через /login не работает"
echo ""

echo "5. Извлечение токена и тест защищенного endpoint:"
TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.token' 2>/dev/null)
if [ "$TOKEN" != "null" ] && [ -n "$TOKEN" ]; then
    echo "Токен получен: ${TOKEN:0:30}..."
    echo "Тест защищенного endpoint /expenses:"
    curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/expenses" | jq '.[0:2]' || echo "Защищенный endpoint не работает"
else
    echo "Не удалось получить токен"
fi
echo ""

echo "6. Проверка nginx конфигурации:"
docker-compose exec proxy nginx -t 2>/dev/null || echo "Nginx конфигурация неверна"
echo ""

echo "=== Тестирование завершено ==="
