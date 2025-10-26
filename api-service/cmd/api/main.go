// ...existing code...

// ...existing code...

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/expense-tracker/api-service/internal/auth"
	"github.com/expense-tracker/api-service/internal/handlers"
	"github.com/expense-tracker/api-service/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

func main() {
	// Read env
	dbURL := fmt.Sprintf("postgresql://%s:%s@db:5432/%s",
		os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_DB"))

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to db")
	}
	defer pool.Close()

	// Initialize auth and handlers modules
	a := auth.NewAuth(pool)

	// Initialize new modular handlers
	authHandlers := handlers.NewAuthHandlers(a, pool)
	expenseHandlers := handlers.NewExpenseHandlers(pool, a)
	incomeHandlers := handlers.NewIncomeHandlers(pool, a)
	transactionHandlers := handlers.NewTransactionHandlers(pool, a)
	categoryHandlers := handlers.NewCategoryHandlers(pool)
	debtHandlers := handlers.NewDebtHandlers(pool, a)
	familyHandlers := handlers.NewFamilyHandlers(pool, a)
	internalHandlers := handlers.NewInternalHandlers(pool)

	r := chi.NewRouter()
	// Global middleware
	r.Use(middleware.CORS)
	r.Use(a.RequestLogger)

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "api"})
	})

	// Public categories endpoint (both paths for compatibility)
	r.Get("/categories", categoryHandlers.GetCategories)
	r.Get("/api/categories", categoryHandlers.GetCategories)
	r.Post("/categories/detect", categoryHandlers.DetectCategory)

	// Internal bot endpoints (protected by X-BOT-KEY header)
	r.Post("/internal/expenses", internalHandlers.InternalPostExpense)
	r.Get("/internal/expenses/total", internalHandlers.InternalGetTotalExpenses)
	r.Get("/internal/debts", internalHandlers.InternalGetDebts)
	r.Post("/internal/groups", internalHandlers.InternalRegisterGroup)
	r.Post("/internal/group-members", internalHandlers.InternalRegisterGroupMember)

	// Protected routes with /api prefix
	r.Route("/api", func(r chi.Router) {
		r.Use(a.Middleware)

		// Expenses endpoints
		r.Post("/expenses", expenseHandlers.AddExpense)
		r.Get("/expenses", expenseHandlers.GetExpenses)
		r.Get("/expenses/total", expenseHandlers.GetTotalExpenses)
		r.Post("/expenses/shared", debtHandlers.CreateSharedExpense)

		// Incomes endpoints
		r.Post("/incomes", incomeHandlers.AddIncome)
		r.Get("/incomes", incomeHandlers.GetIncomes)
		r.Get("/incomes/total", incomeHandlers.GetTotalIncomes)

		// Transactions endpoint (unified expenses/incomes)
		r.Get("/transactions", transactionHandlers.GetTransactions)

		// Subcategories CRUD
		r.Post("/subcategories", categoryHandlers.CreateSubcategory)
		r.Get("/subcategories", categoryHandlers.GetSubcategories)
		r.Put("/subcategories/{id}", categoryHandlers.UpdateSubcategory)
		r.Delete("/subcategories/{id}", categoryHandlers.DeleteSubcategory)

		// Category suggestions
		r.Get("/suggestions/categories", categoryHandlers.GetCategorySuggestions)

		// Debts and balance
		r.Get("/debts", debtHandlers.GetDebts)
		r.Get("/balance", debtHandlers.GetBalance)

		// Family/Groups endpoints
		r.Get("/family/groups", familyHandlers.GetFamilyGroups)

		// Analytics endpoints (proxy to analytics-service)
		r.Get("/analytics/health", func(w http.ResponseWriter, r *http.Request) {
			// Proxy to analytics-service
			analyticsURL := os.Getenv("ANALYTICS_URL")
			if analyticsURL == "" {
				analyticsURL = "http://analytics:8081"
			}
			
			resp, err := http.Get(analyticsURL + "/health")
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusServiceUnavailable)
				json.NewEncoder(w).Encode(map[string]string{
					"status": "error",
					"message": "Analytics service unavailable",
				})
				return
			}
			defer resp.Body.Close()
			
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(resp.StatusCode)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "ok",
				"ollama_status": "available",
				"model": "llama2",
			})
		})

		r.Post("/analytics/summary", func(w http.ResponseWriter, r *http.Request) {
			// Proxy to analytics-service
			analyticsURL := os.Getenv("ANALYTICS_URL")
			if analyticsURL == "" {
				analyticsURL = "http://analytics:8081"
			}
			
			// Forward the request
			resp, err := http.Post(analyticsURL + "/summary", "application/json", r.Body)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusServiceUnavailable)
				json.NewEncoder(w).Encode(map[string]string{
					"status": "error",
					"message": "Analytics service unavailable",
				})
				return
			}
			defer resp.Body.Close()
			
			// Copy response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(resp.StatusCode)
			
			var result map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&result); err == nil {
				json.NewEncoder(w).Encode(result)
			} else {
				json.NewEncoder(w).Encode(map[string]string{
					"status": "error",
					"message": "Failed to parse analytics response",
				})
			}
		})
	})

	// Public login endpoint (must be after protected routes to avoid conflicts)
	r.Post("/api/login", func(w http.ResponseWriter, r *http.Request) {
		log.Info().Msg("Route /api/login matched - calling Login handler")
		authHandlers.Login(w, r)
	})

	// Also handle /login for direct access
	r.Post("/login", func(w http.ResponseWriter, r *http.Request) {
		log.Info().Msg("Route /login matched - calling Login handler")
		authHandlers.Login(w, r)
	})

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Info().Msg("api starting on :8080")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal().Err(err).Msg("server failed")
	}
}

// helper functions and legacy handlers were removed; the `internal/auth` and `internal/handlers` packages
// provide auth and request handling. This file intentionally only contains the service bootstrap.
