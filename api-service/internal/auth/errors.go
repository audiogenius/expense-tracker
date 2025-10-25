package auth

import "errors"

// Auth errors for better error handling
var (
	ErrInvalidToken        = errors.New("invalid token")
	ErrTokenExpired        = errors.New("token expired")
	ErrUserNotFound        = errors.New("user not found")
	ErrUserNotWhitelisted  = errors.New("user not in whitelist")
	ErrInvalidTelegramAuth = errors.New("invalid telegram authentication")
	ErrMissingJWTSecret    = errors.New("JWT secret not configured")
	ErrMissingBotToken     = errors.New("telegram bot token not configured")
)

// AuthError represents an authentication error with context
type AuthError struct {
	Err     error
	Code    string
	Message string
}

func (e *AuthError) Error() string {
	return e.Message
}

func (e *AuthError) Unwrap() error {
	return e.Err
}

// NewAuthError creates a new authentication error
func NewAuthError(err error, code, message string) *AuthError {
	return &AuthError{
		Err:     err,
		Code:    code,
		Message: message,
	}
}
