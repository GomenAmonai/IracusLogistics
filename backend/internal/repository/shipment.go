package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"icaris-logistic/backend/internal/domain"
)

type ShipmentRepository struct {
	db *gorm.DB
}

func NewShipmentRepository(db *gorm.DB) *ShipmentRepository {
	return &ShipmentRepository{db: db}
}

// Create вставляет груз и первую запись истории статуса одной транзакцией: груз без
// начального события оставил бы клиентский таймлайн неполным, поэтому атомарно.
func (r *ShipmentRepository) Create(ctx context.Context, shipment *domain.Shipment) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(shipment).Error; err != nil {
			return err
		}

		event := &domain.ShipmentStatusEvent{
			ShipmentID: shipment.ID,
			Status:     shipment.Status,
			Comment:    shipment.StatusComment,
			ChangedBy:  &shipment.ManagerID,
		}

		return tx.Create(event).Error
	})
}

func (r *ShipmentRepository) List(ctx context.Context) ([]domain.Shipment, error) {
	var shipments []domain.Shipment
	if err := r.db.WithContext(ctx).Order("created_at desc").Find(&shipments).Error; err != nil {
		return nil, err
	}

	return shipments, nil
}

func (r *ShipmentRepository) ListByClient(ctx context.Context, clientID uuid.UUID) ([]domain.Shipment, error) {
	var shipments []domain.Shipment
	err := r.db.WithContext(ctx).
		Where("client_id = ?", clientID).
		Order("created_at desc").
		Find(&shipments).Error
	if err != nil {
		return nil, err
	}

	return shipments, nil
}

func (r *ShipmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Shipment, error) {
	var shipment domain.Shipment
	err := r.db.WithContext(ctx).First(&shipment, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &shipment, nil
}

// UpdateStatus меняет статус груза и добавляет запись истории одной транзакцией. При
// первом переходе в delivered проставляет delivered_at. Несуществующий груз →
// domain.ErrNotFound. Возвращает обновлённый груз (перечитанный после апдейта).
func (r *ShipmentRepository) UpdateStatus(
	ctx context.Context,
	id uuid.UUID,
	status domain.ShipmentStatus,
	comment string,
	changedBy uuid.UUID,
) (*domain.Shipment, error) {
	var updated domain.Shipment

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current domain.Shipment
		if err := tx.First(&current, "id = ?", id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.ErrNotFound
			}
			return err
		}

		updates := map[string]any{
			"status":         status,
			"status_comment": comment,
			"updated_at":     gorm.Expr("now()"),
		}
		// delivered_at синхронизируем со статусом: ставим при первом переходе в delivered,
		// снимаем при уходе из него (переходы свободные — менеджер может откатить статус, и
		// тогда штамп доставки не должен «зависнуть»).
		if status == domain.ShipmentStatusDelivered && current.DeliveredAt == nil {
			updates["delivered_at"] = gorm.Expr("now()")
		} else if status != domain.ShipmentStatusDelivered && current.DeliveredAt != nil {
			updates["delivered_at"] = nil
		}
		if err := tx.Model(&current).Updates(updates).Error; err != nil {
			return err
		}

		event := &domain.ShipmentStatusEvent{
			ShipmentID: id,
			Status:     status,
			Comment:    comment,
			ChangedBy:  &changedBy,
		}
		if err := tx.Create(event).Error; err != nil {
			return err
		}

		// Перечитываем: Updates с map не пишет now()-значения обратно в структуру.
		return tx.First(&updated, "id = ?", id).Error
	})
	if err != nil {
		return nil, err
	}

	return &updated, nil
}

func (r *ShipmentRepository) StatusHistory(ctx context.Context, shipmentID uuid.UUID) ([]domain.ShipmentStatusEvent, error) {
	var events []domain.ShipmentStatusEvent
	// id как вторичный ключ — детерминированный порядок при совпадении created_at (события
	// в отдельных транзакциях могут разделить микросекунду); иначе строки с равным временем
	// БД вернёт в произвольном порядке и таймлайн «прыгал» бы между запросами.
	err := r.db.WithContext(ctx).
		Where("shipment_id = ?", shipmentID).
		Order("created_at asc, id asc").
		Find(&events).Error
	if err != nil {
		return nil, err
	}

	return events, nil
}

func (r *ShipmentRepository) ExistsTrackingKey(ctx context.Context, key string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Shipment{}).
		Where("tracking_key = ?", key).
		Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
