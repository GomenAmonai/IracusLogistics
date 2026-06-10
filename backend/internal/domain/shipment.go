package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Shipment — груз клиента в работе. Создаётся менеджером после подтверждения сделки,
// отслеживается клиентом по tracking_key.
type Shipment struct {
	ID            uuid.UUID           `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ClientID      uuid.UUID           `gorm:"type:uuid;not null;index" json:"client_id"`
	ManagerID     uuid.UUID           `gorm:"type:uuid;not null;index" json:"manager_id"`
	TrackingKey   string              `gorm:"type:varchar(64);uniqueIndex;not null" json:"tracking_key"`
	Lane          Lane                `gorm:"type:varchar(10);not null;default:cargo" json:"lane"`
	Status        ShipmentStatus      `gorm:"type:varchar(20);not null;default:pending" json:"status"`
	StatusComment string              `gorm:"type:text" json:"status_comment"`
	Weight        decimal.NullDecimal `gorm:"type:numeric" json:"weight"`
	Volume        decimal.NullDecimal `gorm:"type:numeric" json:"volume"`
	FromCity      string              `gorm:"type:varchar(255)" json:"from_city"`
	ToCity        string              `gorm:"type:varchar(255)" json:"to_city"`
	Price         decimal.NullDecimal `gorm:"type:numeric" json:"price"`
	Currency      string              `gorm:"type:varchar(3);not null;default:USD" json:"currency"`
	EstimatedAt   *time.Time          `json:"estimated_at"`
	DeliveredAt   *time.Time          `json:"delivered_at"`
	CreatedAt     time.Time           `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt     time.Time           `gorm:"not null;default:now()" json:"updated_at"`
}

// Lane — полоса доставки: каким режимом везём груз. Ратифицировано Hiki 2026-06-10 как
// метка на грузе (не три раздельных потока): процессы полос в MVP совпадают, различия —
// в документах и цене; раздельные статусные машины отложены до реальной необходимости.
type Lane string

const (
	LaneCargo  Lane = "cargo"  // карго: быстрая доставка через Казахстан/Азербайджан
	LaneWhite  Lane = "white"  // белый импорт: с документами, НДС, Честным знаком
	LaneBuyout Lane = "buyout" // выкуп: 找货, оплата поставщику, проверка на складе в КНР
)

// IsValid сообщает, входит ли полоса в допустимый набор (CHECK в БД дублирует на своём
// уровне; сервис проверяет раньше, чтобы вернуть 400, а не 500).
func (l Lane) IsValid() bool {
	switch l {
	case LaneCargo, LaneWhite, LaneBuyout:
		return true
	default:
		return false
	}
}

type ShipmentStatus string

const (
	ShipmentStatusPending        ShipmentStatus = "pending"
	ShipmentStatusPickedUp       ShipmentStatus = "picked_up"
	ShipmentStatusInTransit      ShipmentStatus = "in_transit"
	ShipmentStatusCustomsClear   ShipmentStatus = "customs_clear"
	ShipmentStatusInWarehouse    ShipmentStatus = "in_warehouse"
	ShipmentStatusOutForDelivery ShipmentStatus = "out_for_delivery"
	ShipmentStatusDelivered      ShipmentStatus = "delivered"
	ShipmentStatusCancelled      ShipmentStatus = "cancelled"
)

// IsValid сообщает, входит ли статус в допустимый набор. Сервис проверяет это перед
// записью, чтобы вернуть 400, а не словить 500 от CHECK-ограничения в БД.
func (s ShipmentStatus) IsValid() bool {
	switch s {
	case ShipmentStatusPending, ShipmentStatusPickedUp, ShipmentStatusInTransit,
		ShipmentStatusCustomsClear, ShipmentStatusInWarehouse, ShipmentStatusOutForDelivery,
		ShipmentStatusDelivered, ShipmentStatusCancelled:
		return true
	default:
		return false
	}
}

// ShipmentStatusEvent — запись в истории смены статуса груза. Создаётся при заведении
// груза (начальный статус) и при каждой смене статуса; из этих записей клиент видит
// таймлайн в WebApp. ChangedBy nullable — менеджера могли удалить.
type ShipmentStatusEvent struct {
	ID         uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ShipmentID uuid.UUID      `gorm:"type:uuid;not null;index" json:"shipment_id"`
	Status     ShipmentStatus `gorm:"type:varchar(20);not null" json:"status"`
	Comment    string         `gorm:"type:text" json:"comment"`
	ChangedBy  *uuid.UUID     `gorm:"type:uuid" json:"changed_by"`
	CreatedAt  time.Time      `gorm:"not null;default:now()" json:"created_at"`
}
