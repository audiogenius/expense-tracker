#!/bin/bash

echo "=== ДИАГНОСТИКА 403 ОШИБКИ ==="

echo "1. Статус контейнеров:"
docker-compose ps

echo -e "\n2. Логи nginx (последние 10 строк):"
docker-compose logs proxy | tail -10

echo -e "\n3. Логи API (последние 10 строк):"
docker-compose logs api | tail -10

echo -e "\n4. Проверка конфигурации nginx:"
docker-compose exec proxy nginx -t

echo -e "\n5. Тест API напрямую:"
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"id":"260144148","first_name":"Test","last_name":"User","username":"testuser","photo_url":"","auth_date":"1732560000","hash":"test_hash"}' \
  -v

echo -e "\n6. Тест через nginx HTTP:"
curl -X POST http://localhost/api/login \
  -H "Content-Type: application/json" \
  -d '{"id":"260144148","first_name":"Test","last_name":"User","username":"testuser","photo_url":"","auth_date":"1732560000","hash":"test_hash"}' \
  -v

echo -e "\n7. Тест через nginx HTTPS:"
curl -X POST https://rd-expense-tracker-bot.ru/api/login \
  -H "Content-Type: application/json" \
  -d '{"id":"260144148","first_name":"Test","last_name":"User","username":"testuser","photo_url":"","auth_date":"1732560000","hash":"test_hash"}' \
  -v

echo -e "\n8. Проверка CORS заголовков:"
curl -X OPTIONS https://rd-expense-tracker-bot.ru/api/login \
  -H "Origin: https://rd-expense-tracker-bot.ru" \
  -H "Access-Control-Request-Method: POST" \
  -H "Access-Control-Request-Headers: Content-Type" \
  -v

echo -e "\n=== ДИАГНОСТИКА ЗАВЕРШЕНА ==="
