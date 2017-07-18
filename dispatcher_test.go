package dispatcher

import (
	"context"
	"errors"
	"testing"
	"time"
)

var (
	workerCount = 2
	queueSize   = 10
)

func TestNew(t *testing.T) {

	d := New(workerCount, queueSize)

	if d == nil {
		t.Fatal("New returned nil")
	}

	if d.pool == nil {
		t.Fatal("Worker pool is not initialized.")
	}
	if d.queue == nil {
		t.Fatal("Tasker queue is not initialized.")
	}
	if d.quit == nil {
		t.Fatal("Quit channel is not initialized.")
	}
	if d.workers == nil {
		t.Fatal("Workers not initialized.")
	}

	poolCap := cap(d.pool)
	if poolCap != workerCount {
		t.Fatalf("want %v\ngot %v", workerCount, poolCap)
	}

	queueCap := cap(d.queue)
	if queueCap != queueSize {
		t.Fatalf("want %v\ngot %v", queueSize, queueCap)
	}

	workerCnt := len(d.workers)
	if workerCnt != workerCount {
		t.Fatalf("want %v\ngot %v", workerCount, workerCnt)
	}

	for i, w := range d.workers {
		if w == nil {
			t.Fatal("%d : Worker is nil.", i)
		}

		if w.dispatcher != d {
			t.Errorf("%d : Invalid dispatcher.", i)
		}

		if w.quit == nil {
			t.Errorf("%d : Quit channel is not initialized.", i)
		}

		if w.task == nil {
			t.Errorf("%d : Tasker channel is not initialized.", i)
		}
	}
}

func TestNewDefault(t *testing.T) {
	d := NewDefault()

	if d == nil {
		t.Fatal("New returned nil")
	}
	if d.pool == nil {
		t.Fatal("Worker pool is not initialized.")
	}
	if d.queue == nil {
		t.Fatal("Tasker queue is not initialized.")
	}
	if d.quit == nil {
		t.Fatal("Quit channel is not initialized.")
	}
	if d.workers == nil {
		t.Fatal("Workers not initialized.")
	}

	poolCap := cap(d.pool)
	if poolCap != DefaultWorkerCount {
		t.Fatalf("want %v\ngot %v", DefaultWorkerCount, poolCap)
	}

	queueCap := cap(d.queue)
	if queueCap != DefaultQueueSize {
		t.Fatalf("want %v\ngot %v", DefaultQueueSize, queueCap)
	}

	workerCnt := len(d.workers)
	if workerCnt != DefaultWorkerCount {
		t.Fatalf("want %v\ngot %v", DefaultWorkerCount, workerCnt)
	}

	for i, w := range d.workers {
		if w == nil {
			t.Fatal("%d : Worker is nil.", i)
		}

		if w.dispatcher != d {
			t.Errorf("%d : Invalid dispatcher.", i)
		}

		if w.quit == nil {
			t.Errorf("%d : Quit channel is not initialized.", i)
		}

		if w.task == nil {
			t.Errorf("%d : Tasker channel is not initialized.", i)
		}
	}
}

var errTestTasker = errors.New("test tasker error")

type testTasker struct {
	done bool
	ch   <-chan error
}

func (t *testTasker) Run(ctx context.Context) error {
	t.done = true

	return errTestTasker
}

func TestDispatcher(t *testing.T) {
	d := New(workerCount, queueSize)

	ctx := context.Background()

	tasks := make([]*testTasker, queueSize)
	for i := 0; i < len(tasks); i++ {
		task := &testTasker{done: false}
		task.ch = d.Enqueue(ctx, task)
		tasks[i] = task
	}

	quit := make(chan struct{})

	d.Start()

	go func() {
		for i, task := range tasks {
			err := <-task.ch
			if err == nil || err != errTestTasker {
				t.Errorf("Tasker.Run %d method returns invalid error %#v.", i, err)
			}

			if !task.done {
				t.Errorf("Tasker %d has never run.", i)
			}
		}
		close(quit)
	}()

	d.Wait()

	<-quit
}

var (
	taskTime    = 5 * time.Second
	taskTimeout = 10 * time.Millisecond
)

type cancelTask struct{}

func (t *cancelTask) Run(ctx context.Context) error {
	select {
	case <-time.After(taskTime):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func TestCancel(t *testing.T) {
	d := New(workerCount, queueSize)

	ctx, cancel := context.WithTimeout(context.Background(), taskTimeout)
	defer cancel()

	task := &cancelTask{}
	ch := d.Enqueue(ctx, task)

	quit := make(chan struct{})

	d.Start()

	go func() {
		err := <-ch
		if err != context.DeadlineExceeded {
			t.Error("task was not canceled")
		}
		close(quit)
	}()

	d.Wait()

	<-quit
}
