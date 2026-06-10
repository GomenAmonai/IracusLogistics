package bot

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"icaris-logistic/backend/internal/domain"
)

// OutboxStore — что боту нужно от хранилища outbox-уведомлений (интерфейс на стороне
// потребителя; реализует repository.NotificationRepository).
type OutboxStore interface {
	Enqueue(ctx context.Context, notification *domain.Notification) error
	Due(ctx context.Context, limit int) ([]domain.Notification, error)
	MarkSent(ctx context.Context, id uuid.UUID) error
	MarkFailed(ctx context.Context, id uuid.UUID, lastError string, nextAttemptAt time.Time, final bool) error
}

const (
	// outboxPollInterval — пауза между выборками очереди. Уведомления не латентно-критичны
	// (смена статуса груза), 3 секунды незаметны пользователю и не грузят БД.
	outboxPollInterval = 3 * time.Second
	outboxBatchSize    = 20
	// outboxMaxAttempts: с backoff 30s·2^n десятая попытка уходит за ~4 часа суммарно —
	// дальше Telegram-чат, скорее всего, недоступен навсегда (бот заблокирован), фиксируем failed.
	outboxMaxAttempts = 10
	outboxBaseBackoff = 30 * time.Second
	outboxMaxBackoff  = 30 * time.Minute
)

// UseOutbox переводит уведомления в режим outbox: Notify*-методы ставят сообщение в
// очередь вместо прямой отправки, RunOutbox доставляет с ретраями. Вызывается один раз
// при сборке процесса, до запуска горутин.
func (b *Bot) UseOutbox(store OutboxStore) {
	b.outbox = store
}

// RunOutbox доставляет уведомления из очереди до отмены ctx. Блокирующий — вызывается в
// отдельной горутине (через Background, как Run). Без outbox или в no-op режиме сразу
// возвращается.
func (b *Bot) RunOutbox(ctx context.Context) {
	if b.disabled || b.outbox == nil {
		return
	}

	ticker := time.NewTicker(outboxPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			b.dispatchDue(ctx)
		}
	}
}

// dispatchDue отправляет готовые к доставке уведомления. Ошибка отправки одного не
// останавливает остальные: у каждого свой счётчик попыток и своё расписание ретраев.
func (b *Bot) dispatchDue(ctx context.Context) {
	batch, err := b.outbox.Due(ctx, outboxBatchSize)
	if err != nil {
		slog.Error("outbox: load due notifications", "error", err)
		return
	}

	// Пометки идут без отмены: shutdown между отправкой и MarkSent отменил бы запись —
	// и уже доставленное сообщение ушло бы дублем на следующем старте. Дренаж bg.Wait
	// дождётся завершения текущего батча, пул БД закрывается позже (defer в main).
	markCtx := context.WithoutCancel(ctx)

	for _, n := range batch {
		if ctx.Err() != nil {
			return
		}

		if err := b.send(n.ChatID, n.Text); err != nil {
			attempt := n.Attempts + 1
			final := attempt >= outboxMaxAttempts
			next := time.Now().Add(outboxBackoff(attempt))
			if markErr := b.outbox.MarkFailed(markCtx, n.ID, err.Error(), next, final); markErr != nil {
				slog.Error("outbox: mark failed", "notification_id", n.ID, "error", markErr)
			}
			if final {
				slog.Error("outbox: notification gave up after max attempts", "notification_id", n.ID, "kind", n.Kind)
			}
			continue
		}

		if err := b.outbox.MarkSent(markCtx, n.ID); err != nil {
			// Сообщение ушло, а пометка не записалась — на следующем тике уйдёт дубль.
			// At-least-once: дубль уведомления приемлемее потери.
			slog.Error("outbox: mark sent", "notification_id", n.ID, "error", err)
		}
	}
}

// outboxBackoff — экспоненциальная задержка ретрая: 30s, 1m, 2m, ... с потолком 30m.
func outboxBackoff(attempt int) time.Duration {
	backoff := outboxBaseBackoff
	for i := 1; i < attempt; i++ {
		backoff *= 2
		if backoff >= outboxMaxBackoff {
			return outboxMaxBackoff
		}
	}

	return backoff
}

// deliver — единая точка исходящих уведомлений: в режиме outbox ставит в очередь
// (доставит RunOutbox с ретраями), иначе шлёт напрямую. В no-op режиме — ничего.
func (b *Bot) deliver(ctx context.Context, kind string, chatID int64, text string) error {
	if b.disabled {
		return nil
	}
	if b.outbox != nil {
		return b.outbox.Enqueue(ctx, &domain.Notification{Kind: kind, ChatID: chatID, Text: text})
	}

	return b.send(chatID, text)
}
