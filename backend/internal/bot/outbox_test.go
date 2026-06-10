package bot

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"icaris-logistic/backend/internal/domain"
)

type fakeOutboxStore struct {
	enqueued *domain.Notification
}

func (f *fakeOutboxStore) Enqueue(_ context.Context, n *domain.Notification) error {
	f.enqueued = n
	return nil
}

func (f *fakeOutboxStore) Due(_ context.Context, _ int) ([]domain.Notification, error) {
	return nil, nil
}

func (f *fakeOutboxStore) MarkSent(_ context.Context, _ uuid.UUID) error { return nil }

func (f *fakeOutboxStore) MarkFailed(_ context.Context, _ uuid.UUID, _ string, _ time.Time, _ bool) error {
	return nil
}

func TestNotifyNewLeadEnqueuesWhenOutboxConfigured(t *testing.T) {
	store := &fakeOutboxStore{}
	// api == nil: если бы Notify пошёл в прямую отправку, тест бы упал паникой — это
	// и есть проверка, что в режиме outbox сеть не трогается.
	b := &Bot{chatID: 42, outbox: store}

	if err := b.NotifyNewLead(context.Background(), &domain.Lead{Name: "Иван"}); err != nil {
		t.Fatalf("notify: %v", err)
	}

	if store.enqueued == nil || store.enqueued.ChatID != 42 {
		t.Errorf("expected notification enqueued for manager chat 42, got %+v", store.enqueued)
	}
}

func TestNotifyShipmentStatusEnqueuesForClientChat(t *testing.T) {
	store := &fakeOutboxStore{}
	b := &Bot{chatID: 42, outbox: store}

	err := b.NotifyShipmentStatus(context.Background(), 777, &domain.Shipment{TrackingKey: "ICR-X"})
	if err != nil {
		t.Fatalf("notify: %v", err)
	}

	if store.enqueued == nil || store.enqueued.ChatID != 777 {
		t.Errorf("expected notification enqueued for client chat 777, got %+v", store.enqueued)
	}
}

func TestDeliverSkipsEnqueueWhenDisabled(t *testing.T) {
	store := &fakeOutboxStore{}
	b := &Bot{disabled: true, outbox: store}

	if err := b.NotifyNewLead(context.Background(), &domain.Lead{}); err != nil {
		t.Fatalf("notify: %v", err)
	}

	if store.enqueued != nil {
		t.Errorf("disabled bot must not enqueue, got %+v", store.enqueued)
	}
}

func TestOutboxBackoffGrowsExponentially(t *testing.T) {
	if got := outboxBackoff(3); got != 2*time.Minute {
		t.Errorf("expected attempt 3 backoff 2m, got %v", got)
	}
}

func TestOutboxBackoffIsCapped(t *testing.T) {
	if got := outboxBackoff(20); got != outboxMaxBackoff {
		t.Errorf("expected capped backoff %v, got %v", outboxMaxBackoff, got)
	}
}

func TestProcessUpdateJSONRejectsInvalidJSON(t *testing.T) {
	b := &Bot{chatID: 1, webhookDeps: &RunDeps{}}

	if err := b.ProcessUpdateJSON(context.Background(), []byte("not-json")); err == nil {
		t.Error("expected error for unparseable update body")
	}
}

func TestProcessUpdateJSONIgnoresUpdatesWithoutWebhookMode(t *testing.T) {
	b := &Bot{chatID: 1}

	if err := b.ProcessUpdateJSON(context.Background(), []byte(`{"update_id":1}`)); err != nil {
		t.Errorf("expected update to be ignored without webhook mode, got %v", err)
	}
}

func TestProcessUpdateJSONIgnoresNonCommandUpdate(t *testing.T) {
	b := &Bot{chatID: 1, webhookDeps: &RunDeps{}}

	// Апдейт без message/callback не должен трогать ни API (nil), ни deps (пустые).
	if err := b.ProcessUpdateJSON(context.Background(), []byte(`{"update_id":7}`)); err != nil {
		t.Errorf("expected non-command update to be a no-op, got %v", err)
	}
}
