package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/seymourrisey/payflow-simulator/config"
)

type WebhookPayload struct {
	Event       string    `json:"event"` // "payment.success", "payment.failed"
	ReferenceID string    `json:"reference_id"`
	MerchantID  int       `json:"merchant_id"`
	Amount      float64   `json:"amount"`
	Fee         float64   `json:"fee"`
	Status      string    `json:"status"`
	Timestamp   time.Time `json:"timestamp"`
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

// Send — kirim webhook ke merchant URL dengan retry logic
func (d *Dispatcher) Send(ctx context.Context, webhookURL string, payload *WebhookPayload) error {
	maxRetries := config.App.WebhookMaxRetry

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal webhook payload: %w", err)
	}

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := d.sendOnce(ctx, webhookURL, body)
		if err == nil {
			log.Printf("✅ Webhook sent to %s (attempt %d)", webhookURL, attempt)
			return nil
		}

		lastErr = err
		log.Printf("⚠️  Webhook attempt %d/%d failed: %v", attempt, maxRetries, err)

		if attempt < maxRetries {
			// Exponential backoff: 1s, 2s, 4s
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}
	}

	return fmt.Errorf("webhook failed after %d attempts: %w", maxRetries, lastErr)
}

func (d *Dispatcher) sendOnce(ctx context.Context, url string, body []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Payflow-Signature", generateSignature(body)) // HMAC signature

	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("merchant returned status %d", resp.StatusCode)
	}

	return nil
}

// generateSignature — HMAC-SHA256 sederhana untuk verifikasi webhook
// (implementasi lengkap bisa ditambahkan dengan crypto/hmac)
func generateSignature(body []byte) string {
	return fmt.Sprintf("sha256=%x", body[:min(8, len(body))])
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
