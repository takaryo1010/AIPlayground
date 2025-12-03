package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"gopkg.in/gomail.v2"
)

// SendErrorEmail はメールでエラー通知を送信します（フォールバック用）。
func SendErrorEmail(err error, method, uri string) {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPortStr := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	toEmail := smtpUser // SMTP_USERを宛先に設定

	if smtpHost == "" || smtpPortStr == "" || smtpUser == "" || smtpPass == "" {
		log.Println("ERROR: SMTP configuration is incomplete for fallback email. Email not sent.")
		return
	}

	smtpPort, convErr := strconv.Atoi(smtpPortStr)
	if convErr != nil {
		log.Printf("ERROR: Invalid SMTP_PORT: %v\n", convErr)
		return
	}

	m := gomail.NewMessage()
	m.SetHeader("From", smtpUser)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", "[ALERT] API Server Error Detected")
	body := fmt.Sprintf("<h1>API Server Error</h1><p>An error occurred in the API server.</p><ul><li><b>Error:</b> %v</li><li><b>Request URI:</b> %s</li><li><b>Request Method:</b> %s</li></ul>", err, uri, method)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPass)

	log.Printf("INFO: Sending fallback error email to %s...\n", toEmail)
	if err := d.DialAndSend(m); err != nil {
		log.Printf("ERROR: Failed to send fallback email: %v\n", err)
	} else {
		log.Println("INFO: Fallback email sent successfully.")
	}
}

// SendDiscordNotification はDiscord Webhookにエラー通知を送信します。
func SendDiscordNotification(err error, method, uri string) {
	discordWebhookURL := os.Getenv("DISCORD_WEBHOOK_URL")
	if discordWebhookURL == "" {
		log.Println("ERROR: DISCORD_WEBHOOK_URL is not set. Should not be called.")
		return
	}

	payload := map[string]interface{}{
		"username": "API Server Alertだよ♡",
		"embeds": []map[string]interface{}{
			{
				"title":       "❌ API Server Error Detected",
				"description": fmt.Sprintf("APIサーバーでエラーが発生しちゃった(´;ω;｀)\n早く直してね♡"),
				"color":       15158332, // 赤色
				"fields": []map[string]interface{}{
					{"name": "Error", "value": err.Error(), "inline": false},
					{"name": "Request URI", "value": uri, "inline": true},
					{"name": "Method", "value": method, "inline": true},
				},
				"timestamp": time.Now().Format(time.RFC3339),
			},
		},
	}

	jsonPayload, _ := json.Marshal(payload)
	resp, httpErr := http.Post(discordWebhookURL, "application/json", bytes.NewBuffer(jsonPayload))
	if httpErr != nil {
		log.Printf("ERROR: Failed to send Discord notification: %v\n", httpErr)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		log.Println("INFO: Discord notification sent successfully.")
	} else {
		log.Printf("ERROR: Failed to send Discord notification. Status code: %d\n", resp.StatusCode)
	}
}
