package state

import (
	"bytes"
	"testing"

	"github.com/ChainSafe/gossamer/lib/common"
	"github.com/ChainSafe/gossamer/lib/trie"

	database "github.com/ChainSafe/chaindb"
)

func newTestStorageState(t *testing.T) *StorageState {
	db := database.NewMemDatabase()

	s, err := NewStorageState(db, trie.NewEmptyTrie())
	if err != nil {
		t.Fatal(err)
	}

	return s
}

func TestLoadCodeHash(t *testing.T) {
	storage := newTestStorageState(t)
	testCode := []byte("asdf")

	err := storage.SetStorage(codeKey, testCode)
	if err != nil {
		t.Fatal(err)
	}

	resCode, err := storage.LoadCode()
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(resCode, testCode) {
		t.Fatalf("Fail: got %s expected %s", resCode, testCode)
	}

	resHash, err := storage.LoadCodeHash()
	if err != nil {
		t.Fatal(err)
	}

	expectedHash, err := common.HexToHash("0xb91349ff7c99c3ae3379dd49c2f3208e202c95c0aac5f97bb24ded899e9a2e83")
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(resHash[:], expectedHash[:]) {
		t.Fatalf("Fail: got %s expected %s", resHash, expectedHash)
	}
}

func TestSetAndGetBalance(t *testing.T) {
	storage := newTestStorageState(t)

	key := [32]byte{1, 2, 3, 4, 5, 6, 7}
	bal := uint64(99)

	err := storage.SetBalance(key, bal)
	if err != nil {
		t.Fatal(err)
	}

	res, err := storage.GetBalance(key)
	if err != nil {
		t.Fatal(err)
	}

	if res != bal {
		t.Fatalf("Fail: got %d expected %d", res, bal)
	}
}
