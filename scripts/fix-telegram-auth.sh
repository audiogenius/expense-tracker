#!/bin/bash

echo "=== ИСПРАВЛЕНИЕ TELEGRAM АУТЕНТИФИКАЦИИ ==="

echo "1. Проверяем переменные окружения:"
docker-compose exec api env | grep -E "(TELEGRAM_BOT_TOKEN|JWT_SECRET)"

echo -e "\n2. Проверяем время на сервере:"
date

echo -e "\n3. Проверяем время в API контейнере:"
docker-compose exec api date

echo -e "\n4. Тестируем с правильными данными Telegram:"
# Создаем тестовые данные с правильной подписью
TELEGRAM_BOT_TOKEN=$(docker-compose exec api printenv TELEGRAM_BOT_TOKEN | tr -d '\r')
echo "Bot Token: $TELEGRAM_BOT_TOKEN"

# Генерируем правильную подпись для теста
AUTH_DATA="id=260144148&first_name=Test&last_name=User&username=gmmmpls&photo_url=&auth_date=1732560000"
SECRET_KEY=$(echo -n "$TELEGRAM_BOT_TOKEN" | openssl dgst -sha256 -binary | openssl base64)
HASH=$(echo -n "$AUTH_DATA" | openssl dgst -sha256 -hmac "$SECRET_KEY" | cut -d' ' -f2)

echo "Auth Data: $AUTH_DATA"
echo "Generated Hash: $HASH"

echo -e "\n5. Тест с правильной подписью:"
curl -X POST http://localhost/api/login \
  -H "Content-Type: application/json" \
  -d "{\"id\":\"260144148\",\"first_name\":\"Test\",\"last_name\":\"User\",\"username\":\"gmmmpls\",\"photo_url\":\"\",\"auth_date\":\"1732560000\",\"hash\":\"$HASH\"}" \
  -v

echo -e "\n6. Проверяем пользователя в базе:"
docker-compose exec db psql -U expense_user -d expense_tracker -c "SELECT * FROM users WHERE telegram_id = 260144148;"

echo -e "\n7. Добавляем пользователя если его нет:"
docker-compose exec db psql -U expense_user -d expense_tracker -c "
INSERT INTO users (telegram_id, username) 
VALUES (260144148, 'gmmmpls') 
ON CONFLICT (telegram_id) DO UPDATE SET username = EXCLUDED.username;
"

echo -e "\n=== ИСПРАВЛЕНИЕ ЗАВЕРШЕНО ==="
