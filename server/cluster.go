package server

import (
	"fmt"
	"math"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/hashicorp/serf/serf"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/DrmagicE/gmqtt/config"
	"github.com/DrmagicE/gmqtt/server/event"
)

type Cluster struct {
	mu       sync.Mutex
	wg       sync.WaitGroup
	cfg      config.Config
	nodeName string
	members  map[string]*serf.Member
	// store nodeName => eventProducer mapping
	producerQueue map[string]*eventProducer

	serfEventCh chan serf.Event

	exit chan struct{}
	srv  *server
}

func (c *Cluster) Run() {
	serfCfg := serf.DefaultConfig()
	serfCfg.NodeName = c.cfg.Cluster.NodeName
	serfCfg.EventCh = c.serfEventCh
	host, port, _ := net.SplitHostPort(c.cfg.Cluster.GossipAddr)
	if host != "" {
		serfCfg.MemberlistConfig.BindAddr = host
	}
	p, _ := strconv.Atoi(port)
	serfCfg.MemberlistConfig.BindPort = p
	//serfCfg.ReapInterval = 1 * time.Second
	//serfCfg.ReconnectTimeout = 10 * time.Second
	serfCfg.Tags = map[string]string{"internal_addr": c.cfg.Cluster.InternalAddr}
	serfCfg.Logger, _ = zap.NewStdLogAt(zaplog, zapcore.InfoLevel)
	serfCfg.MemberlistConfig.Logger, _ = zap.NewStdLogAt(zaplog, zapcore.InfoLevel)
	s, err := serf.Create(serfCfg)
	if err != nil {
		//TODO error handling
		fmt.Println(err)
		return
	}
	s.Join(c.cfg.Cluster.Join, true)
	c.wg.Add(2)
	go c.eventHandler()
	go c.serveTCP()
	c.wg.Wait()
}

func (c *Cluster) serveTCP() {
	defer c.wg.Done()
	l, err := net.Listen("tcp", c.cfg.Cluster.InternalAddr)
	if err != nil {
		panic(err)
	}
	for {
		conn, e := l.Accept()
		var tempDelay time.Duration
		if e != nil {
			if ne, ok := e.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				time.Sleep(tempDelay)
				continue
			}
			return
		}
		consumer := newEventConsumer(conn, c.srv.publishService, c.srv.subscriptionsDB)
		go consumer.serve()
	}
}

// BroadCast broadcasts the event to all known nodes in cluster.
func (c *Cluster) Broadcast(evt event.Event) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, v := range c.producerQueue {
		v.appendEvent(evt)
	}
}

// Multicast sends the event to specified nodes.
func (c *Cluster) Multicast(evt event.Event, nodeName []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, name := range nodeName {
		if q, ok := c.producerQueue[name]; ok {
			q.appendEvent(evt)
		}
	}
}

func (c *Cluster) eventHandler() {
	defer func() {
		c.wg.Done()
	}()
	for {
		select {
		case evt := <-c.serfEventCh:
			switch evt.EventType() {
			case serf.EventMemberJoin:
				c.nodeJoin(evt.(serf.MemberEvent))
			case serf.EventMemberLeave, serf.EventMemberFailed, serf.EventMemberReap:
				c.nodeFail(evt.(serf.MemberEvent))
			case serf.EventUser:
				//c.localEvent(e.(serf.UserEvent))
			case serf.EventMemberUpdate: // Ignore
				//c.nodeUpdate(e.(serf.MemberEvent))
			case serf.EventQuery: // Ignore
			default:
			}
		case <-c.exit:
			c.mu.Lock()
			for _, v := range c.producerQueue {
				v.stop()
			}
			c.mu.Unlock()
			return
		}
	}
}

func (c *Cluster) nodeJoin(member serf.MemberEvent) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, v := range member.Members {
		if v.Name == c.nodeName {
			continue
		}
		if _, ok := c.members[v.Name]; !ok {
			addr := v.Tags["internal_addr"]
			q := event.NewQueue(addr, math.MaxInt64, zaplog)
			zaplog.Info("node joined, create event producer", zap.String("remote_addr", addr), zap.String("node_name", v.Name))
			producer := newEventProducer(addr, c.srv.publishService, q)
			go producer.startProducer()
			c.producerQueue[v.Name] = producer
		}
	}
}

func (c *Cluster) nodeFail(member serf.MemberEvent) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, v := range member.Members {
		if v.Name == c.nodeName {
			// TODO
			continue
		}
		if p, ok := c.producerQueue[v.Name]; ok {
			zaplog.Info("node failed, stop event producer", zap.String("node_name", v.Name))
			p.stop()
		}
		delete(c.producerQueue, v.Name)
	}
}
