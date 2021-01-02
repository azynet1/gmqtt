package event_queue

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DrmagicE/gmqtt"
)

func TestSubscribeEvent_Pack_Unpack(t *testing.T) {
	a := assert.New(t)
	var tt = []SubscribeEvent{
		{
			EventID:  1,
			ClientID: "cid",
			Subscriptions: []*gmqtt.Subscription{
				{
					ShareName:         "share",
					TopicFilter:       "topic",
					ID:                1,
					QoS:               2,
					NoLocal:           true,
					RetainAsPublished: false,
					RetainHandling:    2,
				}, {
					ShareName:         "share1",
					TopicFilter:       "topic2",
					ID:                0,
					QoS:               1,
					NoLocal:           false,
					RetainAsPublished: true,
					RetainHandling:    1,
				},
			},
		}, {
			EventID:  2,
			ClientID: "cid",
			Subscriptions: []*gmqtt.Subscription{
				{
					ShareName:         "share",
					TopicFilter:       "topic",
					ID:                1,
					QoS:               2,
					NoLocal:           true,
					RetainAsPublished: false,
					RetainHandling:    2,
				},
			},
		},
	}
	for _, v := range tt {
		b := &bytes.Buffer{}
		err := v.Pack(b)
		a.NoError(err)
		r := bufio.NewReader(bytes.NewBuffer(b.Bytes()))
		rd := NewReader(r)
		ev, err := rd.ReadEvent()
		a.NoError(err)
		a.IsType(&SubscribeEvent{}, ev)
		a.Equal(&v, ev)
	}
}
