package event_queue

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAckEvent_Pack_Unpack(t *testing.T) {
	a := assert.New(t)
	var tt = []AckEvent{
		{
			EventID: 1,
		}, {
			EventID: 2,
		}, {
			EventID: 3,
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
		a.IsType(&AckEvent{}, ev)
		a.Equal(&v, ev)
	}
}
