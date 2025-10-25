package types

import "time"

// FinancialData represents aggregated financial data for analysis
type FinancialData struct {
	Period     string             `json:"period"`
	StartDate  time.Time          `json:"start_date"`
	EndDate    time.Time          `json:"end_date"`
	Expenses   float64            `json:"expenses"`
	Incomes    float64            `json:"incomes"`
	Balance    float64            `json:"balance"`
	Categories map[string]float64 `json:"categories"`
}

// ComparisonData represents comparison between periods
type ComparisonData struct {
	Current  FinancialData `json:"current"`
	Previous FinancialData `json:"previous"`
	Change   ChangeData    `json:"change"`
}

// ChangeData represents changes between periods
type ChangeData struct {
	ExpensesChange  float64 `json:"expenses_change"`
	IncomesChange   float64 `json:"incomes_change"`
	BalanceChange   float64 `json:"balance_change"`
	ExpensesPercent float64 `json:"expenses_percent"`
	IncomesPercent  float64 `json:"incomes_percent"`
	BalancePercent  float64 `json:"balance_percent"`
}

// AnomalyData represents detected spending anomalies
type AnomalyData struct {
	Type        string  `json:"type"` // "high_spending", "unusual_category", "low_income"
	Category    string  `json:"category"`
	Amount      float64 `json:"amount"`
	Average     float64 `json:"average"`
	Multiplier  float64 `json:"multiplier"`
	Description string  `json:"description"`
	Severity    string  `json:"severity"` // "low", "medium", "high"
}

// TrendData represents spending trends
type TrendData struct {
	Type        string  `json:"type"`      // "saving", "spending_increase", "stable"
	Direction   string  `json:"direction"` // "up", "down", "stable"
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"` // 0.0 to 1.0
}

// AnalysisResult represents complete analysis result
type AnalysisResult struct {
	Period      string         `json:"period"`
	Data        FinancialData  `json:"data"`
	Comparison  ComparisonData `json:"comparison"`
	Anomalies   []AnomalyData  `json:"anomalies"`
	Trends      []TrendData    `json:"trends"`
	Insights    []string       `json:"insights"`
	GeneratedAt time.Time      `json:"generated_at"`
}

// MessageTemplate represents a message template for Telegram
type MessageTemplate struct {
	Type       string   `json:"type"` // "daily_report", "anomaly_alert", "trend_notification"
	Title      string   `json:"title"`
	Message    string   `json:"message"`
	Emoji      string   `json:"emoji"`
	Priority   string   `json:"priority"`   // "low", "medium", "high"
	Conditions []string `json:"conditions"` // Conditions when to send
}

// OllamaRequest represents request to Ollama API
type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// OllamaResponse represents response from Ollama API
type OllamaResponse struct {
	Model    string `json:"model"`
	Response string `json:"response"`
	Done     bool   `json:"done"`
	Error    string `json:"error,omitempty"`
}

// TelegramMessage represents message to be sent to Telegram
type TelegramMessage struct {
	ChatID    int64  `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
}

// HealthStatus represents service health status
type HealthStatus struct {
	Service   string    `json:"service"`
	Status    string    `json:"status"` // "healthy", "unhealthy", "degraded"
	Ollama    bool      `json:"ollama"`
	Database  bool      `json:"database"`
	LastCheck time.Time `json:"last_check"`
	Uptime    string    `json:"uptime"`
	Version   string    `json:"version"`
}

// ScheduledJob represents a scheduled job
type ScheduledJob struct {
	ID       int       `json:"id"`
	Next     time.Time `json:"next"`
	Prev     time.Time `json:"prev"`
	Schedule string    `json:"schedule"`
}
