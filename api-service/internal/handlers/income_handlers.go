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

// IncomeHandlers handles all income-related endpoints
type IncomeHandlers struct {
	DB   *pgxpool.Pool
	Auth *auth.Auth
}

// NewIncomeHandlers creates a new IncomeHandlers instance
func NewIncomeHandlers(db *pgxpool.Pool, auth *auth.Auth) *IncomeHandlers {
	return &IncomeHandlers{
		DB:   db,
		Auth: auth,
	}
}

type incomeRequest struct {
	AmountCents   int    `json:"amount_cents"`
	IncomeType    string `json:"income_type"` // salary, debt_return, prize, gift, refund, other
	Description   string `json:"description"`
	RelatedDebtID *int   `json:"related_debt_id"` // optional, for debt_return type
	Timestamp     string `json:"timestamp"`       // RFC3339 optional
}

// AddIncome creates an income for the authenticated user
func (h *IncomeHandlers) AddIncome(w http.ResponseWriter, r *http.Request) {
	var req incomeRequest
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

	// Validate income type
	validTypes := map[string]bool{
		"salary": true, "debt_return": true, "prize": true,
		"gift": true, "refund": true, "other": true,
	}
	if !validTypes[req.IncomeType] {
		req.IncomeType = "other"
	}

	// parse timestamp if provided
	ts := time.Now().UTC()
	if req.Timestamp != "" {
		if parsed, err := time.Parse(time.RFC3339, req.Timestamp); err == nil {
			ts = parsed.UTC()
		}
	}

	// If this is a debt_return and related_debt_id is provided, mark debt as paid
	if req.IncomeType == "debt_return" && req.RelatedDebtID != nil {
		_, err := h.DB.Exec(r.Context(),
			`UPDATE debts SET is_paid = true, paid_at = NOW() WHERE id = $1`,
			*req.RelatedDebtID)
		if err != nil {
			log.Error().Err(err).Msg("update debt status")
		}
	}

	var incomeID int
	err := h.DB.QueryRow(r.Context(),
		`INSERT INTO incomes (user_id, amount_cents, income_type, description, related_debt_id, timestamp) 
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		userID, req.AmountCents, req.IncomeType, req.Description, req.RelatedDebtID, ts).Scan(&incomeID)

	if err != nil {
		log.Error().Err(err).Msg("insert income")
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int{"id": incomeID})
	log.Info().Int64("user_id", userID).Int("amount_cents", req.AmountCents).Str("type", req.IncomeType).Msg("income added")
}

// GetIncomes returns recent incomes for ALL family members (whitelist)
func (h *IncomeHandlers) GetIncomes(w http.ResponseWriter, r *http.Request) {
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

	// Get all telegram_ids from whitelist to fetch family incomes
	whitelistIDs := make([]int64, 0)
	for _, idStr := range h.Auth.Whitelist {
		if idStr != "*" {
			var tgID int64
			fmt.Sscan(idStr, &tgID)
			whitelistIDs = append(whitelistIDs, tgID)
		}
	}

	query := `
		SELECT i.id, i.user_id, i.amount_cents, i.income_type, i.description, i.related_debt_id, i.timestamp, u.username 
		FROM incomes i
		LEFT JOIN users u ON i.user_id = u.id
		WHERE u.telegram_id = ANY($1)
		ORDER BY i.timestamp DESC 
		LIMIT 200
	`

	rows, err := h.DB.Query(r.Context(), query, whitelistIDs)
	if err != nil {
		log.Error().Err(err).Msg("select incomes")
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type income struct {
		ID            int     `json:"id"`
		UserID        int64   `json:"user_id"`
		AmountCents   int     `json:"amount_cents"`
		IncomeType    string  `json:"income_type"`
		Description   *string `json:"description"`
		RelatedDebtID *int    `json:"related_debt_id"`
		Timestamp     string  `json:"timestamp"`
		Username      string  `json:"username"`
	}

	var res []income
	for rows.Next() {
		var it income
		var ts time.Time
		var username *string
		if err := rows.Scan(&it.ID, &it.UserID, &it.AmountCents, &it.IncomeType, &it.Description, &it.RelatedDebtID, &ts, &username); err == nil {
			it.Timestamp = ts.UTC().Format(time.RFC3339)
			if username != nil {
				it.Username = *username
			}
			res = append(res, it)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
	log.Info().Int64("user_id", userID).Int("count", len(res)).Msg("returned family incomes")
}

// GetTotalIncomes returns total incomes for ALL family members with optional period filter
func (h *IncomeHandlers) GetTotalIncomes(w http.ResponseWriter, r *http.Request) {
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
		timeFilter = "AND i.timestamp >= NOW() - INTERVAL '7 days'"
	case "month":
		timeFilter = "AND i.timestamp >= NOW() - INTERVAL '30 days'"
	default:
		timeFilter = ""
	}

	query := fmt.Sprintf(`
		SELECT COALESCE(SUM(i.amount_cents), 0) 
		FROM incomes i
		LEFT JOIN users u ON i.user_id = u.id
		WHERE u.telegram_id = ANY($1) %s
	`, timeFilter)

	var totalCents int
	err := h.DB.QueryRow(r.Context(), query, whitelistIDs).Scan(&totalCents)
	if err != nil {
		log.Error().Err(err).Msg("select total incomes")
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
	log.Info().Int64("user_id", userID).Str("period", period).Int("total_cents", totalCents).Msg("returned family total incomes")
}
