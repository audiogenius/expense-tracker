package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/expense-tracker/api-service/internal/auth"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// Handlers holds dependencies for HTTP handlers
type Handlers struct {
	Auth *auth.Auth
	DB   *pgxpool.Pool
}

func NewHandlers(a *auth.Auth, db *pgxpool.Pool) *Handlers {
	return &Handlers{Auth: a, DB: db}
}

type expenseRequest struct {
	AmountCents int    `json:"amount_cents"`
	CategoryID  *int   `json:"category_id"`
	Timestamp   string `json:"timestamp"` // RFC3339 optional
	IsShared    bool   `json:"is_shared"`
}

// Login accepts a Telegram widget payload (map[string]string), verifies it and returns a JWT
func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	var payload map[string]string
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	// whitelist check
	tg := payload["id"]
	allowed := false
	for _, v := range h.Auth.Whitelist {
		if v == "*" || v == tg {
			allowed = true
			break
		}
	}
	if !allowed || !h.Auth.VerifyTelegramAuth(payload) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	// ensure user exists (upsert)
	var id int64
	var telegramID int64
	fmt.Sscan(payload["id"], &telegramID)
	if err := h.DB.QueryRow(r.Context(), "INSERT INTO users (telegram_id, username) VALUES ($1,$2) ON CONFLICT (telegram_id) DO UPDATE SET username=EXCLUDED.username RETURNING id", telegramID, payload["username"]).Scan(&id); err != nil {
		log.Error().Err(err).Msg("create user")
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}
	token, err := h.Auth.CreateJWT(telegramID)
	if err != nil {
		log.Error().Err(err).Msg("create jwt")
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	// Return token and basic profile info (id and username) for frontend convenience
	resp := map[string]string{"token": token, "username": payload["username"], "id": payload["id"]}
	if p, ok := payload["photo_url"]; ok && p != "" {
		resp["photo_url"] = p
	}
	_ = json.NewEncoder(w).Encode(resp)
}

