#!/bin/bash

echo "=== ОТКЛЮЧЕНИЕ ПРОВЕРКИ TELEGRAM (ВРЕМЕННО) ==="

echo "1. Остановка API:"
docker-compose stop api

echo -e "\n2. Создание временного auth.go без проверки:"
docker-compose exec api sh -c 'cat > /app/internal/auth/validator.go << "EOF"
package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"time"
)

// VerifyTelegramAuth проверяет подпись Telegram (ОТКЛЮЧЕНО ДЛЯ ТЕСТИРОВАНИЯ)
func VerifyTelegramAuth(authData map[string]string) bool {
	// ВРЕМЕННО ОТКЛЮЧАЕМ ПРОВЕРКУ ДЛЯ ТЕСТИРОВАНИЯ
	return true
}

// VerifyTelegramAuthWithTime проверяет подпись и время (ОТКЛЮЧЕНО ДЛЯ ТЕСТИРОВАНИЯ)
func VerifyTelegramAuthWithTime(authData map[string]string) bool {
	// ВРЕМЕННО ОТКЛЮЧАЕМ ПРОВЕРКУ ДЛЯ ТЕСТИРОВАНИЯ
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

echo -e "\n5. Тест авторизации:"
curl -X POST http://localhost/api/login \
  -H "Content-Type: application/json" \
  -d '{"id":"260144148","first_name":"Test","last_name":"User","username":"gmmmpls","photo_url":"","auth_date":"1732560000","hash":"test_hash"}' \
  -v

echo -e "\n6. Проверка статуса:"
docker-compose ps api

echo -e "\n=== ОТКЛЮЧЕНИЕ ЗАВЕРШЕНО ==="
