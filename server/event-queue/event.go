package event_queue

import (
	"bufio"
	"fmt"
	"io"
)

// Type is the event type
type Type = byte

const (
	TypeSubscribe Type = iota + 1
	TypeUnsubscribe
	TypePublish
	TypeSessionCreate
	TypeSessionRemove
	TypeAck
)

// Event is used to exchange information with other nodes in the cluster.
type Event interface {
	// ID is a auto increment and unique identifier for events.
	ID() uint64
	// Type returns the type of the event.
	Type() Type
	// Pack encodes the event struct into bytes and writes it into io.Writer.
	Pack(w io.Writer) error
	// Unpack read bytes from io.Reader and decodes it into struct.
	Unpack(r io.Reader) error
	// String is mainly used in logging, debugging and testing.
	String() string
}

// Reader is used to read data from bufio.Reader and create event instance.
type Reader struct {
	bufr *bufio.Reader
}

// Writer is used to encode event into bytes and write it to bufio.Writer.
type Writer struct {
	bufw *bufio.Writer
}

// Flush writes any buffered data to the underlying io.Writer.
func (w *Writer) Flush() error {
	return w.bufw.Flush()
}

// ReadWriter warps Reader and Writer.
type ReadWriter struct {
	*Reader
	*Writer
}

// NewReader returns a new Reader.
func NewReader(r io.Reader) *Reader {
	if bufr, ok := r.(*bufio.Reader); ok {
		return &Reader{bufr: bufr}
	}
	return &Reader{bufr: bufio.NewReaderSize(r, 2048)}
}

// NewWriter returns a new Writer.
func NewWriter(w io.Writer) *Writer {
	if bufw, ok := w.(*bufio.Writer); ok {
		return &Writer{bufw: bufw}
	}
	return &Writer{bufw: bufio.NewWriterSize(w, 2048)}
}

// ReadPacket reads data from Reader and returns a Event instance.
// If any errors occurs, returns nil, error
func (r *Reader) ReadEvent() (Event, error) {
	eventType, err := r.bufr.ReadByte()
	if err != nil {
		return nil, err
	}
	// TODO
	var eventPacket Event
	switch eventType {
	case TypeSubscribe:
		eventPacket = &SubscribeEvent{}
	case TypeUnsubscribe:
	case TypePublish:
	case TypeSessionRemove:
	case TypeSessionCreate:
	case TypeAck:
		eventPacket = &AckEvent{}
	default:
		return nil, fmt.Errorf("invalid event type: %b", eventType)
	}
	err = eventPacket.Unpack(r.bufr)
	if err != nil {
		return nil, err
	}
	return eventPacket, nil
}

// WriteEvent writes the event to the Writer.
// Call Flush after WriteEvent to flush buffered data to the underlying io.Writer.
func (w *Writer) WriteEvent(ev Event) error {
	err := ev.Pack(w.bufw)
	if err != nil {
		return err
	}
	return nil
}

// WriteAndFlush writes and flush buffered data to the underlying io.Writer.
func (w *Writer) WriteAndFlush(ev Event) error {
	err := ev.Pack(w.bufw)
	if err != nil {
		return err
	}
	return w.Flush()
}
