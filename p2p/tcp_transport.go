package p2p

import (
	"fmt"
	"net"
	"sync"
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
}

type TCPTransport struct {
	options  TCPTransportOptions
	listener net.Listener

	mu    sync.Mutex
	peers map[net.Addr]Peer
}

func NewTCPTransport(opts TCPTransportOptions) *TCPTransport {
	return &TCPTransport{
		options: opts,
	}
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

		go t.handleConn(conn)
	}
}

func (t *TCPTransport) handleConn(conn net.Conn) {
	peer := &TCPPeer{
		conn:     conn,
		outbound: false,
	}

	if err := t.options.HandshakeFunc(peer); err != nil {
		conn.Close()
		fmt.Printf("TCP handshake error: %v", err)
		return
	}

	// Read loop
	msg := Message{}
	for {
		if err := t.options.Decoder.Decode(conn, &msg); err != nil {
			fmt.Printf("TCP error: %v", err)
			continue
		}

		msg.From = conn.RemoteAddr()
		fmt.Printf("message: %v", msg)

	}

}
