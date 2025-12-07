package test

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/amitprakash/queue/queue"
	"github.com/amitprakash/queue/worker"
)

func TestNewPool(t *testing.T) {
	p := worker.NewPool(4)
	if p == nil {
		t.Fatal("NewPool returned nil")
	}
}

func TestWorkerPool_ProcessJobs(t *testing.T) {
	p := worker.NewPool(4)

	var processed atomic.Int64
	var wg sync.WaitGroup

	numJobs := 100
	wg.Add(numJobs)

	p.Start(func(job queue.Job) {
		processed.Add(1)
		wg.Done()
	})

	for i := 0; i < numJobs; i++ {
		p.Submit(queue.Job{ID: int64(i)})
	}

	// Wait for all jobs with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for jobs to complete")
	}

	if processed.Load() != int64(numJobs) {
		t.Errorf("expected %d processed, got %d", numJobs, processed.Load())
	}
}

func TestWorkerPool_JobOrder(t *testing.T) {
	p := worker.NewPool(1) // Single worker to preserve order

	var results []int64
	var mu sync.Mutex
	var wg sync.WaitGroup

	numJobs := 10
	wg.Add(numJobs)

	p.Start(func(job queue.Job) {
		mu.Lock()
		results = append(results, job.ID)
		mu.Unlock()
		wg.Done()
	})

	for i := 0; i < numJobs; i++ {
		p.Submit(queue.Job{ID: int64(i)})
	}

	wg.Wait()

	// With single worker, order should be preserved
	for i, id := range results {
		if id != int64(i) {
			t.Errorf("expected job %d at position %d, got %d", i, i, id)
		}
	}
}

func TestWorkerPool_ConcurrentProcessing(t *testing.T) {
	numWorkers := 4
	p := worker.NewPool(numWorkers)

	var concurrent atomic.Int32
	var maxConcurrent atomic.Int32
	var wg sync.WaitGroup

	numJobs := 20
	wg.Add(numJobs)

	p.Start(func(job queue.Job) {
		cur := concurrent.Add(1)
		// Track max concurrency
		for {
			old := maxConcurrent.Load()
			if cur <= old || maxConcurrent.CompareAndSwap(old, cur) {
				break
			}
		}

		time.Sleep(10 * time.Millisecond) // Simulate work
		concurrent.Add(-1)
		wg.Done()
	})

	for i := 0; i < numJobs; i++ {
		p.Submit(queue.Job{ID: int64(i)})
	}

	wg.Wait()

	// Should have achieved some concurrency
	if maxConcurrent.Load() < 2 {
		t.Errorf("expected concurrent processing, max concurrent was %d", maxConcurrent.Load())
	}
}

func BenchmarkWorkerPool_Submit(b *testing.B) {
	p := worker.NewPool(8)

	var wg sync.WaitGroup
	wg.Add(b.N)

	p.Start(func(job queue.Job) {
		wg.Done()
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.Submit(queue.Job{ID: int64(i)})
	}

	wg.Wait()
}
