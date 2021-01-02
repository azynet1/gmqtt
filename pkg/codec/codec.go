package codec

import (
	"encoding/binary"
	"io"
)

func ReadUint16(r io.Reader) (uint16, error) {
	b := make([]byte, 2)
	_, err := r.Read(b)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(b), nil
}

func WriteUint16(w io.Writer, i uint16) error {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, i)
	_, err := w.Write(b)
	return err
}

func ReadUint32(r io.Reader) (uint32, error) {
	b := make([]byte, 4)
	_, err := r.Read(b)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(b), nil
}

func WriteUint32(w io.Writer, i uint32) error {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, i)
	_, err := w.Write(b)
	return err
}

func ReadUint64(r io.Reader) (uint64, error) {
	b := make([]byte, 8)
	_, err := r.Read(b)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(b), nil
}

func WriteUint64(w io.Writer, i uint64) (err error) {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, i)
	_, err = w.Write(b)
	return
}

func ReadBool(r io.Reader) (bool, error) {
	b := make([]byte, 1)
	_, err := r.Read(b)
	if err != nil {
		return false, err
	}
	if b[0] == 0 {
		return false, nil
	}
	return true, nil
}

func WriteBool(w io.Writer, b bool) (err error) {
	var bt []byte
	if b {
		bt = []byte{1}
	} else {
		bt = []byte{0}
	}
	_, err = w.Write(bt)
	return err
}

func ReadBinary(r io.Reader) ([]byte, error) {
	l, err := ReadUint16(r)
	if err != nil {
		return nil, err
	}
	b := make([]byte, l)
	_, err = r.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func WriteBinary(w io.Writer, b []byte) error {
	err := WriteUint16(w, uint16(len(b)))
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	return err
}
