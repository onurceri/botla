// Package testutils provides testing utilities for deterministic, parallel test execution.
//
// This package helps eliminate two major blockers for parallel Go tests:
// 1. t.Setenv() - Process-global environment variable changes that prevent parallel execution
// 2. time.Sleep() - Fixed delays that cause flakiness and slow tests.
package testutils

import (
	"context"
	"sync"
	"testing"
	"time"
)

// Completion provides a channel-based signal for deterministic async operation testing.
// Eliminates time.Sleep() calls that cause flaky tests.
type Completion struct {
	done chan struct{}
	t    *testing.T
}

// NewCompletion creates a new Completion tracker.
func NewCompletion(t *testing.T) *Completion {
	return &Completion{
		done: make(chan struct{}),
		t:    t,
	}
}

// Signal marks the operation as complete.
// Safe to call multiple times and from any goroutine.
func (c *Completion) Signal() {
	select {
	case <-c.done:
	default:
		close(c.done)
	}
}

// Wait blocks until Signal() is called or timeout expires.
func (c *Completion) Wait() {
	select {
	case <-c.done:
		return
	case <-time.After(5 * time.Second):
		c.t.Fatal("completion signal timeout after 5s")
	}
}

// WaitWithTimeout waits with a custom timeout duration.
func (c *Completion) WaitWithTimeout(timeout time.Duration) {
	select {
	case <-c.done:
		return
	case <-time.After(timeout):
		c.t.Fatalf("completion signal timeout after %v", timeout)
	}
}

// Await returns the done channel for use in select statements.
func (c *Completion) Await() <-chan struct{} {
	return c.done
}

// WaitContext waits for completion or context cancellation.
func (c *Completion) WaitContext(ctx context.Context) error {
	select {
	case <-c.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// ErrGroup provides a sync.WaitGroup-like interface with error collection.
type ErrGroup struct {
	wg   sync.WaitGroup
	errs []error
	mu   sync.Mutex
}

// NewErrGroup creates a new ErrGroup.
func NewErrGroup() *ErrGroup {
	return &ErrGroup{}
}

// Go launches a goroutine with automatic completion tracking and error capture.
func (eg *ErrGroup) Go(fn func() error) {
	eg.wg.Add(1)
	go func() {
		defer eg.wg.Done()
		if err := fn(); err != nil {
			eg.mu.Lock()
			eg.errs = append(eg.errs, err)
			eg.mu.Unlock()
		}
	}()
}

// Wait blocks until all goroutines complete and returns collected errors.
func (eg *ErrGroup) Wait() []error {
	eg.wg.Wait()
	return eg.errs
}

// WaitWithTimeout blocks with a timeout and returns errors.
func (eg *ErrGroup) WaitWithTimeout(timeout time.Duration) ([]error, bool) {
	done := make(chan struct{})
	go func() {
		eg.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return eg.errs, true
	case <-time.After(timeout):
		return eg.errs, false
	}
}

// Poller provides a flexible polling helper with configurable intervals.
type Poller struct {
	interval   time.Duration
	maxRetries int
}

// NewPoller creates a Poller with the specified interval and max retries.
func NewPoller(interval time.Duration, maxRetries int) *Poller {
	return &Poller{
		interval:   interval,
		maxRetries: maxRetries,
	}
}

// Poll executes fn until it returns true or max retries reached.
func (p *Poller) Poll(fn func() (bool, error)) error {
	for i := 0; i <= p.maxRetries; i++ {
		done, err := fn()
		if err != nil {
			return err
		}
		if done {
			return nil
		}
		if i < p.maxRetries {
			time.Sleep(p.interval)
		}
	}
	return nil
}
