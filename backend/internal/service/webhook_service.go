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

		event := "payment." + lowercase(status)

		// Ambil webhook URL merchant
		webhookURL, err := s.webhookRepo.FindMerchantWebhookURL(ctx, merchantID)
		if err != nil {
			log.Printf("(╥‸╥) Webhook: gagal ambil URL merchant %s: %v", merchantID, err)
			// Tetap simpan log dengan status failed
			s.saveLog(ctx, txID, merchantID, event, nil, &model.WebhookLog{
				RetryCount:  0,
				IsDelivered: false,
			})
			return
		}

		// Build payload
		payload := &webhook.WebhookPayload{
			Event:       event,
			ReferenceID: refID,
			MerchantID:  merchantID,
			Amount:      amount,
			Fee:         fee,
			Status:      status,
			Timestamp:   time.Now().UTC(),
		}
		payloadBytes, _ := json.Marshal(payload)

		// Kalau tidak ada webhook URL, tetap catat sebagai skipped
		if webhookURL == "" {
			log.Printf("( ˶°ㅁ°) !! Webhook skipped: merchant %s tidak punya webhook URL", merchantID)
			s.saveLog(ctx, txID, merchantID, event, payloadBytes, &model.WebhookLog{
				RetryCount:  0,
				IsDelivered: false,
			})
			return
		}

		// Kirim webhook dengan retry
		result := s.dispatcher.Send(ctx, webhookURL, payload)

		logEntry := &model.WebhookLog{
			RetryCount:  result.Attempts - 1,
			IsDelivered: result.Delivered,
		}
		if result.StatusCode > 0 {
			logEntry.ResponseStatus = &result.StatusCode
		}
		if result.Body != "" {
			logEntry.ResponseBody = &result.Body
		}

		s.saveLog(ctx, txID, merchantID, event, payloadBytes, logEntry)

		if result.Delivered {
			log.Printf("ദ്ദി ˉ͈̀꒳ˉ͈́ )✧ Webhook delivered → merchant %s | ref %s", merchantID, refID)
		} else {
			log.Printf("(╥﹏╥) Webhook FAILED → merchant %s | ref %s | attempts: %d", merchantID, refID, result.Attempts)
		}
	}()
}

// saveLog — helper internal untuk simpan webhook log ke DB
func (s *WebhookService) saveLog(ctx context.Context, txID, merchantID, event string, payloadBytes []byte, result *model.WebhookLog) {
	if payloadBytes == nil {
		payloadBytes = []byte(`{}`)
	}

	logEntry := &model.WebhookLog{
		MerchantID:     merchantID,
		TransactionID:  txID,
		Event:          event,
		Payload:        json.RawMessage(payloadBytes),
		RetryCount:     result.RetryCount,
		IsDelivered:    result.IsDelivered,
		ResponseStatus: result.ResponseStatus,
		ResponseBody:   result.ResponseBody,
	}

	if err := s.webhookRepo.InsertLog(ctx, logEntry); err != nil {
		log.Printf("(╥‸╥) Gagal simpan webhook log: %v", err)
	} else {
		log.Printf("ദ്ദി ˉ͈̀꒳ˉ͈́ )✧ Webhook log tersimpan: %s | event: %s | delivered: %v",
			logEntry.ID, event, result.IsDelivered)
	}
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
