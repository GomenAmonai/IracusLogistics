package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

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

func (r *LeadRepository) List(ctx context.Context, limit, offset int) ([]domain.Lead, error) {
	var leads []domain.Lead
	err := r.db.WithContext(ctx).
		Order("created_at desc").
		Limit(limit).
		Offset(offset).
		Find(&leads).Error
	if err != nil {
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

// UpdateStatus меняет статус лида и возвращает обновлённую запись. UPDATE ... RETURNING *
// (clause.Returning) делает апдейт и чтение результата одним атомарным statement'ом — без
// гонки read-after-write, которая возможна при двух отдельных запросах. Транзакция здесь не
// нужна: правим одну строку, в отличие от ShipmentRepository.UpdateStatus, где пишутся две
// таблицы. GORM на Update без совпадений ошибку не даёт, поэтому RowsAffected == 0 трактуем
// как «лида с таким id нет» → domain.ErrNotFound.
func (r *LeadRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.LeadStatus) (*domain.Lead, error) {
	var lead domain.Lead
	result := r.db.WithContext(ctx).
		Model(&lead).
		Clauses(clause.Returning{}).
		Where("id = ?", id).
		Update("status", status)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, domain.ErrNotFound
	}

	return &lead, nil
}
