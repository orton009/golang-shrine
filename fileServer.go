package main

import (
	"sync"

	"github.com/orton009/shrine/p2p"
)

type FileServerOpts struct {
	StorageRoot string
	Transport   p2p.TCPTransport
}

type FileServer struct {
	FileServerOpts

	peerLock sync.Mutex
	peers    map[string]p2p.TCPPeer
	storage  *p2p.Storage
}

func NewFileServer(opts FileServerOpts) *FileServer {

	storageOpts := p2p.StorageOpts{
		Root: opts.StorageRoot,
	}
	storage := p2p.NewStorage(storageOpts)

	return &FileServer{
		FileServerOpts: opts,
		peers:          map[string]p2p.TCPPeer{},
		storage:        &storage,
		peerLock:       sync.Mutex{},
	}
}

func (fs *FileServer) Start() error {
	if err := fs.Transport. (); err != nil {
		return err
	}
	return nil
}
