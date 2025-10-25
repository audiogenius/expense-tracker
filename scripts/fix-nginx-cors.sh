#!/bin/bash

echo "=== Исправление nginx с CORS для HTTPS ==="

echo "1. Остановка nginx:"
docker-compose stop proxy
echo ""

echo "2. Создание правильной конфигурации:"
docker-compose exec proxy sh -c 'cat > /etc/nginx/conf.d/default.conf << "EOF"
# HTTP server (for development and API access)
server {
  listen 80;
  server_name rd-expense-tracker-bot.ru www.rd-expense-tracker-bot.ru;

  location /.well-known/acme-challenge/ {
    root /var/www/certbot;
  }

  # API endpoints - serve directly without HTTPS redirect
  location /api/ {
    proxy_pass http://api:8080;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_set_header X-Forwarded-Host $host;
    proxy_set_header X-Forwarded-Port $server_port;
    
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

  # Frontend - redirect to HTTPS
  location / {
    return 301 https://$server_name$request_uri;
  }
}

# HTTPS server
server {
  listen 443 ssl http2;
  server_name rd-expense-tracker-bot.ru www.rd-expense-tracker-bot.ru;

  ssl_certificate /etc/letsencrypt/live/rd-expense-tracker-bot.ru/fullchain.pem;
  ssl_certificate_key /etc/letsencrypt/live/rd-expense-tracker-bot.ru/privkey.pem;

  ssl_protocols TLSv1.2 TLSv1.3;
  ssl_ciphers HIGH:!aNULL:!MD5;
  ssl_prefer_server_ciphers on;

  location / {
    proxy_pass http://frontend:80;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
  }

  # API endpoints для HTTPS - ДОБАВЛЯЕМ CORS!
  location /api/ {
    proxy_pass http://api:8080;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_set_header X-Forwarded-Host $host;
    proxy_set_header X-Forwarded-Port $server_port;
    
    # CORS headers для HTTPS
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
echo "=== Исправление завершено ==="
