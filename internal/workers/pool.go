package workers

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/onurceri/botla-co/pkg/logger"
)

type WorkerPool struct {
	wg      sync.WaitGroup
	jobs    chan func(context.Context)
	ctx     context.Context
	cancel  context.CancelFunc
	logger  *logger.Logger
	maxSize int
}

func NewWorkerPool(logger *logger.Logger, size int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	pool := &WorkerPool{
		jobs:    make(chan func(context.Context), size),
		ctx:     ctx,
		cancel:  cancel,
		logger:  logger,
		maxSize: size,
	}

	for i := 0; i < size; i++ {
		pool.wg.Add(1)
		go pool.worker()
	}
	return pool
}

func (p *WorkerPool) worker() {
	defer p.wg.Done()
	for {
		select {
		case job := <-p.jobs:
			p.executeJob(job)
		case <-p.ctx.Done():
			// Drain remaining jobs
			for {
				select {
				case job := <-p.jobs:
					p.executeJob(job)
				default:
					return
				}
			}
		}
	}
}

func (p *WorkerPool) executeJob(job func(context.Context)) {
	defer func() {
		if r := recover(); r != nil {
			p.logger.Error("worker_panic", map[string]any{
				"panic": fmt.Sprintf("%v", r),
				"stack": string(debug.Stack()),
			})
		}
	}()

	ctx, cancel := context.WithTimeout(p.ctx, 5*time.Second)
	defer cancel()
	job(ctx)
}

func (p *WorkerPool) Submit(job func(context.Context)) bool {
	select {
	case p.jobs <- job:
		return true
	default:
		p.logger.Warn("worker_pool_full", nil)
		return false
	}
}

func (p *WorkerPool) Shutdown(timeout time.Duration) {
	p.cancel()

	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(timeout):
		p.logger.Warn("worker_pool_shutdown_timeout", nil)
	}
}

// WaitPending waits for all currently pending jobs to complete.
// This is useful for tests that need to ensure all submitted work is finished.
func (p *WorkerPool) WaitPending() {
	done := make(chan struct{})
	p.Submit(func(ctx context.Context) {
		close(done)
	})
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		p.logger.Warn("worker_pool_wait_pending_timeout", nil)
	}
}
