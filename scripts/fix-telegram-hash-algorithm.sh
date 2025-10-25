#!/bin/bash

echo "=== ИСПРАВЛЕНИЕ АЛГОРИТМА TELEGRAM ХЕШИРОВАНИЯ ==="

echo "1. Остановка API:"
docker-compose stop api

echo -e "\n2. Создание правильного validator.go:"
docker-compose exec api sh -c 'cat > /app/internal/auth/validator.go << "EOF"
package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// VerifyTelegramAuth проверяет подпись Telegram
func VerifyTelegramAuth(authData map[string]string) bool {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		return false
	}

	// Получаем хеш из данных
	hash, exists := authData["hash"]
	if !exists {
		return false
	}

	// Создаем секретный ключ из токена бота
	secretKey := sha256.Sum256([]byte(botToken))

	// Сортируем ключи и создаем строку для проверки
	var keys []string
	for k := range authData {
		if k != "hash" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	// Создаем строку для проверки
	var dataCheckArr []string
	for _, k := range keys {
		dataCheckArr = append(dataCheckArr, fmt.Sprintf("%s=%s", k, authData[k]))
	}
	dataCheckString := strings.Join(dataCheckArr, "\n")

	// Вычисляем HMAC-SHA256
	mac := hmac.New(sha256.New, secretKey[:])
	mac.Write([]byte(dataCheckString))
	expectedHash := hex.EncodeToString(mac.Sum(nil))

	// Сравниваем хеши
	return hmac.Equal([]byte(hash), []byte(expectedHash))
}

// VerifyTelegramAuthWithTime проверяет подпись и время
func VerifyTelegramAuthWithTime(authData map[string]string) bool {
	// Сначала проверяем подпись
	if !VerifyTelegramAuth(authData) {
		return false
	}

	// Проверяем время (не старше 24 часов)
	authDateStr, exists := authData["auth_date"]
	if !exists {
		return false
	}

	authDate, err := strconv.ParseInt(authDateStr, 10, 64)
	if err != nil {
		return false
	}

	// Проверяем, что время не старше 24 часов
	now := time.Now().Unix()
	if now-authDate > 86400 { // 24 часа в секундах
		return false
	}

	return true
}

// IsUserInWhitelist проверяет, есть ли пользователь в whitelist
func IsUserInWhitelist(telegramID int64, whitelist []string) bool {
	if len(whitelist) == 0 {
		return true // Если whitelist пустой, разрешаем всем
	}
	
	telegramIDStr := strconv.FormatInt(telegramID, 10)
	for _, allowedID := range whitelist {
		if allowedID == "*" || allowedID == telegramIDStr {
			return true
		}
	}
	return false
}
EOF'

echo -e "\n3. Пересборка и запуск API:"
docker-compose up --build -d api

echo -e "\n4. Ожидание запуска (5 секунд):"
sleep 5

echo -e "\n5. Тест с правильным алгоритмом:"
# Генерируем правильный хеш
TELEGRAM_BOT_TOKEN=$(docker-compose exec api printenv TELEGRAM_BOT_TOKEN | tr -d '\r')
SECRET_KEY=$(echo -n "$TELEGRAM_BOT_TOKEN" | openssl dgst -sha256 -binary)
AUTH_DATA="id=260144148&first_name=Test&last_name=User&username=gmmmpls&photo_url=&auth_date=1732560000"
HASH=$(echo -n "$AUTH_DATA" | openssl dgst -sha256 -hmac "$SECRET_KEY" -binary | xxd -p)

echo "Generated Hash: $HASH"

curl -X POST http://localhost/api/login \
  -H "Content-Type: application/json" \
  -d "{\"id\":\"260144148\",\"first_name\":\"Test\",\"last_name\":\"User\",\"username\":\"gmmmpls\",\"photo_url\":\"\",\"auth_date\":\"1732560000\",\"hash\":\"$HASH\"}" \
  -v

echo -e "\n=== ИСПРАВЛЕНИЕ ЗАВЕРШЕНО ==="
