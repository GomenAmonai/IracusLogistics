package repository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"icaris-logistic/backend/internal/domain"
	"icaris-logistic/backend/internal/repository"
)

func setupPaymentRepo(t *testing.T) (*repository.PaymentRepository, *repository.ShipmentRepository, *gorm.DB) {
	t.Helper()

	shipmentRepo, gdb := setupShipmentRepo(t)

	return repository.NewPaymentRepository(gdb), shipmentRepo, gdb
}

// createTestPayment заводит платёж по существующему грузу с уборкой после теста.
func createTestPayment(t *testing.T, repo *repository.PaymentRepository, gdb *gorm.DB, shipment *domain.Shipment) *domain.Payment {
	t.Helper()

	payment := &domain.Payment{
		ShipmentID: shipment.ID,
		Amount:     decimal.NewFromInt(150),
		Currency:   "USD",
		Channel:    domain.PaymentChannelBankTransfer,
		Status:     domain.PaymentStatusPending,
		CreatedBy:  &shipment.ManagerID,
	}
	if err := repo.Create(context.Background(), payment); err != nil {
		t.Fatalf("create payment: %v", err)
	}

	t.Cleanup(func() {
		gdb.Delete(&domain.Payment{}, "id = ?", payment.ID)
	})

	return payment
}

func TestPaymentRepository_CreateFillsDefaultsFromDB(t *testing.T) {
	paymentRepo, shipmentRepo, gdb := setupPaymentRepo(t)
	shipment := createTestShipment(t, shipmentRepo, gdb)

	payment := createTestPayment(t, paymentRepo, gdb, shipment)

	if payment.ID == uuid.Nil {
		t.Error("expected database to generate payment id")
	}
}

func TestPaymentRepository_CreateUnknownShipmentReturnsNotFound(t *testing.T) {
	paymentRepo, _, _ := setupPaymentRepo(t)

	err := paymentRepo.Create(context.Background(), &domain.Payment{
		ShipmentID: uuid.New(),
		Amount:     decimal.NewFromInt(10),
		Currency:   "USD",
		Channel:    domain.PaymentChannelCash,
		Status:     domain.PaymentStatusPending,
	})

	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound for unknown shipment (FK violation), got %v", err)
	}
}

func TestPaymentRepository_ListByShipmentReturnsCreated(t *testing.T) {
	paymentRepo, shipmentRepo, gdb := setupPaymentRepo(t)
	shipment := createTestShipment(t, shipmentRepo, gdb)
	createTestPayment(t, paymentRepo, gdb, shipment)
	createTestPayment(t, paymentRepo, gdb, shipment)

	payments, err := paymentRepo.ListByShipment(context.Background(), shipment.ID)
	if err != nil {
		t.Fatalf("list payments: %v", err)
	}

	if len(payments) != 2 {
		t.Errorf("expected two payments, got %d", len(payments))
	}
}

func TestPaymentRepository_UpdateStatusReturnsUpdatedRow(t *testing.T) {
	paymentRepo, shipmentRepo, gdb := setupPaymentRepo(t)
	shipment := createTestShipment(t, shipmentRepo, gdb)
	payment := createTestPayment(t, paymentRepo, gdb, shipment)

	updated, err := paymentRepo.UpdateStatus(context.Background(), payment.ID, domain.PaymentStatusConfirmed)
	if err != nil {
		t.Fatalf("update status: %v", err)
	}

	if updated.Status != domain.PaymentStatusConfirmed {
		t.Errorf("expected status confirmed, got %q", updated.Status)
	}
}

func TestPaymentRepository_UpdateStatusBumpsUpdatedAt(t *testing.T) {
	paymentRepo, shipmentRepo, gdb := setupPaymentRepo(t)
	shipment := createTestShipment(t, shipmentRepo, gdb)
	payment := createTestPayment(t, paymentRepo, gdb, shipment)

	updated, err := paymentRepo.UpdateStatus(context.Background(), payment.ID, domain.PaymentStatusConfirmed)
	if err != nil {
		t.Fatalf("update status: %v", err)
	}

	if !updated.UpdatedAt.After(payment.UpdatedAt) {
		t.Errorf("expected updated_at to move forward: was %v, now %v", payment.UpdatedAt, updated.UpdatedAt)
	}
}

func TestPaymentRepository_UpdateStatusUnknownIDReturnsNotFound(t *testing.T) {
	paymentRepo, _, _ := setupPaymentRepo(t)

	_, err := paymentRepo.UpdateStatus(context.Background(), uuid.New(), domain.PaymentStatusConfirmed)

	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound for unknown payment, got %v", err)
	}
}
