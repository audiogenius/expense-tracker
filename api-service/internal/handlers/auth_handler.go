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

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	auth *auth.Auth
	db   *pgxpool.Pool
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(auth *auth.Auth, db *pgxpool.Pool) *AuthHandler {
	return &AuthHandler{
		auth: auth,
		db:   db,
	}
}

// NewValidator creates a new validator instance
func NewValidator() *auth.Validator {
	return auth.NewValidator()
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

// Login handles user authentication via Telegram
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Login handler called")

	// Parse request payload
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().Err(err).Msg("failed to decode login request")
		auth.WriteSimpleError(w, http.StatusBadRequest, "Invalid request format")
		return
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

	// Check whitelist
	validator := NewValidator()
	if !validator.IsUserWhitelisted(req.ID) {
		log.Error().
			Str("telegram_id", req.ID).
			Strs("whitelist", validator.GetWhitelist()).
			Msg("user not in whitelist")
		auth.WriteSimpleError(w, http.StatusForbidden, "User not authorized")
		return
	}

	// Verify Telegram authentication if hash is provided
	if req.Hash != "" {
		authData := map[string]string{
			"id":         req.ID,
			"username":   req.Username,
			"first_name": req.FirstName,
			"last_name":  req.LastName,
			"photo_url":  req.PhotoURL,
			"hash":       req.Hash,
		}

		if !h.auth.VerifyTelegramAuth(authData) {
			log.Error().Str("telegram_id", req.ID).Msg("telegram auth verification failed")
			auth.WriteSimpleError(w, http.StatusForbidden, "Invalid Telegram authentication")
			return
		}
	}

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
func (h *AuthHandler) createOrUpdateUser(ctx context.Context, telegramID int64, username string) (int64, error) {
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
