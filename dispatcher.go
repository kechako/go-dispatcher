package dispatcher

import (
	"runtime"
	"sync"
)

var (
	MaxWorkers = runtime.NumCPU()
	MaxQueues  = 10000
)

type Dispatcher struct {
	pool    chan *worker
	queue   chan Tasker
	workers []*worker
	wg      sync.WaitGroup
	quit    chan struct{}
}

func New() *Dispatcher {
	d := &Dispatcher{
		pool:  make(chan *worker, MaxWorkers),
		queue: make(chan Tasker, MaxQueues),
		quit:  make(chan struct{}),
	}

	d.workers = make([]*worker, cap(d.pool))
	for i := 0; i < len(d.workers); i++ {
		d.workers[i] = &worker{
			dispatcher: d,
			task:       make(chan Tasker),
			quit:       make(chan struct{}),
		}
	}

	return d
}

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

func (d *Dispatcher) Enqueue(t Tasker) {
	d.wg.Add(1)
	d.queue <- t
}

func (d *Dispatcher) Wait() {
	d.wg.Wait()
}

type worker struct {
	dispatcher *Dispatcher
	task       chan Tasker
	quit       chan struct{}
}

func (w *worker) start() {
	go func() {
		for {
			w.dispatcher.pool <- w

			select {
			case t := <-w.task:
				t.Run()

				w.dispatcher.wg.Done()
			}
		}
	}()
}

type Tasker interface {
	Run()
}
