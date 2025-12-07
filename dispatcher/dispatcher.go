package dispatcher

import (
	"github.com/amitprakash/queue/queue"
	"github.com/amitprakash/queue/worker"
)

type Dispatcher struct {
	ring *queue.Ring
	pool *worker.WorkerPool
}

func NewDispatcher(queueSize, workers int) *Dispatcher {
	return &Dispatcher{
		ring: queue.NewRing(queueSize),
		pool: worker.NewPool(workers),
	}
}

func (d *Dispatcher) Start(handler func(queue.Job)) {
	d.pool.Start(handler)

	go func() {
		var job queue.Job
		for {
			if d.ring.Pop(&job) {
				d.pool.Submit(job)
			}
		}
	}()
}

func (d *Dispatcher) Submit(j queue.Job) bool {
	return d.ring.Push(j)
}

