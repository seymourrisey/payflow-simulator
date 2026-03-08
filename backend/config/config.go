package config

import (
	"log"
	"os"
	"path/filepath"
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
	AllowOrigins    string
}

var App *Config

func Load() {
	// Cari .env file di parent directories
	envPath := findEnvFile()
	if envPath != "" {
		if err := godotenv.Load(envPath); err != nil {
			log.Printf("Error loading .env file: %v", err)
			log.Fatal("Failed to load .env file")
		}
	} else {
		log.Println("Warning: .env file not found, using environment variables")
	}

	App = &Config{
		AppPort:         getEnv("APP_PORT", "8080"),
		AppEnv:          getEnv("APP_ENV", "development"),
		DBUrl:           mustGetEnv("DATABASE_URL"),
		JWTSecret:       mustGetEnv("JWT_SECRET"),
		JWTExpiry:       getEnvInt("JWT_EXPIRY_HOURS", 24),
		WebhookTimeout:  getEnvInt("WEBHOOK_TIMEOUT_SECONDS", 10),
		WebhookMaxRetry: getEnvInt("WEBHOOK_MAX_RETRIES", 3),
		AllowOrigins:    getEnv("ALLOW_ORIGINS", "http://localhost:5173"),
	}

	log.Printf("Config loaded | ENV: %s | PORT: %s", App.AppEnv, App.AppPort)
}

func findEnvFile() string {
	// Mulai dari working directory, cari .env ke atas
	dir, _ := os.Getwd()

	for {
		envPath := filepath.Join(dir, ".env")
		if _, err := os.Stat(envPath); err == nil {
			log.Printf("Found .env at: %s", envPath)
			return envPath
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Sudah sampai root
			break
		}
		dir = parent
	}

	return ""
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
