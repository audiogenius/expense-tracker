package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// InternalHandlers handles all internal bot-related endpoints
type InternalHandlers struct {
	DB *pgxpool.Pool
}

// NewInternalHandlers creates a new InternalHandlers instance
func NewInternalHandlers(db *pgxpool.Pool) *InternalHandlers {
	return &InternalHandlers{DB: db}
}

// InternalPostExpense accepts a trusted request from the bot service to create an expense
// Payload: { telegram_id: number|string, username?: string, amount_cents: number, timestamp?: string }
// Protected by header X-BOT-KEY matching env BOT_API_KEY
func (h *InternalHandlers) InternalPostExpense(w http.ResponseWriter, r *http.Request) {
	botKey := os.Getenv("BOT_API_KEY")
	if botKey == "" {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if got := r.Header.Get("X-BOT-KEY"); got != botKey {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	// extract telegram id
	var telegramID int64
	switch v := payload["telegram_id"].(type) {
	case float64:
		telegramID = int64(v)
	case string:
		fmt.Sscan(v, &telegramID)
	default:
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	username := ""
	if u, ok := payload["username"].(string); ok {
		username = u
	}
	// ensure user exists
	var internalID int64
	if err := h.DB.QueryRow(r.Context(), "INSERT INTO users (telegram_id, username) VALUES ($1,$2) ON CONFLICT (telegram_id) DO UPDATE SET username=EXCLUDED.username RETURNING id", telegramID, username).Scan(&internalID); err != nil {
		log.Error().Err(err).Msg("create user internal")
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}
	// parse amount_cents
	var amountCents int
	switch v := payload["amount_cents"].(type) {
	case float64:
		amountCents = int(v)
	case int64:
		amountCents = int(v)
	case int:
		amountCents = v
	default:
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// parse category_id (optional)
	var categoryID *int
	if catID, ok := payload["category_id"]; ok && catID != nil {
		switch v := catID.(type) {
		case float64:
			id := int(v)
			categoryID = &id
		case int64:
			id := int(v)
			categoryID = &id
		case int:
			categoryID = &v
		}
	}

	// parse group_id (optional)
	var groupID *int64
	if gID, ok := payload["group_id"]; ok && gID != nil {
		switch v := gID.(type) {
		case float64:
			id := int64(v)
			groupID = &id
		case int64:
			groupID = &v
		case int:
			id := int64(v)
			groupID = &id
		}
	}

	// parse is_private (optional, default false)
	isPrivate := false
	if priv, ok := payload["is_private"].(bool); ok {
		isPrivate = priv
	}

	ts := time.Now().UTC()
	if t, ok := payload["timestamp"].(string); ok && t != "" {
		if parsed, err := time.Parse(time.RFC3339, t); err == nil {
			ts = parsed.UTC()
		}
	}
	if _, err := h.DB.Exec(r.Context(), `INSERT INTO expenses (user_id, amount_cents, category_id, timestamp, is_shared, group_id, is_private) VALUES ($1,$2,$3,$4,$5,$6,$7)`, internalID, amountCents, categoryID, ts, false, groupID, isPrivate); err != nil {
		log.Error().Err(err).Msg("insert expense internal")
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// InternalGetTotalExpenses returns total expenses for a user by telegram_id (for bot)
func (h *InternalHandlers) InternalGetTotalExpenses(w http.ResponseWriter, r *http.Request) {
	botKey := os.Getenv("BOT_API_KEY")
	if botKey == "" {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if got := r.Header.Get("X-BOT-KEY"); got != botKey {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	telegramIDStr := r.URL.Query().Get("telegram_id")
	if telegramIDStr == "" {
		http.Error(w, "telegram_id required", http.StatusBadRequest)
		return
	}

	var telegramID int64
	if _, err := fmt.Sscanf(telegramIDStr, "%d", &telegramID); err != nil {
		http.Error(w, "invalid telegram_id", http.StatusBadRequest)
		return
	}

	// Get internal user ID
	var userID int64
	if err := h.DB.QueryRow(r.Context(), "SELECT id FROM users WHERE telegram_id=$1", telegramID).Scan(&userID); err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	period := r.URL.Query().Get("period")
	var whereClause string
	var args []interface{}

	switch period {
	case "week":
		whereClause = "WHERE user_id=$1 AND timestamp >= NOW() - INTERVAL '7 days'"
		args = []interface{}{userID}
	case "month":
		whereClause = "WHERE user_id=$1 AND timestamp >= NOW() - INTERVAL '30 days'"
		args = []interface{}{userID}
	default:
		whereClause = "WHERE user_id=$1"
		args = []interface{}{userID}
	}

	var totalCents int
	err := h.DB.QueryRow(r.Context(), fmt.Sprintf("SELECT COALESCE(SUM(amount_cents), 0) FROM expenses %s", whereClause), args...).Scan(&totalCents)
	if err != nil {
		log.Error().Err(err).Msg("select total expenses internal")
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
}

// InternalGetDebts returns debts for a user by telegram_id (for bot)
func (h *InternalHandlers) InternalGetDebts(w http.ResponseWriter, r *http.Request) {
	botKey := os.Getenv("BOT_API_KEY")
	if botKey == "" {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if got := r.Header.Get("X-BOT-KEY"); got != botKey {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	telegramIDStr := r.URL.Query().Get("telegram_id")
	if telegramIDStr == "" {
		http.Error(w, "telegram_id required", http.StatusBadRequest)
		return
	}

	var telegramID int64
	if _, err := fmt.Sscanf(telegramIDStr, "%d", &telegramID); err != nil {
		http.Error(w, "invalid telegram_id", http.StatusBadRequest)
		return
	}

	// Get internal user ID
	var userID int64
	if err := h.DB.QueryRow(r.Context(), "SELECT id FROM users WHERE telegram_id=$1", telegramID).Scan(&userID); err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
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
		log.Error().Err(err).Msg("select debts internal")
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
		log.Error().Err(err).Msg("select debts owed internal")
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
}

// InternalRegisterGroup registers or updates a Telegram group
func (h *InternalHandlers) InternalRegisterGroup(w http.ResponseWriter, r *http.Request) {
	botKey := os.Getenv("BOT_API_KEY")
	if botKey == "" || r.Header.Get("X-BOT-KEY") != botKey {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	var payload struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
		Type string `json:"type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// Upsert group
	_, err := h.DB.Exec(r.Context(), `
		INSERT INTO telegram_groups (id, name, type, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			type = EXCLUDED.type,
			updated_at = NOW()
	`, payload.ID, payload.Name, payload.Type)

	if err != nil {
		log.Error().Err(err).Msg("failed to register group")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// InternalRegisterGroupMember registers a user in a group
func (h *InternalHandlers) InternalRegisterGroupMember(w http.ResponseWriter, r *http.Request) {
	botKey := os.Getenv("BOT_API_KEY")
	if botKey == "" || r.Header.Get("X-BOT-KEY") != botKey {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	var payload struct {
		GroupID  int64  `json:"group_id"`
		UserID   int64  `json:"user_id"`
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// First, ensure user exists
	_, err := h.DB.Exec(r.Context(), `
		INSERT INTO users (telegram_id, username)
		VALUES ($1, $2)
		ON CONFLICT (telegram_id) DO UPDATE SET
			username = EXCLUDED.username
	`, payload.UserID, payload.Username)

	if err != nil {
		log.Error().Err(err).Msg("failed to register user")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Then, add to group_members
	_, err = h.DB.Exec(r.Context(), `
		INSERT INTO group_members (group_id, user_id, role, joined_at)
		VALUES ($1, $2, 'member', NOW())
		ON CONFLICT (group_id, user_id) DO NOTHING
	`, payload.GroupID, payload.UserID)

	if err != nil {
		log.Error().Err(err).Msg("failed to register group member")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
