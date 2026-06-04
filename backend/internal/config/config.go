package config

import (
	"os"
	"time"
)

type Config struct {
	AppEnv      string
	HTTPAddr    string
	DatabaseURL string

	JWTSecret string
	JWTTTL    time.Duration

	TelegramBotToken string
	ManagerChatID    string
}

func Load() Config {
	return Config{
		AppEnv:      getEnv("APP_ENV", "development"),
		HTTPAddr:    getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://icaris:icaris@localhost:5433/icaris_logistic?sslmode=disable"),

		// NOTE: MVP — dev-дефолт секрета; в проде JWT_SECRET обязателен; см. docs/tech-debt.md
		JWTSecret: getEnv("JWT_SECRET", "dev-insecure-secret-change-me"),
		JWTTTL:    getEnvDuration("JWT_TTL", 24*time.Hour),

		TelegramBotToken: getEnv("TELEGRAM_BOT_TOKEN", ""),
		ManagerChatID:    getEnv("MANAGER_CHAT_ID", ""),
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

// getEnvDuration парсит длительность в формате time.ParseDuration ("24h", "30m").
// Пустое или некорректное значение → fallback: config отдаёт наружу готовый тип, чтобы
// потребители не парсили строки повторно.
func getEnvDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	d, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}

	return d
}
