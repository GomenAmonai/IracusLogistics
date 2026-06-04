package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"icaris-logistic/backend/internal/domain"
	"icaris-logistic/backend/internal/service"
)

type fakeClientStore struct {
	byTelegram  map[int64]*domain.Client
	createCalls int
	// failIfLeadIDSet имитирует нарушение partial-unique uq_clients_lead_id: вставка с уже
	// занятым lead_id падает (как в БД), вставка без lead_id проходит.
	failIfLeadIDSet bool
}

func (f *fakeClientStore) GetByTelegramID(ctx context.Context, telegramID int64) (*domain.Client, error) {
	if c, ok := f.byTelegram[telegramID]; ok {
		return c, nil
	}
	return nil, domain.ErrNotFound
}

func (f *fakeClientStore) Create(ctx context.Context, client *domain.Client) error {
	f.createCalls++
	if f.failIfLeadIDSet && client.LeadID != nil {
		return errors.New("duplicate key value violates unique constraint uq_clients_lead_id")
	}
	client.ID = uuid.New() // имитируем генерацию id базой
	if f.byTelegram == nil {
		f.byTelegram = map[int64]*domain.Client{}
	}
	f.byTelegram[client.TelegramID] = client
	return nil
}

func (f *fakeClientStore) List(ctx context.Context) ([]domain.Client, error) {
	return nil, nil
}

func TestClientService_RegisterCreatesClientWithTelegramID(t *testing.T) {
	svc := service.NewClientService(&fakeClientStore{}, "token", "secret", time.Hour)

	client, err := svc.Register(context.Background(), 42, "ivan", "Иван", nil)
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	if client.TelegramID != 42 {
		t.Errorf("expected telegram id 42, got %d", client.TelegramID)
	}
}

func TestClientService_RegisterIsIdempotentForSameTelegramID(t *testing.T) {
	store := &fakeClientStore{}
	svc := service.NewClientService(store, "token", "secret", time.Hour)

	_, _ = svc.Register(context.Background(), 42, "ivan", "Иван", nil)
	_, _ = svc.Register(context.Background(), 42, "ivan", "Иван", nil)

	if store.createCalls != 1 {
		t.Errorf("expected exactly one create for repeated telegram id, got %d", store.createCalls)
	}
}

func TestClientService_RegisterUnbindsLeadWhenLeadAlreadyClaimed(t *testing.T) {
	// Двое разных Telegram-юзеров открыли один deep-link /start=<lead_id>: lead уже занят.
	// Второй должен зарегистрироваться без привязки, а не залочиться навсегда.
	store := &fakeClientStore{failIfLeadIDSet: true}
	svc := service.NewClientService(store, "token", "secret", time.Hour)
	leadID := uuid.New()

	client, err := svc.Register(context.Background(), 55, "u", "Имя", &leadID)
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	if client.LeadID != nil {
		t.Errorf("expected lead to be unbound after lead_id collision, got %v", client.LeadID)
	}
}

func TestClientService_RegisterFallsBackToDefaultNameWhenEmpty(t *testing.T) {
	svc := service.NewClientService(&fakeClientStore{}, "token", "secret", time.Hour)

	client, err := svc.Register(context.Background(), 7, "", "  ", nil)
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	if client.Name != "Клиент" {
		t.Errorf("expected fallback name %q, got %q", "Клиент", client.Name)
	}
}
