package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	event_queue "github.com/DrmagicE/gmqtt/server/event-queue"
)

func TestQueue(t *testing.T) {
	a := assert.New(t)
	var err error
	s := NewQueue("nodename", 3, zap.NewNop())
	err = s.Append(&event_queue.SubscribeEvent{
		EventID: 1,
	})
	a.NoError(err)
	err = s.Append(&event_queue.SubscribeEvent{
		EventID: 2,
	})
	a.NoError(err)
	// test duplicated events
	err = s.Append(&event_queue.SubscribeEvent{
		EventID: 2,
	})
	a.Equal(event_queue.ErrDuplicated, err)

	err = s.Append(&event_queue.SubscribeEvent{
		EventID: 3,
	})
	a.NoError(err)
	err = s.Append(&event_queue.SubscribeEvent{
		EventID: 4,
	})
	a.Equal(event_queue.ErrEventBufferFull, err)

	ev, err := s.Read(5)
	a.NoError(err)
	a.Len(ev, 3)

	err = s.Ack(6)
	a.Equal(event_queue.ErrIDNotFound, err)

	err = s.Ack(3)
	a.NoError(err)

	a.Equal(0, s.Len())

	a.NoError(s.Close())

	_, err = s.Read(10)
	a.Equal(event_queue.ErrClosed, err)
}
