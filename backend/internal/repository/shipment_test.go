package repository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"icaris-logistic/backend/internal/config"
	"icaris-logistic/backend/internal/db"
	"icaris-logistic/backend/internal/domain"
	"icaris-logistic/backend/internal/repository"
)

func setupShipmentRepo(t *testing.T) (*repository.ShipmentRepository, *gorm.DB) {
	t.Helper()

	gdb, err := db.Connect(context.Background(), config.Load().DatabaseURL)
	if err != nil {
		t.Skipf("database unavailable, skipping integration test: %v", err)
	}

	return repository.NewShipmentRepository(gdb), gdb
}

// createTestShipment заводит менеджера, клиента и груз с уборкой после теста. Возвращает
// созданный груз (с проставленным базой id и начальным статусом pending).
func createTestShipment(t *testing.T, repo *repository.ShipmentRepository, gdb *gorm.DB) *domain.Shipment {
	t.Helper()

	manager := &domain.Manager{Email: "ship-" + uuid.NewString() + "@icaris.io", Name: "М", Password: "x"}
	if err := gdb.Create(manager).Error; err != nil {
		t.Fatalf("create manager: %v", err)
	}
	client := &domain.Client{TelegramID: int64(uuid.New().ID()), Name: "Клиент"}
	if err := gdb.Create(client).Error; err != nil {
		t.Fatalf("create client: %v", err)
	}

	shipment := &domain.Shipment{
		ClientID:    client.ID,
		ManagerID:   manager.ID,
		TrackingKey: "TEST-" + uuid.NewString()[:8],
		Status:      domain.ShipmentStatusPending,
		Currency:    "USD",
	}
	if err := repo.Create(context.Background(), shipment); err != nil {
		t.Fatalf("create shipment: %v", err)
	}

	t.Cleanup(func() {
		gdb.Delete(&domain.Shipment{}, "id = ?", shipment.ID) // события каскадно
		gdb.Delete(&domain.Client{}, "id = ?", client.ID)
		gdb.Delete(&domain.Manager{}, "id = ?", manager.ID)
	})

	return shipment
}

func TestShipmentRepository_CreateWritesInitialStatusEvent(t *testing.T) {
	repo, gdb := setupShipmentRepo(t)
	shipment := createTestShipment(t, repo, gdb)

	history, err := repo.StatusHistory(context.Background(), shipment.ID)
	if err != nil {
		t.Fatalf("status history: %v", err)
	}

	if len(history) != 1 {
		t.Errorf("expected one initial status event, got %d", len(history))
	}
}

func TestShipmentRepository_UpdateStatusSetsDeliveredAt(t *testing.T) {
	repo, gdb := setupShipmentRepo(t)
	shipment := createTestShipment(t, repo, gdb)

	updated, err := repo.UpdateStatus(context.Background(), shipment.ID, domain.ShipmentStatusDelivered, "вручён", shipment.ManagerID)
	if err != nil {
		t.Fatalf("update status: %v", err)
	}

	if updated.DeliveredAt == nil {
		t.Error("expected delivered_at to be set on delivery")
	}
}

func TestShipmentRepository_UpdateStatusClearsDeliveredAtWhenLeavingDelivered(t *testing.T) {
	repo, gdb := setupShipmentRepo(t)
	shipment := createTestShipment(t, repo, gdb)

	if _, err := repo.UpdateStatus(context.Background(), shipment.ID, domain.ShipmentStatusDelivered, "", shipment.ManagerID); err != nil {
		t.Fatalf("deliver: %v", err)
	}
	reverted, err := repo.UpdateStatus(context.Background(), shipment.ID, domain.ShipmentStatusInTransit, "откат", shipment.ManagerID)
	if err != nil {
		t.Fatalf("revert: %v", err)
	}

	if reverted.DeliveredAt != nil {
		t.Error("expected delivered_at to be cleared when leaving delivered status")
	}
}

func TestShipmentRepository_UpdateStatusAppendsToHistory(t *testing.T) {
	repo, gdb := setupShipmentRepo(t)
	shipment := createTestShipment(t, repo, gdb)

	if _, err := repo.UpdateStatus(context.Background(), shipment.ID, domain.ShipmentStatusInTransit, "в пути", shipment.ManagerID); err != nil {
		t.Fatalf("update status: %v", err)
	}

	history, err := repo.StatusHistory(context.Background(), shipment.ID)
	if err != nil {
		t.Fatalf("status history: %v", err)
	}

	if len(history) != 2 {
		t.Errorf("expected two events after one update, got %d", len(history))
	}
}

func TestShipmentRepository_UpdateStatusUnknownIDReturnsNotFound(t *testing.T) {
	repo, _ := setupShipmentRepo(t)

	_, err := repo.UpdateStatus(context.Background(), uuid.New(), domain.ShipmentStatusInTransit, "", uuid.New())

	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound for unknown shipment, got %v", err)
	}
}
