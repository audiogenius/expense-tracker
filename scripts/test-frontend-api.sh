#!/bin/bash

echo "=== Тестирование Frontend -> API подключения ==="

echo "1. Проверка доступности фронтенда:"
FRONTEND_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:3000)
echo "Фронтенд статус: $FRONTEND_RESPONSE"
if [ "$FRONTEND_RESPONSE" = "200" ]; then
    echo "✅ Фронтенд доступен"
else
    echo "❌ Фронтенд недоступен"
fi
echo ""

echo "2. Проверка API через nginx (как видит фронтенд):"
API_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost/api/health)
echo "API через nginx статус: $API_RESPONSE"
if [ "$API_RESPONSE" = "200" ]; then
    echo "✅ API через nginx доступен"
else
    echo "❌ API через nginx недоступен"
fi
echo ""

echo "3. Проверка API напрямую:"
DIRECT_API_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health)
echo "API напрямую статус: $DIRECT_API_RESPONSE"
if [ "$DIRECT_API_RESPONSE" = "200" ]; then
    echo "✅ API напрямую доступен"
else
    echo "❌ API напрямую недоступен"
fi
echo ""

echo "4. Тест логина через nginx (как фронтенд):"
LOGIN_RESPONSE=$(curl -s -X POST http://localhost/api/login \
  -H "Content-Type: application/json" \
  -d '{
    "id": "260144148",
    "username": "testuser",
    "first_name": "Test",
    "last_name": "User"
  }')

echo "Ответ логина через nginx: $LOGIN_RESPONSE"
if [[ "$LOGIN_RESPONSE" == *"token"* ]]; then
    echo "✅ Логин через nginx работает"
else
    echo "❌ Логин через nginx не работает"
fi
echo ""

echo "5. Проверка nginx конфигурации:"
docker-compose exec proxy nginx -t
echo ""

echo "6. Проверка логов nginx:"
docker-compose logs --tail=10 proxy
echo ""

echo "=== Тестирование завершено ==="
