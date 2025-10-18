package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func envOr(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

// createHS256Token manually creates a simple JWT with numeric sub claim
func postExpense(apiURL string, botKey string, telegramID int64, username string, amount float64) (int, error) {
	return postExpenseWithCategory(apiURL, botKey, telegramID, username, amount, nil)
}

func postExpenseWithCategory(apiURL string, botKey string, telegramID int64, username string, amount float64, categoryID *int) (int, error) {
	// Convert amount to cents (multiply by 100 and round)
	amountCents := int(amount * 100)
	payload := map[string]interface{}{
		"telegram_id":  telegramID,
		"username":     username,
		"amount_cents": amountCents,
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
	}
	if categoryID != nil {
		payload["category_id"] = *categoryID
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", apiURL+"/internal/expenses", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if botKey != "" {
		req.Header.Set("X-BOT-KEY", botKey)
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	return resp.StatusCode, nil
}

func detectCategory(apiURL, description string) *int {
	if description == "" {
		return nil
	}

	payload := map[string]string{"description": description}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", apiURL+"/categories/detect", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil
	}

	if id, ok := result["id"].(float64); ok && id > 0 {
		categoryID := int(id)
		return &categoryID
	}

	return nil
}

func main() {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		fmt.Println("TELEGRAM_BOT_TOKEN not set")
		os.Exit(2)
	}
	apiURL := envOr("API_URL", "http://api:8080")
	botKey := os.Getenv("BOT_API_KEY")
	if botKey == "" {
		fmt.Println("warning: BOT_API_KEY not set; internal endpoint may reject requests")
	}

	// poll getUpdates
	offset := 0
	re := regexp.MustCompile(`^\s*([0-9]+(?:[.,][0-9]{1,2})?)\s*$`)

	for {
		uurl := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?offset=%d&timeout=30", botToken, offset)
		resp, err := http.Get(uurl)
		if err != nil {
			fmt.Println("getUpdates err:", err)
			time.Sleep(3 * time.Second)
			continue
		}
		var data struct {
			Ok     bool                     `json:"ok"`
			Result []map[string]interface{} `json:"result"`
		}
		dec := json.NewDecoder(resp.Body)
		if err := dec.Decode(&data); err != nil {
			resp.Body.Close()
			fmt.Println("decode updates err:", err)
			time.Sleep(3 * time.Second)
			continue
		}
		resp.Body.Close()
		for _, item := range data.Result {
			if updateIDf, ok := item["update_id"]; ok {
				if uid, ok := updateIDf.(float64); ok {
					offset = int(uid) + 1
				}
			}
			msg, ok := item["message"].(map[string]interface{})
			if !ok {
				continue
			}

			// get user info
			var fromID int64
			username := ""
			if from, ok := msg["from"].(map[string]interface{}); ok {
				if idf, ok := from["id"].(float64); ok {
					fromID = int64(idf)
				}
				if u, ok := from["username"].(string); ok {
					username = u
				} else {
					// fallback to first_name + last_name
					fn, _ := from["first_name"].(string)
					ln, _ := from["last_name"].(string)
					if fn != "" && ln != "" {
						username = fn + " " + ln
					} else {
						username = fn + ln
					}
				}
			}

			chatID := int64(0)
			if chatf, ok := msg["chat"].(map[string]interface{}); ok {
				if cid, ok := chatf["id"].(float64); ok {
					chatID = int64(cid)
				}
			}

			// handle text messages
			text, _ := msg["text"].(string)
			if text != "" {
				handleTextMessage(botToken, apiURL, botKey, fromID, username, chatID, text, re)
			}

			// handle photo messages
			if photos, ok := msg["photo"].([]interface{}); ok && len(photos) > 0 {
				handlePhotoMessage(botToken, apiURL, botKey, fromID, username, chatID, photos)
			}
		}
	}
}

