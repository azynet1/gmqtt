package server

import (
	"bufio"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	"go.uber.org/zap"

	"github.com/DrmagicE/gmqtt/persistence/subscription"
	"github.com/DrmagicE/gmqtt/server/event"
)

type eventConsumer struct {
	mu           sync.Mutex
	evtQueue     event.Queue
	conn         net.Conn
	remoteAddr   string
	exit         chan struct{}
	publisher    Publisher
	bufr         *bufio.Reader
	bufw         *bufio.Writer
	encoder      *gob.Encoder
	decoder      *gob.Decoder
	in           chan event.Event
	out          chan event.Event
	ackedEventID uint64
	wg           *sync.WaitGroup
	errOnce      *sync.Once
	subStore     subscription.Store
	sessionID    string
}

func newEventConsumer(c net.Conn, publisher Publisher, subStore subscription.Store) *eventConsumer {
	e := &eventConsumer{
		conn:      c,
		bufr:      newBufioReaderSize(c, readBufferSize),
		bufw:      newBufioWriterSize(c, writeBufferSize),
		in:        make(chan event.Event, readBufferSize),
		out:       make(chan event.Event, writeBufferSize),
		publisher: publisher,
		wg:        &sync.WaitGroup{},
		errOnce:   &sync.Once{},
		subStore:  subStore,
		sessionID: getRandomUUID(),
		exit:      make(chan struct{}),
	}
	e.encoder = gob.NewEncoder(e.bufw)
	e.decoder = gob.NewDecoder(e.bufr)
	return e
}

func (e *eventConsumer) serve() {
	defer func() {
		e.conn.Close()
	}()
	if err := e.handshake(); err == nil {
		e.wg.Add(3)
		go e.readLoop()
		go e.readHandler()
		go e.writeLoop()
		e.wg.Wait()
	}
}

func (e *eventConsumer) readEvent() (event.Event, error) {
	var evt event.Event
	err := e.decoder.Decode(&evt)
	if err != nil {
		return nil, err
	}
	zaplog.Info("event received", zap.String("event", evt.String()))
	return evt, nil
}

func (e *eventConsumer) encodeAndFlush(evt event.Event) (err error) {
	err = e.encoder.Encode(&evt)
	if err != nil {
		return err
	}
	zaplog.Info("sending event", zap.String("event", evt.String()))
	return e.bufw.Flush()
}

func (e *eventConsumer) handshake() error {
	evt, err := e.readEvent()
	if err != nil {
		return err
	}
	if _, ok := evt.(*event.ClientHello); ok {
		req := evt.(*event.ClientHello)
		resp := &event.ServerHello{}
		if e.sessionID == req.SessionID {
			resp.SessionID = req.SessionID
			resp.LastEventID = e.ackedEventID
		} else {
			// If the sessionID is not equal, the client node may just recover from a crash,
			// the server node will assign a new session id and request a full sync.
			resp.SessionID = getRandomUUID()
			resp.LastEventID = 0
		}
		err := e.encodeAndFlush(resp)
		return err
	}
	return errors.New("invalid packet")
}

func (e *eventConsumer) readHandler() {
	var err error
	defer func() {
		if re := recover(); re != nil {
			err = errors.New(fmt.Sprint(re))
		}
		e.setError(err)
		close(e.exit)
		e.wg.Done()
	}()
	for evt := range e.in {
		switch evt.(type) {
		// TODO other events.
		case *event.PublishEvent:
			pubEvt := evt.(*event.PublishEvent)
			e.publisher.Publish(pubEvt.Message)
			e.write(&event.AckEvent{EventID: evt.ID()})
		case *event.SubscribeEvent:
			subEvt := evt.(*event.SubscribeEvent)
			//TODO send retained message
			_, err := e.subStore.Subscribe(subEvt.ClientID, subEvt.Subscriptions...)
			if err != nil {
				return
			}
			e.write(&event.AckEvent{EventID: evt.ID()})
		}
		e.ackedEventID = evt.ID()
	}
}

func (e *eventConsumer) readLoop() {
	var err error

	defer func() {
		if re := recover(); re != nil {
			err = errors.New(fmt.Sprint(re))
		}
		e.setError(err)
		close(e.in)
		e.wg.Done()
	}()
	for {
		var evt event.Event
		evt, err = e.readEvent()
		if err != nil {
			if err != io.EOF && evt != nil {
				zaplog.Error("read error", zap.String("event", evt.String()))
			}
			return
		}
		e.in <- evt
	}
}

func (e *eventConsumer) writeLoop() {
	var err error
	defer func() {
		if re := recover(); re != nil {
			err = errors.New(fmt.Sprint(re))
		}
		e.setError(err)
		e.wg.Done()
	}()
	for {
		select {
		case <-e.exit:
			return
		case ev := <-e.out:
			err := e.encodeAndFlush(ev)
			if err != nil {
				return
			}
		}
	}
}

func (e *eventConsumer) setError(err error) {
	e.errOnce.Do(func() {
		if err != nil && err != io.EOF {
			zaplog.Error("consumer connection error", zap.Error(err))
		}
	})
}

func (e *eventConsumer) write(event event.Event) {
	select {
	case <-e.exit:
		return
	case e.out <- event:
	}
}
