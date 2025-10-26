package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"analytics-service/internal/analytics"
	"analytics-service/internal/messaging"
	"analytics-service/internal/ollama"
	"analytics-service/internal/scheduler"
	"analytics-service/internal/types"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// Handlers represents HTTP handlers
type Handlers struct {
	analytics *analytics.Engine
	messaging *messaging.Generator
	ollama    *ollama.Client
	scheduler *scheduler.Scheduler
	db        *pgxpool.Pool
	startTime time.Time
}

// NewHandlers creates new handlers
func NewHandlers(
	analytics *analytics.Engine,
	messaging *messaging.Generator,
	ollama *ollama.Client,
	scheduler *scheduler.Scheduler,
	db *pgxpool.Pool,
) *Handlers {
	return &Handlers{
		analytics: analytics,
		messaging: messaging,
		ollama:    ollama,
		scheduler: scheduler,
		db:        db,
		startTime: time.Now(),
	}
}

// HealthCheck handles health check endpoint
func (h *Handlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Check Ollama
	ollamaHealthy := false
	if h.ollama != nil {
		if err := h.ollama.HealthCheck(); err == nil {
			ollamaHealthy = true
		}
	}

	// Check database
	dbHealthy := false
	if err := h.db.Ping(ctx); err == nil {
		dbHealthy = true
	}

	// Determine overall status
	status := "healthy"
	if !dbHealthy {
		status = "unhealthy"
	} else if !ollamaHealthy {
		status = "degraded"
	}

	health := types.HealthStatus{
		Service:   "analytics-service",
		Status:    status,
		Ollama:    ollamaHealthy,
		Database:  dbHealthy,
		LastCheck: time.Now(),
		Uptime:    time.Since(h.startTime).String(),
		Version:   "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(health)
}

// AnalyzePeriod handles period analysis endpoint
func (h *Handlers) AnalyzePeriod(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "day"
	}

	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	if days == 0 {
		days = 1
	}

	// Calculate date range
	now := time.Now()
	endDate := now
	startDate := now.AddDate(0, 0, -days)

	// Perform analysis
	analysis, err := h.analytics.AnalyzePeriod(ctx, period, startDate, endDate)
	if err != nil {
		log.Error().Err(err).Msg("Failed to analyze period")
		http.Error(w, "Failed to analyze period", http.StatusInternalServerError)
		return
	}

	// Try to enhance with AI if available
	if h.ollama != nil {
		if err := h.ollama.HealthCheck(); err == nil {
			aiMessage, err := h.ollama.GenerateFinancialInsight(ctx, *analysis)
			if err != nil {
				log.Warn().Err(err).Msg("Failed to generate AI insights, using fallback")
			} else {
				analysis.Insights = append(analysis.Insights, aiMessage)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(analysis)
}

// SendMessage handles sending messages to Telegram
func (h *Handlers) SendMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		ChatIDs []int64 `json:"chat_ids"`
		Type    string  `json:"type"` // "daily", "anomaly", "trend"
		Period  string  `json:"period,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.ChatIDs) == 0 {
		http.Error(w, "Chat IDs required", http.StatusBadRequest)
		return
	}

	// Calculate date range
	now := time.Now()
	var startDate, endDate time.Time

	switch req.Period {
	case "week":
		startDate = now.AddDate(0, 0, -7)
		endDate = now
	case "month":
		startDate = now.AddDate(0, -1, 0)
		endDate = now
	default:
		startDate = now.AddDate(0, 0, -1)
		endDate = now
	}

	// Perform analysis
	analysis, err := h.analytics.AnalyzePeriod(ctx, req.Period, startDate, endDate)
	if err != nil {
		log.Error().Err(err).Msg("Failed to analyze period for message")
		http.Error(w, "Failed to analyze period", http.StatusInternalServerError)
		return
	}

	// Send appropriate message type
	var sendErr error
	switch req.Type {
	case "daily":
		sendErr = h.messaging.GenerateDailyReport(ctx, analysis, req.ChatIDs)
	case "anomaly":
		sendErr = h.messaging.GenerateAnomalyAlert(ctx, analysis, req.ChatIDs)
	case "trend":
		sendErr = h.messaging.GenerateTrendNotification(ctx, analysis, req.ChatIDs)
	default:
		http.Error(w, "Invalid message type", http.StatusBadRequest)
		return
	}

	if sendErr != nil {
		log.Error().Err(sendErr).Msg("Failed to send message")
		http.Error(w, "Failed to send message", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "sent"})
}

// GetScheduledJobs returns list of scheduled jobs
func (h *Handlers) GetScheduledJobs(w http.ResponseWriter, r *http.Request) {
	jobs := h.scheduler.GetScheduledJobs()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(jobs)
}

// TriggerAnalysis manually triggers analysis
func (h *Handlers) TriggerAnalysis(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "day"
	}

	// Calculate date range
	now := time.Now()
	var startDate, endDate time.Time

	switch period {
	case "week":
		startDate = now.AddDate(0, 0, -7)
		endDate = now
	case "month":
		startDate = now.AddDate(0, -1, 0)
		endDate = now
	default:
		startDate = now.AddDate(0, 0, -1)
		endDate = now
	}

	// Perform analysis
	analysis, err := h.analytics.AnalyzePeriod(ctx, period, startDate, endDate)
	if err != nil {
		log.Error().Err(err).Msg("Failed to trigger analysis")
		http.Error(w, "Failed to perform analysis", http.StatusInternalServerError)
		return
	}

	// Try to enhance with AI if available
	if h.ollama != nil {
		if err := h.ollama.HealthCheck(); err == nil {
			aiMessage, err := h.ollama.GenerateFinancialInsight(ctx, *analysis)
			if err != nil {
				log.Warn().Err(err).Msg("Failed to generate AI insights")
			} else {
				analysis.Insights = append(analysis.Insights, aiMessage)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(analysis)
}

// GetOllamaStatus returns Ollama service status
func (h *Handlers) GetOllamaStatus(w http.ResponseWriter, r *http.Request) {

	if h.ollama == nil {
		http.Error(w, "Ollama not configured", http.StatusServiceUnavailable)
		return
	}

	err := h.ollama.HealthCheck()
	if err != nil {
		log.Error().Err(err).Msg("Ollama health check failed")
		http.Error(w, "Ollama health check failed", http.StatusInternalServerError)
		return
	}

	status := map[string]interface{}{
		"healthy": true,
		"model":   "qwen2.5:0.5b",
		"url":     "http://ollama:11434",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(status)
}

// GetSummary generates AI summary for a user's expenses
func (h *Handlers) GetSummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		TelegramID int64  `json:"telegram_id"`
		Period     string `json:"period"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.TelegramID == 0 {
		http.Error(w, "telegram_id required", http.StatusBadRequest)
		return
	}

	if req.Period == "" {
		req.Period = "day"
	}

	// Calculate date range
	now := time.Now()
	var startDate, endDate time.Time

	switch req.Period {
	case "week":
		startDate = now.AddDate(0, 0, -7)
		endDate = now
	case "month":
		startDate = now.AddDate(0, -1, 0)
		endDate = now
	default:
		startDate = now.AddDate(0, 0, -1)
		endDate = now
	}

	// Perform analysis
	analysis, err := h.analytics.AnalyzePeriod(ctx, req.Period, startDate, endDate)
	if err != nil {
		log.Error().Err(err).Msg("Failed to analyze period for summary")
		http.Error(w, "Failed to analyze period", http.StatusInternalServerError)
		return
	}

	// Generate AI summary if available
	summary := "Анализ за период:\n\n"
	if analysis.TotalExpenses > 0 {
		summary += "💸 Расходы: " + strconv.FormatFloat(analysis.TotalExpenses/100, 'f', 2, 64) + " руб.\n"
	}
	if analysis.TotalIncome > 0 {
		summary += "💰 Доходы: " + strconv.FormatFloat(analysis.TotalIncome/100, 'f', 2, 64) + " руб.\n"
	}
	balance := analysis.TotalIncome - analysis.TotalExpenses
	if balance >= 0 {
		summary += "✅ Баланс: +" + strconv.FormatFloat(balance/100, 'f', 2, 64) + " руб.\n"
	} else {
		summary += "⚠️ Баланс: " + strconv.FormatFloat(balance/100, 'f', 2, 64) + " руб.\n"
	}

	// Try to enhance with AI if available
	if h.ollama != nil {
		if err := h.ollama.HealthCheck(); err == nil {
			// Use context-aware generation for better memory
			aiMessage, err := h.ollama.GenerateWithContext(ctx, h.buildFinancialPrompt(*analysis))
			if err != nil {
				log.Warn().Err(err).Msg("Failed to generate AI insights, using fallback")
			} else {
				summary += "\n🤖 AI Инсайты:\n" + aiMessage
			}
		}
	}

	response := map[string]interface{}{
		"summary":  summary,
		"period":   req.Period,
		"analysis": analysis,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// buildFinancialPrompt builds a financial analysis prompt
func (h *Handlers) buildFinancialPrompt(analysis types.AnalysisResult) string {
	return fmt.Sprintf(`Проанализируй финансовые данные и дай краткие рекомендации на русском языке:

Период: %s
Расходы: %.2f руб
Доходы: %.2f руб
Баланс: %.2f руб

Изменения:
- Расходы: %.1f%% (%s)
- Доходы: %.1f%% (%s)
- Баланс: %.1f%% (%s)

Аномалии: %d
Тренды: %d

Дай 2-3 кратких совета по управлению финансами. Будь позитивным и мотивирующим.`,
		analysis.Period,
		analysis.TotalExpenses/100,
		analysis.TotalIncome/100,
		(analysis.TotalIncome-analysis.TotalExpenses)/100,
		analysis.Comparison.Change.ExpensesPercent,
		getChangeDirection(analysis.Comparison.Change.ExpensesChange),
		analysis.Comparison.Change.IncomesPercent,
		getChangeDirection(analysis.Comparison.Change.IncomesChange),
		analysis.Comparison.Change.BalancePercent,
		getChangeDirection(analysis.Comparison.Change.BalanceChange),
		len(analysis.Anomalies),
		len(analysis.Trends))
}

// getChangeDirection returns direction emoji for change
func getChangeDirection(change float64) string {
	if change > 0 {
		return "📈"
	} else if change < 0 {
		return "📉"
	}
	return "➡️"
}
