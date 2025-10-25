#!/bin/bash

echo "=== Отключение HTTPS перенаправлений ==="

echo "1. Остановка nginx:"
docker-compose stop proxy
echo ""

echo "2. Создание резервной копии:"
docker-compose run --rm proxy cp /etc/nginx/conf.d/default.conf /etc/nginx/conf.d/default.conf.backup
echo ""

echo "3. Отключение HTTPS перенаправлений:"
docker-compose run --rm proxy sh -c 'sed -i "s/return 301 https/#return 301 https/" /etc/nginx/conf.d/default.conf'
echo ""

echo "4. Добавление API блока в HTTP сервер:"
docker-compose run --rm proxy sh -c 'cat >> /etc/nginx/conf.d/default.conf << EOF

  # API endpoints - serve directly without HTTPS redirect
  location /api/ {
    proxy_pass http://api:8080;
    proxy_set_header Host \$host;
    proxy_set_header X-Real-IP \$remote_addr;
    proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto \$scheme;
  }
EOF'
echo ""

echo "5. Проверка синтаксиса:"
docker-compose run --rm proxy nginx -t
echo ""

echo "6. Запуск nginx:"
docker-compose up -d proxy
echo ""

echo "7. Ожидание запуска (3 секунды):"
sleep 3
echo ""

echo "8. Тест API:"
curl -s http://localhost/api/health
echo ""
echo ""

echo "9. Тест логина:"
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

echo "=== Отключение завершено ==="
