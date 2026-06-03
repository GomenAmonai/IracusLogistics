package repository_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"iracus-logistic/backend/internal/config"
	"iracus-logistic/backend/internal/db"
	"iracus-logistic/backend/internal/domain"
	"iracus-logistic/backend/internal/repository"
)

func setupLeadRepo(t *testing.T) (*repository.LeadRepository, *gorm.DB) {
	t.Helper()

	gdb, err := db.Connect(context.Background(), config.Load().DatabaseURL)
	if err != nil {
		t.Skipf("database unavailable, skipping integration test: %v", err)
	}

	return repository.NewLeadRepository(gdb), gdb
}

func createTestLead(t *testing.T, repo *repository.LeadRepository, gdb *gorm.DB) *domain.Lead {
	t.Helper()

	lead := &domain.Lead{
		Name:     "Тест",
		Phone:    "+79990001122",
		FromCity: "Guangzhou",
		ToCity:   "Moscow",
	}
	if err := repo.Create(context.Background(), lead); err != nil {
		t.Fatalf("create lead: %v", err)
	}
	t.Cleanup(func() {
		gdb.Delete(&domain.Lead{}, "id = ?", lead.ID)
	})

	return lead
}

func TestLeadRepository_CreateAssignsID(t *testing.T) {
	repo, gdb := setupLeadRepo(t)

	lead := createTestLead(t, repo, gdb)

	if lead.ID == uuid.Nil {
		t.Errorf("expected database-generated ID, got nil UUID")
	}
}

func TestLeadRepository_CreateDefaultsStatusToNew(t *testing.T) {
	repo, gdb := setupLeadRepo(t)

	lead := createTestLead(t, repo, gdb)

	if lead.Status != domain.LeadStatusNew {
		t.Errorf("expected default status %q, got %q", domain.LeadStatusNew, lead.Status)
	}
}

func TestLeadRepository_GetByIDReturnsCreatedLead(t *testing.T) {
	repo, gdb := setupLeadRepo(t)
	lead := createTestLead(t, repo, gdb)

	got, err := repo.GetByID(context.Background(), lead.ID)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}

	if got.Name != lead.Name {
		t.Errorf("expected name %q, got %q", lead.Name, got.Name)
	}
}

func TestLeadRepository_GetByIDUnknownReturnsNotFound(t *testing.T) {
	repo, _ := setupLeadRepo(t)

	_, err := repo.GetByID(context.Background(), uuid.New())

	if err != repository.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
