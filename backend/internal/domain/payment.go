package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Payment — платёж по грузу. Бизнес мультиканальный: один груз может закрываться
// несколькими платежами разными способами (часть безналом по счёту, часть наличными),
// поэтому платёж — отдельная запись с собственным каналом, валютой и статусом, а не
// поле на грузе. CreatedBy nullable — менеджера могли удалить, платёж остаётся.
type Payment struct {
	ID         uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ShipmentID uuid.UUID       `gorm:"type:uuid;not null;index" json:"shipment_id"`
	Amount     decimal.Decimal `gorm:"type:numeric;not null" json:"amount"`
	Currency   string          `gorm:"type:varchar(3);not null" json:"currency"`
	Channel    PaymentChannel  `gorm:"type:varchar(20);not null" json:"channel"`
	Status     PaymentStatus   `gorm:"type:varchar(20);not null;default:pending" json:"status"`
	Comment    string          `gorm:"type:text" json:"comment"`
	CreatedBy  *uuid.UUID      `gorm:"type:uuid" json:"created_by"`
	CreatedAt  time.Time       `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt  time.Time       `gorm:"not null;default:now()" json:"updated_at"`
}

// PaymentChannel — канал оплаты. Набор отражает текущую мультиканальную модель MVP;
// доступные клиенту способы должны быть подтверждены до production-запуска.
type PaymentChannel string

const (
	PaymentChannelBankTransfer PaymentChannel = "bank_transfer" // безнал на расчётный счёт юрлица
	PaymentChannelCardSBP      PaymentChannel = "card_sbp"      // карта РФ или СБП
	PaymentChannelCash         PaymentChannel = "cash"          // наличные при выдаче / в офисе
	PaymentChannelCrypto       PaymentChannel = "crypto"        // USDT и прочая крипта
)

// IsValid сообщает, входит ли канал в допустимый набор. Сервис проверяет это перед
// записью, чтобы вернуть 400, а не словить 500 от CHECK-ограничения в БД.
func (c PaymentChannel) IsValid() bool {
	switch c {
	case PaymentChannelBankTransfer, PaymentChannelCardSBP, PaymentChannelCash, PaymentChannelCrypto:
		return true
	default:
		return false
	}
}

// PaymentStatus — состояние платежа: ожидаем (зафиксирована договорённость), получен,
// возвращён. Переходы свободные (без стейт-машины) — как у статусов груза, решение MVP.
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusConfirmed PaymentStatus = "confirmed"
	PaymentStatusRefunded  PaymentStatus = "refunded"
)

func (s PaymentStatus) IsValid() bool {
	switch s {
	case PaymentStatusPending, PaymentStatusConfirmed, PaymentStatusRefunded:
		return true
	default:
		return false
	}
}
