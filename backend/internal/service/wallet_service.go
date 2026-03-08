package service

import (
	"context"

	"github.com/seymourrisey/payflow-simulator/internal/dto"
)

// GetWallet — dipanggil dari PaymentService (walletRepo sudah ada di PaymentService)
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
