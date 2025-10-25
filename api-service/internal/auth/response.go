package auth

import (
	"encoding/json"
	"net/http"
)

// LoginResponse represents the response structure for login
type LoginResponse struct {
	Token    string `json:"token"`
	Username string `json:"username"`
	ID       string `json:"id"`
	PhotoURL string `json:"photo_url,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
}

// WriteLoginResponse writes a successful login response
func WriteLoginResponse(w http.ResponseWriter, token, username, id, photoURL string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := LoginResponse{
		Token:    token,
		Username: username,
		ID:       id,
	}

	if photoURL != "" {
		response.PhotoURL = photoURL
	}

	json.NewEncoder(w).Encode(response)
}

// WriteErrorResponse writes an error response
func WriteErrorResponse(w http.ResponseWriter, statusCode int, err *AuthError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Error:   err.Code,
		Message: err.Message,
	}

	json.NewEncoder(w).Encode(response)
}

// WriteSimpleError writes a simple error response
func WriteSimpleError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Error:   "auth_error",
		Message: message,
	}

	json.NewEncoder(w).Encode(response)
}
