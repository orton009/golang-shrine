package p2p

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"os"
	"strings"
)

type StorageOpts struct {
	PathTransformFunc func(path string) string
	Root              string
}

func DefaultPathTransformFunc(path string) string {

	hashBytes := sha1.Sum([]byte(path))
	hash := string(hex.EncodeToString(hashBytes[:]))
	return hash
}

type Storage struct {
	StorageOpts
}

func NewStorage(opts StorageOpts) Storage {

	if opts.PathTransformFunc == nil {
		opts.PathTransformFunc = DefaultPathTransformFunc
	}
	opts.Root = strings.TrimSuffix(opts.Root, "/")
	return Storage{
		StorageOpts: opts,
	}
}

func (s *Storage) writeStream(path string, r io.Reader) (int64, error) {
	var size int64

	if err := os.Mkdir(s.Root, os.ModePerm); !errors.Is(err, os.ErrExist) {
		return size, err
	}

	file, err := os.Create(path)
	if err != nil {
		return size, err
	}
	defer file.Close()

	size, err = io.Copy(file, r)
	if err != nil {
		return size, err
	}

	log.Printf("written %d bytes to disk", size)

	return size, nil
}

func (s *Storage) readStream(path string) (io.Reader, error) {

	r, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, r)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func (s *Storage) Read(key string) (io.Reader, error) {

	fullPath := s.FullPath(key)

	r, err := s.readStream(fullPath)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (s *Storage) Write(key string, r io.Reader) (int64, error) {

	fullPath := s.FullPath(key)

	size, err := s.writeStream(fullPath, r)

	if err != nil {
		return size, err
	}
	return size, nil
}

func (s *Storage) Delete(key string) error {

	return os.Remove(s.FullPath(key))
}

func (s *Storage) FullPath(key string) string {

	return s.Root + "/" + s.PathTransformFunc(key)
}
