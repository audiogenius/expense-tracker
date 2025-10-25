#!/bin/bash

echo "=== Полная очистка и пересоздание Nginx ==="

echo "1. Полная остановка и удаление nginx контейнера:"
docker-compose stop proxy
docker-compose rm -f proxy
echo ""

echo "2. Удаление всех nginx образов и кэша:"
docker system prune -f
echo ""

echo "3. Создание новой простой конфигурации:"
cat > nginx-simple.conf << 'EOF'
server {
    listen 80;
    server_name _;
    
    # API endpoints - proxy to API service
    location /api/ {
        proxy_pass http://api:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    
    # Frontend - proxy to frontend service
    location / {
        proxy_pass http://frontend:80;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
EOF
echo ""

echo "4. Копирование простой конфигурации в контейнер:"
docker-compose run --rm proxy sh -c 'cat > /etc/nginx/conf.d/default.conf << EOF
server {
    listen 80;
    server_name _;
    
    location /api/ {
        proxy_pass http://api:8080;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
    
    location / {
        proxy_pass http://frontend:80;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
}
EOF'
echo ""

echo "5. Проверка синтаксиса новой конфигурации:"
docker-compose run --rm proxy nginx -t
echo ""

echo "6. Запуск nginx с чистой конфигурацией:"
docker-compose up -d proxy
echo ""

echo "7. Ожидание запуска (5 секунд):"
sleep 5
echo ""

echo "8. Проверка статуса контейнеров:"
docker-compose ps proxy
echo ""

echo "9. Тест API:"
curl -s http://localhost/api/health
echo ""
echo ""

echo "10. Тест логина:"
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

echo "11. Финальная проверка:"
curl -s -o /dev/null -w "API health: %{http_code}\n" http://localhost/api/health
curl -s -o /dev/null -w "API login: %{http_code}\n" http://localhost/api/login
curl -s -o /dev/null -w "Frontend: %{http_code}\n" http://localhost/

echo ""
echo "=== Очистка завершена ==="