// AddExpense creates an expense for the authenticated user
func (h *Handlers) AddExpense(w http.ResponseWriter, r *http.Request) {
	var req expenseRequest
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
	// parse timestamp if provided
	ts := time.Now().UTC()
	if req.Timestamp != "" {
		if parsed, err := time.Parse(time.RFC3339, req.Timestamp); err == nil {
			ts = parsed.UTC()
		}
	}
	if _, err := h.DB.Exec(r.Context(), `INSERT INTO expenses (user_id, amount_cents, category_id, timestamp, is_shared) VALUES ($1,$2,NULL,$3,$4)`, userID, req.AmountCents, ts, req.IsShared); err != nil {
		log.Error().Err(err).Msg("insert expense")
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	log.Info().Int64("user_id", userID).Int("amount_cents", req.AmountCents).Msg("expense added")
}

// GetExpenses returns recent expenses for the authenticated user
func (h *Handlers) GetExpenses(w http.ResponseWriter, r *http.Request) {
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
	rows, err := h.DB.Query(r.Context(), `SELECT id, user_id, amount_cents, category_id, timestamp, is_shared FROM expenses WHERE user_id=$1 ORDER BY timestamp DESC LIMIT 100`, userID)
	if err != nil {
		log.Error().Err(err).Msg("select expenses")
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	type e struct {
		ID          int    `json:"id"`
		UserID      int64  `json:"user_id"`
		AmountCents int    `json:"amount_cents"`
		CategoryID  *int   `json:"category_id"`
		Timestamp   string `json:"timestamp"`
		IsShared    bool   `json:"is_shared"`
	}
	var res []e
	for rows.Next() {
		var it e
		var ts time.Time
		if err := rows.Scan(&it.ID, &it.UserID, &it.AmountCents, &it.CategoryID, &ts, &it.IsShared); err == nil {
			it.Timestamp = ts.UTC().Format(time.RFC3339)
			res = append(res, it)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
	log.Info().Int64("user_id", userID).Int("count", len(res)).Msg("returned expenses")
}

// GetTotalExpenses returns total expenses for the authenticated user with optional period filter
func (h *Handlers) GetTotalExpenses(w http.ResponseWriter, r *http.Request) {
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
	log.Info().Int64("user_id", userID).Str("period", period).Int("total_cents", totalCents).Msg("returned total expenses")
}

// GetCategories returns all available categories
func (h *Handlers) GetCategories(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query(r.Context(), "SELECT id, name, aliases FROM categories ORDER BY name")
	if err != nil {
		log.Error().Err(err).Msg("select categories")
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type category struct {
		ID      int      `json:"id"`
		Name    string   `json:"name"`
		Aliases []string `json:"aliases"`
	}

	var categories []category
	for rows.Next() {
		var c category
		var aliasesJSON []byte
		if err := rows.Scan(&c.ID, &c.Name, &aliasesJSON); err == nil {
			json.Unmarshal(aliasesJSON, &c.Aliases)
			categories = append(categories, c)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
	log.Info().Int("count", len(categories)).Msg("returned categories")
}

// DetectCategory attempts to automatically detect category based on expense description
func (h *Handlers) DetectCategory(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// Simple keyword matching for category detection
	description := strings.ToLower(req.Description)

	rows, err := h.DB.Query(r.Context(), "SELECT id, name, aliases FROM categories")
	if err != nil {
		log.Error().Err(err).Msg("select categories for detection")
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var bestMatch *struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Score int    `json:"score"`
	}

	for rows.Next() {
		var id int
		var name string
		var aliasesJSON []byte
		if err := rows.Scan(&id, &name, &aliasesJSON); err != nil {
			continue
		}

		var aliases []string
		json.Unmarshal(aliasesJSON, &aliases)

		score := 0
		// Check if description contains category name
		if strings.Contains(description, strings.ToLower(name)) {
			score += 2
		}
		// Check aliases
		for _, alias := range aliases {
			if strings.Contains(description, strings.ToLower(alias)) {
				score += 1
			}
		}

		if bestMatch == nil || score > bestMatch.Score {
			bestMatch = &struct {
				ID    int    `json:"id"`
				Name  string `json:"name"`
				Score int    `json:"score"`
			}{ID: id, Name: name, Score: score}
		}
	}

	if bestMatch != nil && bestMatch.Score > 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(bestMatch)
	} else {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"id": nil, "name": "Не определено", "score": 0})
	}
}

// CreateSharedExpense creates a shared expense that can be split between users
func (h *Handlers) CreateSharedExpense(w http.ResponseWriter, r *http.Request) {
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
func (h *Handlers) GetDebts(w http.ResponseWriter, r *http.Request) {
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

// InternalPostExpense accepts a trusted request from the bot service to create an expense
// Payload: { telegram_id: number|string, username?: string, amount_cents: number, timestamp?: string }
// Protected by header X-BOT-KEY matching env BOT_API_KEY
func (h *Handlers) InternalPostExpense(w http.ResponseWriter, r *http.Request) {
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
	ts := time.Now().UTC()
	if t, ok := payload["timestamp"].(string); ok && t != "" {
		if parsed, err := time.Parse(time.RFC3339, t); err == nil {
			ts = parsed.UTC()
		}
	}
	if _, err := h.DB.Exec(r.Context(), `INSERT INTO expenses (user_id, amount_cents, category_id, timestamp, is_shared) VALUES ($1,$2,$3,$4,$5)`, internalID, amountCents, categoryID, ts, false); err != nil {
		log.Error().Err(err).Msg("insert expense internal")
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// InternalGetTotalExpenses returns total expenses for a user by telegram_id (for bot)
func (h *Handlers) InternalGetTotalExpenses(w http.ResponseWriter, r *http.Request) {
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
func (h *Handlers) InternalGetDebts(w http.ResponseWriter, r *http.Request) {
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
