#!/bin/bash

echo "=== Простое отключение HTTPS ==="

echo "1. Остановка nginx:"
docker-compose stop proxy
echo ""

echo "2. Создание простой HTTP-only конфигурации:"
docker-compose exec proxy sh -c 'cat > /etc/nginx/conf.d/default.conf << "EOF"
server {
    listen 80;
    server_name _;
    
    # API endpoints
    location /api/ {
        proxy_pass http://api:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # CORS headers
        add_header Access-Control-Allow-Origin *;
        add_header Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS";
        add_header Access-Control-Allow-Headers "Content-Type, Authorization, X-Requested-With";
        
        # Обработка preflight запросов
        if ($request_method = OPTIONS) {
            add_header Access-Control-Allow-Origin *;
            add_header Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS";
            add_header Access-Control-Allow-Headers "Content-Type, Authorization, X-Requested-With";
            add_header Content-Length 0;
            add_header Content-Type text/plain;
            return 200;
        }
    }
    
    # Frontend
    location / {
        proxy_pass http://frontend:80;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
EOF'
echo ""

echo "3. Проверка синтаксиса:"
docker-compose exec proxy nginx -t
echo ""

echo "4. Запуск nginx:"
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
curl -s -o /dev/null -w "Frontend: %{http_code}\n" http://localhost/

echo ""
echo "=== Отключение HTTPS завершено ==="
echo "Теперь сайт доступен только по HTTP: http://rd-expense-tracker-bot.ru"
