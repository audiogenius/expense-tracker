package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"analytics-service/internal/types"
)

// Client represents Ollama API client
type Client struct {
	baseURL    string
	httpClient *http.Client
	model      string
	timeout    time.Duration
	context    []string // Store conversation context for memory
}

// NewClient creates new Ollama client
func NewClient(baseURL, model string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		model:   model,
		timeout: 30 * time.Second,
		context: make([]string, 0),
	}
}

// HealthCheck checks if Ollama service is healthy
func (c *Client) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/tags", nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("health check request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status: %d", resp.StatusCode)
	}

	// Check if our model is available
	var tagsResponse struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tagsResponse); err != nil {
		return fmt.Errorf("failed to decode health check response: %w", err)
	}

	// Check if our model is loaded
	modelFound := false
	for _, model := range tagsResponse.Models {
		if model.Name == c.model {
			modelFound = true
			break
		}
	}

	if !modelFound {
		return fmt.Errorf("model %s not found in Ollama", c.model)
	}

	return nil
}

// GenerateText generates text using Ollama API
func (c *Client) GenerateText(ctx context.Context, prompt string) (string, error) {
	req := types.OllamaRequest{
		Model:  c.model,
		Prompt: prompt,
		Stream: false,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/generate", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama API error: %d - %s", resp.StatusCode, string(body))
	}

	var ollamaResp types.OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if ollamaResp.Error != "" {
		return "", fmt.Errorf("ollama error: %s", ollamaResp.Error)
	}

	return ollamaResp.Response, nil
}

// GenerateFinancialInsight generates financial insight using AI
func (c *Client) GenerateFinancialInsight(ctx context.Context, data types.AnalysisResult) (string, error) {
	prompt := c.buildFinancialPrompt(data)
	return c.GenerateText(ctx, prompt)
}

// GenerateDailyReport generates daily report message
func (c *Client) GenerateDailyReport(ctx context.Context, data types.AnalysisResult) (string, error) {
	prompt := c.buildDailyReportPrompt(data)
	return c.GenerateText(ctx, prompt)
}

// buildFinancialPrompt builds prompt for financial analysis
func (c *Client) buildFinancialPrompt(data types.AnalysisResult) string {
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
		data.Period,
		data.Data.Expenses,
		data.Data.Incomes,
		data.Data.Balance,
		data.Comparison.Change.ExpensesPercent,
		getChangeDirection(data.Comparison.Change.ExpensesChange),
		data.Comparison.Change.IncomesPercent,
		getChangeDirection(data.Comparison.Change.IncomesChange),
		data.Comparison.Change.BalancePercent,
		getChangeDirection(data.Comparison.Change.BalanceChange),
		len(data.Anomalies),
		len(data.Trends))
}

// buildDailyReportPrompt builds prompt for daily report
func (c *Client) buildDailyReportPrompt(data types.AnalysisResult) string {
	return fmt.Sprintf(`Создай ежедневный финансовый отчет на русском языке:

Сегодня потрачено: %.2f руб
Сегодня заработано: %.2f руб
Баланс: %.2f руб

По сравнению с вчера:
- Расходы: %.1f%% (%s)
- Доходы: %.1f%% (%s)

Создай мотивирующее сообщение с эмодзи. Если сэкономили - похвали, если потратили больше - дай совет.`,
		data.Data.Expenses,
		data.Data.Incomes,
		data.Data.Balance,
		data.Comparison.Change.ExpensesPercent,
		getChangeDirection(data.Comparison.Change.ExpensesChange),
		data.Comparison.Change.IncomesPercent,
		getChangeDirection(data.Comparison.Change.IncomesChange))
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

// AddToContext adds a message to the conversation context
func (c *Client) AddToContext(message string) {
	c.context = append(c.context, message)
	// Keep only last 10 messages to prevent context from growing too large
	if len(c.context) > 10 {
		c.context = c.context[len(c.context)-10:]
	}
}

// ClearContext clears the conversation context
func (c *Client) ClearContext() {
	c.context = make([]string, 0)
}

// GetContext returns the current conversation context
func (c *Client) GetContext() []string {
	return c.context
}

// GenerateWithContext generates text with conversation context
func (c *Client) GenerateWithContext(ctx context.Context, prompt string) (string, error) {
	// Build context-aware prompt
	fullPrompt := c.buildContextPrompt(prompt)

	// Generate response
	response, err := c.GenerateText(ctx, fullPrompt)
	if err != nil {
		return "", err
	}

	// Add to context
	c.AddToContext("User: " + prompt)
	c.AddToContext("Assistant: " + response)

	return response, nil
}

// buildContextPrompt builds a prompt with conversation context
func (c *Client) buildContextPrompt(prompt string) string {
	if len(c.context) == 0 {
		return prompt
	}

	contextStr := "Предыдущий контекст разговора:\n"
	for i, msg := range c.context {
		contextStr += fmt.Sprintf("%d. %s\n", i+1, msg)
	}
	contextStr += "\nТекущий вопрос: " + prompt

	return contextStr
}
