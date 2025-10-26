package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/expense-tracker/api-service/internal/auth"
	"github.com/expense-tracker/api-service/internal/cache"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// TransactionHandlers handles all transaction-related endpoints
type TransactionHandlers struct {
	DB               *pgxpool.Pool
	Cache            *cache.MemoryCache
	SuggestionsCache map[int]suggestionsCache // User ID -> Cache
	Auth             *auth.Auth
	Queries          *TransactionQueries
	Validator        *TransactionValidator
}

// NewTransactionHandlers creates a new TransactionHandlers instance
func NewTransactionHandlers(db *pgxpool.Pool, auth *auth.Auth) *TransactionHandlers {
	return &TransactionHandlers{
		DB:               db,
		Cache:            cache.NewMemoryCache(),
		SuggestionsCache: make(map[int]suggestionsCache),
		Auth:             auth,
		Queries:          &TransactionQueries{DB: db},
		Validator:        &TransactionValidator{},
	}
}

type transactionResponse struct {
	ID              int     `json:"id"`
	UserID          int64   `json:"user_id"`
	AmountCents     int     `json:"amount_cents"`
	CategoryID      *int    `json:"category_id"`
	SubcategoryID   *int    `json:"subcategory_id"`
	OperationType   string  `json:"operation_type"`
	Timestamp       string  `json:"timestamp"`
	IsShared        bool    `json:"is_shared"`
	Username        string  `json:"username"`
	CategoryName    *string `json:"category_name"`
	SubcategoryName *string `json:"subcategory_name"`
}

