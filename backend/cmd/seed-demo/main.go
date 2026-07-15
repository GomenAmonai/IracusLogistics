package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/shopspring/decimal"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"icaris-logistic/backend/internal/config"
	"icaris-logistic/backend/internal/db"
	"icaris-logistic/backend/internal/domain"
	"icaris-logistic/backend/internal/repository"
	"icaris-logistic/backend/internal/service"
	"icaris-logistic/backend/internal/token"
)

const (
	demoManagerEmail    = "demo@icaris.local"
	demoManagerPassword = "demo-local-only"
	demoTelegramID      = int64(9000000001)
	demoMarker          = "[icaris-local-demo]"
	demoTrackingKey     = "ICR-DEM0000000"
)

func main() {
	frontendURL := flag.String("frontend-url", "http://127.0.0.1:5173", "URL локального frontend")
	flag.Parse()

	cfg := config.Load()
	if cfg.AppEnv != "development" {
		log.Fatal("seed-demo разрешён только при APP_ENV=development")
	}

	ctx := context.Background()
	gdb, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}

	manager, err := ensureManager(ctx, gdb)
	if err != nil {
		log.Fatalf("seed manager: %v", err)
	}

	lead, err := ensureLead(ctx, gdb)
	if err != nil {
		log.Fatalf("seed lead: %v", err)
	}

	clientService := service.NewClientService(
		repository.NewClientRepository(gdb),
		"",
		cfg.JWTSecret,
		cfg.JWTTTL,
	)
	client, err := clientService.Register(ctx, demoTelegramID, "icaris_demo", "Демо-клиент", &lead.ID)
	if err != nil {
		log.Fatalf("seed client: %v", err)
	}
	if client.LeadID == nil {
		if err := gdb.WithContext(ctx).Model(client).Update("lead_id", lead.ID).Error; err != nil {
			log.Fatalf("link client to lead: %v", err)
		}
		client.LeadID = &lead.ID
	}

	shipment, err := ensureShipment(ctx, gdb, manager, client)
	if err != nil {
		log.Fatalf("seed shipment: %v", err)
	}
	if err := ensureConversation(ctx, gdb, manager, client, shipment); err != nil {
		log.Fatalf("seed conversation: %v", err)
	}
	if err := ensurePayment(ctx, gdb, manager, shipment); err != nil {
		log.Fatalf("seed payment: %v", err)
	}

	clientToken, err := token.Issue([]byte(cfg.JWTSecret), client.ID, domain.RoleClient, cfg.JWTTTL)
	if err != nil {
		log.Fatalf("issue client token: %v", err)
	}

	baseURL := strings.TrimRight(*frontendURL, "/")
	fmt.Println("Демо-сценарий готов.")
	fmt.Printf("Менеджер: %s/manager.html\n", baseURL)
	fmt.Printf("  email: %s\n", demoManagerEmail)
	fmt.Printf("  password: %s\n", demoManagerPassword)
	fmt.Printf("Клиент: %s/webapp.html?token=%s\n", baseURL, url.QueryEscape(clientToken))
	fmt.Printf("Груз: %s (%s → %s)\n", shipment.TrackingKey, shipment.FromCity, shipment.ToCity)
}

func ensureManager(ctx context.Context, gdb *gorm.DB) (*domain.Manager, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(demoManagerPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	var manager domain.Manager
	err = gdb.WithContext(ctx).First(&manager, "email = ?", demoManagerEmail).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		manager = domain.Manager{
			Email:    demoManagerEmail,
			Name:     "Демо-менеджер",
			Password: string(hash),
		}
		if err := gdb.WithContext(ctx).Create(&manager).Error; err != nil {
			return nil, err
		}
	case err != nil:
		return nil, err
	default:
		if err := gdb.WithContext(ctx).Model(&manager).Updates(map[string]any{
			"name":     "Демо-менеджер",
			"password": string(hash),
		}).Error; err != nil {
			return nil, err
		}
	}

	return &manager, nil
}

