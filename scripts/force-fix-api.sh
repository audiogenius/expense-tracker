#!/bin/bash

echo "=== ПРИНУДИТЕЛЬНОЕ ИСПРАВЛЕНИЕ API ==="

echo "1. Полная остановка всех сервисов:"
docker-compose down

echo -e "\n2. Удаление старого образа API:"
docker rmi expense-tracker-api || true

echo -e "\n3. Создание правильного auth_handler.go:"
mkdir -p temp_fix
cat > temp_fix/auth_handler.go << 'EOF'
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/expense-tracker/api-service/internal/auth"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	auth *auth.Auth
	db   *pgxpool.Pool
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(a *auth.Auth, db *pgxpool.Pool) *AuthHandler {
	return &AuthHandler{auth: a, db: db}
}

// Login handles Telegram authentication
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Login handler called")

	// Парсим данные из URL-encoded формата
	if err := r.ParseForm(); err != nil {
		log.Error().Err(err).Msg("parse form")
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// Создаем map из form данных
	authData := make(map[string]string)
	for key, values := range r.Form {
		if len(values) > 0 {
			authData[key] = values[0]
		}
	}

	log.Info().
		Str("telegram_id", authData["id"]).
		Str("username", authData["username"]).
		Msg("login attempt")

	// Проверяем Telegram аутентификацию
	if !auth.VerifyTelegramAuth(authData) {
		log.Error().Str("telegram_id", authData["id"]).Msg("telegram auth verification failed")
		http.Error(w, "Invalid Telegram authentication", http.StatusForbidden)
		return
	}

	// Получаем telegram_id
	telegramID, err := strconv.ParseInt(authData["id"], 10, 64)
	if err != nil {
		log.Error().Err(err).Msg("parse telegram_id")
		http.Error(w, "invalid telegram_id", http.StatusBadRequest)
		return
	}

	// Проверяем whitelist
	if !auth.IsUserInWhitelist(telegramID, h.auth.Whitelist) {
		log.Error().Int64("telegram_id", telegramID).Msg("user not in whitelist")
		http.Error(w, "user not authorized", http.StatusForbidden)
		return
	}

	// Создаем или обновляем пользователя
	userID, err := h.createOrUpdateUser(telegramID, authData["username"])
	if err != nil {
		log.Error().Err(err).Msg("create or update user")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Генерируем JWT токен
	token, err := h.auth.CreateJWT(userID)
	if err != nil {
		log.Error().Err(err).Msg("create JWT")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Возвращаем успешный ответ
	response := map[string]interface{}{
		"id":        authData["id"],
		"username":  authData["username"],
		"photo_url": authData["photo_url"],
		"token":     token,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	log.Info().
		Int64("user_id", userID).
		Str("telegram_id", authData["id"]).
		Str("username", authData["username"]).
		Msg("login successful")
}

// createOrUpdateUser создает или обновляет пользователя в базе данных
func (h *AuthHandler) createOrUpdateUser(telegramID int64, username string) (int64, error) {
	var userID int64
	
	// Пытаемся найти существующего пользователя
	err := h.db.QueryRow(context.Background(),
		"SELECT id FROM users WHERE telegram_id = $1", telegramID).Scan(&userID)
	
	if err != nil {
		// Пользователь не найден, создаем нового
		err = h.db.QueryRow(context.Background(),
			"INSERT INTO users (telegram_id, username, created_at) VALUES ($1, $2, $3) RETURNING id",
			telegramID, username, time.Now()).Scan(&userID)
		
		if err != nil {
			return 0, fmt.Errorf("failed to create user: %w", err)
		}
		
		log.Info().Int64("user_id", userID).Int64("telegram_id", telegramID).Msg("user created")
	} else {
		// Пользователь найден, обновляем username если нужно
		_, err = h.db.Exec(context.Background(),
			"UPDATE users SET username = $1, updated_at = $2 WHERE id = $3",
			username, time.Now(), userID)
		
		if err != nil {
			log.Error().Err(err).Msg("failed to update user")
		}
		
		log.Info().Int64("user_id", userID).Int64("telegram_id", telegramID).Msg("user updated")
	}
	
	return userID, nil
}
EOF

echo -e "\n4. Копируем исправленный файл в контейнер:"
docker cp temp_fix/auth_handler.go expense_api:/app/internal/handlers/auth_handler.go

echo -e "\n5. Перезапуск API:"
docker-compose up -d api

echo -e "\n6. Ожидание запуска (5 секунд):"
sleep 5

echo -e "\n7. Тест исправленного API:"
curl -X POST http://localhost/api/login \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "id=260144148&first_name=Test&last_name=User&username=gmmmpls&photo_url=&auth_date=1732560000&hash=test_hash" \
  -v

echo -e "\n8. Очистка:"
rm -rf temp_fix

echo -e "\n=== ИСПРАВЛЕНИЕ ЗАВЕРШЕНО ==="
