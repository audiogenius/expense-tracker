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

// Validator handles authentication validation logic
type Validator struct {
	whitelist []string
	botToken  string
	jwtSecret string
}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	whitelist := strings.Split(os.Getenv("TELEGRAM_WHITELIST"), ",")
	// Clean up whitelist entries
	var cleanWhitelist []string
	for _, entry := range whitelist {
		entry = strings.TrimSpace(entry)
		if entry != "" {
			cleanWhitelist = append(cleanWhitelist, entry)
		}
	}

	return &Validator{
		whitelist: cleanWhitelist,
		botToken:  os.Getenv("TELEGRAM_BOT_TOKEN"),
		jwtSecret: os.Getenv("JWT_SECRET"),
	}
}

// ValidateConfiguration checks if all required environment variables are set
func (v *Validator) ValidateConfiguration() error {
	if v.jwtSecret == "" {
		return ErrMissingJWTSecret
	}
	if v.botToken == "" {
		return ErrMissingBotToken
	}
	return nil
}

// IsUserWhitelisted checks if user is in whitelist
func (v *Validator) IsUserWhitelisted(telegramID string) bool {
	// Allow all users if whitelist contains "*"
	for _, allowedID := range v.whitelist {
		if allowedID == "*" || allowedID == telegramID {
			return true
		}
	}
	return false
}

// GetWhitelist returns current whitelist for debugging
func (v *Validator) GetWhitelist() []string {
	return v.whitelist
}

// VerifyTelegramAuth verifies Telegram authentication data
func (v *Validator) VerifyTelegramAuth(authData map[string]string) bool {
	// Check if we have required fields
	hash, exists := authData["hash"]
	if !exists || hash == "" {
		return false
	}

	// Remove hash from data for verification
	dataToVerify := make(map[string]string)
	for k, v := range authData {
		if k != "hash" {
			dataToVerify[k] = v
		}
	}

	// Create data string for verification
	var dataCheckArr []string
	for k, v := range dataToVerify {
		if v != "" {
			dataCheckArr = append(dataCheckArr, fmt.Sprintf("%s=%s", k, v))
		}
	}
	sort.Strings(dataCheckArr)
	dataCheckString := strings.Join(dataCheckArr, "\n")

	// Create secret key
	secretKey := sha256.Sum256([]byte(v.botToken))

	// Calculate HMAC
	h := hmac.New(sha256.New, secretKey[:])
	h.Write([]byte(dataCheckString))
	calculatedHash := hex.EncodeToString(h.Sum(nil))

	// Compare hashes
	return calculatedHash == hash
}

// VerifyTelegramAuthWithTime verifies Telegram auth with time validation
func (v *Validator) VerifyTelegramAuthWithTime(authData map[string]string) bool {
	// First verify the hash
	if !v.VerifyTelegramAuth(authData) {
		return false
	}

	// Check auth_date if provided
	if authDateStr, exists := authData["auth_date"]; exists && authDateStr != "" {
		authDate, err := strconv.ParseInt(authDateStr, 10, 64)
		if err != nil {
			return false
		}

		// Check if auth_date is not too old (within 24 hours)
		currentTime := time.Now().Unix()
		if currentTime-authDate > 86400 { // 24 hours
			return false
		}
	}

	return true
}
