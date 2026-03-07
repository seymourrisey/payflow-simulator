package dto

type RegisterRequest struct {
	FullName string `json:"full_name" validate:"required,min=2"`
	Email    string `json:"email"     validate:"required,email"`
	Password string `json:"password"  validate:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  UserProfile `json:"user"`
}

type UserProfile struct {
	ID       int    `json:"id"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
}

// wallet
type WalletResponse struct {
	ID       int     `json:"id"`
	Balance  float64 `json:"balance"`
	Currency string  `json:"currency"`
}

// topUp
type TopUpRequest struct {
	Amount         float64 `json:"amount"          validate:"required,gt=0"`
	PaymentChannel string  `json:"payment_channel" validate:"required,oneof=BANK_TRANSFER VIRTUAL_ACCOUNT"`
}

type TopUpResponse struct {
	TopUpID        int     `json:"top_up_id"`
	Amount         float64 `json:"amount"`
	PaymentChannel string  `json:"payment_channel"`
	Status         string  `json:"status"`
	ExpiredAt      string  `json:"expired_at"`
}

// payment
type PaymentRequest struct {
	MerchantID  int     `json:"merchant_id"  validate:"required"`
	Amount      float64 `json:"amount"       validate:"required,gt=0"`
	Description string  `json:"description"`
	// IdempotencyKey dikirim via header X-Idempotency-Key
}

type PaymentResponse struct {
	ReferenceID string  `json:"reference_id"`
	Amount      float64 `json:"amount"`
	Fee         float64 `json:"fee"`
	Status      string  `json:"status"`
	QRData      string  `json:"qr_data"` // string yang di-encode ke QR
}

// QR generate
type GenerateQRRequest struct {
	MerchantID  int     `json:"merchant_id"  validate:"required"`
	Amount      float64 `json:"amount"       validate:"required,gt=0"`
	Description string  `json:"description"`
}

type GenerateQRResponse struct {
	QRData      string `json:"qr_data"`      // payload string
	ReferenceID string `json:"reference_id"` // unique ref untuk payment ini
	ExpiredAt   string `json:"expired_at"`
}

// transaction history
type TransactionItem struct {
	ReferenceID string  `json:"reference_id"`
	Type        string  `json:"type"`
	Amount      float64 `json:"amount"`
	Fee         float64 `json:"fee"`
	Status      string  `json:"status"`
	Description string  `json:"description,omitempty"`
	CreatedAt   string  `json:"created_at"`
}

type TransactionListResponse struct {
	Data  []TransactionItem `json:"data"`
	Total int               `json:"total"`
}
