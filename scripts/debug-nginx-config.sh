#!/bin/bash

echo "=== Диагностика конфигурации nginx ==="

echo "1. Проверка текущей конфигурации nginx:"
docker-compose exec proxy cat /etc/nginx/conf.d/default.conf
echo ""

echo "2. Проверка синтаксиса nginx:"
docker-compose exec proxy nginx -t
echo ""

echo "3. Проверка активной конфигурации:"
docker-compose exec proxy nginx -T | grep -A 30 "location /api/"
echo ""

echo "4. Тест с подробными заголовками:"
curl -v -X POST https://rd-expense-tracker-bot.ru/api/login \
  -H "Content-Type: application/json" \
  -H "User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36" \
  -d '{
    "id": "260144148",
    "username": "testuser",
    "first_name": "Test",
    "last_name": "User"
  }' 2>&1 | head -20
echo ""

echo "5. Проверка логов nginx в реальном времени:"
echo "Откройте сайт в браузере и попробуйте войти, затем нажмите Ctrl+C"
timeout 10 docker-compose logs -f proxy &
LOGS_PID=$!
sleep 10
kill $LOGS_PID 2>/dev/null
echo ""

echo "=== Диагностика завершена ==="
