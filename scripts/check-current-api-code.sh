#!/bin/bash

echo "=== ПРОВЕРКА ТЕКУЩЕГО КОДА API ==="

echo "1. Проверяем текущий auth_handler.go:"
docker-compose exec api cat /app/internal/handlers/auth_handler.go | head -20

echo -e "\n2. Проверяем логи API:"
docker-compose logs api | tail -10

echo -e "\n3. Проверяем статус API:"
docker-compose ps api

echo -e "\n4. Тест прямого запроса к API:"
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "id=260144148&first_name=Test&last_name=User&username=gmmmpls&photo_url=&auth_date=1732560000&hash=test_hash" \
  -v

echo -e "\n=== ПРОВЕРКА ЗАВЕРШЕНА ==="
