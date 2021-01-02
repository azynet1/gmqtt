package gmqtt

import (
	"bytes"
	"errors"
	"io"

	"github.com/DrmagicE/gmqtt/pkg/codec"
	"github.com/DrmagicE/gmqtt/pkg/packets"
)

// Subscription represents a subscription in gmqtt.
type Subscription struct {
	// ShareName is the share name of a shared subscription.
	// set to "" if it is a non-shared subscription.
	ShareName string
	// TopicFilter is the topic filter which does not include the share name.
	TopicFilter string
	// ID is the subscription identifier
	ID uint32
	// The following fields are Subscription Options.
	// See: https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901169

	// QoS is the qos level of the Subscription.
	QoS packets.QoS
	// NoLocal is the No Local option.
	NoLocal bool
	// RetainAsPublished is the Retain As Published option.
	RetainAsPublished bool
	// RetainHandling the Retain Handling option.
	RetainHandling byte
}

// GetFullTopicName returns the full topic name of the subscription.
func (s *Subscription) GetFullTopicName() string {
	if s.ShareName != "" {
		return "$share/" + s.ShareName + "/" + s.TopicFilter
	}
	return s.TopicFilter
}

// Copy makes a copy of subscription.
func (s *Subscription) Copy() *Subscription {
	return &Subscription{
		ShareName:         s.ShareName,
		TopicFilter:       s.TopicFilter,
		ID:                s.ID,
		QoS:               s.QoS,
		NoLocal:           s.NoLocal,
		RetainAsPublished: s.RetainAsPublished,
		RetainHandling:    s.RetainHandling,
	}
}

// Validate returns whether the subscription is valid.
// If you can ensure the subscription is valid then just skip the validation.
func (s *Subscription) Validate() error {
	if !packets.ValidV5Topic([]byte(s.GetFullTopicName())) {
		return errors.New("invalid topic name")
	}
	if s.QoS > 2 {
		return errors.New("invalid qos")
	}
	if s.RetainHandling != 0 && s.RetainHandling != 1 && s.RetainHandling != 2 {
		return errors.New("invalid retain handling")
	}
	return nil
}

// Encode encodes Subscription into bytes and writes it to w.
func (s *Subscription) Encode(w io.Writer) (err error) {
	var buf bytes.Buffer
	codec.WriteBinary(&buf, []byte(s.ShareName))
	codec.WriteBinary(&buf, []byte(s.TopicFilter))
	codec.WriteUint32(&buf, s.ID)
	buf.Write([]byte{s.QoS})
	codec.WriteBool(&buf, s.NoLocal)
	codec.WriteBool(&buf, s.RetainAsPublished)
	buf.Write([]byte{s.RetainHandling})
	_, err = buf.WriteTo(w)
	return
}

// Decode reads data from r and decodes it into Subscription.
func (s *Subscription) Decode(r io.Reader) (err error) {
	share, err := codec.ReadBinary(r)
	if err != nil {
		return err
	}
	s.ShareName = string(share)
	topic, err := codec.ReadBinary(r)
	if err != nil {
		return err
	}
	s.TopicFilter = string(topic)

	s.ID, err = codec.ReadUint32(r)
	if err != nil {
		return err
	}
	qos := make([]byte, 1)
	_, err = r.Read(qos)
	if err != nil {
		return err
	}
	s.QoS = qos[0]
	s.NoLocal, err = codec.ReadBool(r)
	if err != nil {
		return err
	}
	s.RetainAsPublished, err = codec.ReadBool(r)
	if err != nil {
		return err
	}
	rhl := make([]byte, 1)
	_, err = r.Read(rhl)
	if err != nil {
		return err
	}
	s.RetainHandling = rhl[0]
	return nil
}
