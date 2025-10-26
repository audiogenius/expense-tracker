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
}

// NewTransactionHandlers creates a new TransactionHandlers instance
func NewTransactionHandlers(db *pgxpool.Pool, auth *auth.Auth) *TransactionHandlers {
	return &TransactionHandlers{
		DB:               db,
		Cache:            cache.NewMemoryCache(),
		SuggestionsCache: make(map[int]suggestionsCache),
		Auth:             auth,
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
	operationType := r.URL.Query().Get("operation_type")
	categoryID := r.URL.Query().Get("category_id")
	subcategoryID := r.URL.Query().Get("subcategory_id")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
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
	cacheKey := fmt.Sprintf("transactions_%d_%s_%s_%s_%s_%s_%s",
		userID, operationType, categoryID, subcategoryID, startDate, endDate, cursor)

	if cached, found := h.Cache.Get(cacheKey); found {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cached)
		return
	}

	// Build WHERE clause
	var whereConditions []string
	var args []interface{}
	argIndex := 1

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
	if err == nil {
		defer groupRows.Close()
		for groupRows.Next() {
			var groupID int64
			if err := groupRows.Scan(&groupID); err == nil {
				userGroupIDs = append(userGroupIDs, groupID)
			}
		}
	}

	// Filter by user's own expenses OR group expenses (excluding private expenses from others)
	// Logic: Show my expenses (all) + group expenses that are not private
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	log.Info().Int64("user_id", userID).Int("count", len(transactions)).Msg("returned transactions")
}
