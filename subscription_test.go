package gmqtt

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeDecodeSubscription(t *testing.T) {
	a := assert.New(t)
	tt := []*Subscription{
		{
			ShareName:         "shareName",
			TopicFilter:       "filter",
			ID:                1,
			QoS:               1,
			NoLocal:           false,
			RetainAsPublished: false,
			RetainHandling:    0,
		}, {
			ShareName:         "",
			TopicFilter:       "abc",
			ID:                0,
			QoS:               2,
			NoLocal:           false,
			RetainAsPublished: true,
			RetainHandling:    1,
		},
	}

	for _, v := range tt {
		buf := &bytes.Buffer{}
		err := v.Encode(buf)
		a.NoError(err)

		sub := &Subscription{}
		err = sub.Decode(bytes.NewBuffer(buf.Bytes()))
		a.NoError(err)
		a.Equal(v, sub)
	}
}
