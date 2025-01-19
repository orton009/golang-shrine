package p2p

import (
	"bytes"
	"io"
	"testing"
)

func TestStore(t *testing.T) {

	key := "sample"
	data := "Welcome to the wasteland!"

	storage := NewStorage(StorageOpts{Root: "./testData"})

	r := bytes.NewReader([]byte(data))

	if err := storage.Write(key, r); err != nil {
		t.Error(err)
	}

	reader, err := storage.Read(key)
	if err != nil {
		t.Error(err)
	}

	bytes, err := io.ReadAll(reader)
	if err != nil {
		t.Error(err)
	}

	if data != string(bytes) {
		t.Errorf("Data mismatch! expected: %v \t got: %v", data, string(bytes))
	}

	err = storage.Delete(key)
	if err != nil {
		t.Error(err)
	}

}
