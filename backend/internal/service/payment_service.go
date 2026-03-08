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
	PaymentFeeRate  = 0.007 // 0.7% fee per transaksi (simulasi MDR)
	QRExpiryMinutes = 15    // QR code expired dalam 15 menit
)

type PaymentService struct {
	txRepo     *repository.TransactionRepository
	walletRepo *repository.WalletRepository
}

func NewPaymentService(txRepo *repository.TransactionRepository, walletRepo *repository.WalletRepository) *PaymentService {
	return &PaymentService{txRepo: txRepo, walletRepo: walletRepo}
}

// GenerateQR — buat QR payload untuk merchant
func (s *PaymentService) GenerateQR(ctx context.Context, userID string, req *dto.GenerateQRRequest) (*dto.GenerateQRResponse, error) {
	referenceID := fmt.Sprintf("PAY-%s", uuid.New().String())
	expiredAt := time.Now().Add(QRExpiryMinutes * time.Minute)

	// QR payload berisi info yang cukup untuk memproses payment
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

// Pay — proses pembayaran dari QR scan
// referenceID dipakai sebagai idempotency key — request duplikat tidak akan diproses 2x
func (s *PaymentService) Pay(ctx context.Context, userID string, req *dto.PaymentRequest, idempotencyKey string) (*dto.PaymentResponse, error) {
	// Idempotency check: cek apakah reference ini sudah diproses
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

	// Ambil wallet user
	wallet, err := s.walletRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("wallet not found: %w", err)
	}

	// Hitung fee
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
		Metadata: map[string]any{
			"description": req.Description,
		},
	}

	// Proses ACID transaction (debit saldo + insert record)
	processed, err := s.txRepo.ProcessPayment(ctx, tx)
	if err != nil {
		return nil, err
	}

	return &dto.PaymentResponse{
		ReferenceID: processed.ReferenceID,
		Amount:      processed.Amount,
		Fee:         processed.Fee,
		Status:      string(processed.Status),
	}, nil
}

// TopUp — simulasi top up saldo
func (s *PaymentService) TopUp(ctx context.Context, userID string, req *dto.TopUpRequest) (*dto.TopUpResponse, error) {
	wallet, err := s.walletRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("wallet not found: %w", err)
	}

	topUpID := idgen.NewTopUpID()
	referenceID := fmt.Sprintf("TUP-%s", uuid.New().String())
	expiredAt := time.Now().Add(24 * time.Hour)

	// Langsung SUCCESS (simulasi — real-world: tunggu konfirmasi bank/VA)
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
		TopUpID:        topUpID, // ← fix: pakai topUpID, bukan wallet.ID
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
