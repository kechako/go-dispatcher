// Package dispatcher dispatches tasks.
package dispatcher

import (
	"context"
	"runtime"
	"sync"
)

var (
	// DefaultWorkerCount is a default count of workers.
	DefaultWorkerCount = runtime.NumCPU()

	// DefaultQueueSize is a default size of task queue.
	DefaultQueueSize = 1024
)

// A Dispatcher dispatches tasks.
type Dispatcher struct {
	count int // count of workers
	size  int // size of task queue

	cancel    func()
	wgWorkers sync.WaitGroup

	queue chan Tasker
	wg    sync.WaitGroup

	errOnce sync.Once
	err     error
}

// Start creates a new Dispatcher and starts to dispatch tasks to workers.
func Start(ctx context.Context, opts ...Option) *Dispatcher {
	ctx, cancel := context.WithCancel(ctx)

	d := &Dispatcher{
		count:  DefaultWorkerCount,
		size:   DefaultQueueSize,
		cancel: cancel,
	}

	for _, opt := range opts {
		opt.apply(d)
	}

	d.queue = make(chan Tasker, d.size)

	for i := 0; i < d.count; i++ {
		d.wgWorkers.Add(1)
		go d.worker(ctx)
	}

	return d
}

func (d *Dispatcher) worker(ctx context.Context) {
	defer d.wgWorkers.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case t := <-d.queue:
			if err := t.Run(ctx); err != nil {
				d.errOnce.Do(func() {
					d.err = err
					d.cancel()
				})
			}
			d.wg.Done()
		}
	}
}

func (d *Dispatcher) Wait() error {
	d.wg.Wait()

	d.cancel()
	d.wgWorkers.Wait()

	return d.err
}

// Add adds a new task to the dispatcher.
func (d *Dispatcher) Add(t Tasker) {
	d.wg.Add(1)
	d.queue <- t
}

// Tasker is the interface that wraps the Run method.
type Tasker interface {
	Run(ctx context.Context) error
}

// The TaskerFunc type is an adapter to allow the use of ordinary functions as task runners.
// If f is a function with the appropriate signature, TaskerFunc(f) is a Tasker that calls f.
type TaskerFunc func(ctx context.Context) error

// Run calls f(ctx).
func (f TaskerFunc) Run(ctx context.Context) error {
	return f(ctx)
}
