package event_queue

import (
	"encoding/binary"
	"io"

	"github.com/DrmagicE/gmqtt/pkg/codec"
)

// AckEvent represents the acknowledge event.
// After receive ack event, the associated event must be deleted from the event queue.
type AckEvent struct {
	EventID uint64
}

func (a *AckEvent) ID() uint64 {
	return a.EventID
}

func (a *AckEvent) Type() Type {
	return TypeAck
}

func (a *AckEvent) Pack(w io.Writer) error {
	_, err := w.Write([]byte{TypeAck})
	if err != nil {
		return err
	}
	return codec.WriteUint64(w, a.EventID)
}

func (a *AckEvent) Unpack(r io.Reader) error {
	restBuffer := make([]byte, 8)
	_, err := io.ReadFull(r, restBuffer)
	if err != nil {
		return err
	}
	a.EventID = binary.BigEndian.Uint64(restBuffer)
	return nil
}

func (a *AckEvent) String() string {
	return "Ack"
}
