package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/expense-tracker/api-service/internal/auth"
	"github.com/expense-tracker/api-service/internal/cache"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// Handlers holds dependencies for HTTP handlers
type Handlers struct {
	Auth             *auth.Auth
	DB               *pgxpool.Pool
	Cache            *cache.MemoryCache
	SuggestionsCache map[int]suggestionsCache // User ID -> Cache
}

func NewHandlers(a *auth.Auth, db *pgxpool.Pool) *Handlers {
	return &Handlers{
		Auth:             a,
		DB:               db,
		Cache:            cache.NewMemoryCache(),
		SuggestionsCache: make(map[int]suggestionsCache),
	}
}

type expenseRequest struct {
	AmountCents   int    `json:"amount_cents"`
	CategoryID    *int   `json:"category_id"`
	SubcategoryID *int   `json:"subcategory_id"`
	OperationType string `json:"operation_type"` // "expense" or "income"
	Timestamp     string `json:"timestamp"`      // RFC3339 optional
	IsShared      bool   `json:"is_shared"`
}

type subcategoryRequest struct {
	Name       string   `json:"name"`
	CategoryID int      `json:"category_id"`
	Aliases    []string `json:"aliases"`
}

type subcategoryResponse struct {
	ID         int      `json:"id"`
	Name       string   `json:"name"`
	CategoryID int      `json:"category_id"`
	Aliases    []string `json:"aliases"`
	CreatedAt  string   `json:"created_at"`
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

type transactionsFilter struct {
	OperationType string `json:"operation_type"` // "expense", "income", "both"
	CategoryID    *int   `json:"category_id"`
	SubcategoryID *int   `json:"subcategory_id"`
	StartDate     string `json:"start_date"` // RFC3339
	EndDate       string `json:"end_date"`   // RFC3339
	Page          int    `json:"page"`
	Limit         int    `json:"limit"`
}

type categorySuggestion struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Type  string  `json:"type"` // "category" or "subcategory"
	Score float64 `json:"score"`
	Usage int     `json:"usage"` // Usage frequency
}

type suggestionsCache struct {
	Data      []categorySuggestion `json:"data"`
	ExpiresAt time.Time            `json:"expires_at"`
	UserID    int                  `json:"user_id"`
}

