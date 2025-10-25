#!/bin/bash

echo "=== ТЕСТ TELEGRAM В ФОРМАТЕ URL-ENCODED ==="

# Получаем токен бота
TELEGRAM_BOT_TOKEN=$(docker-compose exec api printenv TELEGRAM_BOT_TOKEN | tr -d '\r')
echo "Bot Token: $TELEGRAM_BOT_TOKEN"

# Создаем секретный ключ из токена бота
SECRET_KEY=$(echo -n "$TELEGRAM_BOT_TOKEN" | openssl dgst -sha256 -binary)

# Тестовые данные в формате URL-encoded
AUTH_DATA="id=260144148&first_name=Test&last_name=User&username=gmmmpls&photo_url=&auth_date=1732560000"

echo "Auth Data: $AUTH_DATA"
echo "Secret Key (hex): $(echo -n "$SECRET_KEY" | xxd -p)"

# Генерируем правильный HMAC-SHA256 хеш
HASH=$(echo -n "$AUTH_DATA" | openssl dgst -sha256 -hmac "$SECRET_KEY" -binary | xxd -p)

echo "Generated Hash: $HASH"

echo -e "\n=== ТЕСТ 1: JSON ФОРМАТ (текущий) ==="
curl -X POST http://localhost/api/login \
  -H "Content-Type: application/json" \
  -d "{\"id\":\"260144148\",\"first_name\":\"Test\",\"last_name\":\"User\",\"username\":\"gmmmpls\",\"photo_url\":\"\",\"auth_date\":\"1732560000\",\"hash\":\"$HASH\"}" \
  -v

echo -e "\n=== ТЕСТ 2: URL-ENCODED ФОРМАТ (правильный) ==="
curl -X POST http://localhost/api/login \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "id=260144148&first_name=Test&last_name=User&username=gmmmpls&photo_url=&auth_date=1732560000&hash=$HASH" \
  -v

echo -e "\n=== ТЕСТ 3: ПРЯМОЙ TELEGRAM ФОРМАТ ==="
curl -X POST http://localhost/api/login \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "id=260144148&first_name=Test&last_name=User&username=gmmmpls&photo_url=&auth_date=1732560000&hash=$HASH" \
  -v

echo -e "\n=== ТЕСТ ЗАВЕРШЕН ==="
