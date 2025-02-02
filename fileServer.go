package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/orton009/shrine/p2p"
)

type Message struct {
	Payload any
}

type MessageStoreFile struct {
	// ID   string
	Key  string
	Size int64
}

type MessageGetFile struct {
	// ID  string
	Key string
}

type FileServerOpts struct {
	StorageRoot    string
	Transport      *p2p.TCPTransport
	BootstrapNodes []string
}

type FileServer struct {
	FileServerOpts

	peerLock sync.Mutex
	peers    map[string]p2p.TCPPeer
	storage  *p2p.Storage
	quitch   chan struct{}
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
		quitch:         make(chan struct{}),
	}
}
func (fs *FileServer) log(logPayload string) {

	prefix := fmt.Sprintf("[%s] ", fs.Transport.Addr())
	log.Println(prefix, logPayload)
}

func (fs *FileServer) bootstrapNetwork() error {
	for _, addr := range fs.BootstrapNodes {
		if len(addr) == 0 || addr == fs.Transport.Addr() {
			continue
		}

		go func(addr string) {

			fs.log(fmt.Sprintf("attempting to connect with remote %s", addr))
			if err := fs.Transport.Dial(addr); err != nil {
				fs.log(fmt.Sprint("dial error: ", err))
			}
		}(addr)
	}

	return nil
}

func (fs *FileServer) broadcast(msg *Message) error {

	buf := new(bytes.Buffer)

	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return fmt.Errorf("error encoding message: %v", err)
	}

	rpcEncoder := p2p.RPCEncoder{}
	encodedBytes := rpcEncoder.EncodeBytes(buf.Bytes())

	for addr, peer := range fs.peers {

		fs.log(fmt.Sprint("writing to peer...", addr))

		if err := peer.Send(encodedBytes); err != nil {
			return err
		}
	}

	return nil
}

func (fs *FileServer) Store(path string, r io.Reader) error {

	fileBuffer := new(bytes.Buffer)
	tee := io.TeeReader(r, fileBuffer)

	size, err := fs.storage.Write(path, tee)
	if err != nil {
		return err
	}

	msg := &Message{Payload: MessageStoreFile{
		Key:  path,
		Size: size,
	}}

	fs.log(fmt.Sprint("broadcasting file to other peers in network, path: ", path))

	if err := fs.broadcast(msg); err != nil {
		return err
	}

	// TODO: delay for peer to process this message before receiving the stream

	time.Sleep(time.Millisecond * 5)

	// send write to other peers in network

	peers := []io.Writer{}
	for _, peer := range fs.peers {
		peers = append(peers, peer)
	}

	mw := io.MultiWriter(peers...)

	encoder := p2p.RPCEncoder{}
	encodedReader, err := encoder.EncodeStream(fileBuffer)
	if err != nil {
		return err
	}

	n, err := io.Copy(mw, encodedReader)
	if err != nil {
		return err
	}

	fs.log(fmt.Sprintf("successfully written %d bytes in network", n))

	return nil
}

func (fs *FileServer) Read(path string) (io.Reader, error) {
	var io io.Reader
	var err error

	if io, err = fs.storage.Read(path); err != nil {
		return io, err
	}

	return io, nil
}

func (fs *FileServer) OnPeer(peer *p2p.TCPPeer) error {
	fs.peerLock.Lock()
	defer fs.peerLock.Unlock()

	fs.peers[peer.RemoteAddr().String()] = *peer

	fs.log(fmt.Sprintf("connected with remote %s %s", peer.RemoteAddr(), peer.LocalAddr().String()))

	return nil

}

func (fs *FileServer) Stop() {
	fs.log("Stoppping the file server programatically...")
	close(fs.quitch)
}

func (fs *FileServer) handleMessage(from string, m *Message) error {

	fs.log(fmt.Sprintf("Handling message from %s \n Message: %v\n", from, m))

	switch m.Payload.(type) {
	case MessageGetFile:
		return nil
	case MessageStoreFile:
		return nil
	default:
		return fmt.Errorf("FROM: [%s] received unknown message format %v", from, m)
	}
}

func (fs *FileServer) loop() {
	defer func() {
		fs.log("file server stopped due to error or user quit action")
		fs.Transport.Close()
	}()

	for {
		select {
		case rpc := <-fs.Transport.Consume():
			var msg Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
				fs.log(fmt.Sprint("decoding error: ", err))
			}

			if err := fs.handleMessage(rpc.From.String(), &msg); err != nil {
				fs.log(fmt.Sprintf("handle message error: %v\n", err))
			}

		case <-fs.quitch:
			return
		}
	}
}

func (fs *FileServer) Start() error {

	fs.log("starting fileserver...")

	if err := fs.Transport.ListenAndAccept(); err != nil {
		return err
	}

	fs.bootstrapNetwork()
	fs.loop()

	return nil
}

func init() {
	gob.Register(MessageGetFile{})
	gob.Register(MessageStoreFile{})
}
