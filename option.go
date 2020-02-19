package dispatcher

// An Option represents an option for Dispatcher
type Option interface {
	apply(d *Dispatcher)
}

type optionFunc func(d *Dispatcher)

func (f optionFunc) apply(d *Dispatcher) {
	f(d)
}

// WithWorkerCount returns an option that the Dispatcher has the count workers.
func WithWorkerCount(count int) Option {
	return optionFunc(func(d *Dispatcher) {
		d.count = count
	})
}

// WithQueueSize returns an option that the Dispatcher has a queue of the size.
func WithQueueSize(size int) Option {
	return optionFunc(func(d *Dispatcher) {
		d.size = size
	})
}
