package server

import (
	"bufio"
	"encoding/gob"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/DrmagicE/gmqtt/server/event"
)

type eventProducer struct {
	remoteAddr string
	exit       chan struct{}
	publisher  Publisher
	errOnce    *sync.Once
	// reconnect timer
	timer     *time.Timer
	client    *clusterClient
	queue     *event.Queue
	sessionID string
	mu        sync.Mutex
}

func (e *eventProducer) stop() {
	select {
	case <-e.exit:
	default:
		close(e.exit)
	}
	e.client.conn.Close()
	e.client.wg.Wait()
}

func (e *eventProducer) appendEvent(evt event.Event) {
	err := e.queue.Append(evt)
	if err != nil {
		e.client.setError(err)
		return
	}
}

func newEventProducer(remoteAddr string, publisher Publisher, queue *event.Queue) *eventProducer {
	return &eventProducer{
		remoteAddr: remoteAddr,
		publisher:  publisher,
		errOnce:    &sync.Once{},
		exit:       make(chan struct{}),
		timer:      time.NewTimer(0),
		queue:      queue,
	}
}

func (e *eventProducer) newClusterClient(conn net.Conn) *clusterClient {
	bufr := newBufioReaderSize(conn, readBufferSize)
	bufw := newBufioWriterSize(conn, writeBufferSize)
	c := &clusterClient{
		conn:      conn,
		publisher: e.publisher,
		in:        make(chan event.Event, 2048),
		out:       make(chan event.Event, 2048),
		wg:        &sync.WaitGroup{},
		errOnce:   &sync.Once{},
		bufr:      bufr,
		bufw:      bufw,
		encoder:   gob.NewEncoder(bufw),
		decoder:   gob.NewDecoder(bufr),
		evtQueue:  e.queue,
		close:     make(chan struct{}),
		read:      make(chan struct{}, 1),
	}
	e.client = c
	return c
}

type clusterClient struct {
	evtQueue  *event.Queue
	ep        *eventProducer
	conn      net.Conn
	close     chan struct{}
	bufr      *bufio.Reader
	bufw      *bufio.Writer
	encoder   *gob.Encoder
	decoder   *gob.Decoder
	publisher Publisher
	in        chan event.Event
	out       chan event.Event
	wg        *sync.WaitGroup
	errOnce   *sync.Once

	read chan struct{}
	// store the last error
	err  error
	idMu sync.Mutex
	// id for the next event
	nextEventID uint64
}

func (c *clusterClient) getEventID() uint64 {
	c.idMu.Lock()
	defer c.idMu.Unlock()
	id := c.nextEventID
	c.nextEventID++
	return id
}

func (c *clusterClient) stop() {

}

func (c *clusterClient) setError(err error) {
	c.errOnce.Do(func() {
		close(c.close)
		c.err = err
		if err != nil {
			zaplog.Error("error", zap.Error(err))
		}
	})
}

func (c *clusterClient) encodeAndFlush(evt event.Event) (err error) {
	err = c.encoder.Encode(&evt)
	if err != nil {
		return err
	}
	zaplog.Info("sending event", zap.String("event", evt.String()))
	return c.bufw.Flush()
}

func (c *clusterClient) handshake(sessionID string) (respSessID string, err error) {
	err = c.encodeAndFlush(&event.ClientHello{
		SessionID: sessionID,
	})
	if err != nil {
		return "", err
	}
	// TODO configurable
	timeout := time.NewTimer(5 * time.Second)
	serverHello := make(chan event.Event, 1)
	errCh := make(chan error, 1)
	go func() {
		hello, err := c.readEvent()
		if err != nil {
			errCh <- err
			return
		}
		serverHello <- hello
	}()
	select {
	case ev := <-serverHello:

		var sh *event.ServerHello
		sh, ok := ev.(*event.ServerHello)
		if !ok {
			return "", errors.New("handshake error")
		}
		respSessID = sh.SessionID
		c.nextEventID = sh.LastEventID
		c.evtQueue.SetReadPos(sh.LastEventID)
		// TODO preform full sync if LastEventID = 0

	case <-timeout.C:
		return "", errors.New("handshake timeout")
	case err := <-errCh:
		return "", fmt.Errorf("handshake error: %s", err)
	}
	return respSessID, nil
}

func (e *eventProducer) startProducer() {
	var err error
	for {
		select {
		case <-e.exit:
			return
		case <-e.timer.C:
			err = e.serveConn()
			if err != nil {
				zaplog.Error("producer connection error", zap.Error(err))
			}
		}
	}
}

func (e *eventProducer) serveConn() (err error) {
	defer func() {
		if err != nil {
			// reconnect timer
			// TODO configurable
			e.timer.Reset(2 * time.Second)
		}
	}()
	conn, err := net.Dial("tcp", e.remoteAddr)
	if err != nil {
		return err
	}
	c := e.newClusterClient(conn)
	c.ep = e
	defer func() {
		conn.Close()
		putBufioReader(c.bufr)
		putBufioWriter(c.bufw)
	}()
	var sessID string
	if sessID, c.err = c.handshake(e.sessionID); c.err == nil {
		c.nextRead()
		e.sessionID = sessID
		zaplog.Info("producer handshake succeed", zap.String("session_id", e.sessionID))
		c.wg.Add(3)
		go c.readLoop()
		go c.writeLoop()
		go c.fetchEvents()
		c.wg.Wait()
	}
	return c.err
}

func (c *clusterClient) readEvent() (event.Event, error) {
	var evt event.Event
	err := c.decoder.Decode(&evt)
	if err != nil {
		return nil, err
	}
	zaplog.Info("event received", zap.String("event", evt.String()))
	return evt, nil
}

func (c *clusterClient) nextRead() {
	select {
	case c.read <- struct{}{}:
	default:
	}
}

func (c *clusterClient) readLoop() {
	var err error
	var ev event.Event
	defer func() {
		if re := recover(); re != nil {
			err = errors.New(fmt.Sprint(re))
		}
		c.setError(err)
		c.wg.Done()
	}()
	for {
		select {
		case <-c.close:
		default:
			ev, err = c.readEvent()
			if err != nil {
				return
			}
			if ack, ok := ev.(*event.AckEvent); ok {
				err = c.evtQueue.Ack(ack.ID())
			} else {
				err = fmt.Errorf("unexpected event type: %s", ev)
			}
			if err != nil {
				return
			}
			c.nextRead()
		}
	}
}

func (c *clusterClient) writeLoop() {
	var err error
	defer func() {
		if re := recover(); re != nil {
			err = errors.New(fmt.Sprint(re))
		}
		c.setError(err)
		c.wg.Done()
	}()
	for {
		select {
		case <-c.close:
			return
		case ev := <-c.out:
			err := c.encodeAndFlush(ev)
			if err != nil {
				return
			}
		}
	}
}

func (c *clusterClient) fetchEvents() {
	var err error
	defer func() {
		if re := recover(); re != nil {
			err = errors.New(fmt.Sprint(re))
		}
		c.setError(err)
		c.wg.Done()
	}()
	for {
		select {
		case <-c.read:
			ev := c.evtQueue.Read(c.getEventID())
			if ev == nil {
				err = errors.New("read nil event")
				return
			}
			c.write(ev)
		case <-c.close:
			return
		}

	}
}

func (c *clusterClient) write(event event.Event) {
	select {
	case <-c.close:
		return
	case c.out <- event:
	}
}
