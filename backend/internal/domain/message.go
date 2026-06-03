package domain

import (
	"time"

	"github.com/google/uuid"
)

// Message — сообщение в переписке клиента и менеджера, опционально привязанное к грузу.
type Message struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ShipmentID *uuid.UUID `gorm:"type:uuid;index" json:"shipment_id"`
	ClientID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"client_id"`
	ManagerID  *uuid.UUID `gorm:"type:uuid" json:"manager_id"`
	Text       string     `gorm:"type:text;not null" json:"text"`
	FromRole   Role       `gorm:"type:varchar(10);not null" json:"from_role"`
	CreatedAt  time.Time  `gorm:"not null;default:now()" json:"created_at"`
}

type Role string

const (
	RoleClient  Role = "client"
	RoleManager Role = "manager"
)
