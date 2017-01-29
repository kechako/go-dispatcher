package dispatcher

import "testing"

func TestNew(t *testing.T) {
	d := New()

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
	if poolCap != MaxWorkers {
		t.Errorf("want %v\ngot %v", MaxWorkers, poolCap)
	}

	queueCap := cap(d.queue)
	if queueCap != MaxQueues {
		t.Errorf("want %v\ngot %v", MaxQueues, queueCap)
	}

	workerCnt := len(d.workers)
	if workerCnt != MaxWorkers {
		t.Errorf("want %v\ngot %v", MaxWorkers, workerCnt)
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
	d := New()

	tasks := make([]*testTasker, MaxQueues)
	for i := 0; i < MaxQueues; i++ {
		task := &testTasker{false}
		d.Queue(task)
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
