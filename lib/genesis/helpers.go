// Copyright 2019 ChainSafe Systems (ON) Corp.
// This file is part of gossamer.
//
// The gossamer library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The gossamer library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the gossamer library. If not, see <http://www.gnu.org/licenses/>.

package genesis

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/ChainSafe/gossamer/dot/types"
	"github.com/ChainSafe/gossamer/lib/common"
	"github.com/ChainSafe/gossamer/lib/crypto"
	"github.com/ChainSafe/gossamer/lib/scale"
	"github.com/ChainSafe/gossamer/lib/trie"
	"github.com/OneOfOne/xxhash"
)

// NewGenesisFromJSON parses a JSON formatted genesis file
func NewGenesisFromJSON(file string) (*Genesis, error) {
	fp, err := filepath.Abs(file)
	if err != nil {
		return nil, err
	}
fmt.Printf("READ GENESIS FILE %v\n", file)
	data, err := ioutil.ReadFile(filepath.Clean(fp))
	if err != nil {
		return nil, err
	}

	g := new(Genesis)
	err = json.Unmarshal(data, g)
	return g, err
}

// NewGenesisFromJSONHR parses Human Readable JSON formatted genesis file
func NewGenesisFromJSONHR(file string) (*Genesis, error) {
	fp, err := filepath.Abs(file)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadFile(filepath.Clean(fp))
	if err != nil {
		return nil, err
	}

	g := new(Genesis)

	err = json.Unmarshal(data, g)

	grt := g.Genesis.Runtime
	res := buildRawMap(grt)
fmt.Printf("res %v\n", len(res))
	//g.Genesis.Raw = [2]map[string]string
	// todo convert interface (res) to string
	//g.Genesis.Raw[0] = res

	return g, err
}

// keyValue struct to hold data regarding entry
type keyValue struct {
	key      []string
	value    string
	valueLen *big.Int
}

func buildRawMap(m map[string]map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	for k, v := range m {
		kv := new(keyValue)
		kv.key = append(kv.key, k)
		buildRawMapInterface(v, kv)

		key := formatKey(kv.key)

		value, err := formatValue(kv)
		if err != nil {
			// todo determine how to handle error
		}
		res[key] = value
	}
	return res
}

func buildRawMapInterface(m map[string]interface{}, kv *keyValue) {
	for k, v := range m {
		kv.key = append(kv.key, k)
		switch v2 := v.(type) {
		case []interface{}:
			kv.valueLen = big.NewInt(int64(len(v2)))
			buildRawArrayInterface(v2, kv)
		case string:
			kv.value = v2
		}
	}
}

func buildRawArrayInterface(a []interface{}, kv *keyValue) {
	for _, v := range a {
		switch v2 := v.(type) {
		case []interface{}:
			buildRawArrayInterface(v2, kv)
		case string:
			// todo check to confirm it's an address
			tba := crypto.PublicAddressToByteArray(common.Address(v2))
			fmt.Printf("KEY VAL %x\n", tba)
			kv.value = kv.value + fmt.Sprintf("%x", tba)
		case float64:
			encVal, err := scale.Encode(uint64(v2))
			if err != nil {
				//todo determine how to handle this error
			}
			kv.value = kv.value + fmt.Sprintf("%x", encVal)
		}
	}
}

func formatKey(key []string) string {
	switch true {
	case reflect.DeepEqual([]string{"grandpa", "authorities"}, key):
		kb := []byte(`:grandpa_authorities`)
		return common.BytesToHex(kb)
	case reflect.DeepEqual([]string{"system", "code"}, key):
		kb := []byte(`:code`)
		return common.BytesToHex(kb)
	default:
		var fKey string
		for _, v := range key {
			fKey = fKey + v + " "
		}
		fKey = strings.Trim(fKey, " ")
		fKey = strings.Title(fKey)
		kb := twoxHash([]byte(fKey))
		return common.BytesToHex(kb)
	}
}

func formatValue(kv *keyValue) (string, error) {
	switch true {
	case reflect.DeepEqual([]string{"grandpa", "authorities"}, kv.key):
		if kv.valueLen != nil {
			lenEnc, err := scale.Encode(kv.valueLen)
			if err != nil {
				return "", err
			}
			// prepend 01 to grandpa_authorities values
			return fmt.Sprintf("0x01%x%v", lenEnc, kv.value), nil
		}
		return "", fmt.Errorf("error formatting value for grandpa authorities")
	case reflect.DeepEqual([]string{"system", "code"}, kv.key):
		return kv.value, nil
	default:
		if kv.valueLen != nil {
			lenEnc, err := scale.Encode(kv.valueLen)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("0x%x%v", lenEnc, kv.value), nil
		}
		return fmt.Sprintf("0x%x", kv.value), nil
	}
}

// NewTrieFromGenesis creates a new trie from the raw genesis data
func NewTrieFromGenesis(g *Genesis) (*trie.Trie, error) {
	t := trie.NewEmptyTrie()

	r := g.GenesisFields().Raw[0]

	err := t.Load(r)
	if err != nil {
		return nil, fmt.Errorf("failed to create trie from genesis: %s", err)
	}

	return t, nil
}

// NewGenesisBlockFromTrie creates a genesis block from the provided trie
func NewGenesisBlockFromTrie(t *trie.Trie) (*types.Header, error) {

	// create state root from trie hash
	stateRoot, err := t.Hash()
	if err != nil {
		return nil, fmt.Errorf("failed to create state root from trie hash: %s", err)
	}

	// create genesis block header
	header, err := types.NewHeader(
		common.NewHash([]byte{0}), // parentHash
		big.NewInt(0),             // number
		stateRoot,                 // stateRoot
		trie.EmptyHash,            // extrinsicsRoot
		[][]byte{},                // digest
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create genesis block header: %s", err)
	}

	return header, nil
}

func twoxHash(msg []byte) []byte {
	// compute xxHash64 twice with seeds 0 and 1 applied on given byte array
	h0 := xxhash.NewS64(0) // create xxHash with 0 seed
	_, err := h0.Write(msg[0:])
	if err != nil {
		return nil
	}
	res0 := h0.Sum64()
	hash0 := make([]byte, 8)
	binary.LittleEndian.PutUint64(hash0, res0)

	h1 := xxhash.NewS64(1) // create xxHash with 1 seed
	_, err = h1.Write(msg[0:])
	if err != nil {
		return nil
	}
	res1 := h1.Sum64()
	hash1 := make([]byte, 8)
	binary.LittleEndian.PutUint64(hash1, res1)

	//concatenated result
	both := append(hash0, hash1...)
	return both
}
