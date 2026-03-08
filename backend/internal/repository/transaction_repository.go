package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

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

// ProcessPayment — inti dari ACID transaction
// Debit saldo + insert transaction record dalam 1 DB transaction
func (r *TransactionRepository) ProcessPayment(ctx context.Context, tx *model.Transaction) (*model.Transaction, error) {
	dbTx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer dbTx.Rollback(ctx)

	// STEP 1: Lock wallet row (SELECT FOR UPDATE)
	// Mencegah race condition jika ada 2 request bayar bersamaan
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

	// STEP 4: Marshal metadata ke JSON string
	// Dengan SimpleProtocol, WAJIB pakai ::jsonb cast di SQL
	// karena PostgreSQL tidak bisa auto-detect tipe dari parameter text
	metadataStr := "{}" // default empty JSON object
	if tx.Metadata != nil {
		metadataBytes, err := json.Marshal(tx.Metadata)
		if err != nil {
			return nil, fmt.Errorf("marshal metadata: %w", err)
		}
		metadataStr = string(metadataBytes)
	}

	// STEP 5: Generate custom ID lalu insert transaction record
	tx.ID = idgen.NewTransactionID()
	err = dbTx.QueryRow(ctx, `
		INSERT INTO transactions
			(id, reference_id, wallet_id, receiver_merchant_id, type, amount, fee, status, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'SUCCESS', $8::jsonb)
		RETURNING created_at
	`, tx.ID, tx.ReferenceID, tx.WalletID, tx.ReceiverMerchantID,
		tx.Type, tx.Amount, tx.Fee, metadataStr).
		Scan(&tx.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert transaction: %w", err)
	}

	// STEP 6: Commit
	if err = dbTx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	tx.Status = model.TxStatusSuccess
	return tx, nil
}

// ProcessTopUp — credit saldo + insert ke top_up_requests + transactions (ACID)
func (r *TransactionRepository) ProcessTopUp(
	ctx context.Context,
	walletID string,
	amount float64,
	referenceID string,
	topUpID string,
	paymentChannel string,
	expiredAt time.Time,
) (*model.Transaction, error) {
	dbTx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin topup: %w", err)
	}
	defer dbTx.Rollback(ctx)

	// STEP 1: Tambah saldo wallet
	_, err = dbTx.Exec(ctx, `
		UPDATE wallets SET balance = balance + $1, updated_at = NOW()
		WHERE id = $2
	`, amount, walletID)
	if err != nil {
		return nil, fmt.Errorf("credit wallet: %w", err)
	}

	// STEP 2: Insert ke top_up_requests
	_, err = dbTx.Exec(ctx, `
		INSERT INTO top_up_requests (id, wallet_id, amount, payment_channel, status, expired_at)
		VALUES ($1, $2, $3, $4, 'SUCCESS', $5)
	`, topUpID, walletID, amount, paymentChannel, expiredAt)
	if err != nil {
		return nil, fmt.Errorf("insert top_up_request: %w", err)
	}

	// STEP 3: Insert ke transactions
	tx := &model.Transaction{}
	tx.ID = idgen.NewTransactionID()
	err = dbTx.QueryRow(ctx, `
		INSERT INTO transactions (id, reference_id, wallet_id, type, amount, fee, status)
		VALUES ($1, $2, $3, 'TOPUP', $4, 0, 'SUCCESS')
		RETURNING id, reference_id, wallet_id, type, amount, fee, status, created_at
	`, tx.ID, referenceID, walletID, amount).
		Scan(&tx.ID, &tx.ReferenceID, &tx.WalletID, &tx.Type, &tx.Amount, &tx.Fee, &tx.Status, &tx.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert topup transaction: %w", err)
	}

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
		if err := rows.Scan(&tx.ID, &tx.ReferenceID, &tx.WalletID, &tx.ReceiverMerchantID,
			&tx.Type, &tx.Amount, &tx.Fee, &tx.Status, &tx.CreatedAt); err != nil {
			return nil, 0, err
		}
		txs = append(txs, tx)
	}

	var total int
	r.db.QueryRow(ctx, `SELECT COUNT(*) FROM transactions WHERE wallet_id = $1`, walletID).Scan(&total)
	return txs, total, nil
}

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
