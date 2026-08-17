package fetch

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestPoolBoundsConcurrency(t *testing.T) {
	const cap = 3
	pool := NewPool(context.Background(), cap)

	var (
		mu       sync.Mutex
		inFlight int
		maxSeen  int
	)
	// Tasks self-complete via a short sleep, so scheduling synchronously is safe:
	// Go blocks on the full semaphore but running tasks free slots on their own.
	for i := 0; i < 12; i++ {
		pool.Go(func(_ context.Context) {
			mu.Lock()
			inFlight++
			if inFlight > maxSeen {
				maxSeen = inFlight
			}
			mu.Unlock()

			time.Sleep(2 * time.Millisecond)

			mu.Lock()
			inFlight--
			mu.Unlock()
		})
	}
	pool.Wait()

	mu.Lock()
	defer mu.Unlock()
	if maxSeen > cap {
		t.Fatalf("max in-flight = %d, want <= %d", maxSeen, cap)
	}
	if maxSeen < 1 {
		t.Fatal("no tasks ran")
	}
}

func TestPoolRunsAllTasks(t *testing.T) {
	pool := NewPool(context.Background(), 2)
	var n int32
	for i := 0; i < 25; i++ {
		pool.Go(func(_ context.Context) { atomic.AddInt32(&n, 1) })
	}
	pool.Wait()
	if n != 25 {
		t.Fatalf("ran %d tasks, want 25", n)
	}
}

func TestPoolStopsSchedulingOnCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already canceled
	pool := NewPool(ctx, 2)
	var n int32
	for i := 0; i < 5; i++ {
		pool.Go(func(_ context.Context) { atomic.AddInt32(&n, 1) })
	}
	pool.Wait()
	if n != 0 {
		t.Fatalf("scheduled %d tasks after cancel, want 0", n)
	}
}

func TestPoolMinLimit(t *testing.T) {
	pool := NewPool(context.Background(), 0) // treated as 1
	var n int32
	for i := 0; i < 3; i++ {
		pool.Go(func(_ context.Context) { atomic.AddInt32(&n, 1) })
	}
	pool.Wait()
	if n != 3 {
		t.Fatalf("ran %d, want 3", n)
	}
}
