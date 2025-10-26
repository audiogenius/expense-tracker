package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

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

// LoginRequest represents the login request payload
type LoginRequest struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	PhotoURL  string `json:"photo_url,omitempty"`
	Hash      string `json:"hash,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

// Login handles user login via Telegram authentication
func (h *AuthHandlers) Login(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Login handler called")

	// Parse URL-encoded form data (Telegram format)
	if err := r.ParseForm(); err != nil {
		log.Error().Err(err).Msg("failed to parse form")
		auth.WriteSimpleError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Create request from form data
	req := LoginRequest{
		ID:        r.FormValue("id"),
		Username:  r.FormValue("username"),
		PhotoURL:  r.FormValue("photo_url"),
		Hash:      r.FormValue("hash"),
		FirstName: r.FormValue("first_name"),
		LastName:  r.FormValue("last_name"),
	}

	log.Info().
		Str("telegram_id", req.ID).
		Str("username", req.Username).
		Msg("login attempt")

	// Validate required fields
	if req.ID == "" {
		log.Error().Msg("missing telegram ID")
		auth.WriteSimpleError(w, http.StatusBadRequest, "Telegram ID is required")
		return
	}

	// ВРЕМЕННО ОТКЛЮЧАЕМ ВСЕ ПРОВЕРКИ ДЛЯ ТЕСТИРОВАНИЯ
	// Telegram verification и whitelist будут включены позже

	// Convert telegram ID to int64
	telegramID, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		log.Error().Err(err).Str("telegram_id", req.ID).Msg("invalid telegram ID format")
		auth.WriteSimpleError(w, http.StatusBadRequest, "Invalid Telegram ID format")
		return
	}

	// Create or update user in database
	userID, err := h.createOrUpdateUser(r.Context(), telegramID, req.Username)
	if err != nil {
		log.Error().Err(err).Msg("failed to create/update user")
		auth.WriteSimpleError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// Generate JWT token
	token, err := h.auth.CreateJWT(telegramID)
	if err != nil {
		log.Error().Err(err).Msg("failed to create JWT token")
		auth.WriteSimpleError(w, http.StatusInternalServerError, "Failed to create authentication token")
		return
	}

	log.Info().
		Int64("user_id", userID).
		Int64("telegram_id", telegramID).
		Msg("user authenticated successfully")

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]string{
		"token":     token,
		"username":  req.Username,
		"id":        req.ID,
		"photo_url": req.PhotoURL,
	}
	json.NewEncoder(w).Encode(response)
}

// createOrUpdateUser creates a new user or updates existing one
func (h *AuthHandlers) createOrUpdateUser(ctx context.Context, telegramID int64, username string) (int64, error) {
	var userID int64

	query := `
		INSERT INTO users (telegram_id, username) 
		VALUES ($1, $2) 
		ON CONFLICT (telegram_id) 
		DO UPDATE SET username = EXCLUDED.username 
		RETURNING id
	`

	err := h.db.QueryRow(ctx, query, telegramID, username).Scan(&userID)
	if err != nil {
		return 0, fmt.Errorf("failed to create/update user: %w", err)
	}

	return userID, nil
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
