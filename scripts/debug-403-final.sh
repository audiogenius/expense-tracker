#!/bin/bash

echo "=== Финальная диагностика 403 ошибки ==="

echo "1. Проверка текущих логов nginx:"
docker-compose logs proxy | tail -10
echo ""

echo "2. Мониторинг логов в реальном времени:"
echo "Откройте сайт в браузере и попробуйте войти, затем нажмите Ctrl+C"
timeout 15 docker-compose logs -f proxy &
LOGS_PID=$!
sleep 15
kill $LOGS_PID 2>/dev/null
echo ""

echo "3. Тест с HTTPS (как браузер):"
curl -k -s -X POST https://rd-expense-tracker-bot.ru/api/login \
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

echo "4. Тест с HTTP (как curl):"
curl -s -X POST http://rd-expense-tracker-bot.ru/api/login \
  -H "Content-Type: application/json" \
  -d '{
    "id": "260144148",
    "username": "testuser",
    "first_name": "Test",
    "last_name": "User"
  }'
echo ""
echo ""

echo "5. Проверка конфигурации nginx:"
docker-compose exec proxy cat /etc/nginx/conf.d/default.conf | grep -A 20 "location /api/"
echo ""

echo "6. Проверка SSL сертификатов:"
docker-compose exec proxy ls -la /etc/letsencrypt/live/rd-expense-tracker-bot.ru/ 2>/dev/null || echo "SSL сертификаты не найдены"
echo ""

echo "=== Диагностика завершена ==="
