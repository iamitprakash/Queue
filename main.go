package main

import (
	"fmt"
	"time"

	"github.com/amitprakash/queue/dispatcher"
	"github.com/amitprakash/queue/queue"
)

func main() {
	d := dispatcher.NewDispatcher(1024, 8)

	d.Start(func(job queue.Job) {
		// process job
		_ = job.ID
	})

	submitted := 0
	for i := 0; i < 1_000_000; i++ {
		// Retry with backoff if queue is full (Bug 1 fix)
		for !d.Submit(queue.Job{ID: int64(i)}) {
			time.Sleep(time.Microsecond)
		}
		submitted++
	}

	// Wait for all jobs to complete before exiting (Bug 2 fix)
	d.Wait()
	d.Stop()

	fmt.Printf("Completed processing %d jobs\n", submitted)
}
