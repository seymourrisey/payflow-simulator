package handler

import (
	"encoding/json"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/seymourrisey/payflow-simulator/internal/service"
	"github.com/seymourrisey/payflow-simulator/pkg/response"
)

type WebhookHandler struct {
	webhookService *service.WebhookService
}

func NewWebhookHandler(webhookService *service.WebhookService) *WebhookHandler {
	return &WebhookHandler{webhookService: webhookService}
}

// GET /api/webhooks — semua webhook logs (panel)
func (h *WebhookHandler) GetLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}

	result, err := h.webhookService.GetLogs(c.Request.Context(), page, limit)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.OK(c, "Webhook logs retrieved", result)
}

// GET /api/webhooks/stats — stats ringkasan
func (h *WebhookHandler) GetStats(c *gin.Context) {
	stats, err := h.webhookService.GetStats(c.Request.Context())
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.OK(c, "Webhook stats retrieved", stats)
}

// GET /api/webhooks/merchants — list merchant untuk filter dropdown
func (h *WebhookHandler) GetMerchants(c *gin.Context) {
	merchants, err := h.webhookService.GetMerchants(c.Request.Context())
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.OK(c, "Merchants retrieved", merchants)
}

// POST /webhook/receive — built-in receiver untuk local testing
// Merchant webhook URL diarahkan ke sini
func (h *WebhookHandler) Receive(c *gin.Context) {
	// Log semua headers
	log.Printf(" ◝(ᵔᗜᵔ)◜ Webhook received!◝(ᵔᗜᵔ)◜")
	log.Printf("   Signature : %s", c.GetHeader("X-Payflow-Signature"))
	log.Printf("   Timestamp : %s", c.GetHeader("X-Payflow-Timestamp"))
	log.Printf("   User-Agent: %s", c.GetHeader("User-Agent"))

	// Get raw body
	body, _ := c.GetRawData()

	// Parse body
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Printf("   Body (raw): %s", string(body))
	} else {
		prettyJSON, _ := json.MarshalIndent(payload, "   ", "  ")
		log.Printf("   Payload:\n   %s", string(prettyJSON))
	}

	// Return 200 agar webhook dianggap delivered
	c.JSON(200, gin.H{
		"received": true,
		"message":  "Webhook received by PayFlow local receiver",
	})
}
