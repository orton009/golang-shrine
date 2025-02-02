package p2p

import "net"

// peer is remote node
type Peer interface {
	net.Conn
	Send([]byte) error
	CloseStream()
}

// this can be tcp, udp, or websockets
type Transport interface {
}
