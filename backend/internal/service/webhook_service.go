package service

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/seymourrisey/payflow-simulator/internal/dto"
	"github.com/seymourrisey/payflow-simulator/internal/model"
	"github.com/seymourrisey/payflow-simulator/internal/repository"
	"github.com/seymourrisey/payflow-simulator/pkg/webhook"
)

type WebhookService struct {
	webhookRepo *repository.WebhookRepository
	dispatcher  *webhook.Dispatcher
}

func NewWebhookService(webhookRepo *repository.WebhookRepository, dispatcher *webhook.Dispatcher) *WebhookService {
	return &WebhookService{webhookRepo: webhookRepo, dispatcher: dispatcher}
}

// DispatchPaymentWebhook — dipanggil async setelah payment berhasil
func (s *WebhookService) DispatchPaymentWebhook(txID, merchantID, refID string, amount, fee float64, status string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		webhookURL, err := s.webhookRepo.FindMerchantWebhookURL(ctx, merchantID)
		if err != nil || webhookURL == "" {
			log.Printf("Webhook skipped: merchant %s has no webhook URL", merchantID)
			return
		}

		payload := &webhook.WebhookPayload{
			Event:       "payment." + lowercase(status),
			ReferenceID: refID,
			MerchantID:  merchantID,
			Amount:      amount,
			Fee:         fee,
			Status:      status,
			Timestamp:   time.Now().UTC(),
		}

		result := s.dispatcher.Send(ctx, webhookURL, payload)

		payloadBytes, _ := json.Marshal(payload)

		logEntry := &model.WebhookLog{
			MerchantID:    merchantID,
			TransactionID: txID,
			Event:         payload.Event,
			Payload:       json.RawMessage(payloadBytes),
			RetryCount:    result.Attempts - 1,
			IsDelivered:   result.Delivered,
		}
		if result.StatusCode > 0 {
			logEntry.ResponseStatus = &result.StatusCode
		}
		if result.Body != "" {
			logEntry.ResponseBody = &result.Body
		}

		if err := s.webhookRepo.InsertLog(ctx, logEntry); err != nil {
			log.Printf("Failed to save webhook log: %v", err)
		}
	}()
}

// GetLogs — ambil semua webhook logs untuk panel
func (s *WebhookService) GetLogs(ctx context.Context, page, limit int) (*dto.WebhookLogsResponse, error) {
	offset := (page - 1) * limit
	logs, total, err := s.webhookRepo.FindAll(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	var items []dto.WebhookLogItem
	for _, l := range logs {
		item := dto.WebhookLogItem{
			ID:            l.ID,
			MerchantID:    l.MerchantID,
			MerchantName:  l.MerchantName,
			TransactionID: l.TransactionID,
			Event:         l.Event,
			RetryCount:    l.RetryCount,
			IsDelivered:   l.IsDelivered,
			SentAt:        l.SentAt.Format(time.RFC3339),
		}
		if l.ResponseStatus != nil {
			item.ResponseStatus = l.ResponseStatus
		}
		if l.ResponseBody != nil {
			item.ResponseBody = l.ResponseBody
		}
		payloadStr := string(l.Payload)
		item.Payload = &payloadStr
		items = append(items, item)
	}

	return &dto.WebhookLogsResponse{Data: items, Total: total}, nil
}

func (s *WebhookService) GetStats(ctx context.Context) (*model.WebhookStats, error) {
	return s.webhookRepo.GetStats(ctx)
}

func (s *WebhookService) GetMerchants(ctx context.Context) ([]model.Merchant, error) {
	return s.webhookRepo.GetAllMerchants(ctx)
}

func lowercase(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		result[i] = c
	}
	return string(result)
}
