package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/seymourrisey/payflow-simulator/config"
	"github.com/seymourrisey/payflow-simulator/internal/handler"
	"github.com/seymourrisey/payflow-simulator/internal/middleware"
	"github.com/seymourrisey/payflow-simulator/internal/repository"
	"github.com/seymourrisey/payflow-simulator/internal/service"
	"github.com/seymourrisey/payflow-simulator/pkg/webhook"
)

func main() {
	// 1. Load config & connect DB
	config.Load()
	config.ConnectDB()
	defer config.CloseDB()

	// 2. Init repositories
	authRepo := repository.NewAuthRepository(config.DB)
	walletRepo := repository.NewWalletRepository(config.DB)
	txRepo := repository.NewTransactionRepository(config.DB)
	webhookRepo := repository.NewWebhookRepository(config.DB)

	// 3. Init services (webhook dulu karena dibutuhkan paymentService)
	dispatcher := webhook.NewDispatcher()
	webhookService := service.NewWebhookService(webhookRepo, dispatcher)
	authService := service.NewAuthService(authRepo)
	paymentService := service.NewPaymentService(txRepo, walletRepo, webhookService)

	// 4. Init handlers
	authHandler := handler.NewAuthHandler(authService)
	payHandler := handler.NewPayHandler(paymentService)
	webhookHandler := handler.NewWebhookHandler(webhookService)

	// 5. Setup Gin (gin.Default() includes logger and recovery middleware)
	router := gin.Default()

	// 6. Global middleware - CORS
	// Parse AllowOrigins dari string comma-separated
	allowOrigins := []string{}
	for _, origin := range strings.Split(config.App.AllowOrigins, ",") {
		allowOrigins = append(allowOrigins, strings.TrimSpace(origin))
	}

	router.Use(cors.New(cors.Config{
		AllowOrigins: allowOrigins,
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{"Origin", "Content-Type", "Authorization", "X-Idempotency-Key"},
	}))

	// 7. Routes
	api := router.Group("/api")

	// Public routes
	auth := api.Group("/auth")
	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)

	// Protected routes (JWT required)
	protected := api.Group("", middleware.Protected())

	// Logout — protected karena butuh verifikasi token dulu
	protected.POST("/auth/logout", authHandler.Logout)

	wallet := protected.Group("/wallet")
	wallet.GET("/", payHandler.GetWallet)
	wallet.POST("/topup", payHandler.TopUp)

	payment := protected.Group("/payment")
	payment.POST("/qr", payHandler.GenerateQR)
	payment.POST("/pay", payHandler.Pay)

	protected.GET("/transactions", payHandler.GetHistory)

	// Webhook panel routes
	webhooks := protected.Group("/webhooks")
	webhooks.GET("/", webhookHandler.GetLogs)
	webhooks.GET("/stats", webhookHandler.GetStats)
	webhooks.GET("/merchants", webhookHandler.GetMerchants)

	// ── Built-in Webhook Receiver (PUBLIC - tidak perlu JWT) ──
	// Merchant webhook URL diarahkan ke sini untuk local testing
	router.POST("/webhook/receive", webhookHandler.Receive)

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "app": "payflow-simulator"})
	})

	// 8. Start server
	log.Printf("(˶˃ᆺ˂˶) PAYFLOW SIMULATOR running on :%s [%s]", config.App.AppPort, config.App.AppEnv)
	if err := router.Run(":" + config.App.AppPort); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
