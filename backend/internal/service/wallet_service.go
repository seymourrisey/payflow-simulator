package service

import (
	"context"

	"github.com/seymourrisey/payflow-simulator/internal/dto"
	"github.com/seymourrisey/payflow-simulator/internal/repository"
)

// WalletService — bisa juga dipisah jadi file sendiri
// Untuk simplisitas, method GetWallet ditambahkan di sini

type WalletService struct {
	walletRepo *repository.WalletRepository
}

func NewWalletService(walletRepo *repository.WalletRepository) *WalletService {
	return &WalletService{walletRepo: walletRepo}
}

// GetWallet juga diakses via PaymentService
func (s *PaymentService) GetWallet(ctx context.Context, userID string) (*dto.WalletResponse, error) {
	wallet, err := s.walletRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &dto.WalletResponse{
		ID:       wallet.ID,
		Balance:  wallet.Balance,
		Currency: wallet.Currency,
	}, nil
}
