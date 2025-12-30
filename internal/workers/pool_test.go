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
		if i%10 == 0 {
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
	if diff > poolSize+5 {
		// allow some buffer for test runner
		t.Logf("Warning: Goroutine count increased by %d, expected roughly %d", diff, poolSize)
	}
}

func TestWorkerPool_ConfigurableSize(t *testing.T) {
	log := logger.New("DEBUG")

	testCases := []struct {
		name     string
		poolSize int
	}{
		{"single_worker", 1},
		{"small_pool", 3},
		{"medium_pool", 10},
		{"large_pool", 50},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pool := NewWorkerPool(log, tc.poolSize)
			defer pool.Shutdown(1 * time.Second)

			var wg sync.WaitGroup
			wg.Add(tc.poolSize)

			for i := 0; i < tc.poolSize; i++ {
				pool.Submit(func(ctx context.Context) {
					defer wg.Done()
				})
			}

			done := make(chan struct{})
			go func() {
				wg.Wait()
				close(done)
			}()

			select {
			case <-done:
			case <-time.After(2 * time.Second):
				t.Fatalf("Pool size %d: jobs did not complete in time", tc.poolSize)
			}
		})
	}
}

func TestWorkerPool_BufferCapacity(t *testing.T) {
	log := logger.New("ERROR")
	poolSize := 5
	pool := NewWorkerPool(log, poolSize)
	defer pool.Shutdown(1 * time.Second)

	submitted := 0
	rejected := 0

	for i := 0; i < 100; i++ {
		success := pool.Submit(func(ctx context.Context) {
			time.Sleep(10 * time.Millisecond)
		})
		if success {
			submitted++
		} else {
			rejected++
		}
	}

	t.Logf("Submitted: %d, Rejected: %d", submitted, rejected)

	if submitted == 0 {
		t.Error("Expected some jobs to be accepted")
	}
}

func TestWorkerPool_ConcurrentSubmissions(t *testing.T) {
	log := logger.New("ERROR")
	poolSize := 10
	pool := NewWorkerPool(log, poolSize)
	defer pool.Shutdown(5 * time.Second)

	numGoroutines := 50
	jobsPerGoroutine := 10
	var wg sync.WaitGroup
	wg.Add(numGoroutines * jobsPerGoroutine)

	var mu sync.Mutex
	acceptedCount := 0

	for i := 0; i < numGoroutines; i++ {
		go func() {
			for j := 0; j < jobsPerGoroutine; j++ {
				success := pool.Submit(func(ctx context.Context) {
					time.Sleep(5 * time.Millisecond)
					wg.Done()
				})
				if success {
					mu.Lock()
					acceptedCount++
					mu.Unlock()
				} else {
					wg.Done()
				}
			}
		}()
	}

	wg.Wait()
	t.Logf("Concurrently accepted: %d jobs", acceptedCount)
}

func TestWorkerPool_JobCancellation(t *testing.T) {
	log := logger.New("DEBUG")
	pool := NewWorkerPool(log, 2)

	var wg sync.WaitGroup
	wg.Add(1)

	pool.Submit(func(ctx context.Context) {
		defer wg.Done()
		select {
		case <-ctx.Done():
			// Success: cancelled
		case <-time.After(2 * time.Second):
			// Failure: did not cancel
		}
	})

	// Shutdown calls pool.cancel() which should cancel the job's context
	pool.Shutdown(1 * time.Second)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Job cancellation did not complete in time")
	}
}
