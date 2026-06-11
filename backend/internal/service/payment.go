package service

import (
	"context"
	"fmt"
	"log/slog"
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

// PaymentNotifier уведомляет клиента о счёте и полученном платеже (реализует бот).
type PaymentNotifier interface {
	NotifyPaymentCreated(ctx context.Context, telegramID int64, shipment *domain.Shipment, payment *domain.Payment) error
	NotifyPaymentConfirmed(ctx context.Context, telegramID int64, shipment *domain.Shipment, payment *domain.Payment) error
}

// PaymentService использует общий ShipmentReader (см. message.go): проверяет
// существование груза и берёт его валюту как дефолт для платежа.
type PaymentService struct {
	store     PaymentStore
	shipments ShipmentReader
	clients   ClientReader
	notifier  PaymentNotifier
	bg        *Background
}

func NewPaymentService(
	store PaymentStore,
	shipments ShipmentReader,
	clients ClientReader,
	notifier PaymentNotifier,
	bg *Background,
) *PaymentService {
	return &PaymentService{store: store, shipments: shipments, clients: clients, notifier: notifier, bg: bg}
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

	// pending → клиенту выставлен счёт; сразу confirmed (менеджер зафиксировал уже
	// полученные деньги, например наличные) → уведомление о получении, без стадии счёта.
	switch payment.Status {
	case domain.PaymentStatusPending:
		s.notify(shipment, payment, false)
	case domain.PaymentStatusConfirmed:
		s.notify(shipment, payment, true)
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

	updated, err := s.store.UpdateStatus(ctx, paymentID, status)
	if err != nil {
		return nil, err
	}

	// Уведомляем только при реальном переходе в confirmed: повторное «подтверждение» уже
	// подтверждённого платежа не должно слать клиенту дубль.
	if status == domain.PaymentStatusConfirmed && payment.Status != domain.PaymentStatusConfirmed {
		shipment, err := s.shipments.GetByID(ctx, shipmentID)
		if err != nil {
			// Платёж обновлён, уведомление потеряно — это деградация, не сбой операции.
			slog.Error("payment confirmed: load shipment for notify", "payment_id", paymentID, "error", err)
			return updated, nil
		}
		s.notify(shipment, updated, true)
	}

	return updated, nil
}

// ListForClient возвращает платежи по грузу клиента. Чужой груз → domain.ErrNotFound
// (как DetailForClient: не раскрываем существование чужих грузов).
func (s *PaymentService) ListForClient(ctx context.Context, shipmentID, clientID uuid.UUID) ([]domain.Payment, error) {
	shipment, err := s.shipments.GetByID(ctx, shipmentID)
	if err != nil {
		return nil, err
	}
	if shipment.ClientID != clientID {
		return nil, domain.ErrNotFound
	}

	return s.store.ListByShipment(ctx, shipmentID)
}

// notify шлёт клиенту уведомление о счёте (confirmed=false) или полученном платеже
// (confirmed=true) в фоне — по паттерну ShipmentService.notifyClient: латентность Telegram
// не тормозит ответ менеджеру, ошибка только логируется (в режиме outbox Notify* лишь
// ставит сообщение в очередь, доставку с ретраями ведёт диспетчер).
func (s *PaymentService) notify(shipment *domain.Shipment, payment *domain.Payment, confirmed bool) {
	if s.notifier == nil {
		return
	}
	s.bg.Go(func() {
		ctx := context.Background()
		client, err := s.clients.GetByID(ctx, shipment.ClientID)
		if err != nil {
			slog.Error("notify payment: load client", "payment_id", payment.ID, "error", err)
			return
		}
		if confirmed {
			err = s.notifier.NotifyPaymentConfirmed(ctx, client.TelegramID, shipment, payment)
		} else {
			err = s.notifier.NotifyPaymentCreated(ctx, client.TelegramID, shipment, payment)
		}
		if err != nil {
			slog.Error("notify payment", "payment_id", payment.ID, "confirmed", confirmed, "error", err)
		}
	})
}
