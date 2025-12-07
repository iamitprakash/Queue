package test

import (
	"testing"

	"github.com/amitprakash/queue/queue"
)

func TestNewRing(t *testing.T) {
	r := queue.NewRing(10)
	if r == nil {
		t.Fatal("NewRing returned nil")
	}
}

func TestRing_PushPop(t *testing.T) {
	r := queue.NewRing(4)

	// Push items
	if !r.Push(queue.Job{ID: 1}) {
		t.Error("Push should succeed")
	}
	if !r.Push(queue.Job{ID: 2}) {
		t.Error("Push should succeed")
	}
	if !r.Push(queue.Job{ID: 3}) {
		t.Error("Push should succeed")
	}

	// Ring of size 4 can hold 3 items (one slot reserved)
	if r.Push(queue.Job{ID: 4}) {
		t.Error("Push should fail when queue is full")
	}

	// Pop items and verify order (FIFO)
	var job queue.Job
	if !r.Pop(&job) || job.ID != 1 {
		t.Errorf("expected job ID 1, got %d", job.ID)
	}
	if !r.Pop(&job) || job.ID != 2 {
		t.Errorf("expected job ID 2, got %d", job.ID)
	}
	if !r.Pop(&job) || job.ID != 3 {
		t.Errorf("expected job ID 3, got %d", job.ID)
	}

	// Queue should be empty now
	if r.Pop(&job) {
		t.Error("Pop should fail on empty queue")
	}
}

func TestRing_Empty(t *testing.T) {
	r := queue.NewRing(4)
	var job queue.Job

	if r.Pop(&job) {
		t.Error("Pop on empty ring should return false")
	}
}

func TestRing_Wraparound(t *testing.T) {
	r := queue.NewRing(4)

	// Fill and empty multiple times to test wraparound
	for round := 0; round < 5; round++ {
		for i := 0; i < 3; i++ {
			if !r.Push(queue.Job{ID: int64(round*10 + i)}) {
				t.Errorf("Push failed at round %d, item %d", round, i)
			}
		}

		var job queue.Job
		for i := 0; i < 3; i++ {
			if !r.Pop(&job) {
				t.Errorf("Pop failed at round %d, item %d", round, i)
			}
			expected := int64(round*10 + i)
			if job.ID != expected {
				t.Errorf("expected ID %d, got %d", expected, job.ID)
			}
		}
	}
}

func TestRing_SingleItem(t *testing.T) {
	r := queue.NewRing(2)

	if !r.Push(queue.Job{ID: 42}) {
		t.Error("Push should succeed")
	}

	// Queue full with size 2 (1 usable slot)
	if r.Push(queue.Job{ID: 43}) {
		t.Error("Push should fail - queue full")
	}

	var job queue.Job
	if !r.Pop(&job) || job.ID != 42 {
		t.Errorf("expected job ID 42, got %d", job.ID)
	}
}

func BenchmarkRing_Push(b *testing.B) {
	r := queue.NewRing(1024)
	job := queue.Job{ID: 1, Data: []byte("test")}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Push(job)
		var j queue.Job
		r.Pop(&j)
	}
}

func BenchmarkRing_PushPop(b *testing.B) {
	r := queue.NewRing(1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Push(queue.Job{ID: int64(i)})
		var j queue.Job
		r.Pop(&j)
	}
}
