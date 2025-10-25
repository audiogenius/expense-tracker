#!/bin/bash

echo "=== Проверка конфигурации nginx ==="

echo "1. Текущая конфигурация nginx:"
docker-compose exec proxy cat /etc/nginx/conf.d/default.conf
echo ""

echo "2. Проверка синтаксиса nginx:"
docker-compose exec proxy nginx -t
echo ""

echo "3. Проверка активных процессов nginx:"
docker-compose exec proxy ps aux | grep nginx
echo ""

echo "4. Проверка портов nginx:"
docker-compose exec proxy netstat -tlnp | grep nginx
echo ""

echo "5. Тест API напрямую:"
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

echo "6. Тест API через nginx:"
curl -s -X POST http://localhost/api/login \
  -H "Content-Type: application/json" \
  -d '{
    "id": "260144148",
    "username": "testuser",
    "first_name": "Test",
    "last_name": "User"
  }'
echo ""
echo ""

echo "7. Тест API через HTTPS:"
curl -k -s -X POST https://localhost/api/login \
  -H "Content-Type: application/json" \
  -d '{
    "id": "260144148",
    "username": "testuser",
    "first_name": "Test",
    "last_name": "User"
  }'
echo ""
echo ""

echo "=== Проверка завершена ==="