// Login accepts a Telegram widget payload (map[string]string), verifies it and returns a JWT
func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Login handler called")
	var payload map[string]string
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Error().Err(err).Msg("failed to decode JSON payload")
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	log.Info().Interface("payload", payload).Msg("decoded payload")
	// whitelist check
	tg := payload["id"]
	log.Info().Str("telegram_id", tg).Strs("whitelist", h.Auth.Whitelist).Msg("checking whitelist")
	allowed := false
	for _, v := range h.Auth.Whitelist {
		if v == "*" || v == tg {
			allowed = true
			break
		}
	}
	if !allowed {
		log.Error().Str("telegram_id", tg).Strs("whitelist", h.Auth.Whitelist).Msg("user not in whitelist")
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	log.Info().Str("telegram_id", tg).Msg("user authorized")
	// For development/testing, skip Telegram auth verification if no hash is provided
	if hash, hasHash := payload["hash"]; hasHash && hash != "" {
		if !h.Auth.VerifyTelegramAuth(payload) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
	}
	// ensure user exists (upsert)
	var id int64
	var telegramID int64
	fmt.Sscan(payload["id"], &telegramID)
	log.Info().Int64("telegram_id", telegramID).Str("username", payload["username"]).Msg("creating/updating user")
	if err := h.DB.QueryRow(r.Context(), "INSERT INTO users (telegram_id, username) VALUES ($1,$2) ON CONFLICT (telegram_id) DO UPDATE SET username=EXCLUDED.username RETURNING id", telegramID, payload["username"]).Scan(&id); err != nil {
		log.Error().Err(err).Msg("create user")
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}
	log.Info().Int64("user_id", id).Msg("user created/updated")
	token, err := h.Auth.CreateJWT(telegramID)
	if err != nil {
		log.Error().Err(err).Msg("create jwt")
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}
	log.Info().Str("token", token[:20]+"...").Msg("jwt token created")
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

	// Get all telegram_ids from whitelist to fetch family expenses
	whitelistIDs := make([]int64, 0)
	for _, idStr := range h.Auth.Whitelist {
		if idStr != "*" {
			var tgID int64
			fmt.Sscan(idStr, &tgID)
			whitelistIDs = append(whitelistIDs, tgID)
		}
	}

	// Fetch ALL expenses from whitelist users (family members)
	query := `
		SELECT e.id, e.user_id, e.amount_cents, e.category_id, e.timestamp, e.is_shared, u.username 
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

	type e struct {
		ID          int    `json:"id"`
		UserID      int64  `json:"user_id"`
		AmountCents int    `json:"amount_cents"`
		CategoryID  *int   `json:"category_id"`
		Timestamp   string `json:"timestamp"`
		IsShared    bool   `json:"is_shared"`
		Username    string `json:"username"`
	}
	var res []e
	for rows.Next() {
		var it e
		var ts time.Time
		var username *string
		if err := rows.Scan(&it.ID, &it.UserID, &it.AmountCents, &it.CategoryID, &ts, &it.IsShared, &username); err == nil {
			it.Timestamp = ts.UTC().Format(time.RFC3339)
			if username != nil {
				it.Username = *username
			}
			res = append(res, it)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
	log.Info().Int64("user_id", userID).Int("count", len(res)).Msg("returned family expenses")
}

// GetTotalExpenses returns total expenses for ALL family members with optional period filter
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

// ========== INCOMES HANDLERS ==========

type incomeRequest struct {
	AmountCents   int    `json:"amount_cents"`
	IncomeType    string `json:"income_type"` // salary, debt_return, prize, gift, refund, other
	Description   string `json:"description"`
	RelatedDebtID *int   `json:"related_debt_id"` // optional, for debt_return type
	Timestamp     string `json:"timestamp"`       // RFC3339 optional
}

// AddIncome creates an income for the authenticated user
func (h *Handlers) AddIncome(w http.ResponseWriter, r *http.Request) {
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
func (h *Handlers) GetIncomes(w http.ResponseWriter, r *http.Request) {
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
func (h *Handlers) GetTotalIncomes(w http.ResponseWriter, r *http.Request) {
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

// GetBalance returns family balance (total incomes - total expenses)
func (h *Handlers) GetBalance(w http.ResponseWriter, r *http.Request) {
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

// ========== SUBCATEGORIES CRUD ==========

// CreateSubcategory creates a new subcategory
func (h *Handlers) CreateSubcategory(w http.ResponseWriter, r *http.Request) {
	var req subcategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	if req.CategoryID <= 0 {
		http.Error(w, "valid category_id is required", http.StatusBadRequest)
		return
	}

	// Check if category exists
	var categoryExists bool
	err := h.DB.QueryRow(r.Context(), "SELECT EXISTS(SELECT 1 FROM categories WHERE id = $1)", req.CategoryID).Scan(&categoryExists)
	if err != nil || !categoryExists {
		http.Error(w, "category not found", http.StatusBadRequest)
		return
	}

	// Convert aliases to JSON
	aliasesJSON, err := json.Marshal(req.Aliases)
	if err != nil {
		http.Error(w, "invalid aliases format", http.StatusBadRequest)
		return
	}

	var subcategoryID int
	err = h.DB.QueryRow(r.Context(),
		`INSERT INTO subcategories (name, category_id, aliases) VALUES ($1, $2, $3) RETURNING id`,
		req.Name, req.CategoryID, aliasesJSON).Scan(&subcategoryID)

	if err != nil {
		log.Error().Err(err).Msg("insert subcategory")
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}

	response := subcategoryResponse{
		ID:         subcategoryID,
		Name:       req.Name,
		CategoryID: req.CategoryID,
		Aliases:    req.Aliases,
		CreatedAt:  time.Now().UTC().Format(time.RFC3339),
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
	log.Info().Int("subcategory_id", subcategoryID).Str("name", req.Name).Int("category_id", req.CategoryID).Msg("subcategory created")
}

// GetSubcategories returns all subcategories, optionally filtered by category
func (h *Handlers) GetSubcategories(w http.ResponseWriter, r *http.Request) {
	categoryID := r.URL.Query().Get("category_id")

	var query string
	var args []interface{}

	if categoryID != "" {
		query = `SELECT s.id, s.name, s.category_id, s.aliases, s.created_at, c.name as category_name 
				 FROM subcategories s 
				 JOIN categories c ON s.category_id = c.id 
				 WHERE s.category_id = $1 
				 ORDER BY s.name`
		args = []interface{}{categoryID}
	} else {
		query = `SELECT s.id, s.name, s.category_id, s.aliases, s.created_at, c.name as category_name 
				 FROM subcategories s 
				 JOIN categories c ON s.category_id = c.id 
				 ORDER BY c.name, s.name`
		args = []interface{}{}
	}

	rows, err := h.DB.Query(r.Context(), query, args...)
	if err != nil {
		log.Error().Err(err).Msg("select subcategories")
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type subcategoryWithCategory struct {
		ID           int      `json:"id"`
		Name         string   `json:"name"`
		CategoryID   int      `json:"category_id"`
		CategoryName string   `json:"category_name"`
		Aliases      []string `json:"aliases"`
		CreatedAt    string   `json:"created_at"`
	}

	var subcategories []subcategoryWithCategory
	for rows.Next() {
		var s subcategoryWithCategory
		var aliasesJSON []byte
		var createdAt time.Time

		if err := rows.Scan(&s.ID, &s.Name, &s.CategoryID, &aliasesJSON, &createdAt, &s.CategoryName); err == nil {
			json.Unmarshal(aliasesJSON, &s.Aliases)
			s.CreatedAt = createdAt.UTC().Format(time.RFC3339)
			subcategories = append(subcategories, s)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subcategories)
	log.Info().Int("count", len(subcategories)).Msg("returned subcategories")
}

// UpdateSubcategory updates an existing subcategory
func (h *Handlers) UpdateSubcategory(w http.ResponseWriter, r *http.Request) {
	subcategoryID := chi.URLParam(r, "id")
	if subcategoryID == "" {
		http.Error(w, "subcategory id required", http.StatusBadRequest)
		return
	}

	var req subcategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	if req.CategoryID <= 0 {
		http.Error(w, "valid category_id is required", http.StatusBadRequest)
		return
	}

	// Convert aliases to JSON
	aliasesJSON, err := json.Marshal(req.Aliases)
	if err != nil {
		http.Error(w, "invalid aliases format", http.StatusBadRequest)
		return
	}

	// Check if subcategory exists
	var exists bool
	err = h.DB.QueryRow(r.Context(), "SELECT EXISTS(SELECT 1 FROM subcategories WHERE id = $1)", subcategoryID).Scan(&exists)
	if err != nil || !exists {
		http.Error(w, "subcategory not found", http.StatusNotFound)
		return
	}

	// Check if category exists
	var categoryExists bool
	err = h.DB.QueryRow(r.Context(), "SELECT EXISTS(SELECT 1 FROM categories WHERE id = $1)", req.CategoryID).Scan(&categoryExists)
	if err != nil || !categoryExists {
		http.Error(w, "category not found", http.StatusBadRequest)
		return
	}

	_, err = h.DB.Exec(r.Context(),
		`UPDATE subcategories SET name = $1, category_id = $2, aliases = $3 WHERE id = $4`,
		req.Name, req.CategoryID, aliasesJSON, subcategoryID)

	if err != nil {
		log.Error().Err(err).Msg("update subcategory")
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}

	response := subcategoryResponse{
		ID:         int(subcategoryID[0] - '0'), // Simple conversion for demo
		Name:       req.Name,
		CategoryID: req.CategoryID,
		Aliases:    req.Aliases,
		CreatedAt:  time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	log.Info().Str("subcategory_id", subcategoryID).Str("name", req.Name).Msg("subcategory updated")
}

// DeleteSubcategory deletes a subcategory
func (h *Handlers) DeleteSubcategory(w http.ResponseWriter, r *http.Request) {
	subcategoryID := chi.URLParam(r, "id")
	if subcategoryID == "" {
		http.Error(w, "subcategory id required", http.StatusBadRequest)
		return
	}

	// Check if subcategory exists
	var exists bool
	err := h.DB.QueryRow(r.Context(), "SELECT EXISTS(SELECT 1 FROM subcategories WHERE id = $1)", subcategoryID).Scan(&exists)
	if err != nil || !exists {
		http.Error(w, "subcategory not found", http.StatusNotFound)
		return
	}

	// Check if subcategory is used in expenses
	var usedInExpenses bool
	err = h.DB.QueryRow(r.Context(), "SELECT EXISTS(SELECT 1 FROM expenses WHERE subcategory_id = $1)", subcategoryID).Scan(&usedInExpenses)
	if err == nil && usedInExpenses {
		http.Error(w, "cannot delete subcategory that is used in expenses", http.StatusBadRequest)
		return
	}

	_, err = h.DB.Exec(r.Context(), "DELETE FROM subcategories WHERE id = $1", subcategoryID)
	if err != nil {
		log.Error().Err(err).Msg("delete subcategory")
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	log.Info().Str("subcategory_id", subcategoryID).Msg("subcategory deleted")
}

// ========== TRANSACTIONS ENDPOINT ==========

// GetTransactions returns paginated transactions with filters using keyset pagination
func (h *Handlers) GetTransactions(w http.ResponseWriter, r *http.Request) {
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

	// Get whitelist IDs for family transactions
	whitelistIDs := make([]int64, 0)
	for _, idStr := range h.Auth.Whitelist {
		if idStr != "*" {
			var tgID int64
			fmt.Sscan(idStr, &tgID)
			whitelistIDs = append(whitelistIDs, tgID)
		}
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

	// User filter (whitelist)
	whereConditions = append(whereConditions, fmt.Sprintf("u.telegram_id = ANY($%d)", argIndex))
	args = append(args, whitelistIDs)
	argIndex++

	// Build WHERE clause
	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Add keyset pagination condition
	if cursor != "" {
		if cursorTime, err := time.Parse(time.RFC3339, cursor); err == nil {
			whereConditions = append(whereConditions, fmt.Sprintf("e.timestamp < $%d", argIndex))
			args = append(args, cursorTime)
			argIndex++
		}
	}

	// Update where clause with keyset condition
	whereClause = ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Optimized query with keyset pagination (no COUNT for performance)
	query := fmt.Sprintf(`
		SELECT e.id, e.user_id, e.amount_cents, e.category_id, e.subcategory_id, 
			   e.operation_type, e.timestamp, e.is_shared, u.username,
			   c.name as category_name, s.name as subcategory_name
		FROM expenses e
		LEFT JOIN users u ON e.user_id = u.id
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

// ========== CATEGORY SUGGESTIONS ==========

// GetCategorySuggestions returns smart category suggestions based on query with caching and usage statistics
func (h *Handlers) GetCategorySuggestions(w http.ResponseWriter, r *http.Request) {
	// Get user ID from JWT
	userID, err := h.Auth.GetUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	query := r.URL.Query().Get("query")
	if query == "" {
		http.Error(w, "query parameter is required", http.StatusBadRequest)
		return
	}

	// Check cache first
	if cache, exists := h.SuggestionsCache[int(userID)]; exists && time.Now().Before(cache.ExpiresAt) {
		// Filter cached results by query
		filteredSuggestions := h.filterSuggestionsByQuery(cache.Data, query)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(filteredSuggestions)
		return
	}

	// Generate new suggestions with usage statistics
	suggestions, err := h.generateSmartSuggestions(r.Context(), int(userID), query)
	if err != nil {
		log.Error().Err(err).Msg("failed to generate suggestions")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Cache the results for 1 hour
	h.SuggestionsCache[int(userID)] = suggestionsCache{
		Data:      suggestions,
		ExpiresAt: time.Now().Add(1 * time.Hour),
		UserID:    int(userID),
	}

	// Filter by query and return
	filteredSuggestions := h.filterSuggestionsByQuery(suggestions, query)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(filteredSuggestions)
	log.Info().Str("query", query).Int("count", len(filteredSuggestions)).Msg("returned smart category suggestions")
}

// generateSmartSuggestions generates suggestions with usage statistics and similarity search
func (h *Handlers) generateSmartSuggestions(ctx context.Context, userID int, query string) ([]categorySuggestion, error) {
	var suggestions []categorySuggestion

	// Get category suggestions with usage frequency and similarity
	categoryQuery := `
		WITH user_category_usage AS (
			SELECT 
				c.id,
				c.name,
				COUNT(e.id) as usage_count,
				similarity(c.name, $2) as similarity_score
			FROM categories c
			LEFT JOIN expenses e ON c.id = e.category_id 
				AND e.user_id = $1 
				AND e.timestamp >= NOW() - INTERVAL '30 days'
			WHERE c.name ILIKE '%' || $2 || '%' 
				OR similarity(c.name, $2) > 0.3
			GROUP BY c.id, c.name
		)
		SELECT 
			id, name, usage_count, similarity_score
		FROM user_category_usage
		ORDER BY usage_count DESC, similarity_score DESC
		LIMIT 20
	`

	rows, err := h.DB.Query(ctx, categoryQuery, userID, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name string
		var usage int
		var similarity float64
		if err := rows.Scan(&id, &name, &usage, &similarity); err != nil {
			continue
		}

		// Calculate score based on usage and similarity
		score := float64(usage)*0.7 + similarity*0.3

		suggestions = append(suggestions, categorySuggestion{
			ID:    id,
			Name:  name,
			Type:  "category",
			Score: score,
			Usage: usage,
		})
	}

	// Get subcategory suggestions with usage frequency and similarity
	subcategoryQuery := `
		WITH user_subcategory_usage AS (
			SELECT 
				s.id,
				s.name,
				c.name as category_name,
				COUNT(e.id) as usage_count,
				similarity(s.name, $2) as similarity_score
			FROM subcategories s
			JOIN categories c ON s.category_id = c.id
			LEFT JOIN expenses e ON s.id = e.subcategory_id 
				AND e.user_id = $1 
				AND e.timestamp >= NOW() - INTERVAL '30 days'
			WHERE s.name ILIKE '%' || $2 || '%' 
				OR similarity(s.name, $2) > 0.3
			GROUP BY s.id, s.name, c.name
		)
		SELECT 
			id, name, category_name, usage_count, similarity_score
		FROM user_subcategory_usage
		ORDER BY usage_count DESC, similarity_score DESC
		LIMIT 20
	`

	subRows, err := h.DB.Query(ctx, subcategoryQuery, userID, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query subcategories: %w", err)
	}
	defer subRows.Close()

	for subRows.Next() {
		var id int
		var name string
		var categoryName string
		var usage int
		var similarity float64
		if err := subRows.Scan(&id, &name, &categoryName, &usage, &similarity); err != nil {
			continue
		}

		// Calculate score based on usage and similarity
		score := float64(usage)*0.7 + similarity*0.3

		suggestions = append(suggestions, categorySuggestion{
			ID:    id,
			Name:  name,
			Type:  "subcategory",
			Score: score,
			Usage: usage,
		})
	}

	// Sort by score (highest first)
	sort.Slice(suggestions, func(i, j int) bool {
		return suggestions[i].Score > suggestions[j].Score
	})

	return suggestions, nil
}

// filterSuggestionsByQuery filters cached suggestions by query
func (h *Handlers) filterSuggestionsByQuery(suggestions []categorySuggestion, query string) []categorySuggestion {
	if query == "" {
		return suggestions
	}

	queryLower := strings.ToLower(query)
	var filtered []categorySuggestion

	for _, suggestion := range suggestions {
		if strings.Contains(strings.ToLower(suggestion.Name), queryLower) {
			filtered = append(filtered, suggestion)
		}
	}

	// Limit to 10 suggestions
	if len(filtered) > 10 {
		filtered = filtered[:10]
	}

	return filtered
}
