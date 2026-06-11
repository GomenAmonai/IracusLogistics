package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"icaris-logistic/backend/internal/domain"
	"icaris-logistic/backend/internal/service"
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

// fakePaymentNotifier запоминает последнее уведомление: "" — не было, иначе created/confirmed.
type fakePaymentNotifier struct {
	kind       string
	telegramID int64
}

func (f *fakePaymentNotifier) NotifyPaymentCreated(_ context.Context, telegramID int64, _ *domain.Shipment, _ *domain.Payment) error {
	f.kind = "created"
	f.telegramID = telegramID
	return nil
}

func (f *fakePaymentNotifier) NotifyPaymentConfirmed(_ context.Context, telegramID int64, _ *domain.Shipment, _ *domain.Payment) error {
	f.kind = "confirmed"
	f.telegramID = telegramID
	return nil
}

func validPaymentInput() service.CreatePaymentInput {
	return service.CreatePaymentInput{
		Amount:  decimal.NewFromInt(100),
		Channel: domain.PaymentChannelCash,
	}
}

func paymentServiceWithShipment() (*service.PaymentService, *fakePaymentStore) {
	store := &fakePaymentStore{}
	shipments := &fakeShipmentReader{shipment: &domain.Shipment{ID: uuid.New(), Currency: "USD"}}
	return service.NewPaymentService(store, shipments, nil, nil, nil), store
}

// notifyingPaymentService — сервис с фейковым нотификатором и реальным Background:
// тест дренирует фон через bg.Wait и смотрит, что отправлено.
func notifyingPaymentService(store *fakePaymentStore, shipment *domain.Shipment) (*service.PaymentService, *fakePaymentNotifier, *service.Background) {
	notifier := &fakePaymentNotifier{}
	bg := service.NewBackground()
	clients := &fakeClientReader{client: &domain.Client{ID: shipment.ClientID, TelegramID: 4242}}
	svc := service.NewPaymentService(store, &fakeShipmentReader{shipment: shipment}, clients, notifier, bg)
	return svc, notifier, bg
}

func TestPaymentServiceCreateRejectsZeroAmount(t *testing.T) {
	svc, _ := paymentServiceWithShipment()
	input := validPaymentInput()
	input.Amount = decimal.Zero

	_, err := svc.Create(context.Background(), uuid.New(), uuid.New(), input)

	if !errors.Is(err, service.ErrValidation) {
		t.Errorf("expected service.ErrValidation for zero amount, got %v", err)
	}
}

func TestPaymentServiceCreateRejectsNegativeAmount(t *testing.T) {
	svc, _ := paymentServiceWithShipment()
	input := validPaymentInput()
	input.Amount = decimal.NewFromInt(-5)

	_, err := svc.Create(context.Background(), uuid.New(), uuid.New(), input)

	if !errors.Is(err, service.ErrValidation) {
		t.Errorf("expected service.ErrValidation for negative amount, got %v", err)
	}
}

func TestPaymentServiceCreateRejectsUnknownChannel(t *testing.T) {
	svc, _ := paymentServiceWithShipment()
	input := validPaymentInput()
	input.Channel = "carrier_pigeon"

	_, err := svc.Create(context.Background(), uuid.New(), uuid.New(), input)

	if !errors.Is(err, service.ErrValidation) {
		t.Errorf("expected service.ErrValidation for unknown channel, got %v", err)
	}
}

func TestPaymentServiceCreateRejectsUnknownStatus(t *testing.T) {
	svc, _ := paymentServiceWithShipment()
	input := validPaymentInput()
	input.Status = "maybe"

	_, err := svc.Create(context.Background(), uuid.New(), uuid.New(), input)

	if !errors.Is(err, service.ErrValidation) {
		t.Errorf("expected service.ErrValidation for unknown status, got %v", err)
	}
}

