package config

import (
	"errors"
	"os"
	"strings"
	"time"
)

const (
	envDevelopment        = "development"
	devDefaultJWTSecret   = "dev-insecure-secret-change-me"
	devDefaultDatabaseURL = "postgres://icaris:icaris@localhost:5433/icaris_logistic?sslmode=disable"
)

type Config struct {
	AppEnv      string
	HTTPAddr    string
	DatabaseURL string

	JWTSecret string
	JWTTTL    time.Duration

	TelegramBotToken string
	ManagerChatID    string

	// TelegramWebhookURL — публичный HTTPS-URL ручки /api/telegram/webhook. Задан =>
	// бот получает апдейты webhook'ом вместо long polling. Secret обязателен вместе с URL:
	// Telegram шлёт его в заголовке X-Telegram-Bot-Api-Secret-Token, по нему ручка
	// отличает Telegram от постороннего трафика.
	TelegramWebhookURL    string
	TelegramWebhookSecret string

	// AllowedOrigins — белый список origin'ов для CORS (env ALLOWED_ORIGINS, через запятую).
	// Пусто вне development => браузерные кросс-доменные запросы запрещены (безопасный дефолт).
	AllowedOrigins []string

	// TrustedProxies — IP/CIDR прокси, чьим X-Forwarded-For верим (env TRUSTED_PROXIES,
	// через запятую). Пусто => приватные диапазоны (Railway/докер-сеть), см. router.
	TrustedProxies []string
}

func Load() Config {
	return Config{
		AppEnv:      getEnv("APP_ENV", envDevelopment),
		HTTPAddr:    httpAddr(),
		DatabaseURL: getEnv("DATABASE_URL", devDefaultDatabaseURL),

		// Dev-дефолт секрета удобен локально; Validate() запрещает его вне development.
		JWTSecret: getEnv("JWT_SECRET", devDefaultJWTSecret),
		JWTTTL:    getEnvDuration("JWT_TTL", 24*time.Hour),

		// TrimSpace: секреты часто прилетают из env/дашбордов с хвостовым "\n" (копипаст,
		// переменные Railway) — для токена и числового chat_id это ломает парсинг на старте.
		TelegramBotToken: strings.TrimSpace(getEnv("TELEGRAM_BOT_TOKEN", "")),
		ManagerChatID:    strings.TrimSpace(getEnv("MANAGER_CHAT_ID", "")),

		TelegramWebhookURL:    strings.TrimSpace(getEnv("TELEGRAM_WEBHOOK_URL", "")),
		TelegramWebhookSecret: strings.TrimSpace(getEnv("TELEGRAM_WEBHOOK_SECRET", "")),

		AllowedOrigins: getEnvList("ALLOWED_ORIGINS"),
		TrustedProxies: getEnvList("TRUSTED_PROXIES"),
	}
}

// Validate отвергает небезопасные dev-дефолты вне development-окружения.
// Возвращает все найденные проблемы разом (errors.Join), чтобы их чинили списком,
// а не по одной перезапуском. В development дефолты допустимы намеренно.
func (c Config) Validate() error {
	var errs []error

	// Webhook — опциональная фича, но заданная наполовину конфигурация это ошибка в любом
	// окружении: без секрета ручку откроет кто угодно, без https Telegram URL не примет.
	if c.TelegramWebhookURL != "" {
		if !strings.HasPrefix(c.TelegramWebhookURL, "https://") {
			errs = append(errs, errors.New("TELEGRAM_WEBHOOK_URL must be an https URL"))
		}
		if c.TelegramWebhookSecret == "" {
			errs = append(errs, errors.New("TELEGRAM_WEBHOOK_SECRET is required when TELEGRAM_WEBHOOK_URL is set"))
		}
	}

	if c.AppEnv == envDevelopment {
		return errors.Join(errs...)
	}
	if c.JWTSecret == "" || c.JWTSecret == devDefaultJWTSecret {
		errs = append(errs, errors.New("JWT_SECRET must be set to a non-default value outside development"))
	}
	if c.DatabaseURL == "" || c.DatabaseURL == devDefaultDatabaseURL {
		errs = append(errs, errors.New("DATABASE_URL must be set outside development"))
	}
	if strings.Contains(c.DatabaseURL, "sslmode=disable") {
		errs = append(errs, errors.New("DATABASE_URL must not use sslmode=disable outside development"))
	}
	if c.TelegramBotToken == "" || c.ManagerChatID == "" {
		errs = append(errs, errors.New("TELEGRAM_BOT_TOKEN and MANAGER_CHAT_ID are required outside development"))
	}

	return errors.Join(errs...)
}

// getEnvList парсит env как список через запятую, отбрасывая пустые элементы. Пустой env → nil.
func getEnvList(key string) []string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}

	return out
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

// httpAddr выбирает адрес прослушивания. Приоритет: явный HTTP_ADDR (полный адрес вида ":8080"),
// затем PORT (его назначают облачные платформы — Render/Heroku — и сервис ОБЯЗАН слушать именно
// его, иначе health-check платформы не достучится), затем dev-дефолт.
func httpAddr() string {
	if addr := os.Getenv("HTTP_ADDR"); addr != "" {
		return addr
	}
	if port := os.Getenv("PORT"); port != "" {
		return ":" + port
	}

	return ":8080"
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
