package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"icaris-logistic/backend/internal/bot"
	"icaris-logistic/backend/internal/config"
	"icaris-logistic/backend/internal/db"
	apphttp "icaris-logistic/backend/internal/http"
	"icaris-logistic/backend/internal/repository"
	"icaris-logistic/backend/internal/service"
)

func main() {
	// В dev подхватываем backend/.env, чтобы `go run` видел переменные без ручного export.
	// godotenv НЕ переопределяет уже заданные env-переменные, поэтому в проде (где их инжектит
	// платформа или docker-compose) вызов — безопасный no-op. Отсутствие файла ошибкой не считаем.
	_ = godotenv.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		logger.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

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
	clientRepo := repository.NewClientRepository(gdb)
	shipmentRepo := repository.NewShipmentRepository(gdb)
	messageRepo := repository.NewMessageRepository(gdb)

	notifier, err := bot.New(cfg.TelegramBotToken, cfg.ManagerChatID)
	if err != nil {
		logger.Error("init bot", "error", err)
		os.Exit(1)
	}

	// bg учитывает фоновые задачи (уведомления, цикл бота), чтобы дренировать их при остановке.
	bg := service.NewBackground()

	leadService := service.NewLeadService(leadRepo, notifier, bg)
	authService := service.NewAuthService(managerRepo, cfg.JWTSecret, cfg.JWTTTL)
	clientService := service.NewClientService(clientRepo, cfg.TelegramBotToken, cfg.JWTSecret, cfg.JWTTTL)
	shipmentService := service.NewShipmentService(shipmentRepo, clientRepo, notifier, bg)
	messageService := service.NewMessageService(messageRepo, shipmentRepo, clientRepo, notifier, notifier, bg)

	// Бот принимает команды клиентов (/start, /status) в long polling, пока жив ctx. Через bg,
	// чтобы при остановке дождаться завершения текущей команды (Run выходит по ctx.Done).
	bg.Go(func() {
		notifier.Run(ctx, bot.RunDeps{Registrar: clientService, Lister: shipmentService})
	})

	isDev := cfg.AppEnv == "development"
	router := apphttp.NewRouter(apphttp.RouterDeps{
		DB:              gdb,
		LeadService:     leadService,
		AuthService:     authService,
		ClientService:   clientService,
		ShipmentService: shipmentService,
		MessageService:  messageService,
		JWTSecret:       cfg.JWTSecret,
		AllowedOrigins:  cfg.AllowedOrigins,
		// В dev без явного списка отдаём «*» для удобства; вне dev — строго белый список.
		AllowAnyOrigin: isDev && len(cfg.AllowedOrigins) == 0,
		ReleaseMode:    !isDev,
	})

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
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

	// Сначала останавливаем HTTP-сервер (дожидается in-flight запросов — они и порождают
	// последние фоновые задачи), затем дренируем фон. Порядок важен: новые задачи в bg больше
	// не поступают, поэтому Wait не словит race по WaitGroup.
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown", "error", err)
	}
	if err := bg.Wait(shutdownCtx); err != nil {
		logger.Warn("background work not drained before timeout", "error", err)
	}

	logger.Info("api stopped")
}
