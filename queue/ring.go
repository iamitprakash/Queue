package queue

type Job struct {
	ID   int64
	Data []byte
}

type Ring struct {
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
	next := (r.tail + 1) % r.size
	if next == r.head {
		return false // queue full
	}
	r.buf[r.tail] = j
	r.tail = next
	return true
}

func (r *Ring) Pop(j *Job) bool {
	if r.head == r.tail {
		return false // empty
	}
	*j = r.buf[r.head] // zero-alloc copy
	r.head = (r.head + 1) % r.size
	return true
}

