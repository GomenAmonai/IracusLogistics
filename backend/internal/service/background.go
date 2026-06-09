package service

import (
	"context"
	"sync"
)

// Background выполняет фоновые задачи (уведомления, цикл бота) и учитывает их в WaitGroup,
// чтобы при остановке процесса дождаться их завершения, а не обрывать на полуслове.
//
// Это минимальный дрейн, а НЕ полноценный outbox: ретраев и персистентности нет, незавершённые
// при таймауте задачи теряются. Полный outbox — см. docs/tech-debt.md (#11/#15).
type Background struct {
	wg sync.WaitGroup
}

func NewBackground() *Background {
	return &Background{}
}

// Go запускает задачу в горутине, отслеживая её для последующего дрейна. Nil-приёмник
// допустим (тесты/несконфигурированный путь): тогда задача запускается без учёта.
func (b *Background) Go(task func()) {
	if b == nil {
		go task()
		return
	}

	b.wg.Go(task)
}

// Wait блокируется до завершения всех задач либо до отмены ctx (таймаут shutdown).
// Возвращает ctx.Err(), если не дождались. Вызывать ПОСЛЕ остановки источников новых
// задач (HTTP-сервер, отмена ctx бота) — иначе возможен race по WaitGroup.
func (b *Background) Wait(ctx context.Context) error {
	if b == nil {
		return nil
	}

	done := make(chan struct{})
	go func() {
		b.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
