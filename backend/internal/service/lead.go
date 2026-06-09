package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"icaris-logistic/backend/internal/domain"
)

var ErrValidation = errors.New("validation failed")

// LeadStore — то, что сервису нужно от хранилища лидов. Интерфейс объявлен здесь
// (на стороне потребителя), чтобы сервис не зависел от конкретного repository.
type LeadStore interface {
	Create(ctx context.Context, lead *domain.Lead) error
	List(ctx context.Context) ([]domain.Lead, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Lead, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.LeadStatus) (*domain.Lead, error)
}

type LeadService struct {
	store    LeadStore
	notifier Notifier
	bg       *Background
}

func NewLeadService(store LeadStore, notifier Notifier, bg *Background) *LeadService {
	return &LeadService{store: store, notifier: notifier, bg: bg}
}

// CreateLeadInput — данные публичной формы. weight/volume опциональны (NullDecimal),
// остальные обязательные поля проверяет validate.
type CreateLeadInput struct {
	Name      string              `json:"name"`
	Phone     string              `json:"phone"`
	FromCity  string              `json:"from_city"`
	ToCity    string              `json:"to_city"`
	Weight    decimal.NullDecimal `json:"weight"`
	Volume    decimal.NullDecimal `json:"volume"`
	CargoType string              `json:"cargo_type"`
	Comment   string              `json:"comment"`
}

func (s *LeadService) Create(ctx context.Context, input CreateLeadInput) (*domain.Lead, error) {
	input = input.normalized()
	if err := validateCreateLead(input); err != nil {
		return nil, err
	}

	lead := &domain.Lead{
		Name:      input.Name,
		Phone:     input.Phone,
		FromCity:  input.FromCity,
		ToCity:    input.ToCity,
		Weight:    input.Weight,
		Volume:    input.Volume,
		CargoType: input.CargoType,
		Comment:   input.Comment,
	}

	if err := s.store.Create(ctx, lead); err != nil {
		return nil, err
	}

	s.notifyNewLead(lead)

	return lead, nil
}

func (s *LeadService) List(ctx context.Context) ([]domain.Lead, error) {
	return s.store.List(ctx)
}

func (s *LeadService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Lead, error) {
	return s.store.GetByID(ctx, id)
}

// UpdateStatus меняет статус лида, предварительно проверив, что значение из набора enum.
// Переходы свободные (без стейт-машины) — решение MVP; см. docs/tech-debt.md.
func (s *LeadService) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.LeadStatus) (*domain.Lead, error) {
	if !status.IsValid() {
		return nil, fmt.Errorf("%w: unknown status %q", ErrValidation, status)
	}

	return s.store.UpdateStatus(ctx, id, status)
}

// notifyNewLead уведомляет менеджера в фоне: задача в Background — чтобы латентность Telegram
// не тормозила ответ клиенту; context.Background(), т.к. ctx запроса отменится сразу после
// ответа; ошибку только логируем — создание лида не должно падать из-за уведомления. Через
// Background, а не bare go, чтобы отправка дренировалась при остановке процесса.
//
// NOTE: MVP — без ретраев и персистентности (не полный outbox); см. docs/tech-debt.md
func (s *LeadService) notifyNewLead(lead *domain.Lead) {
	if s.notifier == nil {
		return
	}
	s.bg.Go(func() {
		if err := s.notifier.NotifyNewLead(context.Background(), lead); err != nil {
			slog.Error("notify new lead", "lead_id", lead.ID, "error", err)
		}
	})
}

func (i CreateLeadInput) normalized() CreateLeadInput {
	i.Name = strings.TrimSpace(i.Name)
	i.Phone = strings.TrimSpace(i.Phone)
	i.FromCity = strings.TrimSpace(i.FromCity)
	i.ToCity = strings.TrimSpace(i.ToCity)
	i.CargoType = strings.TrimSpace(i.CargoType)
	i.Comment = strings.TrimSpace(i.Comment)
	return i
}

func validateCreateLead(input CreateLeadInput) error {
	if input.Name == "" {
		return fmt.Errorf("%w: name is required", ErrValidation)
	}
	if input.Phone == "" {
		return fmt.Errorf("%w: phone is required", ErrValidation)
	}
	if input.FromCity == "" {
		return fmt.Errorf("%w: from_city is required", ErrValidation)
	}
	if input.ToCity == "" {
		return fmt.Errorf("%w: to_city is required", ErrValidation)
	}

	return nil
}
