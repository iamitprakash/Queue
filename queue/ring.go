package queue

import "sync"

type Job struct {
	ID   int64
	Data []byte
}

type Ring struct {
	mu   sync.Mutex
	buf  []Job
	size int
	head int
	tail int
}

func NewRing(size int) *Ring {
	return &Ring{
		buf:  make([]Job, size),
		size: size,
	}
}

func (r *Ring) Push(j Job) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	next := (r.tail + 1) % r.size
	if next == r.head {
		return false // queue full
	}
	r.buf[r.tail] = j
	r.tail = next
	return true
}

func (r *Ring) Pop(j *Job) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.head == r.tail {
		return false // empty
	}
	*j = r.buf[r.head] // zero-alloc copy
	r.head = (r.head + 1) % r.size
	return true
}

// Len returns the current number of items in the ring buffer
func (r *Ring) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.tail >= r.head {
		return r.tail - r.head
	}
	return r.size - r.head + r.tail
}
