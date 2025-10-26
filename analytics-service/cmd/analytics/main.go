package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"analytics-service/internal/analytics"
	"analytics-service/internal/handlers"
	"analytics-service/internal/messaging"
	"analytics-service/internal/ollama"
	"analytics-service/internal/scheduler"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	zerologlog "github.com/rs/zerolog/log"
)

func main() {
	// Configure logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerologlog.Logger = zerologlog.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	zerologlog.Info().Msg("Starting Analytics Service")

	// Load configuration from environment
	config := loadConfig()

	// Initialize database connection
	db, err := initDatabase(config.DatabaseURL)
	if err != nil {
		zerologlog.Fatal().Err(err).Msg("Failed to initialize database")
	}
	defer db.Close()

	// Initialize Ollama client
	var ollamaClient *ollama.Client
	if config.OllamaURL != "" {
		ollamaClient = ollama.NewClient(config.OllamaURL, config.OllamaModel)
		zerologlog.Info().Str("url", config.OllamaURL).Str("model", config.OllamaModel).Msg("Ollama client initialized")
	} else {
		zerologlog.Warn().Msg("Ollama not configured, using fallback analytics")
	}

	// Initialize components
	analyticsEngine := analytics.NewEngine(db)
	messagingGenerator := messaging.NewGenerator(config.TelegramToken)

	// Initialize scheduler
	scheduler := scheduler.NewScheduler(db, analyticsEngine, messagingGenerator, ollamaClient, config.ChatIDs)

	// Initialize handlers
	handlers := handlers.NewHandlers(analyticsEngine, messagingGenerator, ollamaClient, scheduler, db)

	// Setup HTTP router
	router := setupRouter(handlers)

	// Start scheduler
	if err := scheduler.Start(context.Background()); err != nil {
		zerologlog.Fatal().Err(err).Msg("Failed to start scheduler")
	}
	defer scheduler.Stop()

	// Start HTTP server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", config.Port),
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		zerologlog.Info().Str("port", config.Port).Msg("Starting HTTP server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zerologlog.Fatal().Err(err).Msg("Failed to start HTTP server")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zerologlog.Info().Msg("Shutting down Analytics Service")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		zerologlog.Error().Err(err).Msg("Failed to shutdown HTTP server")
	}

	zerologlog.Info().Msg("Analytics Service stopped")
}

// Config represents service configuration
type Config struct {
	Port          string
	DatabaseURL   string
	TelegramToken string
	OllamaURL     string
	OllamaModel   string
	ChatIDs       []int64
}

// loadConfig loads configuration from environment variables
func loadConfig() Config {
	config := Config{
		Port:          getEnv("ANALYTICS_PORT", "8081"),
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5432/expense_tracker?sslmode=disable"),
		TelegramToken: getEnv("TELEGRAM_BOT_TOKEN", ""),
		OllamaURL:     getEnv("OLLAMA_URL", "http://ollama:11434"),
		OllamaModel:   getEnv("OLLAMA_MODEL", "qwen2.5:0.5b"),
		ChatIDs:       []int64{}, // Should be loaded from database or config
	}

	// Parse chat IDs from environment (comma-separated)
	if chatIDsStr := getEnv("TELEGRAM_CHAT_IDS", ""); chatIDsStr != "" {
		// This is a simplified version - in production, you'd parse the string properly
		zerologlog.Info().Str("chat_ids", chatIDsStr).Msg("Chat IDs configured")
	}

	return config
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// initDatabase initializes database connection
func initDatabase(databaseURL string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	zerologlog.Info().Msg("Database connection established")
	return pool, nil
}

// setupRouter sets up HTTP router
func setupRouter(handlers *handlers.Handlers) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Health check
	r.Get("/health", handlers.HealthCheck)

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Analysis endpoints
		r.Get("/analyze", handlers.AnalyzePeriod)
		r.Post("/analyze/trigger", handlers.TriggerAnalysis)

		// Messaging endpoints
		r.Post("/messages/send", handlers.SendMessage)

		// Scheduler endpoints
		r.Get("/scheduler/jobs", handlers.GetScheduledJobs)

		// Ollama endpoints
		r.Get("/ollama/status", handlers.GetOllamaStatus)
	})

	// Direct summary endpoint (for bot compatibility)
	r.Post("/summary", handlers.GetSummary)

	return r
}
