package handler

import (
	"encoding/json"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
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
func (h *WebhookHandler) GetLogs(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}

	result, err := h.webhookService.GetLogs(c.Context(), page, limit)
	if err != nil {
		return response.InternalError(c, err)
	}
	return response.OK(c, "Webhook logs retrieved", result)
}

// GET /api/webhooks/stats — stats ringkasan
func (h *WebhookHandler) GetStats(c *fiber.Ctx) error {
	stats, err := h.webhookService.GetStats(c.Context())
	if err != nil {
		return response.InternalError(c, err)
	}
	return response.OK(c, "Webhook stats retrieved", stats)
}

// GET /api/webhooks/merchants — list merchant untuk filter dropdown
func (h *WebhookHandler) GetMerchants(c *fiber.Ctx) error {
	merchants, err := h.webhookService.GetMerchants(c.Context())
	if err != nil {
		return response.InternalError(c, err)
	}
	return response.OK(c, "Merchants retrieved", merchants)
}

// POST /webhook/receive — built-in receiver untuk local testing
// Merchant webhook URL diarahkan ke sini
func (h *WebhookHandler) Receive(c *fiber.Ctx) error {
	// Log semua headers
	log.Printf(" ◝(ᵔᗜᵔ)◜ Webhook received!◝(ᵔᗜᵔ)◜")
	log.Printf("   Signature : %s", c.Get("X-Payflow-Signature"))
	log.Printf("   Timestamp : %s", c.Get("X-Payflow-Timestamp"))
	log.Printf("   User-Agent: %s", c.Get("User-Agent"))

	// Parse body
	var payload map[string]any
	if err := json.Unmarshal(c.Body(), &payload); err != nil {
		log.Printf("   Body (raw): %s", string(c.Body()))
	} else {
		prettyJSON, _ := json.MarshalIndent(payload, "   ", "  ")
		log.Printf("   Payload:\n   %s", string(prettyJSON))
	}

	// Return 200 agar webhook dianggap delivered
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"received": true,
		"message":  "Webhook received by PayFlow local receiver",
	})
}
