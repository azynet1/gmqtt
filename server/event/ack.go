package event

import (
	"fmt"
)

// AckEvent represents the acknowledge event.
// After receive ack event, the associated event must be deleted from the event queue.
type AckEvent struct {
	EventID uint64
}

func (a *AckEvent) SetID(id uint64) {
	a.EventID = id
}

func (a *AckEvent) ID() uint64 {
	return a.EventID
}

func (a *AckEvent) Type() Type {
	return TypeAck
}

func (a *AckEvent) String() string {
	return fmt.Sprintf("Ack, id: %d", a.EventID)
}
