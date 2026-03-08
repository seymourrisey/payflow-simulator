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
)

func main() {
	// 1. Load config & connect DB
	config.Load()
	config.ConnectDB()
	defer config.CloseDB()

	// 2. Init dependencies (manual DI)
	authRepo := repository.NewAuthRepository(config.DB)
	walletRepo := repository.NewWalletRepository(config.DB)
	txRepo := repository.NewTransactionRepository(config.DB)

	authService := service.NewAuthService(authRepo)
	paymentService := service.NewPaymentService(txRepo, walletRepo)

	authHandler := handler.NewAuthHandler(authService)
	payHandler := handler.NewPayHandler(paymentService)

	// 3. Setup Fiber
	app := fiber.New(fiber.Config{
		AppName:      "payflow-simulator v1.0",
		ErrorHandler: customErrorHandler,
	})

	// 4. Global middlewares
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${method} ${path} → ${status} (${latency})\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: config.App.AllowOrigins,
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Authorization, X-Idempotency-Key",
	}))

	// 5. Routes
	api := app.Group("/api")

	// Public routes
	auth := api.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)

	// Protected routes (JWT required)
	protected := api.Group("/", middleware.Protected())

	wallet := protected.Group("/wallet")
	wallet.Get("/", payHandler.GetWallet)
	wallet.Post("/topup", payHandler.TopUp)

	payment := protected.Group("/payment")
	payment.Post("/qr", payHandler.GenerateQR)
	payment.Post("/pay", payHandler.Pay)

	protected.Get("/transactions", payHandler.GetHistory)

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"app":    "payflow-simulator",
		})
	})

	// 6. Start server (graceful shutdown)
	go func() {
		log.Printf(" PAYFLOW SIMULATOR running on :%s [%s]", config.App.AppPort, config.App.AppEnv)
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
