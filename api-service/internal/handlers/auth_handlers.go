package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/expense-tracker/api-service/internal/auth"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// AuthHandlers handles all authentication-related endpoints
type AuthHandlers struct {
	auth *auth.Auth
	db   *pgxpool.Pool
}

// NewAuthHandlers creates a new AuthHandlers instance
func NewAuthHandlers(a *auth.Auth, db *pgxpool.Pool) *AuthHandlers {
	return &AuthHandlers{
		auth: a,
		db:   db,
	}
}

// Login handles user login via Telegram authentication
func (h *AuthHandlers) Login(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Login handler called")

	// Create a temporary AuthHandler for login
	authHandler := NewAuthHandler(h.auth, h.db)
	authHandler.Login(w, r)
}

// Logout handles user logout
func (h *AuthHandlers) Logout(w http.ResponseWriter, r *http.Request) {
	// For JWT-based auth, logout is handled client-side by removing the token
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "logged out"})
	log.Info().Msg("User logged out")
}

// RefreshToken handles token refresh
func (h *AuthHandlers) RefreshToken(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from existing token
	userID, err := h.auth.GetUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Generate new token
	token, err := h.auth.CreateJWT(userID)
	if err != nil {
		log.Error().Err(err).Msg("failed to generate refresh token")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"token": token,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	log.Info().Int64("user_id", userID).Msg("token refreshed")
}

// GetProfile returns user profile information
func (h *AuthHandlers) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := h.auth.GetUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user info from database
	var username string
	var telegramID int64
	err = h.db.QueryRow(r.Context(),
		"SELECT username, telegram_id FROM users WHERE id = $1", userID).Scan(&username, &telegramID)
	if err != nil {
		log.Error().Err(err).Int64("user_id", userID).Msg("failed to get user profile")
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	profile := map[string]interface{}{
		"id":          userID,
		"telegram_id": telegramID,
		"username":    username,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
	log.Info().Int64("user_id", userID).Str("username", username).Msg("profile returned")
}
