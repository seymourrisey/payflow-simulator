package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/seymourrisey/payflow-simulator/internal/dto"
	"github.com/seymourrisey/payflow-simulator/internal/model"
	"github.com/seymourrisey/payflow-simulator/internal/repository"
	"github.com/seymourrisey/payflow-simulator/pkg/idgen"
)

const (
	PaymentFeeRate  = 0.007 // 0.7% MDR fee
	QRExpiryMinutes = 15    // QR expired dalam 15 menit
)

type PaymentService struct {
	txRepo         *repository.TransactionRepository
	walletRepo     *repository.WalletRepository
	webhookService *WebhookService
}

func NewPaymentService(
	txRepo *repository.TransactionRepository,
	walletRepo *repository.WalletRepository,
	webhookService *WebhookService,
) *PaymentService {
	return &PaymentService{
		txRepo:         txRepo,
		walletRepo:     walletRepo,
		webhookService: webhookService,
	}
}

// GenerateQR — buat QR payload, tidak menyentuh DB
func (s *PaymentService) GenerateQR(ctx context.Context, userID string, req *dto.GenerateQRRequest) (*dto.GenerateQRResponse, error) {
	referenceID := fmt.Sprintf("PAY-%s", uuid.New().String())
	expiredAt := time.Now().Add(QRExpiryMinutes * time.Minute)

	qrPayload := map[string]any{
		"reference_id": referenceID,
		"merchant_id":  req.MerchantID,
		"amount":       req.Amount,
		"description":  req.Description,
		"expired_at":   expiredAt.Unix(),
	}

	qrBytes, err := json.Marshal(qrPayload)
	if err != nil {
		return nil, err
	}

	return &dto.GenerateQRResponse{
		QRData:      string(qrBytes),
		ReferenceID: referenceID,
		ExpiredAt:   expiredAt.Format(time.RFC3339),
	}, nil
}

// Pay — proses pembayaran ACID + dispatch webhook setelahnya
func (s *PaymentService) Pay(ctx context.Context, userID string, req *dto.PaymentRequest, idempotencyKey string) (*dto.PaymentResponse, error) {
	// Idempotency check
	if idempotencyKey != "" {
		existing, _ := s.txRepo.FindByReferenceID(ctx, idempotencyKey)
		if existing != nil {
			return &dto.PaymentResponse{
				ReferenceID: existing.ReferenceID,
				Amount:      existing.Amount,
				Fee:         existing.Fee,
				Status:      string(existing.Status),
			}, nil
		}
	}

	wallet, err := s.walletRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("wallet not found: %w", err)
	}

	fee := req.Amount * PaymentFeeRate

	referenceID := idempotencyKey
	if referenceID == "" {
		referenceID = fmt.Sprintf("PAY-%s", uuid.New().String())
	}

	merchantID := req.MerchantID
	tx := &model.Transaction{
		ReferenceID:        referenceID,
		WalletID:           wallet.ID,
		ReceiverMerchantID: &merchantID,
		Type:               model.TxTypePayment,
		Amount:             req.Amount,
		Fee:                fee,
		Metadata:           map[string]any{"description": req.Description},
	}

	processed, err := s.txRepo.ProcessPayment(ctx, tx)
	if err != nil {
		return nil, err
	}

	// Dispatch webhook ke merchant (non-blocking goroutine)
	if s.webhookService != nil {
		s.webhookService.DispatchPaymentWebhook(
			processed.ID,
			merchantID,
			processed.ReferenceID,
			processed.Amount,
			processed.Fee,
			string(processed.Status),
		)
	}

	return &dto.PaymentResponse{
		ReferenceID: processed.ReferenceID,
		Amount:      processed.Amount,
		Fee:         processed.Fee,
		Status:      string(processed.Status),
	}, nil
}

// TopUp — credit saldo + insert top_up_requests + transactions
func (s *PaymentService) TopUp(ctx context.Context, userID string, req *dto.TopUpRequest) (*dto.TopUpResponse, error) {
	wallet, err := s.walletRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("wallet not found: %w", err)
	}

	topUpID := idgen.NewTopUpID() // TUP-20260307-A1B2C3D4
	referenceID := fmt.Sprintf("TUP-%s", uuid.New().String())
	expiredAt := time.Now().Add(24 * time.Hour)

	_, err = s.txRepo.ProcessTopUp(
		ctx,
		wallet.ID,
		req.Amount,
		referenceID,
		topUpID,
		req.PaymentChannel,
		expiredAt,
	)
	if err != nil {
		return nil, err
	}

	return &dto.TopUpResponse{
		TopUpID:        topUpID, // fix: pakai topUpID bukan wallet.ID
		Amount:         req.Amount,
		PaymentChannel: req.PaymentChannel,
		Status:         "SUCCESS",
		ExpiredAt:      expiredAt.Format(time.RFC3339),
	}, nil
}

// GetHistory — riwayat transaksi dengan pagination
func (s *PaymentService) GetHistory(ctx context.Context, userID string, page, limit int) (*dto.TransactionListResponse, error) {
	wallet, err := s.walletRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	offset := (page - 1) * limit
	txs, total, err := s.txRepo.FindByWalletID(ctx, wallet.ID, limit, offset)
	if err != nil {
		return nil, err
	}

	var items []dto.TransactionItem
	for _, tx := range txs {
		items = append(items, dto.TransactionItem{
			ReferenceID: tx.ReferenceID,
			Type:        string(tx.Type),
			Amount:      tx.Amount,
			Fee:         tx.Fee,
			Status:      string(tx.Status),
			CreatedAt:   tx.CreatedAt.Format(time.RFC3339),
		})
	}

	return &dto.TransactionListResponse{
		Data:  items,
		Total: total,
	}, nil
}