func TestPaymentServiceCreateUnknownShipmentReturnsNotFound(t *testing.T) {
	svc := service.NewPaymentService(&fakePaymentStore{}, &fakeShipmentReader{}, nil, nil, nil)

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

	if !errors.Is(err, service.ErrValidation) {
		t.Errorf("expected service.ErrValidation for 4-letter currency, got %v", err)
	}
}

func TestPaymentServiceListUnknownShipmentReturnsNotFound(t *testing.T) {
	svc := service.NewPaymentService(&fakePaymentStore{}, &fakeShipmentReader{}, nil, nil, nil)

	_, err := svc.ListByShipment(context.Background(), uuid.New())

	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound for unknown shipment, got %v", err)
	}
}

func TestPaymentServiceUpdateStatusRejectsUnknownStatus(t *testing.T) {
	svc, _ := paymentServiceWithShipment()

	_, err := svc.UpdateStatus(context.Background(), uuid.New(), uuid.New(), "maybe")

	if !errors.Is(err, service.ErrValidation) {
		t.Errorf("expected service.ErrValidation for unknown status, got %v", err)
	}
}

func TestPaymentServiceUpdateStatusForeignShipmentReturnsNotFound(t *testing.T) {
	store := &fakePaymentStore{payment: &domain.Payment{ID: uuid.New(), ShipmentID: uuid.New()}}
	svc := service.NewPaymentService(store, &fakeShipmentReader{}, nil, nil, nil)

	_, err := svc.UpdateStatus(context.Background(), uuid.New(), store.payment.ID, domain.PaymentStatusConfirmed)

	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound for payment of another shipment, got %v", err)
	}
}

func TestPaymentServiceUpdateStatusUpdatesMatchingPayment(t *testing.T) {
	shipmentID := uuid.New()
	store := &fakePaymentStore{payment: &domain.Payment{ID: uuid.New(), ShipmentID: shipmentID}}
	svc := service.NewPaymentService(store, &fakeShipmentReader{}, nil, nil, nil)

	updated, err := svc.UpdateStatus(context.Background(), shipmentID, store.payment.ID, domain.PaymentStatusConfirmed)
	if err != nil {
		t.Fatalf("update status: %v", err)
	}

	if updated.Status != domain.PaymentStatusConfirmed {
		t.Errorf("expected status confirmed, got %q", updated.Status)
	}
}

func TestPaymentServiceCreatePendingNotifiesInvoice(t *testing.T) {
	shipment := &domain.Shipment{ID: uuid.New(), ClientID: uuid.New(), Currency: "USD"}
	svc, notifier, bg := notifyingPaymentService(&fakePaymentStore{}, shipment)

	if _, err := svc.Create(context.Background(), shipment.ID, uuid.New(), validPaymentInput()); err != nil {
		t.Fatalf("create: %v", err)
	}
	drain(t, bg)

	if notifier.kind != "created" {
		t.Errorf("expected invoice notification, got %q", notifier.kind)
	}
}

func TestPaymentServiceCreateConfirmedNotifiesReceived(t *testing.T) {
	shipment := &domain.Shipment{ID: uuid.New(), ClientID: uuid.New(), Currency: "USD"}
	svc, notifier, bg := notifyingPaymentService(&fakePaymentStore{}, shipment)
	input := validPaymentInput()
	input.Status = domain.PaymentStatusConfirmed

	if _, err := svc.Create(context.Background(), shipment.ID, uuid.New(), input); err != nil {
		t.Fatalf("create: %v", err)
	}
	drain(t, bg)

	if notifier.kind != "confirmed" {
		t.Errorf("expected received notification, got %q", notifier.kind)
	}
}

func TestPaymentServiceCreateNotifiesClientTelegramID(t *testing.T) {
	shipment := &domain.Shipment{ID: uuid.New(), ClientID: uuid.New(), Currency: "USD"}
	svc, notifier, bg := notifyingPaymentService(&fakePaymentStore{}, shipment)

	if _, err := svc.Create(context.Background(), shipment.ID, uuid.New(), validPaymentInput()); err != nil {
		t.Fatalf("create: %v", err)
	}
	drain(t, bg)

	if notifier.telegramID != 4242 {
		t.Errorf("expected notification to client telegram id 4242, got %d", notifier.telegramID)
	}
}

