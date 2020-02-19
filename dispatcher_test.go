package dispatcher

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func Test_Dispatcher(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	d := Start(ctx)

	if d == nil {
		t.Fatal("New() should not return nil")
	}

	if d.count != DefaultWorkerCount {
		t.Errorf("Dispatcher.count must be DefaultWorkerCount (got %d, want %d)", d.count, DefaultWorkerCount)
	}
	if d.size != DefaultQueueSize {
		t.Errorf("Dispatcher.size must be DefaultQueueSize (got %d, want %d)", d.size, DefaultQueueSize)
	}

	if cap(d.queue) != d.size {
		t.Error("size of Dispatcher.queue is not valid")
	}
}

func Test_DispatcherWithOptions(t *testing.T) {
	var (
		count = DefaultWorkerCount * 2
		size  = DefaultQueueSize * 2
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	d := Start(ctx, WithWorkerCount(count), WithQueueSize(size))

	if d == nil {
		t.Fatal("New() should not return nil")
	}

	if d.count != count {
		t.Errorf("Dispatcher.count is %d, want %d", d.count, count)
	}
	if d.size != size {
		t.Errorf("Dispatcher.size is %d, want %d", d.size, size)
	}

	if cap(d.queue) != d.size {
		t.Error("size of Dispatcher.queue is not valid")
	}
}

type Waiter interface {
	Wait() error
}

func WaitWithTimeout(ctx context.Context, w Waiter, timeout time.Duration) error {
	c := make(chan struct{})

	go func() {
		defer close(c)
		w.Wait()
	}()

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	select {
	case <-c:
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

type TestTask struct {
	name string
	done bool
}

func (t *TestTask) Run(ctx context.Context) error {
	t.done = true

	return nil
}

func Test_DispatcherStart(t *testing.T) {
	ctx := context.Background()
	d := Start(ctx, WithWorkerCount(2), WithQueueSize(10))

	if d == nil {
		t.Fatal("New() should not return nil")
	}

	var tasks []*TestTask
	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("task %d", i)
		task := &TestTask{name: name}
		tasks = append(tasks, task)
		d.Add(task)
	}

	if err := WaitWithTimeout(ctx, d, 5*time.Second); err != nil {
		t.Error("Running tasks is timeout")
	}

	for _, task := range tasks {
		if !task.done {
			t.Errorf("%s is not done", task.name)
		}
	}
}
