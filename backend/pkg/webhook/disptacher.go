package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/seymourrisey/payflow-simulator/config"
)

type WebhookPayload struct {
	Event       string    `json:"event"`
	ReferenceID string    `json:"reference_id"`
	MerchantID  string    `json:"merchant_id"` // string: MRC-XXXXXXXX
	Amount      float64   `json:"amount"`
	Fee         float64   `json:"fee"`
	Status      string    `json:"status"`
	Timestamp   time.Time `json:"timestamp"`
}

type SendResult struct {
	StatusCode   int
	Body         string
	Delivered    bool
	Attempts     int
	ErrorMessage string
}

type Dispatcher struct {
	client *http.Client
}

func NewDispatcher() *Dispatcher {
	timeout := time.Duration(config.App.WebhookTimeout) * time.Second
	return &Dispatcher{
		client: &http.Client{Timeout: timeout},
	}
}

// Send — kirim webhook ke merchant URL dengan exponential backoff retry
// Mengembalikan SendResult berisi status code, body, dan jumlah attempts
func (d *Dispatcher) Send(ctx context.Context, webhookURL string, payload *WebhookPayload) *SendResult {
	maxRetries := config.App.WebhookMaxRetry
	result := &SendResult{}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("marshal payload: %v", err)
		return result
	}

	for attempt := 1; attempt <= maxRetries; attempt++ {
		result.Attempts = attempt
		statusCode, respBody, err := d.sendOnce(ctx, webhookURL, bodyBytes)
		result.StatusCode = statusCode
		result.Body = respBody

		if err == nil && statusCode >= 200 && statusCode < 300 {
			result.Delivered = true
			log.Printf("✅ Webhook delivered to %s (attempt %d/%d)", webhookURL, attempt, maxRetries)
			return result
		}

		if err != nil {
			result.ErrorMessage = err.Error()
		}

		log.Printf("⚠️  Webhook attempt %d/%d failed (status=%d): %v", attempt, maxRetries, statusCode, err)

		if attempt < maxRetries {
			// Exponential backoff: 1s → 2s → 4s
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			select {
			case <-ctx.Done():
				result.ErrorMessage = "context cancelled"
				return result
			case <-time.After(backoff):
			}
		}
	}

	log.Printf("❌ Webhook failed after %d attempts to %s", maxRetries, webhookURL)
	return result
}

func (d *Dispatcher) sendOnce(ctx context.Context, url string, body []byte) (int, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return 0, "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Payflow-Signature", generateHMACSignature(body))
	req.Header.Set("X-Payflow-Timestamp", time.Now().UTC().Format(time.RFC3339))
	req.Header.Set("User-Agent", "PayFlow-Webhook/1.0")

	resp, err := d.client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()

	respBodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1024)) // max 1KB
	return resp.StatusCode, string(respBodyBytes), nil
}

// generateHMACSignature — HMAC-SHA256 untuk verifikasi keaslian webhook
// Merchant bisa verifikasi dengan secret key yang di-share saat onboarding
func generateHMACSignature(body []byte) string {
	// Dalam produksi, secret ini per-merchant dan disimpan di DB
	secret := []byte("payflow-webhook-secret")
	mac := hmac.New(sha256.New, secret)
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
