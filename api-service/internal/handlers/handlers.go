package handlers

import (
	"net/http"

	"github.com/expense-tracker/api-service/internal/auth"
	"github.com/expense-tracker/api-service/internal/cache"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Handlers holds dependencies for HTTP handlers
type Handlers struct {
	Auth        *auth.Auth
	AuthHandler *AuthHandler
	DB          *pgxpool.Pool
	Cache       *cache.MemoryCache
}

func NewHandlers(a *auth.Auth, db *pgxpool.Pool) *Handlers {
	authHandler := NewAuthHandler(a, db)
	return &Handlers{
		Auth:        a,
		AuthHandler: authHandler,
		DB:          db,
		Cache:       cache.NewMemoryCache(),
	}
}

// expenseRequest moved to expense_handlers.go

// Types moved to individual handler files to avoid duplication

// Login delegates to AuthHandler for better separation of concerns
func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	h.AuthHandler.Login(w, r)
}

// AddExpense moved to expense_handlers.go

// GetExpenses moved to expense_handlers.go (keeping for compatibility)

// GetTotalExpenses moved to expense_handlers.go

// GetCategories moved to category_handlers.go

// DetectCategory moved to category_handlers.go

// CreateSharedExpense moved to debt_handlers.go

// GetDebts moved to debt_handlers.go

// InternalPostExpense moved to internal_handlers.go

// InternalGetTotalExpenses moved to internal_handlers.go

// InternalGetDebts moved to internal_handlers.go

// ========== INCOMES HANDLERS ==========
// All income functions moved to income_handlers.go

// GetIncomes moved to income_handlers.go

// GetTotalIncomes moved to income_handlers.go

// GetBalance moved to debt_handlers.go

// ========== SUBCATEGORIES CRUD ==========
// All subcategory functions moved to category_handlers.go

// GetSubcategories moved to category_handlers.go

// UpdateSubcategory moved to category_handlers.go

// DeleteSubcategory moved to category_handlers.go

// ========== TRANSACTIONS ENDPOINT ==========
// GetTransactions moved to transaction_handlers.go

// ========== CATEGORY SUGGESTIONS ==========
// GetCategorySuggestions moved to category_handlers.go

// generateSmartSuggestions moved to category_handlers.go

// filterSuggestionsByQuery moved to category_handlers.go
