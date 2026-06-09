package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"icaris-logistic/backend/internal/domain"
)

// trackingKeyPrefix и trackingKeyAlphabet задают формат трек-ключа: ICR-XXXXXXXXXX.
// Алфавит — Crockford base32 без похожих символов (I, L, O, U), чтобы ключ можно было
// надиктовать без путаницы. 10 символов ≈ 50 бит энтропии — коллизии практически исключены.
const (
	trackingKeyPrefix   = "ICR-"
	trackingKeyAlphabet = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"
	trackingKeyLength   = 10
)

// ShipmentStore — что ShipmentService нужно от хранилища грузов (интерфейс на стороне
// потребителя). Create и UpdateStatus транзакционно пишут ещё и запись истории статуса.
type ShipmentStore interface {
	Create(ctx context.Context, shipment *domain.Shipment) error
	List(ctx context.Context) ([]domain.Shipment, error)
	ListByClient(ctx context.Context, clientID uuid.UUID) ([]domain.Shipment, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Shipment, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ShipmentStatus, comment string, changedBy uuid.UUID) (*domain.Shipment, error)
	StatusHistory(ctx context.Context, shipmentID uuid.UUID) ([]domain.ShipmentStatusEvent, error)
	ExistsTrackingKey(ctx context.Context, key string) (bool, error)
}

// ClientReader — чтение клиента. Общий интерфейс для сервисов, которым нужен клиент по id
// (уведомление о статусе) или по telegram_id (команда /status в боте).
type ClientReader interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Client, error)
	GetByTelegramID(ctx context.Context, telegramID int64) (*domain.Client, error)
}

// ClientNotifier уведомляет клиента о смене статуса груза (реализует бот).
type ClientNotifier interface {
	NotifyShipmentStatus(ctx context.Context, telegramID int64, shipment *domain.Shipment) error
}

// ShipmentDetail — груз вместе с историей статусов. Возвращается на экране деталей, чтобы
// клиент/менеджер получил и текущее состояние, и таймлайн за один запрос.
type ShipmentDetail struct {
	Shipment domain.Shipment              `json:"shipment"`
	History  []domain.ShipmentStatusEvent `json:"history"`
}

type ShipmentService struct {
	store    ShipmentStore
	clients  ClientReader
	notifier ClientNotifier
	bg       *Background
}

func NewShipmentService(store ShipmentStore, clients ClientReader, notifier ClientNotifier, bg *Background) *ShipmentService {
	return &ShipmentService{store: store, clients: clients, notifier: notifier, bg: bg}
}

// CreateShipmentInput — данные для заведения груза менеджером. client_id обязателен и
// должен указывать на существующего клиента; остальное опционально и уточняется по ходу.
type CreateShipmentInput struct {
	ClientID   uuid.UUID           `json:"client_id"`
	Weight     decimal.NullDecimal `json:"weight"`
	Volume     decimal.NullDecimal `json:"volume"`
	FromCity   string              `json:"from_city"`
	ToCity     string              `json:"to_city"`
	Price      decimal.NullDecimal `json:"price"`
	Currency   string              `json:"currency"`
	StatusNote string              `json:"status_note"`
}

func (s *ShipmentService) Create(ctx context.Context, managerID uuid.UUID, input CreateShipmentInput) (*domain.Shipment, error) {
	if input.ClientID == uuid.Nil {
		return nil, fmt.Errorf("%w: client_id is required", ErrValidation)
	}
	if _, err := s.clients.GetByID(ctx, input.ClientID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, fmt.Errorf("%w: unknown client_id", ErrValidation)
		}
		return nil, err
	}

	currency := input.Currency
	if currency == "" {
		currency = "USD"
	}

	key, err := s.generateTrackingKey(ctx)
	if err != nil {
		return nil, err
	}

	shipment := &domain.Shipment{
		ClientID:      input.ClientID,
		ManagerID:     managerID,
		TrackingKey:   key,
		Status:        domain.ShipmentStatusPending,
		StatusComment: input.StatusNote,
		Weight:        input.Weight,
		Volume:        input.Volume,
		FromCity:      input.FromCity,
		ToCity:        input.ToCity,
		Price:         input.Price,
		Currency:      currency,
	}
	if err := s.store.Create(ctx, shipment); err != nil {
		return nil, err
	}

	return shipment, nil
}

func (s *ShipmentService) List(ctx context.Context) ([]domain.Shipment, error) {
	return s.store.List(ctx)
}

