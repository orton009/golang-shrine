package main

import (
	"bytes"
	"context"
	"log"
	"time"

	"github.com/orton009/shrine/p2p"
)

func createServer(ctx context.Context, addr string, nodes []string) *FileServer {

	tcpOpts := p2p.TCPTransportOptions{
		ListenAddr:    "127.0.0.1" + addr,
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       &p2p.DefaultDecoder{},
	}

	tr := p2p.NewTCPTransport(tcpOpts)

	fsOpts := FileServerOpts{
		StorageRoot:    "./" + addr + "_root",
		Transport:      tr,
		BootstrapNodes: nodes,
	}
	fs := NewFileServer(fsOpts)
	tr.OnPeer = fs.OnPeer

	go func(ctx context.Context, fs *FileServer) {

		select {

		case <-ctx.Done():
			fs.Stop()
			return

		default:
			if err := fs.Start(); err != nil {
				log.Fatal(err)
				return

			}
		}
	}(ctx, fs)

	return fs

}

const (
	p1 = ":3001"
	p2 = ":3002"
)

func main() {

	// ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	ctx, _ := context.WithCancel(context.Background())

	fs1 := createServer(ctx, p1, []string{p2})

	fs2 := createServer(ctx, p2, []string{})

	// waiting for other peers to start
	time.Sleep(10 * time.Millisecond)

	buf := new(bytes.Buffer)
	buf.Write([]byte("hello world!"))
	if err := fs1.Store("hello", buf); err != nil {
		log.Fatal(err)
	}

	<-ctx.Done()
	fs1.Stop()
	fs2.Stop()

}
