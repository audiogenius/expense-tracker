#!/bin/bash

echo "=== Ручное исправление конфигурации Nginx ==="

echo "1. Остановка nginx:"
docker-compose stop proxy
echo ""

echo "2. Создание простой конфигурации прямо в контейнере:"
docker-compose run --rm proxy sh -c 'cat > /etc/nginx/conf.d/default.conf << EOF
# Simple nginx configuration for development
server {
    listen 80;
    server_name _;
    
    # API endpoints - proxy to API service
    location /api/ {
        proxy_pass http://api:8080;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
    
    # Frontend - proxy to frontend service
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

echo "3. Проверка синтаксиса:"
docker-compose run --rm proxy nginx -t
echo ""

echo "4. Запуск nginx с новой конфигурацией:"
docker-compose up -d proxy
echo ""

echo "5. Ожидание запуска (3 секунды):"
sleep 3
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

echo "8. Проверка статуса:"
curl -s -o /dev/null -w "API health: %{http_code}\n" http://localhost/api/health
curl -s -o /dev/null -w "API login: %{http_code}\n" http://localhost/api/login

echo ""
echo "=== Исправление завершено ==="
