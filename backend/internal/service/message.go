package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"

	"icaris-logistic/backend/internal/domain"
)

// MessageStore — что MessageService нужно от хранилища сообщений.
type MessageStore interface {
	Create(ctx context.Context, message *domain.Message) error
	ListByShipment(ctx context.Context, shipmentID uuid.UUID) ([]domain.Message, error)
}

// ShipmentReader — чтение груза для проверки принадлежности (чей это груз).
type ShipmentReader interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Shipment, error)
}

// ManagerNotifier уведомляет менеджера о новом сообщении клиента (реализует бот).
type ManagerNotifier interface {
	NotifyClientMessage(ctx context.Context, client *domain.Client, shipment *domain.Shipment, text string) error
}

// ClientMessageNotifier уведомляет клиента об ответе менеджера (реализует бот).
type ClientMessageNotifier interface {
	NotifyManagerReply(ctx context.Context, telegramID int64, shipment *domain.Shipment, text string) error
}

type MessageService struct {
	messages      MessageStore
	shipments     ShipmentReader
	clients       ClientReader
	managerNotify ManagerNotifier
	clientNotify  ClientMessageNotifier
	bg            *Background
}

func NewMessageService(
	messages MessageStore,
	shipments ShipmentReader,
	clients ClientReader,
	managerNotify ManagerNotifier,
	clientNotify ClientMessageNotifier,
	bg *Background,
) *MessageService {
	return &MessageService{
		messages:      messages,
		shipments:     shipments,
		clients:       clients,
		managerNotify: managerNotify,
		clientNotify:  clientNotify,
		bg:            bg,
	}
}

// SendFromClient сохраняет сообщение клиента по его грузу и уведомляет менеджера. Чужой
// груз → domain.ErrNotFound (не раскрываем существование).
func (s *MessageService) SendFromClient(ctx context.Context, shipmentID, clientID uuid.UUID, text string) (*domain.Message, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, fmt.Errorf("%w: text is required", ErrValidation)
	}

	shipment, err := s.shipments.GetByID(ctx, shipmentID)
	if err != nil {
		return nil, err
	}
	if shipment.ClientID != clientID {
		return nil, domain.ErrNotFound
	}

	message := &domain.Message{
		ShipmentID: &shipmentID,
		ClientID:   clientID,
		Text:       text,
		FromRole:   domain.RoleClient,
	}
	if err := s.messages.Create(ctx, message); err != nil {
		return nil, err
	}

	s.notifyManager(shipment, clientID, text)

	return message, nil
}

// SendFromManager сохраняет ответ менеджера и уведомляет клиента в Telegram.
func (s *MessageService) SendFromManager(ctx context.Context, shipmentID, managerID uuid.UUID, text string) (*domain.Message, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, fmt.Errorf("%w: text is required", ErrValidation)
	}

	shipment, err := s.shipments.GetByID(ctx, shipmentID)
	if err != nil {
		return nil, err
	}

	message := &domain.Message{
		ShipmentID: &shipmentID,
		ClientID:   shipment.ClientID,
		ManagerID:  &managerID,
		Text:       text,
		FromRole:   domain.RoleManager,
	}
	if err := s.messages.Create(ctx, message); err != nil {
		return nil, err
	}

	s.notifyClient(shipment, text)

	return message, nil
}

// ListForClient возвращает переписку по грузу клиента (с проверкой принадлежности).
func (s *MessageService) ListForClient(ctx context.Context, shipmentID, clientID uuid.UUID) ([]domain.Message, error) {
	shipment, err := s.shipments.GetByID(ctx, shipmentID)
	if err != nil {
		return nil, err
	}
	if shipment.ClientID != clientID {
		return nil, domain.ErrNotFound
	}

	return s.messages.ListByShipment(ctx, shipmentID)
}

// ListForManager возвращает переписку по грузу для менеджера (доступ ко всем грузам).
func (s *MessageService) ListForManager(ctx context.Context, shipmentID uuid.UUID) ([]domain.Message, error) {
	if _, err := s.shipments.GetByID(ctx, shipmentID); err != nil {
		return nil, err
	}

	return s.messages.ListByShipment(ctx, shipmentID)
}

func (s *MessageService) notifyManager(shipment *domain.Shipment, clientID uuid.UUID, text string) {
	if s.managerNotify == nil {
		return
	}
	s.bg.Go(func() {
		ctx := context.Background()
		client, err := s.clients.GetByID(ctx, clientID)
		if err != nil {
			slog.Error("notify client message: load client", "shipment_id", shipment.ID, "error", err)
			return
		}
		if err := s.managerNotify.NotifyClientMessage(ctx, client, shipment, text); err != nil {
			slog.Error("notify client message", "shipment_id", shipment.ID, "error", err)
		}
	})
}

func (s *MessageService) notifyClient(shipment *domain.Shipment, text string) {
	if s.clientNotify == nil {
		return
	}
	s.bg.Go(func() {
		ctx := context.Background()
		client, err := s.clients.GetByID(ctx, shipment.ClientID)
		if err != nil {
			slog.Error("notify manager reply: load client", "shipment_id", shipment.ID, "error", err)
			return
		}
		if err := s.clientNotify.NotifyManagerReply(ctx, client.TelegramID, shipment, text); err != nil {
			slog.Error("notify manager reply", "shipment_id", shipment.ID, "error", err)
		}
	})
}