func TestPaymentServiceConfirmNotifiesReceived(t *testing.T) {
	shipment := &domain.Shipment{ID: uuid.New(), ClientID: uuid.New(), Currency: "USD"}
	store := &fakePaymentStore{payment: &domain.Payment{
		ID: uuid.New(), ShipmentID: shipment.ID, Status: domain.PaymentStatusPending,
	}}
	svc, notifier, bg := notifyingPaymentService(store, shipment)

	if _, err := svc.UpdateStatus(context.Background(), shipment.ID, store.payment.ID, domain.PaymentStatusConfirmed); err != nil {
		t.Fatalf("update status: %v", err)
	}
	drain(t, bg)

	if notifier.kind != "confirmed" {
		t.Errorf("expected received notification, got %q", notifier.kind)
	}
}

func TestPaymentServiceReconfirmDoesNotNotifyAgain(t *testing.T) {
	shipment := &domain.Shipment{ID: uuid.New(), ClientID: uuid.New(), Currency: "USD"}
	store := &fakePaymentStore{payment: &domain.Payment{
		ID: uuid.New(), ShipmentID: shipment.ID, Status: domain.PaymentStatusConfirmed,
	}}
	svc, notifier, bg := notifyingPaymentService(store, shipment)

	if _, err := svc.UpdateStatus(context.Background(), shipment.ID, store.payment.ID, domain.PaymentStatusConfirmed); err != nil {
		t.Fatalf("update status: %v", err)
	}
	drain(t, bg)

	if notifier.kind != "" {
		t.Errorf("expected no notification on re-confirm, got %q", notifier.kind)
	}
}

func TestPaymentServiceRefundDoesNotNotify(t *testing.T) {
	shipment := &domain.Shipment{ID: uuid.New(), ClientID: uuid.New(), Currency: "USD"}
	store := &fakePaymentStore{payment: &domain.Payment{
		ID: uuid.New(), ShipmentID: shipment.ID, Status: domain.PaymentStatusConfirmed,
	}}
	svc, notifier, bg := notifyingPaymentService(store, shipment)

	if _, err := svc.UpdateStatus(context.Background(), shipment.ID, store.payment.ID, domain.PaymentStatusRefunded); err != nil {
		t.Fatalf("update status: %v", err)
	}
	drain(t, bg)

	if notifier.kind != "" {
		t.Errorf("expected no notification on refund, got %q", notifier.kind)
	}
}

func TestPaymentServiceListForClientOwnShipmentReturnsPayments(t *testing.T) {
	clientID := uuid.New()
	shipment := &domain.Shipment{ID: uuid.New(), ClientID: clientID}
	svc := service.NewPaymentService(&fakePaymentStore{}, &fakeShipmentReader{shipment: shipment}, nil, nil, nil)

	_, err := svc.ListForClient(context.Background(), shipment.ID, clientID)

	if err != nil {
		t.Errorf("expected own shipment payments to be listed, got %v", err)
	}
}

func TestPaymentServiceListForClientForeignShipmentReturnsNotFound(t *testing.T) {
	shipment := &domain.Shipment{ID: uuid.New(), ClientID: uuid.New()}
	svc := service.NewPaymentService(&fakePaymentStore{}, &fakeShipmentReader{shipment: shipment}, nil, nil, nil)

	_, err := svc.ListForClient(context.Background(), shipment.ID, uuid.New())

	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound for foreign shipment, got %v", err)
	}
}

// drain дожидается фоновых задач сервиса (уведомления уходят через Background).
func drain(t *testing.T, bg *service.Background) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := bg.Wait(ctx); err != nil {
		t.Fatalf("drain background: %v", err)
	}
}
