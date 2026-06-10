package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"icaris-logistic/backend/internal/domain"
)

type fakePaymentStore struct {
	created *domain.Payment
	payment *domain.Payment
	updated *domain.Payment
}

func (f *fakePaymentStore) Create(_ context.Context, payment *domain.Payment) error {
	f.created = payment
	return nil
}

func (f *fakePaymentStore) ListByShipment(_ context.Context, _ uuid.UUID) ([]domain.Payment, error) {
	return nil, nil
}

func (f *fakePaymentStore) GetByID(_ context.Context, _ uuid.UUID) (*domain.Payment, error) {
	if f.payment == nil {
		return nil, domain.ErrNotFound
	}
	return f.payment, nil
}

func (f *fakePaymentStore) UpdateStatus(_ context.Context, id uuid.UUID, status domain.PaymentStatus) (*domain.Payment, error) {
	f.updated = &domain.Payment{ID: id, Status: status}
	return f.updated, nil
}

type fakeShipmentReader struct {
	shipment *domain.Shipment
}

func (f *fakeShipmentReader) GetByID(_ context.Context, _ uuid.UUID) (*domain.Shipment, error) {
	if f.shipment == nil {
		return nil, domain.ErrNotFound
	}
	return f.shipment, nil
}

func validPaymentInput() CreatePaymentInput {
	return CreatePaymentInput{
		Amount:  decimal.NewFromInt(100),
		Channel: domain.PaymentChannelCash,
	}
}

func paymentServiceWithShipment() (*PaymentService, *fakePaymentStore) {
	store := &fakePaymentStore{}
	shipments := &fakeShipmentReader{shipment: &domain.Shipment{ID: uuid.New(), Currency: "USD"}}
	return NewPaymentService(store, shipments), store
}

func TestPaymentServiceCreateRejectsZeroAmount(t *testing.T) {
	svc, _ := paymentServiceWithShipment()
	input := validPaymentInput()
	input.Amount = decimal.Zero

	_, err := svc.Create(context.Background(), uuid.New(), uuid.New(), input)

	if !errors.Is(err, ErrValidation) {
		t.Errorf("expected ErrValidation for zero amount, got %v", err)
	}
}

func TestPaymentServiceCreateRejectsNegativeAmount(t *testing.T) {
	svc, _ := paymentServiceWithShipment()
	input := validPaymentInput()
	input.Amount = decimal.NewFromInt(-5)

	_, err := svc.Create(context.Background(), uuid.New(), uuid.New(), input)

	if !errors.Is(err, ErrValidation) {
		t.Errorf("expected ErrValidation for negative amount, got %v", err)
	}
}

func TestPaymentServiceCreateRejectsUnknownChannel(t *testing.T) {
	svc, _ := paymentServiceWithShipment()
	input := validPaymentInput()
	input.Channel = "carrier_pigeon"

	_, err := svc.Create(context.Background(), uuid.New(), uuid.New(), input)

	if !errors.Is(err, ErrValidation) {
		t.Errorf("expected ErrValidation for unknown channel, got %v", err)
	}
}

func TestPaymentServiceCreateRejectsUnknownStatus(t *testing.T) {
	svc, _ := paymentServiceWithShipment()
	input := validPaymentInput()
	input.Status = "maybe"

	_, err := svc.Create(context.Background(), uuid.New(), uuid.New(), input)

	if !errors.Is(err, ErrValidation) {
		t.Errorf("expected ErrValidation for unknown status, got %v", err)
	}
}

func TestPaymentServiceCreateUnknownShipmentReturnsNotFound(t *testing.T) {
	svc := NewPaymentService(&fakePaymentStore{}, &fakeShipmentReader{})

	_, err := svc.Create(context.Background(), uuid.New(), uuid.New(), validPaymentInput())

	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound for unknown shipment, got %v", err)
	}
}

func TestPaymentServiceCreateDefaultsStatusToPending(t *testing.T) {
	svc, store := paymentServiceWithShipment()

	if _, err := svc.Create(context.Background(), uuid.New(), uuid.New(), validPaymentInput()); err != nil {
		t.Fatalf("create: %v", err)
	}

	if store.created.Status != domain.PaymentStatusPending {
		t.Errorf("expected default status pending, got %q", store.created.Status)
	}
}

func TestPaymentServiceCreateDefaultsCurrencyFromShipment(t *testing.T) {
	svc, store := paymentServiceWithShipment()

	if _, err := svc.Create(context.Background(), uuid.New(), uuid.New(), validPaymentInput()); err != nil {
		t.Fatalf("create: %v", err)
	}

	if store.created.Currency != "USD" {
		t.Errorf("expected currency defaulted from shipment (USD), got %q", store.created.Currency)
	}
}

func TestPaymentServiceCreateNormalizesCurrencyToUpper(t *testing.T) {
	svc, store := paymentServiceWithShipment()
	input := validPaymentInput()
	input.Currency = " rub "

	if _, err := svc.Create(context.Background(), uuid.New(), uuid.New(), input); err != nil {
		t.Fatalf("create: %v", err)
	}

	if store.created.Currency != "RUB" {
		t.Errorf("expected normalized currency RUB, got %q", store.created.Currency)
	}
}

func TestPaymentServiceCreateRejectsBadCurrencyLength(t *testing.T) {
	svc, _ := paymentServiceWithShipment()
	input := validPaymentInput()
	input.Currency = "RUBL"

	_, err := svc.Create(context.Background(), uuid.New(), uuid.New(), input)

	if !errors.Is(err, ErrValidation) {
		t.Errorf("expected ErrValidation for 4-letter currency, got %v", err)
	}
}

func TestPaymentServiceListUnknownShipmentReturnsNotFound(t *testing.T) {
	svc := NewPaymentService(&fakePaymentStore{}, &fakeShipmentReader{})

	_, err := svc.ListByShipment(context.Background(), uuid.New())

	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound for unknown shipment, got %v", err)
	}
}

func TestPaymentServiceUpdateStatusRejectsUnknownStatus(t *testing.T) {
	svc, _ := paymentServiceWithShipment()

	_, err := svc.UpdateStatus(context.Background(), uuid.New(), uuid.New(), "maybe")

	if !errors.Is(err, ErrValidation) {
		t.Errorf("expected ErrValidation for unknown status, got %v", err)
	}
}

func TestPaymentServiceUpdateStatusForeignShipmentReturnsNotFound(t *testing.T) {
	store := &fakePaymentStore{payment: &domain.Payment{ID: uuid.New(), ShipmentID: uuid.New()}}
	svc := NewPaymentService(store, &fakeShipmentReader{})

	_, err := svc.UpdateStatus(context.Background(), uuid.New(), store.payment.ID, domain.PaymentStatusConfirmed)

	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound for payment of another shipment, got %v", err)
	}
}

func TestPaymentServiceUpdateStatusUpdatesMatchingPayment(t *testing.T) {
	shipmentID := uuid.New()
	store := &fakePaymentStore{payment: &domain.Payment{ID: uuid.New(), ShipmentID: shipmentID}}
	svc := NewPaymentService(store, &fakeShipmentReader{})

	updated, err := svc.UpdateStatus(context.Background(), shipmentID, store.payment.ID, domain.PaymentStatusConfirmed)
	if err != nil {
		t.Fatalf("update status: %v", err)
	}

	if updated.Status != domain.PaymentStatusConfirmed {
		t.Errorf("expected status confirmed, got %q", updated.Status)
	}
}
