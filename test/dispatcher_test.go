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
	numJobs := 100

	d.Start(func(job queue.Job) {
		processed.Add(1)
	})

	for i := 0; i < numJobs; i++ {
		for !d.Submit(queue.Job{ID: int64(i)}) {
			time.Sleep(time.Microsecond)
		}
	}

	// Use the new Wait() method
	d.Wait()
	d.Stop()

	if processed.Load() != int64(numJobs) {
		t.Errorf("expected %d processed, got %d", numJobs, processed.Load())
	}
}

func TestDispatcher_JobData(t *testing.T) {
	d := dispatcher.NewDispatcher(64, 2)

	var receivedIDs []int64
	var mu sync.Mutex
	var processed atomic.Int64

	numJobs := 10

	d.Start(func(job queue.Job) {
		mu.Lock()
		receivedIDs = append(receivedIDs, job.ID)
		mu.Unlock()
		processed.Add(1)
	})

	for i := 0; i < numJobs; i++ {
		for !d.Submit(queue.Job{ID: int64(i * 10)}) {
			time.Sleep(time.Microsecond)
		}
	}

	d.Wait()
	d.Stop()

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

	d.Start(func(job queue.Job) {
		processed.Add(1)
	})

	for i := 0; i < numJobs; i++ {
		for !d.Submit(queue.Job{ID: int64(i)}) {
			time.Sleep(time.Microsecond)
		}
	}

	d.Wait()
	d.Stop()

	if processed.Load() != int64(numJobs) {
		t.Errorf("expected %d processed, got %d", numJobs, processed.Load())
	}
}

func TestDispatcher_Wait(t *testing.T) {
	d := dispatcher.NewDispatcher(1024, 4)

	var processed atomic.Int64
	numJobs := 50

	d.Start(func(job queue.Job) {
		time.Sleep(time.Millisecond) // Simulate work
		processed.Add(1)
	})

	for i := 0; i < numJobs; i++ {
		for !d.Submit(queue.Job{ID: int64(i)}) {
			time.Sleep(time.Microsecond)
		}
	}

	// Wait should block until all jobs are processed
	d.Wait()

	if processed.Load() != int64(numJobs) {
		t.Errorf("Wait() returned before all jobs processed: %d/%d", processed.Load(), numJobs)
	}

	d.Stop()
}

func TestDispatcher_StopRejectsNewJobs(t *testing.T) {
	d := dispatcher.NewDispatcher(1024, 4)

	d.Start(func(job queue.Job) {})

	// Submit a job before stop
	if !d.Submit(queue.Job{ID: 1}) {
		t.Error("Submit should succeed before stop")
	}

	d.Wait()
	d.Stop()

	// Submit after stop should fail
	if d.Submit(queue.Job{ID: 2}) {
		t.Error("Submit should fail after stop")
	}
}

func BenchmarkDispatcher_SubmitProcess(b *testing.B) {
	d := dispatcher.NewDispatcher(8192, 8)

	d.Start(func(job queue.Job) {})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for !d.Submit(queue.Job{ID: int64(i)}) {
			// Spin until submitted
		}
	}

	d.Wait()
	d.Stop()
}
