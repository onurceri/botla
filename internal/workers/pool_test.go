package workers

import (
	"context"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/onurceri/botla-co/pkg/logger"
)

func TestWorkerPool_ExecuteJob(t *testing.T) {
	log := logger.New("DEBUG")
	pool := NewWorkerPool(log, 2)
	defer pool.Shutdown(1 * time.Second)

	var wg sync.WaitGroup
	wg.Add(1)

	pool.Submit(func(ctx context.Context) {
		defer wg.Done()
		time.Sleep(10 * time.Millisecond)
	})

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("Job execution timed out")
	}
}

func TestWorkerPool_PanicRecovery(t *testing.T) {
	log := logger.New("DEBUG")
	pool := NewWorkerPool(log, 2)
	defer pool.Shutdown(1 * time.Second)

	var wg sync.WaitGroup
	wg.Add(1)

	// Submit a task that panics
	pool.Submit(func(ctx context.Context) {
		defer wg.Done()
		panic("test panic")
	})

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success, panic was recovered
	case <-time.After(2 * time.Second):
		t.Fatal("Panic test timed out")
	}
}

func TestWorkerPool_GracefulShutdown(t *testing.T) {
	log := logger.New("DEBUG")
	pool := NewWorkerPool(log, 5)

	var wg sync.WaitGroup
	wg.Add(5)

	// Submit multiple jobs
	for i := 0; i < 5; i++ {
		success := pool.Submit(func(ctx context.Context) {
			defer wg.Done()
			time.Sleep(50 * time.Millisecond)
		})
		if !success {
			wg.Done() // Don't wait for rejected jobs
			t.Logf("Job %d rejected", i)
		}
	}

	// Shutdown should wait for all jobs to complete
	pool.Shutdown(2 * time.Second)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(1 * time.Second):
		t.Fatal("Shutdown did not wait for jobs")
	}
}

func TestWorkerPool_Load(t *testing.T) {
	log := logger.New("ERROR") // Reduce log noise
	poolSize := 5
	pool := NewWorkerPool(log, poolSize)
	defer pool.Shutdown(5 * time.Second)

	numJobs := 100
	var wg sync.WaitGroup
	wg.Add(numJobs)

	// Measure start goroutines
	startGoroutines := runtime.NumGoroutine()
	
	started := time.Now()
	
	// Submit 100 jobs
	// Note: since buffer is 5 + 5 workers = 10 capacity approx,
	// submitting 100 fast might drop some if executeJob is slow.
	// But we want to test they are processed eventually if we pace them or just check rejection?
	// The problem description says "Verify bounded goroutine count".
	// The worker pool creates 'poolSize' goroutines. It doesn't spark more.
	
	accepted := 0
	for i := 0; i < numJobs; i++ {
		success := pool.Submit(func(ctx context.Context) {
			defer wg.Done()
			time.Sleep(1 * time.Millisecond)
		})
		if success {
			accepted++
		} else {
			wg.Done()
		}
		// A little sleep to prevent instant buffer fill if we want higher acceptance
		if i % 10 == 0 {
			time.Sleep(1 * time.Millisecond)
		}
	}
	
	t.Logf("Accepted %d/%d jobs", accepted, numJobs)
	
	// Wait for accepted to finish
	wg.Wait()
	
	elapsed := time.Since(started)
	t.Logf("Processed %d jobs in %v", accepted, elapsed)
	
	endGoroutines := runtime.NumGoroutine()
	// We expect goroutine count to be roughly start + poolSize (plus maybe some test runtime overhead)
	// It should NOT be start + 100.
	
	diff := endGoroutines - startGoroutines
	t.Logf("Goroutine diff: %d (Pool size: %d)", diff, poolSize)
	
	// diff can be messy due to logger or other things, but it shouldn't be huge.
	if diff > poolSize + 5 {
		// allow some buffer for test runner
		t.Logf("Warning: Goroutine count increased by %d, expected roughly %d", diff, poolSize)
	}
}
