package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"icaris-logistic/backend/internal/domain"
)

type MessageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) Create(ctx context.Context, message *domain.Message) error {
	return r.db.WithContext(ctx).Create(message).Error
}

// maxChildRows — потолок дочерних списков груза (сообщения, платежи): объём на один груз
// мал, offset-пагинация чату не нужна, но неограниченный SELECT — DoS-вектор (техдолг #25).
const maxChildRows = 500

// ListByShipment возвращает переписку по грузу в хронологическом порядке (старые сверху) —
// так её удобно показывать как ленту чата.
func (r *MessageRepository) ListByShipment(ctx context.Context, shipmentID uuid.UUID) ([]domain.Message, error) {
	var messages []domain.Message
	// id — вторичный ключ для детерминированного порядка при равном created_at (см. тот же
	// приём в ShipmentRepository.StatusHistory).
	err := r.db.WithContext(ctx).
		Where("shipment_id = ?", shipmentID).
		Order("created_at asc, id asc").
		Limit(maxChildRows).
		Find(&messages).Error
	if err != nil {
		return nil, err
	}

	return messages, nil
}
