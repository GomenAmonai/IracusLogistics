package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"icaris-logistic/backend/internal/domain"
)

// pgForeignKeyViolation — код Postgres «нарушение внешнего ключа» (класс 23).
const pgForeignKeyViolation = "23503"

type PaymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

// Create вставляет платёж. Сервис проверяет существование груза до вставки, но между
// проверкой и INSERT груз могли удалить — нарушение FK на shipment_id транслируем в
// domain.ErrNotFound, чтобы гонка дала клиенту честный 404, а не 500.
func (r *PaymentRepository) Create(ctx context.Context, payment *domain.Payment) error {
	err := r.db.WithContext(ctx).Create(payment).Error
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgForeignKeyViolation &&
		pgErr.ConstraintName == "payments_shipment_id_fkey" {
		return domain.ErrNotFound
	}

	return err
}

func (r *PaymentRepository) ListByShipment(ctx context.Context, shipmentID uuid.UUID) ([]domain.Payment, error) {
	var payments []domain.Payment
	// id как вторичный ключ — детерминированный порядок при совпадении created_at
	// (как в StatusHistory).
	err := r.db.WithContext(ctx).
		Where("shipment_id = ?", shipmentID).
		Order("created_at asc, id asc").
		Find(&payments).Error
	if err != nil {
		return nil, err
	}

	return payments, nil
}

func (r *PaymentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error) {
	var payment domain.Payment
	err := r.db.WithContext(ctx).First(&payment, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &payment, nil
}

// UpdateStatus меняет статус платежа и возвращает обновлённую запись одним
// UPDATE ... RETURNING * (как LeadRepository.UpdateStatus): одна строка, транзакция не
// нужна. RowsAffected == 0 → платежа с таким id нет → domain.ErrNotFound.
func (r *PaymentRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.PaymentStatus) (*domain.Payment, error) {
	var payment domain.Payment
	result := r.db.WithContext(ctx).
		Model(&payment).
		Clauses(clause.Returning{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":     status,
			"updated_at": gorm.Expr("now()"),
		})
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, domain.ErrNotFound
	}

	return &payment, nil
}