func (s *ShipmentService) ListByClientID(ctx context.Context, clientID uuid.UUID) ([]domain.Shipment, error) {
	return s.store.ListByClient(ctx, clientID)
}

// ListByTelegramID — грузы клиента по его telegram_id (для команды /status в боте).
// Незарегистрированный telegram_id → domain.ErrNotFound.
func (s *ShipmentService) ListByTelegramID(ctx context.Context, telegramID int64) ([]domain.Shipment, error) {
	client, err := s.clients.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return nil, err
	}

	return s.store.ListByClient(ctx, client.ID)
}

// Detail — груз и его история для менеджера (доступ ко всем грузам).
func (s *ShipmentService) Detail(ctx context.Context, id uuid.UUID) (*ShipmentDetail, error) {
	shipment, err := s.store.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.withHistory(ctx, shipment)
}

// DetailForClient — груз и история, но только если груз принадлежит клиенту. Чужой груз →
// domain.ErrNotFound (а не 403): не раскрываем существование чужих грузов.
func (s *ShipmentService) DetailForClient(ctx context.Context, id, clientID uuid.UUID) (*ShipmentDetail, error) {
	shipment, err := s.store.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if shipment.ClientID != clientID {
		return nil, domain.ErrNotFound
	}

	return s.withHistory(ctx, shipment)
}

func (s *ShipmentService) withHistory(ctx context.Context, shipment *domain.Shipment) (*ShipmentDetail, error) {
	history, err := s.store.StatusHistory(ctx, shipment.ID)
	if err != nil {
		return nil, err
	}

	return &ShipmentDetail{Shipment: *shipment, History: history}, nil
}

// UpdateStatus меняет статус груза (проверив, что значение из набора enum) и уведомляет
// клиента в фоне. Переходы свободные (без стейт-машины) — решение MVP, см. docs/tech-debt.md.
func (s *ShipmentService) UpdateStatus(
	ctx context.Context,
	id, managerID uuid.UUID,
	status domain.ShipmentStatus,
	comment string,
) (*domain.Shipment, error) {
	if !status.IsValid() {
		return nil, fmt.Errorf("%w: unknown status %q", ErrValidation, status)
	}

	updated, err := s.store.UpdateStatus(ctx, id, status, comment, managerID)
	if err != nil {
		return nil, err
	}

	s.notifyClient(updated)

	return updated, nil
}

// notifyClient уведомляет клиента о новом статусе в фоне (как notifyNewLead): латентность
// Telegram не тормозит ответ менеджеру, ошибка только логируется. context.Background(),
// т.к. ctx запроса отменится сразу после ответа.
//
// NOTE: MVP — без ретраев и персистентности (не полный outbox); см. docs/tech-debt.md
func (s *ShipmentService) notifyClient(shipment *domain.Shipment) {
	if s.notifier == nil {
		return
	}
	s.bg.Go(func() {
		ctx := context.Background()
		client, err := s.clients.GetByID(ctx, shipment.ClientID)
		if err != nil {
			slog.Error("notify shipment status: load client", "shipment_id", shipment.ID, "error", err)
			return
		}
		if err := s.notifier.NotifyShipmentStatus(ctx, client.TelegramID, shipment); err != nil {
			slog.Error("notify shipment status", "shipment_id", shipment.ID, "error", err)
		}
	})
}

// generateTrackingKey генерирует уникальный трек-ключ, повторяя при маловероятном
// совпадении. Уникальность в БД гарантирует unique-индекс; пред-проверка лишь избавляет
// от 500 на коллизии.
func (s *ShipmentService) generateTrackingKey(ctx context.Context) (string, error) {
	for attempt := 0; attempt < 5; attempt++ {
		key, err := randomTrackingKey()
		if err != nil {
			return "", err
		}
		exists, err := s.store.ExistsTrackingKey(ctx, key)
		if err != nil {
			return "", err
		}
		if !exists {
			return key, nil
		}
	}

	return "", errors.New("could not generate unique tracking key")
}

func randomTrackingKey() (string, error) {
	buf := make([]byte, trackingKeyLength)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("read random: %w", err)
	}
	// len(alphabet) == 32, а 256 % 32 == 0 — деление байта по модулю не даёт смещения.
	for i := range buf {
		buf[i] = trackingKeyAlphabet[int(buf[i])%len(trackingKeyAlphabet)]
	}

	return trackingKeyPrefix + string(buf), nil
}
