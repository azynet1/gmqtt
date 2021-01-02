package codec

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadWriteUint16(t *testing.T) {
	a := assert.New(t)
	tt := []uint16{1, 2, 3, 4, 5, 6}
	for _, v := range tt {
		wr := &bytes.Buffer{}
		err := WriteUint16(wr, v)
		br := bytes.NewBuffer(wr.Bytes())
		a.NoError(err)
		i, err := ReadUint16(br)
		a.NoError(err)
		a.Equal(v, i)
	}
}

func TestReadWriteUint32(t *testing.T) {
	a := assert.New(t)
	tt := []uint32{1, 2, 3, 4, 5, 6}
	for _, v := range tt {
		wr := &bytes.Buffer{}
		err := WriteUint32(wr, v)
		br := bytes.NewBuffer(wr.Bytes())
		a.NoError(err)
		i, err := ReadUint32(br)
		a.NoError(err)
		a.Equal(v, i)
	}
}

func TestReadWriteUint64(t *testing.T) {
	a := assert.New(t)
	tt := []uint64{1, 2, 3, 4, 5, 6}
	for _, v := range tt {
		wr := &bytes.Buffer{}
		err := WriteUint64(wr, v)
		br := bytes.NewBuffer(wr.Bytes())
		a.NoError(err)
		i, err := ReadUint64(br)
		a.NoError(err)
		a.Equal(v, i)
	}
}

func TestReadWriteBool(t *testing.T) {
	a := assert.New(t)
	tt := []bool{true, false}
	for _, v := range tt {
		wr := &bytes.Buffer{}
		err := WriteBool(wr, v)
		br := bytes.NewBuffer(wr.Bytes())
		a.NoError(err)
		i, err := ReadBool(br)
		a.NoError(err)
		a.Equal(v, i)
	}
}

func TestReadWriteBinary(t *testing.T) {
	a := assert.New(t)
	tt := [][]byte{{1, 2, 3, 4}, {}}
	for _, v := range tt {
		wr := &bytes.Buffer{}
		err := WriteBinary(wr, v)
		br := bytes.NewBuffer(wr.Bytes())
		a.NoError(err)
		i, err := ReadBinary(br)
		a.NoError(err)
		a.Equal(v, i)
	}
}
