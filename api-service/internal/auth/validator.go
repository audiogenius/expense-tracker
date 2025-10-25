package auth

import (
	"os"
	"strings"
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
