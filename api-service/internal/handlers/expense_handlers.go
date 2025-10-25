package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/expense-tracker/api-service/internal/auth"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// ExpenseHandlers handles all expense-related endpoints
type ExpenseHandlers struct {
	DB   *pgxpool.Pool
	Auth *auth.Auth
}

// NewExpenseHandlers creates a new ExpenseHandlers instance
func NewExpenseHandlers(db *pgxpool.Pool, auth *auth.Auth) *ExpenseHandlers {
	return &ExpenseHandlers{DB: db, Auth: auth}
}

type ExpenseRequest struct {
	AmountCents   int    `json:"amount_cents"`
	CategoryID    *int   `json:"category_id"`
	SubcategoryID *int   `json:"subcategory_id"`
	OperationType string `json:"operation_type"` // "expense" or "income"
	Timestamp     string `json:"timestamp"`      // RFC3339 optional
	IsShared      bool   `json:"is_shared"`
}

// AddExpense creates an expense for the authenticated user
func (h *ExpenseHandlers) AddExpense(w http.ResponseWriter, r *http.Request) {
	var req ExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

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

	// Validate operation_type
	if req.OperationType == "" {
		req.OperationType = "expense" // default
	}
	if req.OperationType != "expense" && req.OperationType != "income" {
		http.Error(w, "invalid operation_type", http.StatusBadRequest)
		return
	}

	// Validate subcategory_id if provided
	if req.SubcategoryID != nil && req.CategoryID != nil {
		var exists bool
		err := h.DB.QueryRow(r.Context(),
			"SELECT EXISTS(SELECT 1 FROM subcategories WHERE id = $1 AND category_id = $2)",
			*req.SubcategoryID, *req.CategoryID).Scan(&exists)
		if err != nil || !exists {
			http.Error(w, "subcategory does not belong to category", http.StatusBadRequest)
			return
		}
	}

	// parse timestamp if provided
	ts := time.Now().UTC()
	if req.Timestamp != "" {
		if parsed, err := time.Parse(time.RFC3339, req.Timestamp); err == nil {
			ts = parsed.UTC()
		}
	}

	var expenseID int
	err := h.DB.QueryRow(r.Context(),
		`INSERT INTO expenses (user_id, amount_cents, category_id, subcategory_id, operation_type, timestamp, is_shared) 
		 VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`,
		userID, req.AmountCents, req.CategoryID, req.SubcategoryID, req.OperationType, ts, req.IsShared).Scan(&expenseID)

	if err != nil {
		log.Error().Err(err).Msg("insert expense")
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int{"id": expenseID})
	log.Info().Int64("user_id", userID).Int("amount_cents", req.AmountCents).Str("operation_type", req.OperationType).Msg("expense added")
}

// GetExpenses returns recent expenses for ALL family members (whitelist)
func (h *ExpenseHandlers) GetExpenses(w http.ResponseWriter, r *http.Request) {
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

	// Get all telegram_ids from whitelist to fetch family expenses
	whitelistIDs := make([]int64, 0)
	for _, idStr := range h.Auth.Whitelist {
		if idStr != "*" {
			var tgID int64
			fmt.Sscan(idStr, &tgID)
			whitelistIDs = append(whitelistIDs, tgID)
		}
	}

	query := `
		SELECT e.id, e.user_id, e.amount_cents, e.category_id, e.subcategory_id, e.operation_type, e.timestamp, e.is_shared, u.username 
		FROM expenses e
		LEFT JOIN users u ON e.user_id = u.id
		WHERE u.telegram_id = ANY($1)
		ORDER BY e.timestamp DESC 
		LIMIT 200
	`

	rows, err := h.DB.Query(r.Context(), query, whitelistIDs)
	if err != nil {
		log.Error().Err(err).Msg("select expenses")
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type expense struct {
		ID            int    `json:"id"`
		UserID        int64  `json:"user_id"`
		AmountCents   int    `json:"amount_cents"`
		CategoryID    *int   `json:"category_id"`
		SubcategoryID *int   `json:"subcategory_id"`
		OperationType string `json:"operation_type"`
		Timestamp     string `json:"timestamp"`
		IsShared      bool   `json:"is_shared"`
		Username      string `json:"username"`
	}

	var res []expense
	for rows.Next() {
		var e expense
		var ts time.Time
		var username *string
		if err := rows.Scan(&e.ID, &e.UserID, &e.AmountCents, &e.CategoryID, &e.SubcategoryID, &e.OperationType, &ts, &e.IsShared, &username); err == nil {
			e.Timestamp = ts.UTC().Format(time.RFC3339)
			if username != nil {
				e.Username = *username
			}
			res = append(res, e)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
	log.Info().Int64("user_id", userID).Int("count", len(res)).Msg("returned family expenses")
}

// GetTotalExpenses returns total expenses for ALL family members with optional period filter
func (h *ExpenseHandlers) GetTotalExpenses(w http.ResponseWriter, r *http.Request) {
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

	// Get whitelist IDs
	whitelistIDs := make([]int64, 0)
	for _, idStr := range h.Auth.Whitelist {
		if idStr != "*" {
			var tgID int64
			fmt.Sscan(idStr, &tgID)
			whitelistIDs = append(whitelistIDs, tgID)
		}
	}

	period := r.URL.Query().Get("period")
	var timeFilter string

	switch period {
	case "week":
		timeFilter = "AND e.timestamp >= NOW() - INTERVAL '7 days'"
	case "month":
		timeFilter = "AND e.timestamp >= NOW() - INTERVAL '30 days'"
	default:
		timeFilter = ""
	}

	query := fmt.Sprintf(`
		SELECT COALESCE(SUM(e.amount_cents), 0) 
		FROM expenses e
		LEFT JOIN users u ON e.user_id = u.id
		WHERE u.telegram_id = ANY($1) %s
	`, timeFilter)

	var totalCents int
	err := h.DB.QueryRow(r.Context(), query, whitelistIDs).Scan(&totalCents)
	if err != nil {
		log.Error().Err(err).Msg("select total expenses")
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"total_cents":  totalCents,
		"total_rubles": float64(totalCents) / 100.0,
		"period":       period,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	log.Info().Int64("user_id", userID).Str("period", period).Int("total_cents", totalCents).Msg("returned family total expenses")
}
