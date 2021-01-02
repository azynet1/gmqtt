package event_queue

import (
	"errors"
)

var (
	ErrEventBufferFull = errors.New("event buffer full")
	ErrIDNotFound      = errors.New("id not found")
	ErrClosed          = errors.New("queue has been closed")
	ErrDuplicated      = errors.New("duplicated event")
)

// Store represents a event queue store for one node in cluster.
type Store interface {
	// Append appends the event to the queue.
	// Returns ErrEventBufferFull if the event buffer is full.
	// If receives duplicated event, return ErrDuplicated.
	Append(event Event) error
	// Ack remove the event from queue before the given id.
	// If id is not found, return ErrIDNotFound.
	Ack(id uint64) error
	// Read reads at most max events from the queue.
	// Calling this method will be blocked until there are any new events can be read or the store has been closed.
	// If the store has been closed, returns nil, ErrClosed.
	Read(max int) ([]Event, error)
	// Close will be called when the node is considered down.
	// This method must unblock the Read method.。
	Close() error
	// Len returns the length of the queue.
	Len() int
}
