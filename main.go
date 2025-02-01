package main

import (
	"log"

	"github.com/orton009/shrine/p2p"
)

func startServer(addr string) {

	log.Println("Starting server on: ", addr)

	tcpOpts := p2p.TCPTransportOptions{
		ListenAddr:    ":" + addr,
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       &p2p.DefaultDecoder{},
	}

	tr := p2p.NewTCPTransport(tcpOpts)

	storage := p2p.NewStorage(p2p.StorageOpts{
		PathTransformFunc: p2p.DefaultPathTransformFunc,
		Root:              addr + "_root",
	})

	fs := NewFileServer(tr, storage)

	if err := fs.Start(); err != nil {
		log.Fatal(err)
	}

}

func main() {

	startServer("3001")

	for {

	}

}
