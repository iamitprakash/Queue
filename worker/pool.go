package worker

import (
	"github.com/amitprakash/queue/queue"
)

type WorkerPool struct {
	workers int
	jobs    chan queue.Job
}

func NewPool(n int) *WorkerPool {
	return &WorkerPool{
		workers: n,
		jobs:    make(chan queue.Job, n),
	}
}

func (p *WorkerPool) Start(handle func(queue.Job)) {
	for i := 0; i < p.workers; i++ {
		go func() {
			for job := range p.jobs {
				handle(job)
			}
		}()
	}
}

func (p *WorkerPool) Submit(j queue.Job) {
	p.jobs <- j
}

