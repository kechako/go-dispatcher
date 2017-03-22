package main

import (
	"runtime"
	"time"

	dispatcher "github.com/kechako/go-dispatcher"
)

type task struct {
}

func (t *task) Run() {
	// Do something
	time.Sleep(1 * time.Second)
}

func main() {
	d := dispatcher.New(runtime.NumCPU(), 10000)
	d.Start()

	for i := 0; i < 100; i++ {
		d.Enqueue(new(task))
	}

	d.Wait()
}
