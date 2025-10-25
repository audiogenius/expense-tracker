#!/bin/bash

echo "=== Полное пересоздание Nginx контейнера ==="

echo "1. Остановка всех сервисов:"
docker-compose down
echo ""

echo "2. Удаление nginx контейнера и образов:"
docker-compose rm -f proxy
docker rmi expense-tracker-proxy 2>/dev/null || true
echo ""

echo "3. Очистка Docker кэша:"
docker system prune -f
echo ""

echo "4. Создание временной простой конфигурации:"
mkdir -p temp-nginx
cat > temp-nginx/nginx.conf << 'EOF'
events {
    worker_connections 1024;
}

http {
    upstream api {
        server api:8080;
    }
    
    upstream frontend {
        server frontend:80;
    }
    
    server {
        listen 80;
        server_name _;
        
        location /api/ {
            proxy_pass http://api/;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
        
        location / {
            proxy_pass http://frontend/;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
    }
}
EOF
echo ""

echo "5. Запуск всех сервисов кроме nginx:"
docker-compose up -d db api frontend
echo ""

echo "6. Ожидание готовности сервисов (10 секунд):"
sleep 10
echo ""

echo "7. Запуск nginx с новой конфигурацией:"
docker-compose up -d proxy
echo ""

echo "8. Ожидание запуска nginx (5 секунд):"
sleep 5
echo ""

echo "9. Проверка статуса:"
docker-compose ps
echo ""

echo "10. Тест API:"
curl -s http://localhost/api/health
echo ""
echo ""

echo "11. Тест логина:"
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

echo "12. Очистка временных файлов:"
rm -rf temp-nginx
echo ""

echo "=== Пересоздание завершено ==="
