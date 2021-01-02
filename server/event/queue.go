package event

import (
	"container/list"
	"sync"

	"go.uber.org/zap"
)

// Queue is the event queue for nodes internal communication.
type Queue struct {
	cond     *sync.Cond
	nodeName string
	l        *list.List
	// nextRead is the read positions for next Read.
	nextRead *list.Element
	// index store event_id => eventd
	index  map[uint64]*list.Element
	max    int
	log    *zap.Logger
	closed bool
}

func NewQueue(nodeName string, max int, logger *zap.Logger) *Queue {
	return &Queue{
		cond:     sync.NewCond(&sync.Mutex{}),
		nodeName: nodeName,
		l:        list.New(),
		index:    make(map[uint64]*list.Element),
		max:      max,
		log:      logger,
	}
}

// Append appends the event to the queue.
// Returns ErrEventBufferFull if the event buffer is full.
func (q *Queue) Append(evt Event) error {
	q.cond.L.Lock()
	defer func() {
		q.cond.L.Unlock()
		q.cond.Signal()
	}()
	if q.max <= q.l.Len() {
		return ErrEventBufferFull
	}
	e := q.l.PushBack(evt)
	if q.nextRead == nil {
		q.nextRead = e
	}
	return nil
}

// Ack remove the event from queue for the given id.
// If id is not found, return ErrIDNotFound.
func (q *Queue) Ack(id uint64) error {
	q.cond.L.Lock()
	defer func() {
		q.cond.L.Unlock()
		q.cond.Signal()
	}()
	if _, ok := q.index[id]; ok {
		var next *list.Element
		for e := q.l.Front(); e != nil; e = next {
			next = e.Next()
			if eid := e.Value.(Event).ID(); eid <= id {
				q.l.Remove(e)
				if eid == id {
					return nil
				}
			}
		}
		return nil
	}
	return ErrIDNotFound
}

// Read reads the next event and set the event id to given id.
// Calling this method will be blocked until there are any new events can be read or the store has been closed.
// If the store has been closed, returns nil
func (q *Queue) Read(id uint64) (rs Event) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	for (q.l.Len() == 0 || q.nextRead == nil) && !q.closed {
		q.cond.Wait()
	}
	if q.closed {
		return nil
	}
	if e := q.nextRead; e != nil {
		q.nextRead = e.Next()
		e.Value.(Event).SetID(id)
		q.index[id] = e
		return e.Value.(Event)
	}
	return nil
}

// SetReadPos set the next read position.
func (q *Queue) SetReadPos(start uint64) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	elem := q.l.Front()
	if elem == nil {
		return
	}
	for i := uint64(0); i < start; i++ {
		elem = elem.Next()
	}
	q.nextRead = elem
}
