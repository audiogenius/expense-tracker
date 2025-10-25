package analytics

import (
	"context"
	"fmt"
	"math"
	"time"

	"analytics-service/internal/types"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// Engine represents analytics engine
type Engine struct {
	db *pgxpool.Pool
}

// NewEngine creates new analytics engine
func NewEngine(db *pgxpool.Pool) *Engine {
	return &Engine{db: db}
}

// AnalyzePeriod performs comprehensive financial analysis for a period
func (e *Engine) AnalyzePeriod(ctx context.Context, period string, startDate, endDate time.Time) (*types.AnalysisResult, error) {
	log.Info().Str("period", period).Time("start", startDate).Time("end", endDate).Msg("Starting financial analysis")

	// Get current period data
	currentData, err := e.getFinancialData(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get current period data: %w", err)
	}

	// Get previous period data for comparison
	prevStart, prevEnd := e.getPreviousPeriod(startDate, endDate, period)
	previousData, err := e.getFinancialData(ctx, prevStart, prevEnd)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get previous period data, using zero values")
		previousData = &types.FinancialData{
			Period:     "previous",
			StartDate:  prevStart,
			EndDate:    prevEnd,
			Expenses:   0,
			Incomes:    0,
			Balance:    0,
			Categories: make(map[string]float64),
		}
	}

	// Calculate changes
	changes := e.calculateChanges(*currentData, *previousData)

	// Detect anomalies
	anomalies := e.detectAnomalies(*currentData, *previousData)

	// Analyze trends
	trends := e.analyzeTrends(*currentData, *previousData, changes)

	// Generate insights
	insights := e.generateInsights(*currentData, changes, anomalies, trends)

	result := &types.AnalysisResult{
		Period:      period,
		Data:        *currentData,
		Comparison:  types.ComparisonData{Current: *currentData, Previous: *previousData, Change: changes},
		Anomalies:   anomalies,
		Trends:      trends,
		Insights:    insights,
		GeneratedAt: time.Now(),
	}

	log.Info().
		Float64("expenses", currentData.Expenses).
		Float64("incomes", currentData.Incomes).
		Float64("balance", currentData.Balance).
		Int("anomalies", len(anomalies)).
		Int("trends", len(trends)).
		Msg("Analysis completed")

	return result, nil
}