func ensureLead(ctx context.Context, gdb *gorm.DB) (*domain.Lead, error) {
	var lead domain.Lead
	err := gdb.WithContext(ctx).First(&lead, "phone = ?", demoMarker).Error
	if err == nil {
		return &lead, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	leads := service.NewLeadService(repository.NewLeadRepository(gdb), nil, nil)
	return leads.Create(ctx, service.CreateLeadInput{
		Name:                 "Демо-клиент",
		Phone:                demoMarker,
		FromCity:             "Гуанчжоу",
		ToCity:               "Москва",
		Weight:               nullableDecimal("820"),
		Volume:               nullableDecimal("4.2"),
		CargoType:            "Электроника",
		Comment:              "Безопасные локальные данные для проверки полного сценария",
		PrivacyNoticeVersion: "local-demo-v1",
	})
}

func ensureShipment(
	ctx context.Context,
	gdb *gorm.DB,
	manager *domain.Manager,
	client *domain.Client,
) (*domain.Shipment, error) {
	var shipment domain.Shipment
	err := gdb.WithContext(ctx).First(&shipment, "tracking_key = ?", demoTrackingKey).Error
	if err == nil {
		return &shipment, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	clients := repository.NewClientRepository(gdb)
	shipmentRepo := repository.NewShipmentRepository(gdb)
	shipments := service.NewShipmentService(shipmentRepo, clients, nil, nil)

	// Ранние версии seed-demo создавали случайный tracking key. Подхватываем самый первый
	// груз выделенного demo-клиента и мигрируем его на стабильный ключ вместо дублирования.
	err = gdb.WithContext(ctx).
		Where("client_id = ?", client.ID).
		Order("created_at asc").
		First(&shipment).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		shipment = domain.Shipment{
			ClientID:      client.ID,
			ManagerID:     manager.ID,
			TrackingKey:   demoTrackingKey,
			Lane:          domain.LaneWhite,
			Status:        domain.ShipmentStatusPending,
			StatusComment: demoMarker,
			Weight:        nullableDecimal("820"),
			Volume:        nullableDecimal("4.2"),
			FromCity:      "Гуанчжоу",
			ToCity:        "Москва",
			Price:         nullableDecimal("2450"),
			Currency:      "USD",
		}
		if err := shipmentRepo.Create(ctx, &shipment); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	} else if err := gdb.WithContext(ctx).Model(&shipment).
		Update("tracking_key", demoTrackingKey).Error; err != nil {
		return nil, err
	}

	if shipment.Status != domain.ShipmentStatusPending {
		return &shipment, nil
	}
	updated, err := shipments.UpdateStatus(
		ctx,
		shipment.ID,
		manager.ID,
		domain.ShipmentStatusInTransit,
		"Груз передан перевозчику и следует по маршруту",
	)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func ensureConversation(
	ctx context.Context,
	gdb *gorm.DB,
	manager *domain.Manager,
	client *domain.Client,
	shipment *domain.Shipment,
) error {
	var count int64
	if err := gdb.WithContext(ctx).Model(&domain.Message{}).
		Where("shipment_id = ? AND text LIKE ?", shipment.ID, demoMarker+"%").
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	shipmentRepo := repository.NewShipmentRepository(gdb)
	clients := repository.NewClientRepository(gdb)
	messages := service.NewMessageService(
		repository.NewMessageRepository(gdb),
		shipmentRepo,
		clients,
		nil,
		nil,
		nil,
	)
	if _, err := messages.SendFromClient(ctx, shipment.ID, client.ID, demoMarker+" Когда ожидается прибытие?"); err != nil {
		return err
	}
	_, err := messages.SendFromManager(
		ctx,
		shipment.ID,
		manager.ID,
		demoMarker+" Ориентировочное прибытие — через 18 дней.",
	)
	return err
}

func ensurePayment(
	ctx context.Context,
	gdb *gorm.DB,
	manager *domain.Manager,
	shipment *domain.Shipment,
) error {
	var count int64
	if err := gdb.WithContext(ctx).Model(&domain.Payment{}).
		Where("shipment_id = ? AND comment = ?", shipment.ID, demoMarker).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	shipmentRepo := repository.NewShipmentRepository(gdb)
	payments := service.NewPaymentService(
		repository.NewPaymentRepository(gdb),
		shipmentRepo,
		repository.NewClientRepository(gdb),
		nil,
		nil,
	)
	_, err := payments.Create(ctx, shipment.ID, manager.ID, service.CreatePaymentInput{
		Amount:   decimal.NewFromInt(1200),
		Currency: "USD",
		Channel:  domain.PaymentChannelBankTransfer,
		Status:   domain.PaymentStatusPending,
		Comment:  demoMarker,
	})
	return err
}

func nullableDecimal(raw string) decimal.NullDecimal {
	return decimal.NullDecimal{Decimal: decimal.RequireFromString(raw), Valid: true}
}
