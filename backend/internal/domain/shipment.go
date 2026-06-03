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
