package model

import (
	"time"
)

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // never expose in json
	CreatedAt time.Time `json:"created_at"`
}

type Wallet struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Balance   float64   `json:"balance"`
	Currency  string    `json:"currency"`
	CreatedAt time.Time `json:"created_at"`
}

type Merchant struct {
	ID           int       `json:"id"`
	MerchantName string    `json:"merchant_name"`
	WebhookURL   string    `json:"webhook_url,omitempty"`
	APIKey       string    `json:"-"` // sensitive
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
	ID                 int               `json:"id"`
	ReferenceID        string            `json:"reference_id"`
	WalletID           int               `json:"wallet_id"`
	SenderWalletID     *int              `json:"sender_wallet_id,omitempty"`
	ReceiverMerchantID *int              `json:"receiver_merchant_id,omitempty"`
	Type               TransactionType   `json:"type"`
	Amount             float64           `json:"amount"`
	Fee                float64           `json:"fee"`
	Status             TransactionStatus `json:"status"`
	Metadata           map[string]any    `json:"metadata,omitempty"`
	ExpiredAt          *time.Time        `json:"expired_at,omitempty"`
	CreatedAt          time.Time         `json:"created_at"`
}

type TopUpRequest struct {
	ID             int               `json:"id"`
	WalletID       int               `json:"wallet_id"`
	Amount         float64           `json:"amount"`
	PaymentChannel string            `json:"payment_channel"`
	Status         TransactionStatus `json:"status"`
	ExpiredAt      *time.Time        `json:"expired_at,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
}

type WebhookLog struct {
	ID             int       `json:"id"`
	MerchantID     int       `json:"merchant_id"`
	TransactionID  int       `json:"transaction_id"`
	Payload        any       `json:"payload"`
	ResponseStatus *int      `json:"response_status,omitempty"`
	RetryCount     int       `json:"retry_count"`
	SentAt         time.Time `json:"sent_at"`
}
