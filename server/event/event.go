package event

import (
	"encoding/gob"
	"errors"
)

// Type is the event type
type Type = byte

const (
	TypeClientHello Type = iota + 1
	TypeServerHello
	TypeSubscribe
	TypeUnsubscribe
	TypePublish
	TypeSessionCreate
	TypeSessionRemove
	TypeAck
)

var (
	ErrEventBufferFull = errors.New("event buffer full")
	ErrIDNotFound      = errors.New("id not found")
	ErrClosed          = errors.New("queue has been closed")
)

func init() {
	gob.Register(&AckEvent{})
	gob.Register(&ClientHello{})
	gob.Register(&ServerHello{})
	gob.Register(&PublishEvent{})
	gob.Register(&SubscribeEvent{})
}

// Event is used to exchange information with other nodes in the cluster.
type Event interface {
	// ID is a auto increment and unique identifier for events.
	ID() uint64
	SetID(uint64)
	// Type returns the type of the event.
	Type() Type
	// String is mainly used in logging, debugging and testing.
	String() string
}