func handleTextMessage(botToken, apiURL, botKey string, fromID int64, username string, chatID int64, text string, re *regexp.Regexp) {
	// handle commands
	if strings.HasPrefix(text, "/") {
		handleCommand(botToken, apiURL, botKey, fromID, username, chatID, text)
		return
	}

	// handle shared expenses
	// Format: "split 100 продукты @username1 @username2"
	sharedRegex := regexp.MustCompile(`^split\s+([0-9]+(?:[.,][0-9]{1,2})?)\s+(.*)$`)
	if sharedRegex.MatchString(text) {
		handleSharedExpense(botToken, apiURL, botKey, fromID, username, chatID, text, sharedRegex)
		return
	}

	// handle expense amounts with optional category
	// Format: "100 продукты" or "50.50 кафе"
	expenseRegex := regexp.MustCompile(`^\s*([0-9]+(?:[.,][0-9]{1,2})?)\s*(.*)$`)
	if !expenseRegex.MatchString(text) {
		sendMessage(botToken, chatID, "Отправьте сумму расхода (например: 100 или 50.50 продукты) или используйте команды /help")
		return
	}

	m := expenseRegex.FindStringSubmatch(text)
	if len(m) < 3 {
		return
	}

	// Replace comma with dot for parsing
	numStr := m[1]
	if idx := strings.Index(numStr, ","); idx >= 0 {
		numStr = numStr[:idx] + "." + numStr[idx+1:]
	}
	amount, _ := strconv.ParseFloat(numStr, 64)
	description := strings.TrimSpace(m[2])

	// Try to detect category from description
	categoryID := detectCategory(apiURL, description)

	status, err := postExpenseWithCategory(apiURL, botKey, fromID, username, amount, categoryID)

	// send a reply via sendMessage
	var replyText string
	if err != nil {
		replyText = fmt.Sprintf("❌ Не удалось записать %s: %v", m[1], err)
	} else if status >= 200 && status < 300 {
		categoryText := ""
		if categoryID != nil {
			categoryText = fmt.Sprintf(" (категория: %d)", *categoryID)
		}
		replyText = fmt.Sprintf("✅ Записал расход: %s руб.%s", m[1], categoryText)
	} else {
		replyText = fmt.Sprintf("❌ Не удалось записать %s (ошибка %d)", m[1], status)
	}
	sendMessage(botToken, chatID, replyText)
}

func handleCommand(botToken, apiURL, botKey string, fromID int64, username string, chatID int64, command string) {
	switch command {
	case "/help":
		helpText := "🤖 *Expense Tracker Bot*\n\n" +
			"*📋 Команды:*\n" +
			"/help - показать эту справку\n" +
			"/total - показать общую сумму расходов\n" +
			"/total week - расходы за неделю\n" +
			"/total month - расходы за месяц\n" +
			"/debts - показать долги\n\n" +
			"*💰 Как записать расход:*\n" +
			"• Просто сумма: 100 или 50.50\n" +
			"• С категорией: 100 продукты или 50.50 кафе\n" +
			"• Shared расход: split 300 кафе @username1 @username2\n\n" +
			"*📸 Фото чеков:*\n" +
			"• Отправьте фото чека для автоматического распознавания\n" +
			"• (Функция в разработке)\n\n" +
			"*🏷️ Доступные категории:*\n" +
			"• Продукты (еда, магазин, супермаркет)\n" +
			"• Транспорт (бензин, такси, метро)\n" +
			"• Кафе и рестораны (кафе, ресторан, обед)\n" +
			"• Развлечения (кино, театр, игры)\n" +
			"• Здоровье (аптека, врач, лекарства)\n" +
			"• Одежда (одежда, обувь, шопинг)\n" +
			"• Коммунальные услуги (свет, вода, интернет)\n" +
			"• Прочее (другое, разное)\n\n" +
			"*💡 Примеры:*\n" +
			"• 100 -> Записал расход: 100 руб.\n" +
			"• 100 продукты -> Записал расход: 100 руб. (категория: Продукты)\n" +
			"• split 300 кафе @wife @friend -> Создан shared расход на 3 человек\n" +
			"• /total -> Показать все расходы\n" +
			"• /debts -> Показать долги\n\n" +
			"Все суммы в рублях! 💸"

		sendMessage(botToken, chatID, helpText)

	case "/total":
		getTotalExpenses(botToken, apiURL, botKey, fromID, chatID, "all")

	case "/total week":
		getTotalExpenses(botToken, apiURL, botKey, fromID, chatID, "week")

	case "/total month":
		getTotalExpenses(botToken, apiURL, botKey, fromID, chatID, "month")

	case "/debts":
		getDebts(botToken, apiURL, botKey, fromID, chatID)

	default:
		sendMessage(botToken, chatID, "Неизвестная команда. Используйте /help для справки.")
	}
}

func getTotalExpenses(botToken, apiURL, botKey string, fromID int64, chatID int64, period string) {
	// Make request to internal API
	url := apiURL + "/internal/expenses/total?telegram_id=" + strconv.FormatInt(fromID, 10)
	if period != "all" {
		url += "&period=" + period
	}

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-BOT-KEY", botKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		sendMessage(botToken, chatID, "❌ Ошибка получения данных")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		sendMessage(botToken, chatID, "❌ Ошибка сервера")
		return
	}

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		sendMessage(botToken, chatID, "❌ Ошибка обработки данных")
		return
	}

	totalRubles := data["total_rubles"].(float64)
	periodText := "всего"
	if period == "week" {
		periodText = "за неделю"
	} else if period == "month" {
		periodText = "за месяц"
	}

	sendMessage(botToken, chatID, fmt.Sprintf("📊 Расходы %s: %.2f руб.", periodText, totalRubles))
}

