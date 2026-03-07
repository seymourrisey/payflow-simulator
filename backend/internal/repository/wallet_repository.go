package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/seymourrisey/payflow-simulator/internal/model"
)

type WalletRepository struct {
	db *pgxpool.Pool
}

func NewWalletRepository(db *pgxpool.Pool) *WalletRepository {
	return &WalletRepository{db: db}
}

// FindByUserID — ambil wallet berdasarkan user ID
func (r *WalletRepository) FindByUserID(ctx context.Context, userID string) (*model.Wallet, error) {
	wallet := &model.Wallet{}
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, balance, currency, updated_at
		FROM wallets WHERE user_id = $1
	`, userID).Scan(&wallet.ID, &wallet.UserID, &wallet.Balance, &wallet.Currency, &wallet.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return wallet, nil
}

// FindByIDForUpdate — SELECT FOR UPDATE untuk lock row saat transaksi ACID
// Harus dipanggil di dalam database transaction (pgx.Tx)
func (r *WalletRepository) FindByIDForUpdate(ctx context.Context, tx pgx.Tx, walletID string) (*model.Wallet, error) {
	wallet := &model.Wallet{}
	err := tx.QueryRow(ctx, `
		SELECT id, user_id, balance, currency, updated_at
		FROM wallets WHERE id = $1 FOR UPDATE
	`, walletID).Scan(&wallet.ID, &wallet.UserID, &wallet.Balance, &wallet.Currency, &wallet.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return wallet, nil
}

// TopUpBalance — tambah saldo wallet
func (r *WalletRepository) TopUpBalance(ctx context.Context, walletID string, amount float64) error {
	_, err := r.db.Exec(ctx, `
		UPDATE wallets
		SET balance = balance + $1, updated_at = NOW()
		WHERE id = $2
	`, amount, walletID)
	return err
}
