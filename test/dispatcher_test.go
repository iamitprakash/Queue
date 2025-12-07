package test

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/amitprakash/queue/dispatcher"
	"github.com/amitprakash/queue/queue"
)

func TestNewDispatcher(t *testing.T) {
	d := dispatcher.NewDispatcher(1024, 8)
	if d == nil {
		t.Fatal("NewDispatcher returned nil")
	}
}

func TestDispatcher_Submit(t *testing.T) {
	d := dispatcher.NewDispatcher(4, 2)

	// Should be able to submit jobs
	if !d.Submit(queue.Job{ID: 1}) {
		t.Error("Submit should succeed")
	}
	if !d.Submit(queue.Job{ID: 2}) {
		t.Error("Submit should succeed")
	}
	if !d.Submit(queue.Job{ID: 3}) {
		t.Error("Submit should succeed")
	}

	// Queue should be full now (size 4, 3 usable slots)
	if d.Submit(queue.Job{ID: 4}) {
		t.Error("Submit should fail when queue is full")
	}
}

func TestDispatcher_StartAndProcess(t *testing.T) {
	d := dispatcher.NewDispatcher(1024, 4)

	var processed atomic.Int64
	var wg sync.WaitGroup

	numJobs := 100
	wg.Add(numJobs)

	d.Start(func(job queue.Job) {
		processed.Add(1)
		wg.Done()
	})

	for i := 0; i < numJobs; i++ {
		for !d.Submit(queue.Job{ID: int64(i)}) {
			// Retry if queue is full
			time.Sleep(time.Microsecond)
		}
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
		t.Fatalf("timeout: processed %d/%d jobs", processed.Load(), numJobs)
	}

	if processed.Load() != int64(numJobs) {
		t.Errorf("expected %d processed, got %d", numJobs, processed.Load())
	}
}

func TestDispatcher_JobData(t *testing.T) {
	d := dispatcher.NewDispatcher(64, 2)

	var receivedIDs []int64
	var mu sync.Mutex
	var wg sync.WaitGroup

	numJobs := 10
	wg.Add(numJobs)

	d.Start(func(job queue.Job) {
		mu.Lock()
		receivedIDs = append(receivedIDs, job.ID)
		mu.Unlock()
		wg.Done()
	})

	for i := 0; i < numJobs; i++ {
		d.Submit(queue.Job{ID: int64(i * 10)})
	}

	wg.Wait()

	// Verify all jobs were processed (order may vary)
	if len(receivedIDs) != numJobs {
		t.Errorf("expected %d jobs, got %d", numJobs, len(receivedIDs))
	}

	// Check all expected IDs are present
	expected := make(map[int64]bool)
	for i := 0; i < numJobs; i++ {
		expected[int64(i*10)] = true
	}
	for _, id := range receivedIDs {
		if !expected[id] {
			t.Errorf("unexpected job ID: %d", id)
		}
		delete(expected, id)
	}
	if len(expected) > 0 {
		t.Errorf("missing job IDs: %v", expected)
	}
}

func TestDispatcher_HighVolume(t *testing.T) {
	d := dispatcher.NewDispatcher(4096, 8)

	var processed atomic.Int64
	numJobs := 10000

	var wg sync.WaitGroup
	wg.Add(numJobs)

	d.Start(func(job queue.Job) {
		processed.Add(1)
		wg.Done()
	})

	submitted := 0
	for submitted < numJobs {
		if d.Submit(queue.Job{ID: int64(submitted)}) {
			submitted++
		} else {
			time.Sleep(time.Microsecond)
		}
	}

	// Wait with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(10 * time.Second):
		t.Fatalf("timeout: processed %d/%d jobs", processed.Load(), numJobs)
	}

	if processed.Load() != int64(numJobs) {
		t.Errorf("expected %d processed, got %d", numJobs, processed.Load())
	}
}

func BenchmarkDispatcher_SubmitProcess(b *testing.B) {
	d := dispatcher.NewDispatcher(8192, 8)

	var wg sync.WaitGroup
	wg.Add(b.N)

	d.Start(func(job queue.Job) {
		wg.Done()
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for !d.Submit(queue.Job{ID: int64(i)}) {
			// Spin until submitted
		}
	}

	wg.Wait()
}

