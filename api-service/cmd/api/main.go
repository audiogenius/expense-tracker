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
	h := handlers.NewHandlers(a, pool)

	r := chi.NewRouter()
	// Global request logging
	r.Use(a.RequestLogger)

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "api"})
	})

	// Public login endpoint
	r.Post("/login", h.Login)

	// Public categories endpoint
	r.Get("/categories", h.GetCategories)
	r.Post("/categories/detect", h.DetectCategory)

	// Internal bot endpoints (protected by X-BOT-KEY header)
	r.Post("/internal/expenses", h.InternalPostExpense)
	r.Get("/internal/expenses/total", h.InternalGetTotalExpenses)
	r.Get("/internal/debts", h.InternalGetDebts)

	// Protected routes
	r.Route("/", func(r chi.Router) {
		r.Use(a.Middleware)

		// Expenses endpoints
		r.Post("/expenses", h.AddExpense)
		r.Get("/expenses", h.GetExpenses)
		r.Get("/expenses/total", h.GetTotalExpenses)
		r.Post("/expenses/shared", h.CreateSharedExpense)

		// Incomes endpoints
		r.Post("/incomes", h.AddIncome)
		r.Get("/incomes", h.GetIncomes)
		r.Get("/incomes/total", h.GetTotalIncomes)

		// Transactions endpoint (unified expenses/incomes)
		r.Get("/transactions", h.GetTransactions)

		// Subcategories CRUD
		r.Post("/subcategories", h.CreateSubcategory)
		r.Get("/subcategories", h.GetSubcategories)
		r.Put("/subcategories/{id}", h.UpdateSubcategory)
		r.Delete("/subcategories/{id}", h.DeleteSubcategory)

		// Category suggestions
		r.Get("/suggestions/categories", h.GetCategorySuggestions)

		// Debts and balance
		r.Get("/debts", h.GetDebts)
		r.Get("/balance", h.GetBalance)
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
