# go-dispatcher

go-dispatcher is a task dispatcher for golang.

## Usage

``` go
package main

import (
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
	d := dispatcher.New()
	d.Start()

	for i := 0; i < 100; i++ {
		d.Queue(new(task))
	}

	d.Wait()
}
```

