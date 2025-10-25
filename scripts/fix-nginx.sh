#!/bin/bash

echo "=== Исправление конфигурации Nginx ==="

echo "1. Остановка nginx:"
docker-compose stop proxy
echo ""

echo "2. Копирование простой конфигурации:"
docker-compose run --rm proxy cp /etc/nginx/conf.d/default.conf /etc/nginx/conf.d/default.conf.backup
docker-compose run --rm proxy cp /etc/nginx/conf.d/nginx-simple.conf /etc/nginx/conf.d/default.conf
echo ""

echo "3. Проверка синтаксиса:"
docker-compose run --rm proxy nginx -t
echo ""

echo "4. Запуск nginx с новой конфигурацией:"
docker-compose up -d proxy
echo ""

echo "5. Ожидание запуска (5 секунд):"
sleep 5
echo ""

echo "6. Тест API:"
curl -s http://localhost/api/health
echo ""
echo ""

echo "7. Тест логина:"
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

echo "=== Исправление завершено ==="