// getFinancialData retrieves financial data for a period
func (e *Engine) getFinancialData(ctx context.Context, startDate, endDate time.Time) (*types.FinancialData, error) {
	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN operation_type = 'expense' THEN amount_cents ELSE 0 END), 0) / 100.0 as expenses,
			COALESCE(SUM(CASE WHEN operation_type = 'income' THEN amount_cents ELSE 0 END), 0) / 100.0 as incomes,
			COALESCE(SUM(CASE WHEN operation_type = 'income' THEN amount_cents ELSE -amount_cents END), 0) / 100.0 as balance
		FROM expenses 
		WHERE timestamp >= $1 AND timestamp <= $2
	`

	var expenses, incomes, balance float64
	err := e.db.QueryRow(ctx, query, startDate, endDate).Scan(&expenses, &incomes, &balance)
	if err != nil {
		return nil, fmt.Errorf("failed to query financial data: %w", err)
	}

	// Get category breakdown
	categories, err := e.getCategoryBreakdown(ctx, startDate, endDate)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get category breakdown")
		categories = make(map[string]float64)
	}

	return &types.FinancialData{
		Period:     fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		StartDate:  startDate,
		EndDate:    endDate,
		Expenses:   expenses,
		Incomes:    incomes,
		Balance:    balance,
		Categories: categories,
	}, nil
}

// getCategoryBreakdown gets spending breakdown by categories
func (e *Engine) getCategoryBreakdown(ctx context.Context, startDate, endDate time.Time) (map[string]float64, error) {
	query := `
		SELECT c.name, COALESCE(SUM(e.amount_cents), 0) / 100.0 as amount
		FROM expenses e
		LEFT JOIN categories c ON e.category_id = c.id
		WHERE e.timestamp >= $1 AND e.timestamp <= $2 
		AND e.operation_type = 'expense'
		GROUP BY c.name
		ORDER BY amount DESC
	`

	rows, err := e.db.Query(ctx, query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query category breakdown: %w", err)
	}
	defer rows.Close()

	categories := make(map[string]float64)
	for rows.Next() {
		var name string
		var amount float64
		if err := rows.Scan(&name, &amount); err != nil {
			continue
		}
		if name != "" {
			categories[name] = amount
		}
	}

	return categories, nil
}

// getPreviousPeriod calculates previous period dates
func (e *Engine) getPreviousPeriod(startDate, endDate time.Time, period string) (time.Time, time.Time) {
	duration := endDate.Sub(startDate)

	switch period {
	case "day":
		return startDate.AddDate(0, 0, -1), endDate.AddDate(0, 0, -1)
	case "week":
		return startDate.AddDate(0, 0, -7), endDate.AddDate(0, 0, -7)
	case "month":
		return startDate.AddDate(0, -1, 0), endDate.AddDate(0, -1, 0)
	default:
		return startDate.Add(-duration), endDate.Add(-duration)
	}
}

// calculateChanges calculates percentage changes between periods
func (e *Engine) calculateChanges(current, previous types.FinancialData) types.ChangeData {
	expensesChange := current.Expenses - previous.Expenses
	incomesChange := current.Incomes - previous.Incomes
	balanceChange := current.Balance - previous.Balance

	var expensesPercent, incomesPercent, balancePercent float64

	if previous.Expenses > 0 {
		expensesPercent = (expensesChange / previous.Expenses) * 100
	}
	if previous.Incomes > 0 {
		incomesPercent = (incomesChange / previous.Incomes) * 100
	}
	if previous.Balance != 0 {
		balancePercent = (balanceChange / math.Abs(previous.Balance)) * 100
	}

	return types.ChangeData{
		ExpensesChange:  expensesChange,
		IncomesChange:   incomesChange,
		BalanceChange:   balanceChange,
		ExpensesPercent: expensesPercent,
		IncomesPercent:  incomesPercent,
		BalancePercent:  balancePercent,
	}
}

// detectAnomalies detects spending anomalies
func (e *Engine) detectAnomalies(current, previous types.FinancialData) []types.AnomalyData {
	var anomalies []types.AnomalyData

	// High spending anomaly
	if previous.Expenses > 0 && current.Expenses > previous.Expenses*2 {
		anomalies = append(anomalies, types.AnomalyData{
			Type:        "high_spending",
			Amount:      current.Expenses,
			Average:     previous.Expenses,
			Multiplier:  current.Expenses / previous.Expenses,
			Description: fmt.Sprintf("Расходы увеличились в %.1f раз", current.Expenses/previous.Expenses),
			Severity:    "high",
		})
	}

	// Low income anomaly
	if previous.Incomes > 0 && current.Incomes < previous.Incomes*0.5 {
		anomalies = append(anomalies, types.AnomalyData{
			Type:        "low_income",
			Amount:      current.Incomes,
			Average:     previous.Incomes,
			Multiplier:  current.Incomes / previous.Incomes,
			Description: fmt.Sprintf("Доходы снизились на %.1f%%", (1-current.Incomes/previous.Incomes)*100),
			Severity:    "medium",
		})
	}

	// Category anomalies
	for category, amount := range current.Categories {
		if prevAmount, exists := previous.Categories[category]; exists && prevAmount > 0 {
			if amount > prevAmount*2 {
				anomalies = append(anomalies, types.AnomalyData{
					Type:        "unusual_category",
					Category:    category,
					Amount:      amount,
					Average:     prevAmount,
					Multiplier:  amount / prevAmount,
					Description: fmt.Sprintf("Увеличение трат на %s в %.1f раз", category, amount/prevAmount),
					Severity:    "medium",
				})
			}
		}
	}

	return anomalies
}

// analyzeTrends analyzes spending trends
func (e *Engine) analyzeTrends(current, previous types.FinancialData, changes types.ChangeData) []types.TrendData {
	var trends []types.TrendData

	// Savings trend
	if changes.BalanceChange > 0 {
		confidence := math.Min(math.Abs(changes.BalancePercent)/50, 1.0)
		trends = append(trends, types.TrendData{
			Type:        "saving",
			Direction:   "up",
			Amount:      changes.BalanceChange,
			Description: fmt.Sprintf("Экономия %.2f руб", changes.BalanceChange),
			Confidence:  confidence,
		})
	}

	// Spending increase trend
	if changes.ExpensesChange > 0 && changes.ExpensesPercent > 20 {
		confidence := math.Min(changes.ExpensesPercent/100, 1.0)
		trends = append(trends, types.TrendData{
			Type:        "spending_increase",
			Direction:   "up",
			Amount:      changes.ExpensesChange,
			Description: fmt.Sprintf("Рост расходов на %.1f%%", changes.ExpensesPercent),
			Confidence:  confidence,
		})
	}

	// Stable trend
	if math.Abs(changes.ExpensesPercent) < 10 && math.Abs(changes.IncomesPercent) < 10 {
		trends = append(trends, types.TrendData{
			Type:        "stable",
			Direction:   "stable",
			Amount:      0,
			Description: "Стабильные траты",
			Confidence:  0.8,
		})
	}

	return trends
}

// generateInsights generates actionable insights
func (e *Engine) generateInsights(data types.FinancialData, changes types.ChangeData, anomalies []types.AnomalyData, trends []types.TrendData) []string {
	var insights []string

	// Balance insights
	if data.Balance > 0 {
		insights = append(insights, "💰 Отличная работа! У вас положительный баланс")
	} else if data.Balance < 0 {
		insights = append(insights, "⚠️ Внимание: отрицательный баланс. Рекомендуем сократить расходы")
	}

	// Spending insights
	if changes.ExpensesPercent > 20 {
		insights = append(insights, "📈 Расходы выросли. Проверьте категории с наибольшими тратами")
	} else if changes.ExpensesPercent < -10 {
		insights = append(insights, "🎉 Отлично! Расходы снизились. Продолжайте в том же духе")
	}

	// Income insights
	if changes.IncomesPercent > 20 {
		insights = append(insights, "🚀 Доходы выросли! Отличная работа")
	} else if changes.IncomesPercent < -20 {
		insights = append(insights, "💡 Доходы снизились. Рассмотрите дополнительные источники дохода")
	}

	// Trend insights
	for _, trend := range trends {
		if trend.Type == "saving" && trend.Confidence > 0.7 {
			insights = append(insights, "🏆 Вы формируете хорошие привычки экономии")
		}
	}

	// Anomaly insights
	for _, anomaly := range anomalies {
		if anomaly.Severity == "high" {
			insights = append(insights, fmt.Sprintf("🚨 %s", anomaly.Description))
		}
	}

	return insights
}
