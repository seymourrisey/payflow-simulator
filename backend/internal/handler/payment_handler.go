package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/seymourrisey/payflow-simulator/internal/dto"
	"github.com/seymourrisey/payflow-simulator/internal/middleware"
	"github.com/seymourrisey/payflow-simulator/internal/service"
	"github.com/seymourrisey/payflow-simulator/pkg/response"
)

type PayHandler struct {
	paymentService *service.PaymentService
}

func NewPayHandler(paymentService *service.PaymentService) *PayHandler {
	return &PayHandler{paymentService: paymentService}
}

// GET /api/wallet — ambil saldo
func (h *PayHandler) GetWallet(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	wallet, err := h.paymentService.GetWallet(c.Context(), userID)
	if err != nil {
		return response.NotFound(c, "wallet")
	}

	return response.OK(c, "Wallet retrieved", wallet)
}

// POST /api/wallet/topup
func (h *PayHandler) TopUp(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	var req dto.TopUpRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	result, err := h.paymentService.TopUp(c.Context(), userID, &req)
	if err != nil {
		return response.InternalError(c, err)
	}

	return response.Created(c, "Top up successful", result)
}

// POST /api/payment/qr — generate QR
func (h *PayHandler) GenerateQR(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	var req dto.GenerateQRRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body: "+err.Error())
	}

	// Validasi manual field wajib
	if req.MerchantID == "" {
		return response.BadRequest(c, "merchant_id is required (string, e.g: MRC-TOKOPEDIA0001)")
	}
	if req.Amount <= 0 {
		return response.BadRequest(c, "amount must be greater than 0")
	}

	result, err := h.paymentService.GenerateQR(c.Context(), userID, &req)
	if err != nil {
		return response.InternalError(c, err)
	}

	return response.Created(c, "QR generated", result)
}

// POST /api/payment/pay — proses pembayaran
func (h *PayHandler) Pay(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	idempotencyKey := c.Get("X-Idempotency-Key")

	var req dto.PaymentRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body: "+err.Error())
	}

	// Validasi manual field wajib
	if req.MerchantID == "" {
		return response.BadRequest(c, "merchant_id is required (string, e.g: MRC-TOKOPEDIA0001)")
	}
	if req.Amount <= 0 {
		return response.BadRequest(c, "amount must be greater than 0")
	}

	result, err := h.paymentService.Pay(c.Context(), userID, &req, idempotencyKey)
	if err != nil {
		// Cek insufficient balance dengan aman (tanpa risiko panic)
		errMsg := err.Error()
		if len(errMsg) >= 22 && errMsg[:22] == "insufficient balance: " {
			return response.BadRequest(c, errMsg)
		}
		return response.InternalError(c, err)
	}

	return response.Created(c, "Payment successful", result)
}

// GET /api/transactions?page=1&limit=10
func (h *PayHandler) GetHistory(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}

	result, err := h.paymentService.GetHistory(c.Context(), userID, page, limit)
	if err != nil {
		return response.InternalError(c, err)
	}

	return response.OK(c, "Transactions retrieved", result)
}
