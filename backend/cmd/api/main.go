package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"iracus-logistic/backend/internal/bot"
	"iracus-logistic/backend/internal/config"
	"iracus-logistic/backend/internal/db"
	apphttp "iracus-logistic/backend/internal/http"
	"iracus-logistic/backend/internal/repository"
	"iracus-logistic/backend/internal/service"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	gdb, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("connect database", "error", err)
		os.Exit(1)
	}
	defer func() {
		if sqlDB, err := gdb.DB(); err == nil {
			_ = sqlDB.Close()
		}
	}()

	leadRepo := repository.NewLeadRepository(gdb)
	managerRepo := repository.NewManagerRepository(gdb)

	notifier, err := bot.New(cfg.TelegramBotToken, cfg.ManagerChatID)
	if err != nil {
		logger.Error("init bot", "error", err)
		os.Exit(1)
	}

	leadService := service.NewLeadService(leadRepo, notifier)
	authService := service.NewAuthService(managerRepo, cfg.JWTSecret, cfg.JWTTTL)

	router := apphttp.NewRouter(apphttp.RouterDeps{
		DB:          gdb,
		LeadService: leadService,
		AuthService: authService,
		JWTSecret:   cfg.JWTSecret,
	})

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("api started", "addr", cfg.HTTPAddr, "env", cfg.AppEnv)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("api stopped")
}
