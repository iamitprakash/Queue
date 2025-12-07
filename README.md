# Queue

A high-performance, thread-safe job queue implementation in Go with a lock-free ring buffer, worker pool, and dispatcher.

## Features

- **Ring Buffer**: Lock-free circular buffer for fast, zero-allocation job queuing
- **Worker Pool**: Configurable number of concurrent workers
- **Dispatcher**: Coordinates job submission and worker distribution
- **Thread-Safe**: Mutex-protected operations with race-free design
- **Graceful Shutdown**: Wait for completion and clean shutdown support

## Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Submit    │────▶│ Ring Buffer │────▶│ Worker Pool │
│   (main)    │     │   (queue)   │     │  (workers)  │
└─────────────┘     └─────────────┘     └─────────────┘
                           │
                    ┌──────┴──────┐
                    │  Dispatcher │
                    │ (coordinator)│
                    └─────────────┘
```

## Installation

```bash
go get github.com/amitprakash/queue
```

## Usage

```go
package main

import (
    "fmt"
    "time"

    "github.com/amitprakash/queue/dispatcher"
    "github.com/amitprakash/queue/queue"
)

func main() {
    // Create dispatcher with queue size 1024 and 8 workers
    d := dispatcher.NewDispatcher(1024, 8)

    // Start processing with a handler function
    d.Start(func(job queue.Job) {
        fmt.Printf("Processing job %d\n", job.ID)
    })

    // Submit jobs (retry if queue is full)
    for i := 0; i < 100; i++ {
        for !d.Submit(queue.Job{ID: int64(i)}) {
            time.Sleep(time.Microsecond)
        }
    }

    // Wait for all jobs to complete
    d.Wait()
    
    // Graceful shutdown
    d.Stop()
}
```

## API Reference

### Queue Package

```go
// Job represents a unit of work
type Job struct {
    ID   int64
    Data []byte
}

// Ring is a thread-safe circular buffer
func NewRing(size int) *Ring
func (r *Ring) Push(j Job) bool  // returns false if full
func (r *Ring) Pop(j *Job) bool  // returns false if empty
func (r *Ring) Len() int
```

### Worker Package

```go
// WorkerPool manages concurrent job processing
func NewPool(n int) *WorkerPool
func (p *WorkerPool) Start(handler func(Job))
func (p *WorkerPool) Submit(j Job)
```

### Dispatcher Package

```go
// Dispatcher coordinates queue and workers
func NewDispatcher(queueSize, workers int) *Dispatcher
func (d *Dispatcher) Start(handler func(Job))
func (d *Dispatcher) Submit(j Job) bool
func (d *Dispatcher) Wait()   // blocks until all jobs processed
func (d *Dispatcher) Stop()   // graceful shutdown
```

## Project Structure

```
.
├── dispatcher/
│   └── dispatcher.go    # Job dispatcher/coordinator
├── queue/
│   └── ring.go          # Thread-safe ring buffer
├── worker/
│   └── pool.go          # Worker pool implementation
├── test/
│   ├── dispatcher_test.go
│   ├── pool_test.go
│   └── ring_test.go
├── main.go              # Example usage
├── go.mod
└── README.md
```

## Testing

Run all tests:

```bash
go test ./...
```

Run with race detection:

```bash
go test ./... -race
```

Run with verbose output:

```bash
go test ./test/... -v
```

Run benchmarks:

```bash
go test ./test/... -bench=. -benchmem
```

## Benchmarks

```
BenchmarkRing_PushPop           ~4 ns/op     0 B/op    0 allocs/op
BenchmarkWorkerPool_Submit    ~135 ns/op     0 B/op    0 allocs/op
BenchmarkDispatcher_Submit    ~300 ns/op     0 B/op    0 allocs/op
```

## License

MIT

