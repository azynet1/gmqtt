package event_queue

import (
	"io"

	"github.com/DrmagicE/gmqtt"
	"github.com/DrmagicE/gmqtt/pkg/codec"
)

type SubscribeEvent struct {
	EventID       uint64
	ClientID      string
	Subscriptions []*gmqtt.Subscription
}

func (s *SubscribeEvent) ID() uint64 {
	return s.EventID
}

func (s *SubscribeEvent) Type() Type {
	return TypeSubscribe
}

func (s *SubscribeEvent) Pack(w io.Writer) (err error) {
	_, err = w.Write([]byte{TypeSubscribe})
	if err != nil {
		return err
	}
	err = codec.WriteUint64(w, s.EventID)
	if err != nil {
		return err
	}
	err = codec.WriteBinary(w, []byte(s.ClientID))
	if err != nil {
		return err
	}
	err = codec.WriteUint16(w, uint16(len(s.Subscriptions)))
	if err != nil {
		return err
	}
	for _, v := range s.Subscriptions {
		err = v.Encode(w)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SubscribeEvent) Unpack(r io.Reader) (err error) {
	s.EventID, err = codec.ReadUint64(r)
	if err != nil {
		return err
	}
	cid, err := codec.ReadBinary(r)
	if err != nil {
		return err
	}
	s.ClientID = string(cid)
	l, err := codec.ReadUint16(r)
	if err != nil {
		return err
	}
	for i := 0; i < int(l); i++ {
		sub := &gmqtt.Subscription{}
		err = sub.Decode(r)
		if err != nil {
			return err
		}
		s.Subscriptions = append(s.Subscriptions, sub)
	}
	return nil
}

func (s *SubscribeEvent) String() string {
	return "Subscribe"
}
