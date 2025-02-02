package p2p

import "net"

const (
	IncomingMessage = 0x1
	IncomingStream  = 0x2
)

// RPC holds any arbitrary data that is being sent over the
// each transport between two nodes in the network
type RPC struct {
	From    net.Addr
	Payload []byte
	Stream  bool
}
