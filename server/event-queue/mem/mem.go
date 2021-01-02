package memory

import (
	"container/list"
	"sync"

	"go.uber.org/zap"

	"github.com/DrmagicE/gmqtt/persistence/queue"
	"github.com/DrmagicE/gmqtt/server/event-queue"
)

// Queue implements the event_queue.Store.
type Queue struct {
	cond     *sync.Cond
	nodeName string
	l        *list.List
	// nextRead is the read positions for next Read.
	nextRead *list.Element
	// index store the mapping of event id to event.
	index  map[uint64]*list.Element
	closed bool
	max    int
	log    *zap.Logger
}

func NewQueue(nodeName string, max int, logger *zap.Logger) event_queue.Store {
	return &Queue{
		cond:     sync.NewCond(&sync.Mutex{}),
		nodeName: nodeName,
		l:        list.New(),
		nextRead: nil,
		index:    make(map[uint64]*list.Element),
		closed:   false,
		max:      max,
		log:      logger,
	}
}

func (q *Queue) Append(event event_queue.Event) error {
	q.cond.L.Lock()
	defer func() {
		q.cond.L.Unlock()
		q.cond.Signal()
	}()
	if q.closed {
		return event_queue.ErrClosed
	}
	if q.max <= q.l.Len() {
		return event_queue.ErrEventBufferFull
	}
	if _, ok := q.index[event.ID()]; ok {
		return event_queue.ErrDuplicated
	}
	e := q.l.PushBack(event)
	q.index[event.ID()] = e
	if q.nextRead == nil {
		q.nextRead = e
	}
	return nil
}

func (q *Queue) Ack(id uint64) error {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	if q.closed {
		return event_queue.ErrClosed
	}
	if _, ok := q.index[id]; ok {
		var next *list.Element
		for e := q.l.Front(); e != nil; e = next {
			next = e.Next()
			if eid := e.Value.(event_queue.Event).ID(); eid <= id {
				q.l.Remove(e)
				if eid == id {
					return nil
				}
			}
		}
		return nil
	}
	return event_queue.ErrIDNotFound
}

func (q *Queue) Read(max int) (rs []event_queue.Event, err error) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	for (q.l.Len() == 0 || q.nextRead == nil) && !q.closed {
		q.cond.Wait()
	}
	if q.closed {
		return nil, queue.ErrClosed
	}
	length := q.l.Len()
	if max < length {
		length = max
	}
	for i := 0; i < length && q.nextRead != nil; i++ {
		v := q.nextRead
		rs = append(rs, v.Value.(event_queue.Event))
		q.nextRead = v.Next()
	}

	return rs, nil
}

func (q *Queue) Close() error {
	q.cond.L.Lock()
	defer func() {
		q.cond.L.Unlock()
		q.cond.Signal()
	}()
	q.closed = true
	return nil
}

func (q *Queue) Len() int {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	return q.l.Len()
}
