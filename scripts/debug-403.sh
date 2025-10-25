#!/bin/bash

echo "=== Диагностика 403 ошибки ==="

echo "1. Проверка текущей конфигурации nginx:"
docker-compose exec proxy cat /etc/nginx/conf.d/default.conf
echo ""

echo "2. Проверка логов nginx в реальном времени:"
echo "Откройте сайт в браузере и попробуйте войти, затем нажмите Ctrl+C"
docker-compose logs -f proxy &
LOGS_PID=$!
sleep 10
kill $LOGS_PID 2>/dev/null
echo ""

echo "3. Тест с подробными заголовками:"
curl -v -X POST http://localhost/api/login \
  -H "Content-Type: application/json" \
  -H "User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36" \
  -d '{
    "id": "260144148",
    "username": "testuser",
    "first_name": "Test",
    "last_name": "User"
  }'
echo ""
echo ""

echo "4. Проверка доступности API напрямую:"
curl -s -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{
    "id": "260144148",
    "username": "testuser",
    "first_name": "Test",
    "last_name": "User"
  }'
echo ""
echo ""

echo "5. Проверка nginx конфигурации:"
docker-compose exec proxy nginx -T
echo ""

echo "=== Диагностика завершена ==="
