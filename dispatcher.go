package dispatcher

import (
	"context"
	"runtime"
	"sync"
)

var (
	DefaultWorkerCount = runtime.NumCPU()
	DefaultQueueSize   = 10000
)

type task struct {
	ctx    context.Context
	tasker Tasker
	ch     chan error
}

// A Dispatcher dispatches tasks to workers.
type Dispatcher struct {
	pool    chan *worker
	queue   chan *task
	workers []*worker
	wg      sync.WaitGroup
	quit    chan struct{}
}

// New creates and initializes a new Dispatcher.
// The Dispatcher has count Workers and size queue.
func New(count, size int) *Dispatcher {
	return newDispatcher(count, size)
}

// NewDefault creates and initializes a new Dispatcher
// with DefaultWorkerCount and DefaultQueueSize.
func NewDefault() *Dispatcher {
	return newDispatcher(DefaultWorkerCount, DefaultQueueSize)
}

func newDispatcher(count, size int) *Dispatcher {
	d := &Dispatcher{
		pool:  make(chan *worker, count),
		queue: make(chan *task, size),
		quit:  make(chan struct{}),
	}

	d.workers = make([]*worker, cap(d.pool))
	for i := 0; i < len(d.workers); i++ {
		d.workers[i] = &worker{
			dispatcher: d,
			task:       make(chan *task),
			quit:       make(chan struct{}),
		}
	}

	return d
}

// Start starts to dispatch tasks.
func (d *Dispatcher) Start() {
	for _, w := range d.workers {
		w.start()
	}

	go func() {
		for {
			select {
			case t := <-d.queue:
				(<-d.pool).task <- t
			case <-d.quit:
				return
			}
		}
	}()
}

// Enqueue enqueues a new task to the dispatcher.
func (d *Dispatcher) Enqueue(ctx context.Context, t Tasker) <-chan error {
	d.wg.Add(1)

	ch := make(chan error)

	d.queue <- &task{
		ctx:    ctx,
		tasker: t,
		ch:     ch,
	}

	return ch
}

// Wait waits until all tasks are completed.
func (d *Dispatcher) Wait() {
	d.wg.Wait()
}

type worker struct {
	dispatcher *Dispatcher
	task       chan *task
	quit       chan struct{}
}

func (w *worker) start() {
	go func() {
		for {
			w.dispatcher.pool <- w

			select {
			case t := <-w.task:
				ch := make(chan error)

				go func() {
					ch <- t.tasker.Run(t.ctx)
				}()

				select {
				case err := <-ch:
					t.ch <- err
				case <-t.ctx.Done():
					t.ch <- t.ctx.Err()
				}

				w.dispatcher.wg.Done()
			}
		}
	}()
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
