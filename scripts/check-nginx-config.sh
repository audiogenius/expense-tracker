#!/bin/bash

echo "=== Проверка конфигурации Nginx ==="

echo "1. Проверка текущей конфигурации nginx:"
docker-compose exec proxy cat /etc/nginx/conf.d/default.conf
echo ""

echo "2. Проверка синтаксиса nginx:"
docker-compose exec proxy nginx -t
echo ""

echo "3. Перезагрузка nginx конфигурации:"
docker-compose exec proxy nginx -s reload
echo ""

echo "4. Проверка статуса nginx:"
docker-compose exec proxy nginx -s status 2>/dev/null || echo "Nginx статус недоступен"
echo ""

echo "5. Тест API после перезагрузки:"
curl -s -o /dev/null -w "%{http_code}" http://localhost/api/health
echo " - API health status"

curl -s -o /dev/null -w "%{http_code}" http://localhost/api/login
echo " - API login status"
echo ""

echo "=== Проверка завершена ==="
