package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"icaris-logistic/backend/internal/domain"
)

type LeadRepository struct {
	db *gorm.DB
}

func NewLeadRepository(db *gorm.DB) *LeadRepository {
	return &LeadRepository{db: db}
}

// Create вставляет лид. ID, status и created_at, не заданные в lead,
// проставит база (default-значения из миграции), а GORM вернёт их обратно в lead.
func (r *LeadRepository) Create(ctx context.Context, lead *domain.Lead) error {
	return r.db.WithContext(ctx).Create(lead).Error
}

func (r *LeadRepository) List(ctx context.Context) ([]domain.Lead, error) {
	var leads []domain.Lead
	if err := r.db.WithContext(ctx).Order("created_at desc").Find(&leads).Error; err != nil {
		return nil, err
	}

	return leads, nil
}

func (r *LeadRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Lead, error) {
	var lead domain.Lead
	err := r.db.WithContext(ctx).First(&lead, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &lead, nil
}

// UpdateStatus меняет статус лида. GORM на Update без совпадений ошибку не даёт, поэтому
// RowsAffected == 0 трактуем как «лида с таким id нет» → domain.ErrNotFound.
func (r *LeadRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.LeadStatus) error {
	result := r.db.WithContext(ctx).
		Model(&domain.Lead{}).
		Where("id = ?", id).
		Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}
