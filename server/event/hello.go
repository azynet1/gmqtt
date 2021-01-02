package event

import (
	"fmt"
)

type ClientHello struct {
	SessionID string
}

func (c *ClientHello) SetID(id uint64) {
	return
}
func (c *ClientHello) ID() uint64 {
	return 0
}

func (c *ClientHello) Type() Type {
	return TypeClientHello
}

func (c *ClientHello) String() string {
	return "ClientHello"
}

type ServerHello struct {
	SessionID   string
	LastEventID uint64
}

func (s *ServerHello) ID() uint64 {
	return 0
}

func (s *ServerHello) SetID(id uint64) {
	return
}

func (s *ServerHello) Type() Type {
	return TypeServerHello
}

func (s *ServerHello) String() string {
	return fmt.Sprintf("ServerHello, session_id: %s, last_event_id: %d", s.SessionID, s.LastEventID)
}
