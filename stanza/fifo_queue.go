package stanza

// FIFO queue for string contents
// Implementations have no guarantee regarding thread safety !
type FifoQueue interface {
	// Pop returns the first inserted element still in queue and delete it from queue
	// No guarantee regarding thread safety !
	Pop() Queueable

	// PopN returns the N first inserted elements still in queue and delete them from queue
	// No guarantee regarding thread safety !
	PopN(i int) []Queueable

	// Peek returns a copy of the first inserted element in queue without deleting it
	// No guarantee regarding thread safety !
	Peek() Queueable

	// Peek returns a copy of the first inserted element in queue without deleting it
	// No guarantee regarding thread safety !
	PeekN() []Queueable
	// Push adds an element to the queue
	// No guarantee regarding thread safety !
	Push(s Queueable) error

	// Empty returns true if queue is empty
	// No guarantee regarding thread safety !
	Empty() bool
}

type Queueable interface {
	QueueableName() string
}
