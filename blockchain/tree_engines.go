package blockchain

import (
	"crypto/sha512"
	"encoding/hex"

	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/database"
	"github.com/ellcrys/go-merkle-tree"

	"golang.org/x/net/context"
)

// MemEngine is an in-memory MerkleTree engine, used now mainly for testing.
type MemEngine struct {
	root  merkleTree.Hash
	nodes map[string][]byte
}

// NewMemEngine makes an in-memory storage engine and sets its root.
// We use this to compute new state root after transaction state
// are added to the tree.
func NewMemEngine(rootHash merkleTree.Hash, rootNode []byte) *MemEngine {
	nodes := make(map[string][]byte)
	nodes[hex.EncodeToString(rootHash)] = rootNode
	return &MemEngine{
		root:  rootHash,
		nodes: nodes,
	}
}

// CommitRoot "commits" the root ot the blessed memory slot
func (m *MemEngine) CommitRoot(_ context.Context, prev merkleTree.Hash, curr merkleTree.Hash, txinfo merkleTree.TxInfo) error {
	m.root = curr
	return nil
}

// Hash runs SHA512
func (m *MemEngine) Hash(_ context.Context, d []byte) merkleTree.Hash {
	sum := sha512.Sum512(d)
	return merkleTree.Hash(sum[:])
}

// LookupNode looks up a MerkleTree node by hash
func (m *MemEngine) LookupNode(_ context.Context, h merkleTree.Hash) (b []byte, err error) {
	return m.nodes[hex.EncodeToString(h)], nil
}

// LookupRoot fetches the root of the in-memory tree back out
func (m *MemEngine) LookupRoot(_ context.Context) (merkleTree.Hash, error) {
	return m.root, nil
}

// StoreNode stores the given node under the given key.
func (m *MemEngine) StoreNode(_ context.Context, key merkleTree.Hash, b []byte) error {
	m.nodes[hex.EncodeToString(key)] = b
	return nil
}

// NewTreeStorage creates an instance of TreeStorage
func NewTreeStorage(chainID string, store common.Store) *TreeStorage {
	return &TreeStorage{
		store: store,
	}
}

// CommitRoot "commits" the root ot the blessed memory slot
func (ts *TreeStorage) CommitRoot(_ context.Context, prev merkleTree.Hash, curr merkleTree.Hash, txinfo merkleTree.TxInfo) error {
	rootKey := database.MakePrefix([]string{"tree", "root", ts.chainID})
	return ts.store.Put(rootKey, curr)
}

// Hash runs SHA512
func (ts *TreeStorage) Hash(_ context.Context, d []byte) merkleTree.Hash {
	sum := sha512.Sum512(d)
	return merkleTree.Hash(sum[:])
}

// LookupNode looks up a MerkleTree node by hash
func (ts *TreeStorage) LookupNode(_ context.Context, h merkleTree.Hash) (b []byte, err error) {

	nodeKey := database.MakePrefix([]string{"tree", "node", ts.chainID, hex.EncodeToString(h)})
	var result []*database.KVObject
	ts.store.Get(nodeKey, &result)
	if len(result) == 0 {
		return nil, nil
	}

	return result[0].Value, nil
}

// LookupRoot fetches the root of the in-memory tree back out
func (ts *TreeStorage) LookupRoot(_ context.Context) (merkleTree.Hash, error) {

	rootKey := database.MakePrefix([]string{"tree", "root", ts.chainID})
	var result []*database.KVObject
	ts.store.Get(rootKey, &result)
	if len(result) == 0 {
		return nil, nil
	}

	return result[0].Value, nil
}

// StoreNode stores the given node under the given key.
func (ts *TreeStorage) StoreNode(_ context.Context, key merkleTree.Hash, b []byte) error {
	nodeKey := database.MakePrefix([]string{"tree", "node", ts.chainID, hex.EncodeToString(key)})
	return ts.store.Put(nodeKey, b)
}
