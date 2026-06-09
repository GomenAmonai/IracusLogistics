package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"icaris-logistic/backend/internal/domain"
	"icaris-logistic/backend/internal/service"
)

type fakeShipmentStore struct {
	created    *domain.Shipment
	updated    *domain.Shipment
	updateCall bool
}

func (f *fakeShipmentStore) Create(ctx context.Context, shipment *domain.Shipment) error {
	f.created = shipment
	return nil
}

func (f *fakeShipmentStore) List(ctx context.Context) ([]domain.Shipment, error) {
	return nil, nil
}

func (f *fakeShipmentStore) ListByClient(ctx context.Context, clientID uuid.UUID) ([]domain.Shipment, error) {
	return nil, nil
}

func (f *fakeShipmentStore) GetByID(ctx context.Context, id uuid.UUID) (*domain.Shipment, error) {
	return nil, nil
}

func (f *fakeShipmentStore) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ShipmentStatus, comment string, changedBy uuid.UUID) (*domain.Shipment, error) {
	f.updateCall = true
	f.updated = &domain.Shipment{ID: id, Status: status, StatusComment: comment}
	return f.updated, nil
}

func (f *fakeShipmentStore) StatusHistory(ctx context.Context, shipmentID uuid.UUID) ([]domain.ShipmentStatusEvent, error) {
	return nil, nil
}

func (f *fakeShipmentStore) ExistsTrackingKey(ctx context.Context, key string) (bool, error) {
	return false, nil
}

type fakeClientReader struct {
	client *domain.Client
	getErr error
}

func (f *fakeClientReader) GetByID(ctx context.Context, id uuid.UUID) (*domain.Client, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.client, nil
}

func (f *fakeClientReader) GetByTelegramID(ctx context.Context, telegramID int64) (*domain.Client, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.client, nil
}

func TestShipmentService_CreateRequiresClientID(t *testing.T) {
	svc := service.NewShipmentService(&fakeShipmentStore{}, &fakeClientReader{}, nil, nil)

	_, err := svc.Create(context.Background(), uuid.New(), service.CreateShipmentInput{})

	if !errors.Is(err, service.ErrValidation) {
		t.Errorf("expected ErrValidation for missing client_id, got %v", err)
	}
}

func TestShipmentService_CreateRejectsUnknownClient(t *testing.T) {
	clients := &fakeClientReader{getErr: domain.ErrNotFound}
	svc := service.NewShipmentService(&fakeShipmentStore{}, clients, nil, nil)

	_, err := svc.Create(context.Background(), uuid.New(), service.CreateShipmentInput{ClientID: uuid.New()})

	if !errors.Is(err, service.ErrValidation) {
		t.Errorf("expected ErrValidation for unknown client, got %v", err)
	}
}

func TestShipmentService_CreateGeneratesPrefixedTrackingKey(t *testing.T) {
	clientID := uuid.New()
	store := &fakeShipmentStore{}
	clients := &fakeClientReader{client: &domain.Client{ID: clientID}}
	svc := service.NewShipmentService(store, clients, nil, nil)

	shipment, err := svc.Create(context.Background(), uuid.New(), service.CreateShipmentInput{ClientID: clientID})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if !strings.HasPrefix(shipment.TrackingKey, "ICR-") {
		t.Errorf("expected tracking key with ICR- prefix, got %q", shipment.TrackingKey)
	}
}

func TestShipmentService_CreateStartsAsPending(t *testing.T) {
	clientID := uuid.New()
	clients := &fakeClientReader{client: &domain.Client{ID: clientID}}
	svc := service.NewShipmentService(&fakeShipmentStore{}, clients, nil, nil)

	shipment, err := svc.Create(context.Background(), uuid.New(), service.CreateShipmentInput{ClientID: clientID})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if shipment.Status != domain.ShipmentStatusPending {
		t.Errorf("expected new shipment to be pending, got %q", shipment.Status)
	}
}

func TestShipmentService_UpdateStatusRejectsUnknownStatus(t *testing.T) {
	store := &fakeShipmentStore{}
	svc := service.NewShipmentService(store, &fakeClientReader{}, nil, nil)

	_, err := svc.UpdateStatus(context.Background(), uuid.New(), uuid.New(), domain.ShipmentStatus("bogus"), "")

	if !errors.Is(err, service.ErrValidation) {
		t.Errorf("expected ErrValidation for unknown status, got %v", err)
	}
}

func TestShipmentService_UpdateStatusDoesNotTouchStoreOnInvalidStatus(t *testing.T) {
	store := &fakeShipmentStore{}
	svc := service.NewShipmentService(store, &fakeClientReader{}, nil, nil)

	_, _ = svc.UpdateStatus(context.Background(), uuid.New(), uuid.New(), domain.ShipmentStatus("bogus"), "")

	if store.updateCall {
		t.Error("expected store.UpdateStatus not to be called for invalid status")
	}
}

func TestShipmentService_UpdateStatusReturnsUpdatedStatus(t *testing.T) {
	store := &fakeShipmentStore{}
	svc := service.NewShipmentService(store, &fakeClientReader{client: &domain.Client{}}, nil, nil)

	updated, err := svc.UpdateStatus(context.Background(), uuid.New(), uuid.New(), domain.ShipmentStatusInTransit, "в пути")
	if err != nil {
		t.Fatalf("update status: %v", err)
	}

	if updated.Status != domain.ShipmentStatusInTransit {
		t.Errorf("expected status in_transit, got %q", updated.Status)
	}
}
