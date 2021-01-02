package event

import (
	"fmt"

	"github.com/DrmagicE/gmqtt"
)

type PublishEvent struct {
	EventID uint64
	Message *gmqtt.Message
}

func (p *PublishEvent) SetID(id uint64) {
	p.EventID = id
}

func (p *PublishEvent) ID() uint64 {
	return p.EventID
}

func (p *PublishEvent) Type() Type {
	return TypePublish
}

func (p *PublishEvent) String() string {
	return fmt.Sprintf("Publish, id: %d", p.EventID)
}
