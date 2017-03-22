package dispatcher

import (
	"runtime"
	"testing"
)

var (
	workerCount = 4
	queueSize   = runtime.NumCPU()
)

func TestNew(t *testing.T) {

	d := New(workerCount, queueSize)

	if d == nil {
		t.Error("New returned nil")
	}

	if d.pool == nil {
		t.Error("Worker pool is not initialized.")
	}
	if d.queue == nil {
		t.Error("Tasker queue is not initialized.")
	}
	if d.quit == nil {
		t.Error("Quit channel is not initialized.")
	}
	if d.workers == nil {
		t.Error("Workers not initialized.")
	}

	poolCap := cap(d.pool)
	if poolCap != workerCount {
		t.Errorf("want %v\ngot %v", workerCount, poolCap)
	}

	queueCap := cap(d.queue)
	if queueCap != queueSize {
		t.Errorf("want %v\ngot %v", queueSize, queueCap)
	}

	workerCnt := len(d.workers)
	if workerCnt != workerCount {
		t.Errorf("want %v\ngot %v", workerCount, workerCnt)
	}

	for i, w := range d.workers {
		if w == nil {
			t.Errorf("%d : Worker is nil.", i)
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

type testTasker struct {
	done bool
}

func (t *testTasker) Run() {
	t.done = true
}

func TestDispatcher(t *testing.T) {
	d := New(workerCount, queueSize)

	tasks := make([]*testTasker, queueSize)
	for i := 0; i < queueSize; i++ {
		task := &testTasker{false}
		d.Enqueue(task)
		tasks[i] = task
	}

	d.Start()
	d.Wait()

	for i, task := range tasks {
		if !task.done {
			t.Errorf("Tasker %d has never run.", i)
		}
	}
}
