package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

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

	// 5. Setup Fiber
	app := fiber.New(fiber.Config{
		AppName:      "payflow-simulator v1.0",
		ErrorHandler: customErrorHandler,
	})

	// 6. Global middlewares
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${method} ${path} → ${status} (${latency})\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: config.App.AllowOrigins,
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Authorization, X-Idempotency-Key",
	}))

	// 7. Routes
	api := app.Group("/api")

	// Public
	auth := api.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)

	// Protected (JWT required)
	protected := api.Group("/", middleware.Protected())

	wallet := protected.Group("/wallet")
	wallet.Get("/", payHandler.GetWallet)
	wallet.Post("/topup", payHandler.TopUp)

	payment := protected.Group("/payment")
	payment.Post("/qr", payHandler.GenerateQR)
	payment.Post("/pay", payHandler.Pay)

	protected.Get("/transactions", payHandler.GetHistory)

	// Webhook panel routes
	webhooks := protected.Group("/webhooks")
	webhooks.Get("/", webhookHandler.GetLogs)
	webhooks.Get("/stats", webhookHandler.GetStats)
	webhooks.Get("/merchants", webhookHandler.GetMerchants)

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "app": "payflow-simulator"})
	})

	// 8. Start server (graceful shutdown)
	go func() {
		log.Printf("🚀 PAYFLOW SIMULATOR running on :%s [%s]", config.App.AppPort, config.App.AppEnv)
		if err := app.Listen(":" + config.App.AppPort); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down gracefully...")
	if err := app.Shutdown(); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}
}

func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}
	return c.Status(code).JSON(fiber.Map{
		"success": false,
		"error":   err.Error(),
	})
}
