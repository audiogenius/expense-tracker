package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/expense-tracker/api-service/internal/auth"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// CategoryHandlers handles all category and subcategory-related endpoints
type CategoryHandlers struct {
	DB               *pgxpool.Pool
	SuggestionsCache map[int]suggestionsCache // User ID -> Cache
}

// NewCategoryHandlers creates a new CategoryHandlers instance
func NewCategoryHandlers(db *pgxpool.Pool) *CategoryHandlers {
	return &CategoryHandlers{
		DB:               db,
		SuggestionsCache: make(map[int]suggestionsCache),
	}
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

// GetCategories returns all available categories
func (h *CategoryHandlers) GetCategories(w http.ResponseWriter, r *http.Request) {
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
func (h *CategoryHandlers) DetectCategory(w http.ResponseWriter, r *http.Request) {
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

// CreateSubcategory creates a new subcategory
func (h *CategoryHandlers) CreateSubcategory(w http.ResponseWriter, r *http.Request) {
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
func (h *CategoryHandlers) GetSubcategories(w http.ResponseWriter, r *http.Request) {
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

// GetCategorySuggestions returns smart category suggestions based on query with caching and usage statistics
func (h *CategoryHandlers) GetCategorySuggestions(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
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
func (h *CategoryHandlers) generateSmartSuggestions(ctx context.Context, userID int, query string) ([]categorySuggestion, error) {
	var suggestions []categorySuggestion

	// Get category suggestions with usage frequency
	categoryQuery := `
		WITH user_category_usage AS (
			SELECT 
				c.id,
				c.name,
				COUNT(e.id) as usage_count,
				CASE 
					WHEN LOWER(c.name) LIKE LOWER($2 || '%') THEN 1.0
					WHEN LOWER(c.name) LIKE LOWER('%' || $2 || '%') THEN 0.5
					ELSE 0.3
				END as similarity_score
			FROM categories c
			LEFT JOIN expenses e ON c.id = e.category_id 
				AND e.user_id = $1 
				AND e.timestamp >= NOW() - INTERVAL '30 days'
			WHERE c.name ILIKE '%' || $2 || '%'
			GROUP BY c.id, c.name
		)
		SELECT 
			id, name, usage_count, similarity_score
		FROM user_category_usage
		ORDER BY similarity_score DESC, usage_count DESC
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

	// Get subcategory suggestions with usage frequency
	subcategoryQuery := `
		WITH user_subcategory_usage AS (
			SELECT 
				s.id,
				s.name,
				c.name as category_name,
				COUNT(e.id) as usage_count,
				CASE 
					WHEN LOWER(s.name) LIKE LOWER($2 || '%') THEN 1.0
					WHEN LOWER(s.name) LIKE LOWER('%' || $2 || '%') THEN 0.5
					ELSE 0.3
				END as similarity_score
			FROM subcategories s
			JOIN categories c ON s.category_id = c.id
			LEFT JOIN expenses e ON s.id = e.subcategory_id 
				AND e.user_id = $1 
				AND e.timestamp >= NOW() - INTERVAL '30 days'
			WHERE s.name ILIKE '%' || $2 || '%'
			GROUP BY s.id, s.name, c.name
		)
		SELECT 
			id, name, category_name, usage_count, similarity_score
		FROM user_subcategory_usage
		ORDER BY similarity_score DESC, usage_count DESC
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
func (h *CategoryHandlers) filterSuggestionsByQuery(suggestions []categorySuggestion, query string) []categorySuggestion {
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

// UpdateSubcategory updates an existing subcategory
func (h *CategoryHandlers) UpdateSubcategory(w http.ResponseWriter, r *http.Request) {
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
func (h *CategoryHandlers) DeleteSubcategory(w http.ResponseWriter, r *http.Request) {
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
