package repository_test

import (
	"context"
	"testing"
	"time"

	"gorm.io/gorm"

	"icaris-logistic/backend/internal/config"
	"icaris-logistic/backend/internal/db"
	"icaris-logistic/backend/internal/domain"
	"icaris-logistic/backend/internal/repository"
)

func setupNotificationRepo(t *testing.T) (*repository.NotificationRepository, *gorm.DB) {
	t.Helper()

	gdb, err := db.Connect(context.Background(), config.Load().DatabaseURL)
	if err != nil {
		t.Skipf("database unavailable, skipping integration test: %v", err)
	}

	return repository.NewNotificationRepository(gdb), gdb
}

func enqueueTestNotification(t *testing.T, repo *repository.NotificationRepository, gdb *gorm.DB) *domain.Notification {
	t.Helper()

	notification := &domain.Notification{Kind: "test", ChatID: 1, Text: "проверка"}
	if err := repo.Enqueue(context.Background(), notification); err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	t.Cleanup(func() {
		gdb.Delete(&domain.Notification{}, "id = ?", notification.ID)
	})

	return notification
}

func dueIDs(t *testing.T, repo *repository.NotificationRepository) map[string]bool {
	t.Helper()

	due, err := repo.Due(context.Background(), 100)
	if err != nil {
		t.Fatalf("due: %v", err)
	}
	ids := make(map[string]bool, len(due))
	for _, n := range due {
		ids[n.ID.String()] = true
	}

	return ids
}

func TestNotificationRepository_EnqueuedIsDueImmediately(t *testing.T) {
	repo, gdb := setupNotificationRepo(t)
	notification := enqueueTestNotification(t, repo, gdb)

	if !dueIDs(t, repo)[notification.ID.String()] {
		t.Error("expected freshly enqueued notification to be due")
	}
}

func TestNotificationRepository_MarkSentRemovesFromDue(t *testing.T) {
	repo, gdb := setupNotificationRepo(t)
	notification := enqueueTestNotification(t, repo, gdb)

	if err := repo.MarkSent(context.Background(), notification.ID); err != nil {
		t.Fatalf("mark sent: %v", err)
	}

	if dueIDs(t, repo)[notification.ID.String()] {
		t.Error("expected sent notification to leave the due queue")
	}
}

func TestNotificationRepository_MarkFailedSchedulesRetryInFuture(t *testing.T) {
	repo, gdb := setupNotificationRepo(t)
	notification := enqueueTestNotification(t, repo, gdb)

	err := repo.MarkFailed(context.Background(), notification.ID, "telegram down", time.Now().Add(time.Hour), false)
	if err != nil {
		t.Fatalf("mark failed: %v", err)
	}

	if dueIDs(t, repo)[notification.ID.String()] {
		t.Error("expected retry scheduled in the future to not be due now")
	}
}

func TestNotificationRepository_MarkFailedIncrementsAttempts(t *testing.T) {
	repo, gdb := setupNotificationRepo(t)
	notification := enqueueTestNotification(t, repo, gdb)

	if err := repo.MarkFailed(context.Background(), notification.ID, "x", time.Now(), false); err != nil {
		t.Fatalf("mark failed: %v", err)
	}

	var stored domain.Notification
	if err := gdb.First(&stored, "id = ?", notification.ID).Error; err != nil {
		t.Fatalf("reload: %v", err)
	}
	if stored.Attempts != 1 {
		t.Errorf("expected attempts incremented to 1, got %d", stored.Attempts)
	}
}

func TestNotificationRepository_FinalFailureLeavesQueueForGood(t *testing.T) {
	repo, gdb := setupNotificationRepo(t)
	notification := enqueueTestNotification(t, repo, gdb)

	// next_attempt_at в прошлом: из очереди уведомление держит только терминальный статус.
	err := repo.MarkFailed(context.Background(), notification.ID, "blocked", time.Now().Add(-time.Minute), true)
	if err != nil {
		t.Fatalf("mark failed final: %v", err)
	}

	if dueIDs(t, repo)[notification.ID.String()] {
		t.Error("expected finally-failed notification to never be due again")
	}
}
