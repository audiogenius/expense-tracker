package handlers

import (
	"net/http"
	"strconv"
	"time"
)

// TransactionValidator handles validation for transaction operations
type TransactionValidator struct{}

// NewTransactionValidator creates a new TransactionValidator instance
func NewTransactionValidator() *TransactionValidator {
	return &TransactionValidator{}
}

// ValidateCreateRequest validates create transaction request
func (v *TransactionValidator) ValidateCreateRequest(req createTransactionRequest) (string, int) {
	if req.AmountCents <= 0 {
		return "amount must be positive", http.StatusBadRequest
	}

	if req.OperationType != "expense" && req.OperationType != "income" {
		return "operation_type must be 'expense' or 'income'", http.StatusBadRequest
	}

	if _, err := time.Parse(time.RFC3339, req.Timestamp); err != nil {
		return "invalid timestamp format", http.StatusBadRequest
	}

	return "", 0
}

// ValidateTransactionID validates transaction ID from URL
func (v *TransactionValidator) ValidateTransactionID(transactionIDStr string) (int, string, int) {
	transactionID, err := strconv.Atoi(transactionIDStr)
	if err != nil {
		return 0, "invalid transaction id", http.StatusBadRequest
	}
	return transactionID, "", 0
}

// ValidateLimit validates limit parameter
func (v *TransactionValidator) ValidateLimit(limitStr string, defaultLimit, maxLimit int) int {
	if limitStr == "" {
		return defaultLimit
	}

	if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= maxLimit {
		return limit
	}

	return defaultLimit
}