// GetTransactions returns paginated transactions with filters using keyset pagination
func (h *TransactionHandlers) GetTransactions(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("GetTransactions: Starting request")
	
	// TEMPORARY: Skip auth for debugging
	var userID int64 = 260144148 // Use your Telegram ID for testing
	
	log.Info().Int64("user_id", userID).Msg("GetTransactions: User authenticated (debug mode)")

	// Parse query parameters
	operationType := r.URL.Query().Get("operation_type")
	categoryID := r.URL.Query().Get("category_id")
	subcategoryID := r.URL.Query().Get("subcategory_id")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	scope := r.URL.Query().Get("scope")   // Filter: all, personal, family
	cursor := r.URL.Query().Get("cursor") // timestamp for keyset pagination
	limitStr := r.URL.Query().Get("limit")

	// Set defaults - limit to 20 for memory efficiency
	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	// Check cache first
	cacheKey := fmt.Sprintf("transactions_%d_%s_%s_%s_%s_%s_%s_%s",
		userID, operationType, categoryID, subcategoryID, startDate, endDate, scope, cursor)

	if cached, found := h.Cache.Get(cacheKey); found {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cached)
		return
	}

	// Build WHERE clause
	var whereConditions []string
	var args []interface{}
	argIndex := 1

	// Always exclude soft-deleted transactions
	whereConditions = append(whereConditions, "e.deleted_at IS NULL")

	// Operation type filter
	if operationType != "" && operationType != "both" {
		whereConditions = append(whereConditions, fmt.Sprintf("e.operation_type = $%d", argIndex))
		args = append(args, operationType)
		argIndex++
	}

	// Category filter
	if categoryID != "" {
		if catID, err := strconv.Atoi(categoryID); err == nil {
			whereConditions = append(whereConditions, fmt.Sprintf("e.category_id = $%d", argIndex))
			args = append(args, catID)
			argIndex++
		}
	}

	// Subcategory filter
	if subcategoryID != "" {
		if subID, err := strconv.Atoi(subcategoryID); err == nil {
			whereConditions = append(whereConditions, fmt.Sprintf("e.subcategory_id = $%d", argIndex))
			args = append(args, subID)
			argIndex++
		}
	}

	// Date filters
	if startDate != "" {
		if _, err := time.Parse(time.RFC3339, startDate); err == nil {
			whereConditions = append(whereConditions, fmt.Sprintf("e.timestamp >= $%d", argIndex))
			args = append(args, startDate)
			argIndex++
		}
	}
	if endDate != "" {
		if _, err := time.Parse(time.RFC3339, endDate); err == nil {
			whereConditions = append(whereConditions, fmt.Sprintf("e.timestamp <= $%d", argIndex))
			args = append(args, endDate)
			argIndex++
		}
	}

	// Get user's groups for filtering
	var userGroupIDs []int64
	groupRows, err := h.DB.Query(r.Context(),
		"SELECT group_id FROM group_members WHERE user_id = $1", userID)
	if err != nil {
		// Log error but continue - user might not be in any groups yet
		log.Warn().Err(err).Int64("user_id", userID).Msg("failed to query group_members, continuing without group filtering")
	} else {
		defer groupRows.Close()
		for groupRows.Next() {
			var groupID int64
			if err := groupRows.Scan(&groupID); err != nil {
				log.Warn().Err(err).Int64("user_id", userID).Msg("failed to scan group_id, skipping")
				continue
			}
			userGroupIDs = append(userGroupIDs, groupID)
		}
		// Check for iteration errors
		if err := groupRows.Err(); err != nil {
			log.Warn().Err(err).Int64("user_id", userID).Msg("error during group_members iteration")
		}
	}

	// Apply scope filter (personal, family, all)
	switch scope {
	case "personal":
		// Show only user's own expenses
		whereConditions = append(whereConditions, fmt.Sprintf("e.user_id = $%d", argIndex))
		args = append(args, userID)
		argIndex++
	case "family":
		// Show only group expenses (non-private) that user is a member of
		if len(userGroupIDs) > 0 {
			whereConditions = append(whereConditions, fmt.Sprintf(
				"(e.group_id = ANY($%d) AND e.is_private = false AND e.user_id != $%d)",
				argIndex, argIndex+1))
			args = append(args, userGroupIDs, userID)
			argIndex += 2
		} else {
			// User not in any group, return empty result
			whereConditions = append(whereConditions, "1=0")
		}
	default:
		// "all" or empty: show user's own expenses + group expenses (non-private)
		if len(userGroupIDs) > 0 {
			// User is in groups: show own expenses + non-private group expenses
			whereConditions = append(whereConditions, fmt.Sprintf(
				"(e.user_id = $%d OR (e.group_id = ANY($%d) AND e.is_private = false))",
				argIndex, argIndex+1))
			args = append(args, userID, userGroupIDs)
			argIndex += 2
		} else {
			// User not in any group: show only own expenses
			whereConditions = append(whereConditions, fmt.Sprintf("e.user_id = $%d", argIndex))
			args = append(args, userID)
			argIndex++
		}
	}

	// Add keyset pagination condition
	if cursor != "" {
		if cursorTime, err := time.Parse(time.RFC3339, cursor); err == nil {
			whereConditions = append(whereConditions, fmt.Sprintf("e.timestamp < $%d", argIndex))
			args = append(args, cursorTime)
			argIndex++
		}
	}

	// Build WHERE clause
	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Optimized query with keyset pagination (no COUNT for performance)
	query := fmt.Sprintf(`
		SELECT e.id, e.user_id, e.amount_cents, e.category_id, e.subcategory_id, 
			   e.operation_type, e.timestamp, e.is_shared, u.username,
			   c.name as category_name, s.name as subcategory_name
		FROM expenses e
		LEFT JOIN users u ON u.telegram_id = e.user_id
		LEFT JOIN categories c ON e.category_id = c.id
		LEFT JOIN subcategories s ON e.subcategory_id = s.id
		%s
		ORDER BY e.timestamp DESC, e.id DESC
		LIMIT $%d
	`, whereClause, argIndex)

	args = append(args, limit+1) // Get one extra to check if there are more

	rows, err := h.DB.Query(r.Context(), query, args...)
	if err != nil {
		log.Error().Err(err).Msg("select transactions")
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var transactions []transactionResponse
	var nextCursor string

	for rows.Next() {
		var t transactionResponse
		var ts time.Time
		var username *string
		var categoryName *string
		var subcategoryName *string

		if err := rows.Scan(&t.ID, &t.UserID, &t.AmountCents, &t.CategoryID, &t.SubcategoryID,
			&t.OperationType, &ts, &t.IsShared, &username, &categoryName, &subcategoryName); err == nil {
			t.Timestamp = ts.UTC().Format(time.RFC3339)
			if username != nil {
				t.Username = *username
			}
			if categoryName != nil {
				t.CategoryName = categoryName
			}
			if subcategoryName != nil {
				t.SubcategoryName = subcategoryName
			}
			transactions = append(transactions, t)
		}
	}

	// Check if there are more records (keyset pagination)
	hasMore := len(transactions) > limit
	if hasMore {
		// Remove the extra record and set next cursor
		transactions = transactions[:limit]
		nextCursor = transactions[len(transactions)-1].Timestamp
	}

	response := map[string]interface{}{
		"transactions": transactions,
		"pagination": map[string]interface{}{
			"limit":       limit,
			"has_more":    hasMore,
			"next_cursor": nextCursor,
		},
		"filters": map[string]interface{}{
			"operation_type": operationType,
			"category_id":    categoryID,
			"subcategory_id": subcategoryID,
			"start_date":     startDate,
			"end_date":       endDate,
		},
	}

	// Cache the response
	h.Cache.Set(cacheKey, response, 5*time.Minute)

	log.Info().Int64("user_id", userID).Int("count", len(transactions)).Msg("GetTransactions: Returning response")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	log.Info().Int64("user_id", userID).Int("count", len(transactions)).Msg("returned transactions")
}

