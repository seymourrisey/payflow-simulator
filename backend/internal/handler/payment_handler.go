package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
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
func (h *PayHandler) GetWallet(c *gin.Context) {
	userID := middleware.GetUserID(c)

	wallet, err := h.paymentService.GetWallet(c.Request.Context(), userID)
	if err != nil {
		response.NotFound(c, "wallet")
		return
	}

	response.OK(c, "Wallet retrieved", wallet)
}

// POST /api/wallet/topup
func (h *PayHandler) TopUp(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req dto.TopUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	result, err := h.paymentService.TopUp(c.Request.Context(), userID, &req)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Created(c, "Top up successful", result)
}

// POST /api/payment/qr — generate QR
func (h *PayHandler) GenerateQR(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req dto.GenerateQRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body: "+err.Error())
		return
	}

	// Validasi manual field wajib
	if req.MerchantID == "" {
		response.BadRequest(c, "merchant_id is required (string, e.g: MRC-TOKOPEDIA0001)")
		return
	}
	if req.Amount <= 0 {
		response.BadRequest(c, "amount must be greater than 0")
		return
	}

	result, err := h.paymentService.GenerateQR(c.Request.Context(), userID, &req)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Created(c, "QR generated", result)
}

// POST /api/payment/pay — proses pembayaran
func (h *PayHandler) Pay(c *gin.Context) {
	userID := middleware.GetUserID(c)
	idempotencyKey := c.GetHeader("X-Idempotency-Key")

	var req dto.PaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body: "+err.Error())
		return
	}

	// Validasi manual field wajib
	if req.MerchantID == "" {
		response.BadRequest(c, "merchant_id is required (string, e.g: MRC-TOKOPEDIA0001)")
		return
	}
	if req.Amount <= 0 {
		response.BadRequest(c, "amount must be greater than 0")
		return
	}

	result, err := h.paymentService.Pay(c.Request.Context(), userID, &req, idempotencyKey)
	if err != nil {
		// Cek insufficient balance dengan aman (tanpa risiko panic)
		errMsg := err.Error()
		if len(errMsg) >= 22 && errMsg[:22] == "insufficient balance: " {
			response.BadRequest(c, errMsg)
			return
		}
		response.InternalError(c, err)
		return
	}

	response.Created(c, "Payment successful", result)
}

// GET /api/transactions?page=1&limit=10
func (h *PayHandler) GetHistory(c *gin.Context) {
	userID := middleware.GetUserID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}

	result, err := h.paymentService.GetHistory(c.Request.Context(), userID, page, limit)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.OK(c, "Transactions retrieved", result)
}
