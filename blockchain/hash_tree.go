package blockchain

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"fmt"

	"github.com/ellcrys/elld/blockchain/types"
	"github.com/ellcrys/elld/database"
	"github.com/ellcrys/go-merkle-tree"
)

// ByteVal represents a byte value from an HashTree
type ByteVal string

// Construct returns self/this
func (s ByteVal) Construct() interface{} {
	return s
}

// HashTree uses an underlying merkle tree for the purpose of
// storing object hashes. We use the HashTree to track object
// changes and to verify stateRoot of blocks. The actual
// object values are stored elsewhere.
type HashTree struct {
	chainID   string
	tree      *merkleTree.Tree
	treeStore merkleTree.StorageEngine
}

// TreeStorage implements merkleTree.StorageEngine to provide
// storage mechanism for a merkle tree.
type TreeStorage struct {
	chainID string
	store   types.Store
}

// NewHashTree creates a HashTree instance which includes an initialized
// merkle tree for the chainID specified. Merkle tree is initialized with a
// TreeStorage engine allow access to previously stored tree elements.
func NewHashTree(chainID string, store types.Store) *HashTree {
	ht := new(HashTree)
	ht.chainID = chainID
	ht.treeStore = NewTreeStorage(chainID, store)
	config := merkleTree.NewConfig(merkleTree.SHA512Hasher{}, 256, 512, new(ByteVal))
	ht.tree = merkleTree.NewTree(ht.treeStore, config)
	return ht
}

// Root returns the root of the tree
func (ht *HashTree) Root() ([]byte, error) {
	return ht.treeStore.LookupRoot(context.TODO())
}

// Upsert updates a node on the tree or creates a new node if no matching node is found
func (ht *HashTree) Upsert(key, value []byte) error {
	kvp := merkleTree.KeyValuePair{Key: key, Value: value}
	return ht.tree.Upsert(context.TODO(), kvp, nil)
}

// Find finds a node in the tree by key
func (ht *HashTree) Find(key []byte) (interface{}, error) {
	v, _, err := ht.tree.Find(context.TODO(), key)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, fmt.Errorf("not found")
	}
	return v, nil
}

// NewTreeStorage creates an instance of TreeStorage
func NewTreeStorage(chainID string, store types.Store) *TreeStorage {
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

	var result []database.KVObject

	if err := ts.store.Get(nodeKey, &result); err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("root not found")
	}

	return result[0].Value, nil
}

// LookupRoot fetches the root of the in-memory tree back out
func (ts *TreeStorage) LookupRoot(_ context.Context) (merkleTree.Hash, error) {
	rootKey := database.MakePrefix([]string{"tree", "root", ts.chainID})

	var result []database.KVObject

	if err := ts.store.Get(rootKey, &result); err != nil {
		return nil, err
	}
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
