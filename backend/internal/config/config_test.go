package config

import "testing"

func TestValidateAllowsDefaultsInDevelopment(t *testing.T) {
	cfg := Config{
		AppEnv:      envDevelopment,
		JWTSecret:   devDefaultJWTSecret,
		DatabaseURL: devDefaultDatabaseURL,
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("development should allow dev defaults, got: %v", err)
	}
}

func TestValidateRejectsDefaultSecretOutsideDevelopment(t *testing.T) {
	cfg := Config{
		AppEnv:           "production",
		JWTSecret:        devDefaultJWTSecret,
		DatabaseURL:      "postgres://user:pass@db:5432/app?sslmode=require",
		TelegramBotToken: "token",
		ManagerChatID:    "123",
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("production with dev-default JWT secret should be rejected")
	}
}

func TestValidateRejectsSSLDisabledOutsideDevelopment(t *testing.T) {
	cfg := Config{
		AppEnv:           "production",
		JWTSecret:        "a-real-secret",
		DatabaseURL:      "postgres://user:pass@db:5432/app?sslmode=disable",
		TelegramBotToken: "token",
		ManagerChatID:    "123",
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("production with sslmode=disable should be rejected")
	}
}

func TestValidatePassesWithProperProductionConfig(t *testing.T) {
	cfg := Config{
		AppEnv:           "production",
		JWTSecret:        "a-real-secret",
		DatabaseURL:      "postgres://user:pass@db:5432/app?sslmode=require",
		TelegramBotToken: "token",
		ManagerChatID:    "123",
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("proper production config should pass, got: %v", err)
	}
}

func TestHTTPAddrFallsBackToPortWhenAddrUnset(t *testing.T) {
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("PORT", "10000")

	if got := httpAddr(); got != ":10000" {
		t.Fatalf("expected PORT to drive the listen address, got %q", got)
	}
}

func TestValidateRejectsWebhookURLWithoutSecretEvenInDevelopment(t *testing.T) {
	cfg := Config{
		AppEnv:             envDevelopment,
		JWTSecret:          devDefaultJWTSecret,
		DatabaseURL:        devDefaultDatabaseURL,
		TelegramWebhookURL: "https://api.example.com/api/telegram/webhook",
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("webhook URL without secret should be rejected in any environment")
	}
}

func TestValidateRejectsNonHTTPSWebhookURL(t *testing.T) {
	cfg := Config{
		AppEnv:                envDevelopment,
		JWTSecret:             devDefaultJWTSecret,
		DatabaseURL:           devDefaultDatabaseURL,
		TelegramWebhookURL:    "http://api.example.com/api/telegram/webhook",
		TelegramWebhookSecret: "s3cret",
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("non-https webhook URL should be rejected")
	}
}

func TestValidateAllowsCompleteWebhookConfigInDevelopment(t *testing.T) {
	cfg := Config{
		AppEnv:                envDevelopment,
		JWTSecret:             devDefaultJWTSecret,
		DatabaseURL:           devDefaultDatabaseURL,
		TelegramWebhookURL:    "https://api.example.com/api/telegram/webhook",
		TelegramWebhookSecret: "s3cret",
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("complete webhook config should pass in development, got: %v", err)
	}
}
