package p2p

import (
	"encoding/gob"
	"io"
)

type Decoder interface {
	Decode(io.Reader, *Message) error
}

type GOBDecoder struct{}

func (d *GOBDecoder) Decode(r io.Reader, m *Message) error {
	return gob.NewDecoder(r).Decode(m)
}

type DefaultDecoder struct{}

func (d *DefaultDecoder) Decode(r io.Reader, m *Message) error {
	buf := make([]byte, 1024)
	n, err := r.Read(buf)
	if err != nil {
		return err
	}
	m.Payload = buf[:n]

	return nil

}
