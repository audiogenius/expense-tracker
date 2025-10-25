package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
	"testing"
)

func TestVerifyTelegramAuth(t *testing.T) {
	// prepare auth helper with test bot token
	a := &Auth{BotToken: "test-bot-token"}

	// prepare payload map with fields from Telegram widget
	payload := map[string]string{
		"id":        "12345",
		"auth_date": "1700000000",
		"username":  "testuser",
	}

	// compute expected hash exactly as server does
	var pairs []string
	for k, v := range payload {
		pairs = append(pairs, k+"="+v)
	}
	sort.Strings(pairs)
	dataCheck := strings.Join(pairs, "\n")
	h := sha256.Sum256([]byte(a.BotToken))
	mac := hmac.New(sha256.New, h[:])
	mac.Write([]byte(dataCheck))
	expected := hex.EncodeToString(mac.Sum(nil))

	// attach hash to payload
	payload["hash"] = expected

	if !a.VerifyTelegramAuth(payload) {
		t.Fatalf("expected verification to pass")
	}
}
