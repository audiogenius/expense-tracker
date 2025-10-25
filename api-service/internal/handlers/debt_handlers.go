package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/expense-tracker/api-service/internal/auth"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// DebtHandlers handles all debt and balance-related endpoints
type DebtHandlers struct {
	DB   *pgxpool.Pool
	Auth *auth.Auth
}

// NewDebtHandlers creates a new DebtHandlers instance
func NewDebtHandlers(db *pgxpool.Pool, auth *auth.Auth) *DebtHandlers {
	return &DebtHandlers{
		DB:   db,
		Auth: auth,
	}
}

// CreateSharedExpense creates a shared expense that can be split between users
func (h *DebtHandlers) CreateSharedExpense(w http.ResponseWriter, r *http.Request) {
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

	var req struct {
		AmountCents int     `json:"amount_cents"`
		Description string  `json:"description"`
		CategoryID  *int    `json:"category_id"`
		SplitWith   []int64 `json:"split_with"` // Telegram IDs of users to split with
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// Create the shared expense
	var expenseID int
	err := h.DB.QueryRow(r.Context(), `INSERT INTO expenses (user_id, amount_cents, category_id, timestamp, is_shared) VALUES ($1,$2,$3,NOW(),true) RETURNING id`,
		userID, req.AmountCents, req.CategoryID).Scan(&expenseID)
	if err != nil {
		log.Error().Err(err).Msg("insert shared expense")
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}

	// Calculate split amount per person
	totalPeople := len(req.SplitWith) + 1 // +1 for the creator
	splitAmount := req.AmountCents / totalPeople
	remainder := req.AmountCents % totalPeople

	// Create debt records for each person
	for i, telegramID := range req.SplitWith {
		// Get internal user ID
		var splitUserID int64
		err := h.DB.QueryRow(r.Context(), "SELECT id FROM users WHERE telegram_id=$1", telegramID).Scan(&splitUserID)
		if err != nil {
			log.Error().Err(err).Int64("telegram_id", telegramID).Msg("user not found for split")
			continue
		}

		// Calculate amount for this person (add remainder to first person)
		amount := splitAmount
		if i == 0 {
			amount += remainder
		}

		// Create debt record
		_, err = h.DB.Exec(r.Context(), `INSERT INTO debts (from_user, to_user, amount_cents) VALUES ($1,$2,$3)`,
			splitUserID, userID, amount)
		if err != nil {
			log.Error().Err(err).Msg("insert debt record")
		}
	}

	response := map[string]interface{}{
		"expense_id":   expenseID,
		"split_amount": splitAmount,
		"total_people": totalPeople,
		"split_with":   req.SplitWith,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	log.Info().Int64("user_id", userID).Int("expense_id", expenseID).Int("amount_cents", req.AmountCents).Msg("created shared expense")
}

// GetDebts returns debts for the authenticated user
func (h *DebtHandlers) GetDebts(w http.ResponseWriter, r *http.Request) {
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

	// Get debts where user owes money
	rows, err := h.DB.Query(r.Context(), `
		SELECT d.id, d.amount_cents, u.username, u.telegram_id
		FROM debts d
		JOIN users u ON d.from_user = u.id
		WHERE d.to_user = $1
		ORDER BY d.id DESC
	`, userID)
	if err != nil {
		log.Error().Err(err).Msg("select debts")
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type debt struct {
		ID          int    `json:"id"`
		AmountCents int    `json:"amount_cents"`
		Username    string `json:"username"`
		TelegramID  int64  `json:"telegram_id"`
		Type        string `json:"type"` // "owed_to_me" or "i_owe"
	}

	var debts []debt
	for rows.Next() {
		var d debt
		if err := rows.Scan(&d.ID, &d.AmountCents, &d.Username, &d.TelegramID); err == nil {
			d.Type = "owed_to_me"
			debts = append(debts, d)
		}
	}

	// Get debts where user owes money to others
	rows2, err := h.DB.Query(r.Context(), `
		SELECT d.id, d.amount_cents, u.username, u.telegram_id
		FROM debts d
		JOIN users u ON d.to_user = u.id
		WHERE d.from_user = $1
		ORDER BY d.id DESC
	`, userID)
	if err != nil {
		log.Error().Err(err).Msg("select debts owed")
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}
	defer rows2.Close()

	for rows2.Next() {
		var d debt
		if err := rows2.Scan(&d.ID, &d.AmountCents, &d.Username, &d.TelegramID); err == nil {
			d.Type = "i_owe"
			debts = append(debts, d)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(debts)
	log.Info().Int64("user_id", userID).Int("count", len(debts)).Msg("returned debts")
}

// GetBalance returns family balance (total incomes - total expenses)
func (h *DebtHandlers) GetBalance(w http.ResponseWriter, r *http.Request) {
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
		timeFilter = "AND timestamp >= NOW() - INTERVAL '7 days'"
	case "month":
		timeFilter = "AND timestamp >= NOW() - INTERVAL '30 days'"
	default:
		timeFilter = ""
	}

	// Get total expenses
	expensesQuery := fmt.Sprintf(`
		SELECT COALESCE(SUM(e.amount_cents), 0) 
		FROM expenses e
		LEFT JOIN users u ON e.user_id = u.id
		WHERE u.telegram_id = ANY($1) %s
	`, timeFilter)

	var totalExpensesCents int
	if err := h.DB.QueryRow(r.Context(), expensesQuery, whitelistIDs).Scan(&totalExpensesCents); err != nil {
		log.Error().Err(err).Msg("select expenses for balance")
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}

	// Get total incomes
	incomesQuery := fmt.Sprintf(`
		SELECT COALESCE(SUM(i.amount_cents), 0) 
		FROM incomes i
		LEFT JOIN users u ON i.user_id = u.id
		WHERE u.telegram_id = ANY($1) %s
	`, timeFilter)

	var totalIncomesCents int
	if err := h.DB.QueryRow(r.Context(), incomesQuery, whitelistIDs).Scan(&totalIncomesCents); err != nil {
		log.Error().Err(err).Msg("select incomes for balance")
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}

	balanceCents := totalIncomesCents - totalExpensesCents

	response := map[string]interface{}{
		"balance_cents":         balanceCents,
		"balance_rubles":        float64(balanceCents) / 100.0,
		"total_incomes_cents":   totalIncomesCents,
		"total_incomes_rubles":  float64(totalIncomesCents) / 100.0,
		"total_expenses_cents":  totalExpensesCents,
		"total_expenses_rubles": float64(totalExpensesCents) / 100.0,
		"period":                period,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	log.Info().Int64("user_id", userID).Str("period", period).Int("balance_cents", balanceCents).Msg("returned family balance")
}
