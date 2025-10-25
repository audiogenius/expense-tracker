#!/bin/bash

echo "=== ИСПРАВЛЕНИЕ 403 ОШИБКИ В NGINX ==="

echo "1. Остановка nginx:"
docker-compose stop proxy

echo -e "\n2. Создание правильной конфигурации nginx:"
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

    # CORS headers for API
    add_header Access-Control-Allow-Origin * always;
    add_header Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS" always;
    add_header Access-Control-Allow-Headers "Content-Type, Authorization, X-Requested-With" always;
    add_header Access-Control-Allow-Credentials true always;

    # Handle preflight requests
    if ($request_method = OPTIONS) {
      add_header Access-Control-Allow-Origin * always;
      add_header Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS" always;
      add_header Access-Control-Allow-Headers "Content-Type, Authorization, X-Requested-With" always;
      add_header Access-Control-Allow-Credentials true always;
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
    add_header Access-Control-Allow-Origin "https://rd-expense-tracker-bot.ru" always;
    add_header Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS" always;
    add_header Access-Control-Allow-Headers "Content-Type, Authorization, X-Requested-With" always;
    add_header Access-Control-Allow-Credentials true always;

    # Handle preflight requests
    if ($request_method = OPTIONS) {
      add_header Access-Control-Allow-Origin "https://rd-expense-tracker-bot.ru" always;
      add_header Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS" always;
      add_header Access-Control-Allow-Headers "Content-Type, Authorization, X-Requested-With" always;
      add_header Access-Control-Allow-Credentials true always;
      add_header Content-Length 0;
      add_header Content-Type text/plain;
      return 200;
    }
  }
}
EOF'

echo -e "\n3. Проверка синтаксиса:"
docker-compose exec proxy nginx -t

echo -e "\n4. Запуск nginx:"
docker-compose start proxy

echo -e "\n5. Ожидание запуска (3 секунды):"
sleep 3

echo -e "\n6. Тест API:"
curl -X POST http://localhost/api/login \
  -H "Content-Type: application/json" \
  -d '{"id":"260144148","first_name":"Test","last_name":"User","username":"testuser","photo_url":"","auth_date":"1732560000","hash":"test_hash"}' \
  -v

echo -e "\n=== ИСПРАВЛЕНИЕ ЗАВЕРШЕНО ==="