// SoftDeleteTransaction marks a transaction as deleted
func (h *TransactionHandlers) SoftDeleteTransaction(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value(auth.UserIDKey)
	if uid == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	userID, ok := uid.(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Get transaction ID from URL path
	transactionIDStr := r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
	transactionID, err := strconv.Atoi(transactionIDStr)
	if err != nil {
		http.Error(w, "invalid transaction id", http.StatusBadRequest)
		return
	}

	// Check if transaction exists and belongs to user
	var exists bool
	err = h.DB.QueryRow(r.Context(),
		"SELECT EXISTS(SELECT 1 FROM expenses WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL)",
		transactionID, userID).Scan(&exists)
	if err != nil || !exists {
		http.Error(w, "transaction not found", http.StatusNotFound)
		return
	}

	// Soft delete the transaction
	_, err = h.DB.Exec(r.Context(),
		"UPDATE expenses SET deleted_at = NOW() WHERE id = $1 AND user_id = $2",
		transactionID, userID)
	if err != nil {
		log.Error().Err(err).Int64("user_id", userID).Int("transaction_id", transactionID).Msg("failed to soft delete transaction")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Clear cache
	h.Cache.Clear()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
	log.Info().Int64("user_id", userID).Int("transaction_id", transactionID).Msg("transaction soft deleted")
}

// RestoreTransaction restores a soft-deleted transaction
func (h *TransactionHandlers) RestoreTransaction(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value(auth.UserIDKey)
	if uid == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	userID, ok := uid.(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Get transaction ID from URL path
	transactionIDStr := r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
	transactionID, err := strconv.Atoi(transactionIDStr)
	if err != nil {
		http.Error(w, "invalid transaction id", http.StatusBadRequest)
		return
	}

	// Check if transaction exists and belongs to user
	var exists bool
	err = h.DB.QueryRow(r.Context(),
		"SELECT EXISTS(SELECT 1 FROM expenses WHERE id = $1 AND user_id = $2 AND deleted_at IS NOT NULL)",
		transactionID, userID).Scan(&exists)
	if err != nil || !exists {
		http.Error(w, "deleted transaction not found", http.StatusNotFound)
		return
	}

	// Restore the transaction
	_, err = h.DB.Exec(r.Context(),
		"UPDATE expenses SET deleted_at = NULL WHERE id = $1 AND user_id = $2",
		transactionID, userID)
	if err != nil {
		log.Error().Err(err).Int64("user_id", userID).Int("transaction_id", transactionID).Msg("failed to restore transaction")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Clear cache
	h.Cache.Clear()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "restored"})
	log.Info().Int64("user_id", userID).Int("transaction_id", transactionID).Msg("transaction restored")
}

// GetDeletedTransactions returns soft-deleted transactions for management
func (h *TransactionHandlers) GetDeletedTransactions(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value(auth.UserIDKey)
	if uid == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	userID, ok := uid.(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// Query deleted transactions
	query := `
		SELECT e.id, e.user_id, e.amount_cents, e.category_id, e.subcategory_id, 
			   e.operation_type, e.timestamp, e.is_shared, e.deleted_at, u.username,
			   c.name as category_name, s.name as subcategory_name
		FROM expenses e
		LEFT JOIN users u ON u.telegram_id = e.user_id
		LEFT JOIN categories c ON e.category_id = c.id
		LEFT JOIN subcategories s ON e.subcategory_id = s.id
		WHERE e.user_id = $1 AND e.deleted_at IS NOT NULL
		ORDER BY e.deleted_at DESC
		LIMIT $2
	`

	rows, err := h.DB.Query(r.Context(), query, userID, limit)
	if err != nil {
		log.Error().Err(err).Msg("select deleted transactions")
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var transactions []map[string]interface{}
	for rows.Next() {
		var t transactionResponse
		var ts, deletedAt time.Time
		var username *string
		var categoryName *string
		var subcategoryName *string

		if err := rows.Scan(&t.ID, &t.UserID, &t.AmountCents, &t.CategoryID, &t.SubcategoryID,
			&t.OperationType, &ts, &t.IsShared, &deletedAt, &username, &categoryName, &subcategoryName); err == nil {
			t.Timestamp = ts.UTC().Format(time.RFC3339)
			if username != nil {
				t.Username = *username
			}
			if categoryName != nil {
				t.CategoryName = categoryName
			}
			if subcategoryName != nil {
				t.SubcategoryName = subcategoryName
			}

			transaction := map[string]interface{}{
				"id":               t.ID,
				"user_id":          t.UserID,
				"amount_cents":     t.AmountCents,
				"category_id":      t.CategoryID,
				"subcategory_id":   t.SubcategoryID,
				"operation_type":   t.OperationType,
				"timestamp":        t.Timestamp,
				"is_shared":        t.IsShared,
				"username":         t.Username,
				"category_name":    t.CategoryName,
				"subcategory_name": t.SubcategoryName,
				"deleted_at":       deletedAt.UTC().Format(time.RFC3339),
			}
			transactions = append(transactions, transaction)
		}
	}

	response := map[string]interface{}{
		"transactions": transactions,
		"count":        len(transactions),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	log.Info().Int64("user_id", userID).Int("count", len(transactions)).Msg("returned deleted transactions")
}

type createTransactionRequest struct {
	AmountCents   int    `json:"amount_cents"`
	CategoryID    *int   `json:"category_id"`
	SubcategoryID *int   `json:"subcategory_id"`
	OperationType string `json:"operation_type"` // "expense" or "income"
	Timestamp     string `json:"timestamp"`
	IsShared      bool   `json:"is_shared"`
	GroupID       *int64 `json:"group_id"`
}

// CreateTransaction creates a new transaction
func (h *TransactionHandlers) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value(auth.UserIDKey)
	if uid == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	userID, ok := uid.(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req createTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// Validate request
	if errorMsg, statusCode := h.Validator.ValidateCreateRequest(req); errorMsg != "" {
		http.Error(w, errorMsg, statusCode)
		return
	}

	// Parse timestamp
	timestamp, err := time.Parse(time.RFC3339, req.Timestamp)
	if err != nil {
		http.Error(w, "invalid timestamp format", http.StatusBadRequest)
		return
	}

	// Validate category if provided
	if req.CategoryID != nil {
		exists, err := h.Queries.ValidateCategory(r.Context(), *req.CategoryID)
		if err != nil || !exists {
			http.Error(w, "category not found", http.StatusBadRequest)
			return
		}
	}

	// Validate subcategory if provided
	if req.SubcategoryID != nil {
		exists, err := h.Queries.ValidateSubcategory(r.Context(), *req.SubcategoryID)
		if err != nil || !exists {
			http.Error(w, "subcategory not found", http.StatusBadRequest)
			return
		}
	}

	// Insert transaction
	var transactionID int
	err = h.DB.QueryRow(r.Context(), `
		INSERT INTO expenses (user_id, amount_cents, category_id, subcategory_id, operation_type, timestamp, is_shared, group_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`, userID, req.AmountCents, req.CategoryID, req.SubcategoryID, req.OperationType, timestamp, req.IsShared, req.GroupID).Scan(&transactionID)

	if err != nil {
		log.Error().Err(err).Msg("create transaction")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Clear relevant caches
	h.Cache.ClearPattern("/api/transactions")
	h.Cache.ClearPattern("/api/expenses")
	h.Cache.ClearPattern("/api/incomes")
	h.Cache.ClearPattern("/api/balance")

	response := map[string]interface{}{
		"id":             transactionID,
		"user_id":        userID,
		"amount_cents":   req.AmountCents,
		"category_id":    req.CategoryID,
		"subcategory_id": req.SubcategoryID,
		"operation_type": req.OperationType,
		"timestamp":      req.Timestamp,
		"is_shared":      req.IsShared,
		"group_id":       req.GroupID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
	log.Info().Int64("user_id", userID).Int("transaction_id", transactionID).Str("operation_type", req.OperationType).Msg("transaction created")
}
