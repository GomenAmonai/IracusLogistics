package service_test

import (
	"context"
	"sync/atomic"
	"testing"

	"icaris-logistic/backend/internal/service"
)

func TestBackground_WaitBlocksUntilTaskCompletes(t *testing.T) {
	bg := service.NewBackground()
	release := make(chan struct{})
	var done atomic.Bool
	bg.Go(func() {
		<-release
		done.Store(true)
	})

	close(release)
	_ = bg.Wait(context.Background())

	if !done.Load() {
		t.Fatal("Wait should block until the submitted task has completed")
	}
}

func TestBackground_WaitReturnsErrorWhenTaskOutlastsContext(t *testing.T) {
	bg := service.NewBackground()
	release := make(chan struct{})
	t.Cleanup(func() { close(release) })
	bg.Go(func() { <-release })

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := bg.Wait(ctx)

	if err == nil {
		t.Fatal("Wait should return the context error when tasks do not drain before the deadline")
	}
}
