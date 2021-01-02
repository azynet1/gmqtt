package event

import (
	"fmt"

	"github.com/DrmagicE/gmqtt"
)

type SubscribeEvent struct {
	EventID       uint64
	ClientID      string
	Subscriptions []*gmqtt.Subscription
}

func (s *SubscribeEvent) ID() uint64 {
	return s.EventID
}

func (s *SubscribeEvent) SetID(id uint64) {
	s.EventID = id
}

func (s *SubscribeEvent) Type() Type {
	return TypeSubscribe
}

func (s *SubscribeEvent) String() string {
	return fmt.Sprintf("Subscribe, id: %d", s.EventID)
}