func getDebts(botToken, apiURL, botKey string, fromID int64, chatID int64) {
	// Make request to internal API
	url := apiURL + "/internal/debts?telegram_id=" + strconv.FormatInt(fromID, 10)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-BOT-KEY", botKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		sendMessage(botToken, chatID, "❌ Ошибка получения данных")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		sendMessage(botToken, chatID, "❌ Ошибка сервера")
		return
	}

	var debts []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&debts); err != nil {
		sendMessage(botToken, chatID, "❌ Ошибка обработки данных")
		return
	}

	if len(debts) == 0 {
		sendMessage(botToken, chatID, "💰 У вас нет долгов")
		return
	}

	// Format debts message
	var message strings.Builder
	message.WriteString("💰 *Ваши долги:*\n\n")

	owedToMe := 0
	iOwe := 0

	for _, debt := range debts {
		amountCents := int(debt["amount_cents"].(float64))
		amount := float64(amountCents) / 100.0
		username := debt["username"].(string)
		debtType := debt["type"].(string)

		if debtType == "owed_to_me" {
			owedToMe += amountCents
			message.WriteString(fmt.Sprintf("✅ %s должен вам %.2f руб.\n", username, amount))
		} else if debtType == "i_owe" {
			iOwe += amountCents
			message.WriteString(fmt.Sprintf("❌ Вы должны %s %.2f руб.\n", username, amount))
		}
	}

	message.WriteString("\n📊 *Итого:*\n")
	message.WriteString(fmt.Sprintf("Должны вам: %.2f руб.\n", float64(owedToMe)/100.0))
	message.WriteString(fmt.Sprintf("Вы должны: %.2f руб.\n", float64(iOwe)/100.0))

	balance := owedToMe - iOwe
	if balance > 0 {
		message.WriteString(fmt.Sprintf("💚 Ваш баланс: +%.2f руб.", float64(balance)/100.0))
	} else if balance < 0 {
		message.WriteString(fmt.Sprintf("💸 Ваш баланс: %.2f руб.", float64(balance)/100.0))
	} else {
		message.WriteString("⚖️ Баланс: 0 руб.")
	}

	sendMessage(botToken, chatID, message.String())
}

func handleSharedExpense(botToken, apiURL, botKey string, fromID int64, username string, chatID int64, text string, regex *regexp.Regexp) {
	m := regex.FindStringSubmatch(text)
	if len(m) < 3 {
		sendMessage(botToken, chatID, "❌ Неверный формат. Используйте: split 100 продукты @username1 @username2")
		return
	}

	// Parse amount
	numStr := m[1]
	if idx := strings.Index(numStr, ","); idx >= 0 {
		numStr = numStr[:idx] + "." + numStr[idx+1:]
	}
	amount, _ := strconv.ParseFloat(numStr, 64)

	// Parse description and usernames
	description := strings.TrimSpace(m[2])

	// Extract usernames (format: @username1 @username2)
	usernameRegex := regexp.MustCompile(`@(\w+)`)
	usernames := usernameRegex.FindAllStringSubmatch(description, -1)

	// Remove usernames from description
	description = usernameRegex.ReplaceAllString(description, "")
	description = strings.TrimSpace(description)

	// Get Telegram IDs for usernames (simplified - in real implementation you'd look them up)
	var splitWith []int64
	for range usernames {
		// For now, just add placeholder IDs - in real implementation you'd look up actual Telegram IDs
		// This is a limitation of the current implementation
		splitWith = append(splitWith, 123456789) // Placeholder
	}

	if len(splitWith) == 0 {
		sendMessage(botToken, chatID, "❌ Не найдены пользователи для разделения. Используйте @username")
		return
	}

	// Detect category
	categoryID := detectCategory(apiURL, description)

	// Create shared expense (simplified - just create regular expense for now)
	// TODO: Implement proper shared expense creation
	status, err := postExpenseWithCategory(apiURL, botKey, fromID, username, amount, categoryID)
	if err != nil {
		sendMessage(botToken, chatID, "❌ Ошибка создания расхода")
		return
	}

	if status < 200 || status >= 300 {
		sendMessage(botToken, chatID, "❌ Ошибка сервера при создании расхода")
		return
	}

	sendMessage(botToken, chatID, fmt.Sprintf("✅ Записал расход: %.2f руб. (категория: %d)\n💡 Shared расходы будут добавлены в следующей версии",
		amount, *categoryID))
}

func handlePhotoMessage(botToken, apiURL, botKey string, fromID int64, username string, chatID int64, photos []interface{}) {
	// Get the largest photo (last in array)
	if len(photos) == 0 {
		return
	}

	photo := photos[len(photos)-1].(map[string]interface{})
	_, _ = photo["file_id"].(string) // fileID for future OCR implementation

	sendMessage(botToken, chatID, "📸 Получил фото чека. Обработка OCR будет добавлена в следующей версии.")

	// TODO: Implement OCR processing
	// 1. Get file path from Telegram
	// 2. Download file
	// 3. Send to OCR service
	// 4. Parse results and show inline keyboard for selection
}

func sendMessage(botToken string, chatID int64, text string) {
	smURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	postBody := map[string]interface{}{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "Markdown",
	}
	pb, _ := json.Marshal(postBody)
	http.Post(smURL, "application/json", bytes.NewReader(pb))
}
