package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/seymourrisey/payflow-simulator/internal/model"
	"github.com/seymourrisey/payflow-simulator/pkg/idgen"
)

type TransactionRepository struct {
	db *pgxpool.Pool
}

func NewTransactionRepository(db *pgxpool.Pool) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// ProcessPayment — ACID transaction: debit saldo + insert record dalam 1 DB transaction
func (r *TransactionRepository) ProcessPayment(ctx context.Context, tx *model.Transaction) (*model.Transaction, error) {
	dbTx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer dbTx.Rollback(ctx)

	// STEP 1: Lock wallet row (SELECT FOR UPDATE)
	var currentBalance float64
	err = dbTx.QueryRow(ctx, `
		SELECT balance FROM wallets
		WHERE id = $1 FOR UPDATE
	`, tx.WalletID).Scan(&currentBalance)
	if err != nil {
		return nil, fmt.Errorf("lock wallet: %w", err)
	}

	// STEP 2: Cek saldo cukup
	totalDebit := tx.Amount + tx.Fee
	if currentBalance < totalDebit {
		return nil, fmt.Errorf("insufficient balance: have %.2f, need %.2f", currentBalance, totalDebit)
	}

	// STEP 3: Kurangi saldo
	_, err = dbTx.Exec(ctx, `
		UPDATE wallets
		SET balance = balance - $1, updated_at = NOW()
		WHERE id = $2
	`, totalDebit, tx.WalletID)
	if err != nil {
		return nil, fmt.Errorf("debit wallet: %w", err)
	}

	// STEP 4: Generate custom ID lalu insert transaction record
	tx.ID = idgen.NewTransactionID() // TXN-20260307-A1B2C3D4

	err = dbTx.QueryRow(ctx, `
		INSERT INTO transactions
			(id, reference_id, wallet_id, receiver_merchant_id, type, amount, fee, status, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'SUCCESS', $8)
		RETURNING created_at
	`, tx.ID, tx.ReferenceID, tx.WalletID, tx.ReceiverMerchantID,
		tx.Type, tx.Amount, tx.Fee, tx.Metadata).
		Scan(&tx.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert transaction: %w", err)
	}

	// STEP 5: Commit
	if err = dbTx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	tx.Status = model.TxStatusSuccess
	return tx, nil
}

// ProcessTopUp — credit saldo + insert transaction record
func (r *TransactionRepository) ProcessTopUp(ctx context.Context, walletID string, amount float64, referenceID string) (*model.Transaction, error) {
	dbTx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin topup: %w", err)
	}
	defer dbTx.Rollback(ctx)

	// STEP 1: Tambah saldo
	_, err = dbTx.Exec(ctx, `
		UPDATE wallets SET balance = balance + $1, updated_at = NOW()
		WHERE id = $2
	`, amount, walletID)
	if err != nil {
		return nil, fmt.Errorf("credit wallet: %w", err)
	}

	// STEP 2: Generate custom ID lalu insert transaction record
	tx := &model.Transaction{}
	tx.ID = idgen.NewTransactionID() // TXN-20260307-A1B2C3D4
	err = dbTx.QueryRow(ctx, `
		INSERT INTO transactions (id, reference_id, wallet_id, type, amount, fee, status)
		VALUES ($1, $2, $3, 'TOPUP', $4, 0, 'SUCCESS')
		RETURNING id, reference_id, wallet_id, type, amount, fee, status, created_at
	`, tx.ID, referenceID, walletID, amount).
		Scan(&tx.ID, &tx.ReferenceID, &tx.WalletID, &tx.Type, &tx.Amount, &tx.Fee, &tx.Status, &tx.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert topup transaction: %w", err)
	}

	// STEP 3: Commit
	if err = dbTx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit topup: %w", err)
	}

	return tx, nil
}

// FindByWalletID — riwayat transaksi dengan pagination
func (r *TransactionRepository) FindByWalletID(ctx context.Context, walletID string, limit, offset int) ([]model.Transaction, int, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, reference_id, wallet_id, receiver_merchant_id, type, amount, fee, status, created_at
		FROM transactions
		WHERE wallet_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, walletID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var txs []model.Transaction
	for rows.Next() {
		var tx model.Transaction
		err := rows.Scan(&tx.ID, &tx.ReferenceID, &tx.WalletID, &tx.ReceiverMerchantID,
			&tx.Type, &tx.Amount, &tx.Fee, &tx.Status, &tx.CreatedAt)
		if err != nil {
			return nil, 0, err
		}
		txs = append(txs, tx)
	}

	// Count total untuk pagination
	var total int
	r.db.QueryRow(ctx, `SELECT COUNT(*) FROM transactions WHERE wallet_id = $1`, walletID).Scan(&total)

	return txs, total, nil
}

// FindByReferenceID — cari 1 transaksi berdasarkan nomor referensi
func (r *TransactionRepository) FindByReferenceID(ctx context.Context, refID string) (*model.Transaction, error) {
	tx := &model.Transaction{}
	err := r.db.QueryRow(ctx, `
		SELECT id, reference_id, wallet_id, receiver_merchant_id, type, amount, fee, status, created_at
		FROM transactions WHERE reference_id = $1
	`, refID).Scan(&tx.ID, &tx.ReferenceID, &tx.WalletID, &tx.ReceiverMerchantID,
		&tx.Type, &tx.Amount, &tx.Fee, &tx.Status, &tx.CreatedAt)
	if err != nil {
		return nil, err
	}
	return tx, nil
}
