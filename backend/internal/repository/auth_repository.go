package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/seymourrisey/payflow-simulator/internal/model"
	"github.com/seymourrisey/payflow-simulator/pkg/idgen"
)

type AuthRepository struct {
	db *pgxpool.Pool
}

func NewAuthRepository(db *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) CreateUserWithWallet(ctx context.Context, user *model.User) (*model.User, error) {
	// ── ACID TRANSACTION: insert user + wallet atomically ──
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Generate custom ID sebelum insert
	user.ID = idgen.NewUserID() // USR-A1B2C3D4E5F6

	// 1. Insert user
	err = tx.QueryRow(ctx, `
		INSERT INTO users (id, full_name, email, password_hash)
		VALUES ($1, $2, $3, $4)
		RETURNING full_name, email, created_at
	`, user.ID, user.FullName, user.Email, user.PasswordHash).
		Scan(&user.FullName, &user.Email, &user.CreatedAt)
	if err != nil {
		return nil, err
	}

	// 2. Auto-create wallet untuk user baru
	walletID := idgen.NewWalletID() // WLT-A1B2C3D4E5F6
	_, err = tx.Exec(ctx, `
		INSERT INTO wallets (id, user_id, balance, currency)
		VALUES ($1, $2, 0.00, 'IDR')
	`, walletID, user.ID)
	if err != nil {
		return nil, err
	}

	// 3. Commit — jika gagal di atas, Rollback otomatis via defer
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	return user, nil
}

func (r *AuthRepository) FindUserByEmail(ctx context.Context, email string) (*model.User, error) {
	user := &model.User{}
	err := r.db.QueryRow(ctx, `
		SELECT id, full_name, email, password_hash, created_at
		FROM users WHERE email = $1
	`, email).Scan(&user.ID, &user.FullName, &user.Email, &user.PasswordHash, &user.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil // user not found
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}
