package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort         string
	AppEnv          string
	DBUrl           string
	JWTSecret       string
	JWTExpiry       int
	WebhookTimeout  int
	WebhookMaxRetry int
}

var App *Config

func LoadConfig() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	App = &Config{
		AppPort:         getEnv("APP_PORT", "8080"),
		AppEnv:          getEnv("APP_ENV", "development"),
		DBUrl:           mustGetEnv("DATABASE_URL"),
		JWTSecret:       mustGetEnv("JWT_SECRET"),
		JWTExpiry:       getEnvInt("JWT_EXPIRY_HOURS", 24),
		WebhookTimeout:  getEnvInt("WEBHOOK_TIMEOUT_SECONDS", 10),
		WebhookMaxRetry: getEnvInt("WEBHOOK_MAX_RETRIES", 3),
	}

	log.Printf("Config loaded | ENV: %s | PORT: %s", App.AppEnv, App.AppPort)
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func mustGetEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("Environment variable %s not set", key)
	}
	return val
}

func getEnvInt(key string, fallback int) int {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	num, err := strconv.Atoi(val)
	if err != nil {
		log.Printf("Warning: invalid value for %s, using default %d", key, fallback)
		return fallback
	}
	return num
}
