package messaging

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"analytics-service/internal/types"

	"github.com/rs/zerolog/log"
)

// Generator represents message generator
type Generator struct {
	telegramToken string
	telegramURL   string
	httpClient    *http.Client
}

// NewGenerator creates new message generator
func NewGenerator(telegramToken string) *Generator {
	return &Generator{
		telegramToken: telegramToken,
		telegramURL:   "https://api.telegram.org/bot" + telegramToken,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GenerateDailyReport generates and sends daily report
func (g *Generator) GenerateDailyReport(ctx context.Context, analysis *types.AnalysisResult, chatIDs []int64) error {
	message := g.buildDailyReportMessage(analysis)

	for _, chatID := range chatIDs {
		if err := g.sendMessage(ctx, chatID, message); err != nil {
			log.Error().Err(err).Int64("chat_id", chatID).Msg("Failed to send daily report")
			continue
		}
		log.Info().Int64("chat_id", chatID).Msg("Daily report sent successfully")
	}

	return nil
}

// GenerateAnomalyAlert generates and sends anomaly alert
func (g *Generator) GenerateAnomalyAlert(ctx context.Context, analysis *types.AnalysisResult, chatIDs []int64) error {
	message := g.buildAnomalyAlertMessage(analysis)

	for _, chatID := range chatIDs {
		if err := g.sendMessage(ctx, chatID, message); err != nil {
			log.Error().Err(err).Int64("chat_id", chatID).Msg("Failed to send anomaly alert")
			continue
		}
		log.Info().Int64("chat_id", chatID).Msg("Anomaly alert sent successfully")
	}

	return nil
}

// GenerateTrendNotification generates and sends trend notification
func (g *Generator) GenerateTrendNotification(ctx context.Context, analysis *types.AnalysisResult, chatIDs []int64) error {
	message := g.buildTrendNotificationMessage(analysis)

	for _, chatID := range chatIDs {
		if err := g.sendMessage(ctx, chatID, message); err != nil {
			log.Error().Err(err).Int64("chat_id", chatID).Msg("Failed to send trend notification")
			continue
		}
		log.Info().Int64("chat_id", chatID).Msg("Trend notification sent successfully")
	}

	return nil
}

// buildDailyReportMessage builds daily report message
func (g *Generator) buildDailyReportMessage(analysis *types.AnalysisResult) string {
	var message strings.Builder

	// Header
	message.WriteString("📊 *Ежедневный финансовый отчет*\n\n")

	// Main stats
	message.WriteString(fmt.Sprintf("💰 *Баланс:* %.2f ₽\n", analysis.Data.Balance))
	message.WriteString(fmt.Sprintf("📉 *Расходы:* %.2f ₽\n", analysis.Data.Expenses))
	message.WriteString(fmt.Sprintf("📈 *Доходы:* %.2f ₽\n\n", analysis.Data.Incomes))

	// Changes
	if analysis.Comparison.Change.ExpensesPercent != 0 {
		direction := "📈"
		if analysis.Comparison.Change.ExpensesPercent < 0 {
			direction = "📉"
		}
		message.WriteString(fmt.Sprintf("%s *Расходы:* %.1f%% (%s)\n",
			direction,
			math.Abs(analysis.Comparison.Change.ExpensesPercent),
			g.getChangeDescription(analysis.Comparison.Change.ExpensesChange)))
	}

	if analysis.Comparison.Change.IncomesPercent != 0 {
		direction := "📈"
		if analysis.Comparison.Change.IncomesPercent < 0 {
			direction = "📉"
		}
		message.WriteString(fmt.Sprintf("%s *Доходы:* %.1f%% (%s)\n",
			direction,
			math.Abs(analysis.Comparison.Change.IncomesPercent),
			g.getChangeDescription(analysis.Comparison.Change.IncomesChange)))
	}

	// Insights
	if len(analysis.Insights) > 0 {
		message.WriteString("\n💡 *Рекомендации:*\n")
		for _, insight := range analysis.Insights {
			message.WriteString(fmt.Sprintf("• %s\n", insight))
		}
	}

	// Top categories
	if len(analysis.Data.Categories) > 0 {
		message.WriteString("\n🏷️ *Топ категории:*\n")
		count := 0
		for category, amount := range analysis.Data.Categories {
			if count >= 3 { // Show only top 3
				break
			}
			message.WriteString(fmt.Sprintf("• %s: %.2f ₽\n", category, amount))
			count++
		}
	}

	// Footer
	message.WriteString(fmt.Sprintf("\n⏰ %s", time.Now().Format("15:04, 2 января 2006")))

	return message.String()
}

// buildAnomalyAlertMessage builds anomaly alert message
func (g *Generator) buildAnomalyAlertMessage(analysis *types.AnalysisResult) string {
	var message strings.Builder

	message.WriteString("🚨 *Обнаружены аномалии в тратах*\n\n")

	for _, anomaly := range analysis.Anomalies {
		emoji := "⚠️"
		if anomaly.Severity == "high" {
			emoji = "🚨"
		}

		message.WriteString(fmt.Sprintf("%s *%s*\n", emoji, anomaly.Description))
		message.WriteString(fmt.Sprintf("Сумма: %.2f ₽ (среднее: %.2f ₽)\n\n", anomaly.Amount, anomaly.Average))
	}

	message.WriteString("💡 *Рекомендации:*\n")
	message.WriteString("• Проверьте недавние траты\n")
	message.WriteString("• Убедитесь в корректности записей\n")
	message.WriteString("• Рассмотрите возможность оптимизации расходов\n")

	return message.String()
}

// buildTrendNotificationMessage builds trend notification message
func (g *Generator) buildTrendNotificationMessage(analysis *types.AnalysisResult) string {
	var message strings.Builder

	message.WriteString("📈 *Анализ трендов*\n\n")

	for _, trend := range analysis.Trends {
		emoji := "📊"
		switch trend.Type {
		case "saving":
			emoji = "💰"
		case "spending_increase":
			emoji = "📈"
		case "stable":
			emoji = "➡️"
		}

		message.WriteString(fmt.Sprintf("%s *%s*\n", emoji, trend.Description))
		if trend.Amount != 0 {
			message.WriteString(fmt.Sprintf("Изменение: %.2f ₽\n", trend.Amount))
		}
		message.WriteString(fmt.Sprintf("Уверенность: %.0f%%\n\n", trend.Confidence*100))
	}

	return message.String()
}

// sendMessage sends message to Telegram
func (g *Generator) sendMessage(ctx context.Context, chatID int64, text string) error {
	message := types.TelegramMessage{
		ChatID:    chatID,
		Text:      text,
		ParseMode: "Markdown",
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", g.telegramURL+"/sendMessage", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API error: %d", resp.StatusCode)
	}

	return nil
}

// getChangeDescription returns human-readable change description
func (g *Generator) getChangeDescription(change float64) string {
	if change > 0 {
		return "увеличение"
	} else if change < 0 {
		return "снижение"
	}
	return "без изменений"
}

// GetChatIDs retrieves chat IDs from database (placeholder)
func (g *Generator) GetChatIDs(ctx context.Context, db interface{}) ([]int64, error) {
	// This would typically query the database for active chat IDs
	// For now, return empty slice - should be implemented based on your user management
	return []int64{}, nil
}
