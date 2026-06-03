package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"iracus-logistic/backend/internal/config"
	"iracus-logistic/backend/internal/db"
	apphttp "iracus-logistic/backend/internal/http"
	"iracus-logistic/backend/internal/shipment"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("connect database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	shipmentStore := shipment.NewPGStore(pool)
	if err := shipmentStore.EnsureSchema(ctx); err != nil {
		logger.Error("ensure shipment schema", "error", err)
		os.Exit(1)
	}

	shipmentService := shipment.NewService(shipmentStore)

	router := apphttp.NewRouter(apphttp.RouterDeps{
		DB:              pool,
		ShipmentService: shipmentService,
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
