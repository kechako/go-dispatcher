package main

import (
	"context"
	"runtime"
	"time"

	dispatcher "github.com/kechako/go-dispatcher"
)

type task struct {
}

func (t *task) Run(ctx context.Context) error {
	// Do something
	time.Sleep(1 * time.Second)

	return nil
}

func main() {
	d := dispatcher.New(runtime.NumCPU(), 10000)
	d.Start()

	ctx := context.Background()
	for i := 0; i < 100; i++ {
		d.Enqueue(ctx, new(task))
	}

	d.Wait()
}
