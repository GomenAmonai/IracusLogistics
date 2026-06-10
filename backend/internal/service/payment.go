package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"icaris-logistic/backend/internal/domain"
)

// PaymentStore — что PaymentService нужно от хранилища платежей (интерфейс на стороне
// потребителя).
type PaymentStore interface {
	Create(ctx context.Context, payment *domain.Payment) error
	ListByShipment(ctx context.Context, shipmentID uuid.UUID) ([]domain.Payment, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.PaymentStatus) (*domain.Payment, error)
}

// PaymentService использует общий ShipmentReader (см. message.go): проверяет
// существование груза и берёт его валюту как дефолт для платежа.
type PaymentService struct {
	store     PaymentStore
	shipments ShipmentReader
}

func NewPaymentService(store PaymentStore, shipments ShipmentReader) *PaymentService {
	return &PaymentService{store: store, shipments: shipments}
}

// CreatePaymentInput — данные платежа от менеджера. Сумма обязательна и положительна;
// валюта по умолчанию берётся из груза (платёж может быть и в другой валюте — например,
// рублёвый платёж по грузу с ценой в USD); статус по умолчанию pending.
type CreatePaymentInput struct {
	Amount   decimal.Decimal       `json:"amount"`
	Currency string                `json:"currency"`
	Channel  domain.PaymentChannel `json:"channel"`
	Status   domain.PaymentStatus  `json:"status"`
	Comment  string                `json:"comment"`
}

// Create регистрирует платёж по грузу. Несуществующий груз → domain.ErrNotFound (id
// приходит из пути URL, это не ошибка тела запроса). Гонку «груз удалили после проверки»
// закрывает репозиторий, транслируя нарушение FK в тот же ErrNotFound.
func (s *PaymentService) Create(ctx context.Context, shipmentID, managerID uuid.UUID, input CreatePaymentInput) (*domain.Payment, error) {
	if !input.Amount.IsPositive() {
		return nil, fmt.Errorf("%w: amount must be positive", ErrValidation)
	}
	if !input.Channel.IsValid() {
		return nil, fmt.Errorf("%w: unknown channel %q", ErrValidation, input.Channel)
	}

	status := input.Status
	if status == "" {
		status = domain.PaymentStatusPending
	}
	if !status.IsValid() {
		return nil, fmt.Errorf("%w: unknown status %q", ErrValidation, status)
	}

	shipment, err := s.shipments.GetByID(ctx, shipmentID)
	if err != nil {
		return nil, err
	}

	currency := strings.ToUpper(strings.TrimSpace(input.Currency))
	if currency == "" {
		currency = shipment.Currency
	}
	if len(currency) != 3 {
		return nil, fmt.Errorf("%w: currency must be a 3-letter code", ErrValidation)
	}

	payment := &domain.Payment{
		ShipmentID: shipmentID,
		Amount:     input.Amount,
		Currency:   currency,
		Channel:    input.Channel,
		Status:     status,
		Comment:    strings.TrimSpace(input.Comment),
		CreatedBy:  &managerID,
	}
	if err := s.store.Create(ctx, payment); err != nil {
		return nil, err
	}

	return payment, nil
}

// ListByShipment возвращает платежи груза. Сначала проверяем, что груз существует:
// иначе несуществующий груз отдавал бы пустой список с 200 вместо честного 404.
func (s *PaymentService) ListByShipment(ctx context.Context, shipmentID uuid.UUID) ([]domain.Payment, error) {
	if _, err := s.shipments.GetByID(ctx, shipmentID); err != nil {
		return nil, err
	}

	return s.store.ListByShipment(ctx, shipmentID)
}

// UpdateStatus меняет статус платежа. shipmentID из пути сверяем с платежом — платёж
// чужого груза недоступен по этому URL (→ ErrNotFound, существование не раскрываем).
func (s *PaymentService) UpdateStatus(ctx context.Context, shipmentID, paymentID uuid.UUID, status domain.PaymentStatus) (*domain.Payment, error) {
	if !status.IsValid() {
		return nil, fmt.Errorf("%w: unknown status %q", ErrValidation, status)
	}

	payment, err := s.store.GetByID(ctx, paymentID)
	if err != nil {
		return nil, err
	}
	if payment.ShipmentID != shipmentID {
		return nil, domain.ErrNotFound
	}

	return s.store.UpdateStatus(ctx, paymentID, status)
}
