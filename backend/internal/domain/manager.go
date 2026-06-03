package domain

import (
	"time"

	"github.com/google/uuid"
)

// Manager — сотрудник компании, ведёт лиды и грузы.
// Авторизуется по email и паролю; в Password хранится bcrypt-хеш, а не открытый пароль.
type Manager struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Email     string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Password  string    `gorm:"type:varchar(255);not null" json:"-"`
	Name      string    `gorm:"type:varchar(255);not null" json:"name"`
	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`
}
