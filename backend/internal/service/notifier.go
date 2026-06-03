package service

import (
	"context"

	"iracus-logistic/backend/internal/domain"
)

// Notifier — то, что сервису нужно от уведомлений о новых лидах. Интерфейс объявлен
// здесь (на стороне потребителя), чтобы LeadService не зависел от пакета bot.
type Notifier interface {
	NotifyNewLead(ctx context.Context, lead *domain.Lead) error
}
