package p2p

import (
	"errors"
	"fmt"
	"net"
)

type TCPPeer struct {

	// underlying connection of the peer
	conn net.Conn

	// if we dial and retrieve a connection, true
	// if we accept and retrieve a connection, false
	outbound bool
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outbound: outbound,
	}
}

type TCPTransportOptions struct {
	ListenAddr    string
	HandshakeFunc HandshakeFunc
	Decoder       Decoder
	OnPeer        func(Peer) error
}

type TCPTransport struct {
	options  TCPTransportOptions
	listener net.Listener

	peers map[net.Addr]Peer
}

func NewTCPTransport(opts TCPTransportOptions) *TCPTransport {
	return &TCPTransport{
		options: opts,
	}
}

// func (t *TCPTransport) OnPeer(peer Peer) error {

// }

func (t *TCPTransport) Close() error {
	if err := t.listener.Close(); err != nil {
		return err
	}

	return nil
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error

	t.listener, err = net.Listen("tcp", t.options.ListenAddr)

	if err != nil {
		return err
	}

	go t.startAcceptLoop()

	return nil

}

func (t *TCPTransport) startAcceptLoop() error {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Printf("TCP accept error: %v", err)
		}

		if errors.Is(err, net.ErrClosed) {
			return nil
		}

		go t.handleConn(conn)
	}
}

func (t *TCPTransport) handleConn(conn net.Conn) {
	peer := NewTCPPeer(conn, false)

	if err := t.options.HandshakeFunc(peer); err != nil {
		conn.Close()
		fmt.Printf("TCP handshake error: %v", err)
		return
	}

	if t.options.OnPeer != nil {
		if err := t.options.OnPeer(peer); err != nil {
			conn.Close()
			fmt.Printf("TCP on peer error: %v", err)
			return
		}
	}

	// Read loop
	msg := Message{}
	for {
		if err := t.options.Decoder.Decode(conn, &msg); err != nil {
			fmt.Printf("TCP message decode error: %v", err)
			conn.Close()
			return
		}

		msg.From = conn.RemoteAddr()
		fmt.Printf("message: %v", msg)

	}

}
