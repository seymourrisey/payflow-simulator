package model

import (
	"time"
)

type User struct {
	ID           string    `json:"id"` // USR-A1B2C3D4E5F6
	FullName     string    `json:"full_name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type Wallet struct {
	ID        string    `json:"id"` // WLT-A1B2C3D4E5F6
	UserID    string    `json:"user_id"`
	Balance   float64   `json:"balance"`
	Currency  string    `json:"currency"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Merchant struct {
	ID           string    `json:"id"` // MRC-A1B2C3D4E5F6
	MerchantName string    `json:"merchant_name"`
	WebhookURL   string    `json:"webhook_url,omitempty"`
	APIKey       string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type TransactionType string
type TransactionStatus string

const (
	TxTypePayment  TransactionType = "PAYMENT"
	TxTypeTopUp    TransactionType = "TOPUP"
	TxTypeTransfer TransactionType = "TRANSFER"

	TxStatusPending TransactionStatus = "PENDING"
	TxStatusSuccess TransactionStatus = "SUCCESS"
	TxStatusFailed  TransactionStatus = "FAILED"
	TxStatusExpired TransactionStatus = "EXPIRED"
)

type Transaction struct {
	ID                 string            `json:"id"` // TXN-20260307-A1B2C3D4
	ReferenceID        string            `json:"reference_id"`
	WalletID           string            `json:"wallet_id"`
	SenderWalletID     *string           `json:"sender_wallet_id,omitempty"`
	ReceiverMerchantID *string           `json:"receiver_merchant_id,omitempty"`
	Type               TransactionType   `json:"type"`
	Amount             float64           `json:"amount"`
	Fee                float64           `json:"fee"`
	Status             TransactionStatus `json:"status"`
	Metadata           map[string]any    `json:"metadata,omitempty"`
	ExpiredAt          *time.Time        `json:"expired_at,omitempty"`
	CreatedAt          time.Time         `json:"created_at"`
}

type TopUpRequest struct {
	ID             string            `json:"id"` // TUP-20260307-A1B2C3D4
	WalletID       string            `json:"wallet_id"`
	Amount         float64           `json:"amount"`
	PaymentChannel string            `json:"payment_channel"`
	Status         TransactionStatus `json:"status"`
	ExpiredAt      *time.Time        `json:"expired_at,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
}

type WebhookLog struct {
	ID             string    `json:"id"` // WHL-A1B2C3D4E5F6
	MerchantID     string    `json:"merchant_id"`
	TransactionID  string    `json:"transaction_id"`
	Payload        any       `json:"payload"`
	ResponseStatus *int      `json:"response_status,omitempty"`
	RetryCount     int       `json:"retry_count"`
	SentAt         time.Time `json:"sent_at"`
}
