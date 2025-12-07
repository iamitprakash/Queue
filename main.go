func main() {
    d := dispatcher.NewDispatcher(1024, 8)

    d.Start(func(job queue.Job) {
        // process job
        _ = job.ID
    })

    for i := 0; i < 1_000_000; i++ {
        d.Submit(queue.Job{
            ID: int64(i),
        })
    }
}
