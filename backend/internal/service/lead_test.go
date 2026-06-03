package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"iracus-logistic/backend/internal/domain"
	"iracus-logistic/backend/internal/service"
)

type fakeLeadStore struct {
	lead    *domain.Lead
	updated domain.LeadStatus
}

func (f *fakeLeadStore) Create(ctx context.Context, lead *domain.Lead) error {
	lead.ID = uuid.New()
	f.lead = lead
	return nil
}

func (f *fakeLeadStore) List(ctx context.Context) ([]domain.Lead, error) {
	return nil, nil
}

func (f *fakeLeadStore) GetByID(ctx context.Context, id uuid.UUID) (*domain.Lead, error) {
	if f.lead == nil {
		return nil, domain.ErrNotFound
	}
	return f.lead, nil
}

func (f *fakeLeadStore) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.LeadStatus) error {
	f.updated = status
	if f.lead != nil {
		f.lead.Status = status
	}
	return nil
}

type noopNotifier struct{}

func (noopNotifier) NotifyNewLead(ctx context.Context, lead *domain.Lead) error { return nil }

func TestLeadService_CreateRejectsMissingName(t *testing.T) {
	svc := service.NewLeadService(&fakeLeadStore{}, noopNotifier{})

	_, err := svc.Create(context.Background(), service.CreateLeadInput{
		Phone:    "+79990001122",
		FromCity: "Guangzhou",
		ToCity:   "Moscow",
	})

	if !errors.Is(err, service.ErrValidation) {
		t.Errorf("expected ErrValidation, got %v", err)
	}
}

func TestLeadService_UpdateStatusRejectsUnknownStatus(t *testing.T) {
	svc := service.NewLeadService(&fakeLeadStore{}, noopNotifier{})

	_, err := svc.UpdateStatus(context.Background(), uuid.New(), domain.LeadStatus("bogus"))

	if !errors.Is(err, service.ErrValidation) {
		t.Errorf("expected ErrValidation, got %v", err)
	}
}

func TestLeadService_UpdateStatusAcceptsKnownStatus(t *testing.T) {
	store := &fakeLeadStore{lead: &domain.Lead{Status: domain.LeadStatusNew}}
	svc := service.NewLeadService(store, noopNotifier{})

	lead, err := svc.UpdateStatus(context.Background(), uuid.New(), domain.LeadStatusContacted)
	if err != nil {
		t.Fatalf("update status: %v", err)
	}

	if lead.Status != domain.LeadStatusContacted {
		t.Errorf("expected status %q, got %q", domain.LeadStatusContacted, lead.Status)
	}
}
