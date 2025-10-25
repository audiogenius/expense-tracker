package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// ContextKey is an exported type for context keys used by auth package
type ContextKey string

// UserIDKey is the context key where middleware stores internal user id (int64)
const UserIDKey ContextKey = "user_id"

type Auth struct {
	DB        *pgxpool.Pool
	JWTSecret string
	BotToken  string
	Whitelist []string
}

func NewAuth(db *pgxpool.Pool) *Auth {
	wl := strings.Split(os.Getenv("TELEGRAM_WHITELIST"), ",")
	return &Auth{DB: db, JWTSecret: os.Getenv("JWT_SECRET"), BotToken: os.Getenv("TELEGRAM_BOT_TOKEN"), Whitelist: wl}
}

// VerifyTelegramAuth implements Telegram login verification: compute HMAC-SHA256 over
// the data_check_string using secret = SHA256(bot_token) per Telegram recommendations.
func (a *Auth) VerifyTelegramAuth(data map[string]string) bool {
	hash, ok := data["hash"]
	if !ok {
		return false
	}
	var pairs []string
	for k, v := range data {
		if k == "hash" {
			continue
		}
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(pairs)
	dataCheckString := strings.Join(pairs, "\n")
	secret := sha256.Sum256([]byte(a.BotToken))
	mac := hmac.New(sha256.New, secret[:])
	mac.Write([]byte(dataCheckString))
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(hash))
}

// CreateJWT creates a signed JWT (sub = telegram_id)
func (a *Auth) CreateJWT(telegramID int64) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": telegramID, "exp": time.Now().Add(24 * time.Hour).Unix(), "iat": time.Now().Unix()})
	return token.SignedString([]byte(a.JWTSecret))
}

// Middleware validates Bearer JWT and injects internal user id into context
func (a *Auth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authz := r.Header.Get("Authorization")
		if authz == "" || !strings.HasPrefix(authz, "Bearer ") {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		tokenStr := strings.TrimPrefix(authz, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return []byte(a.JWTSecret), nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		sub, ok := claims["sub"].(float64)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		telegramID := int64(sub)
		var internalID int64
		if err := a.DB.QueryRow(r.Context(), "SELECT id FROM users WHERE telegram_id=$1", telegramID).Scan(&internalID); err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), UserIDKey, internalID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserIDFromRequest extracts user ID from request context
func (a *Auth) GetUserIDFromRequest(r *http.Request) (int64, error) {
	userID, ok := r.Context().Value(UserIDKey).(int64)
	if !ok {
		return 0, fmt.Errorf("user ID not found in context")
	}
	return userID, nil
}

// RequestLogger logs basic request info
func (a *Auth) RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info().Str("method", r.Method).Str("path", r.URL.Path).Msg("incoming request")
		next.ServeHTTP(w, r)
	})
}
