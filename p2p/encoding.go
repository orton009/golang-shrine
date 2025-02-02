package p2p

import (
	"bytes"
	"encoding/gob"
	"io"
)

type Decoder interface {
	Decode(io.Reader, *RPC) error
}

type GOBDecoder struct{}

func (d *GOBDecoder) Decode(r io.Reader, m *RPC) error {
	return gob.NewDecoder(r).Decode(m)
}

type DefaultDecoder struct{}

func (d *DefaultDecoder) Decode(r io.Reader, m *RPC) error {
	peekBuf := make([]byte, 1)
	if _, err := r.Read(peekBuf); err != nil {
		return err
	}

	stream := peekBuf[0] == IncomingStream

	if stream {
		m.Stream = true
		return nil
	}

	buf := make([]byte, 1024)
	n, err := r.Read(buf)
	if err != nil {
		return err
	}
	m.Payload = buf[:n]

	return nil

}

type RPCEncoder struct{}

func (e *RPCEncoder) EncodeStream(r io.Reader) (io.Reader, error) {
	return e.Encode(r, true)
}

func (e *RPCEncoder) EncodeBytes(b []byte) []byte {
	buf := make([]byte, 1)

	buf[0] = IncomingMessage

	buf = append(buf, b...)

	return buf

}

func (e *RPCEncoder) Encode(r io.Reader, isStream bool) (io.Reader, error) {

	peekBuf := make([]byte, 1)

	if isStream {
		peekBuf[0] = IncomingStream
	} else {
		peekBuf[0] = IncomingMessage
	}

	buf := new(bytes.Buffer)
	if _, err := buf.Write(peekBuf); err != nil {
		return buf, err
	}
	if _, err := io.Copy(buf, r); err != nil {
		return buf, err
	}

	return buf, nil
}
