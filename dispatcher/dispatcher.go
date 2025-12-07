package dispatcher

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/amitprakash/queue/queue"
	"github.com/amitprakash/queue/worker"
)

type Dispatcher struct {
	ring       *queue.Ring
	pool       *worker.WorkerPool
	stop       chan struct{}
	stopped    atomic.Bool
	wg         sync.WaitGroup
	inFlight   atomic.Int64
	onComplete chan struct{}
}

func NewDispatcher(queueSize, workers int) *Dispatcher {
	return &Dispatcher{
		ring:       queue.NewRing(queueSize),
		pool:       worker.NewPool(workers),
		stop:       make(chan struct{}),
		onComplete: make(chan struct{}),
	}
}

func (d *Dispatcher) Start(handler func(queue.Job)) {
	d.pool.Start(func(job queue.Job) {
		handler(job)
		// Signal job completion
		if d.inFlight.Add(-1) == 0 {
			select {
			case d.onComplete <- struct{}{}:
			default:
			}
		}
	})

	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		var job queue.Job
		for {
			select {
			case <-d.stop:
				return
			default:
				if d.ring.Pop(&job) {
					d.pool.Submit(job)
				} else {
					// Backoff when queue is empty to avoid CPU spin
					time.Sleep(100 * time.Microsecond)
				}
			}
		}
	}()
}

func (d *Dispatcher) Submit(j queue.Job) bool {
	if d.stopped.Load() {
		return false
	}
	if d.ring.Push(j) {
		d.inFlight.Add(1)
		return true
	}
	return false
}

// Wait blocks until all submitted jobs have been processed
func (d *Dispatcher) Wait() {
	for d.inFlight.Load() > 0 {
		select {
		case <-d.onComplete:
		case <-time.After(10 * time.Millisecond):
		}
	}
}

// Stop signals the dispatcher to stop processing and waits for cleanup
func (d *Dispatcher) Stop() {
	d.stopped.Store(true)
	close(d.stop)
	d.wg.Wait()
}
