package blockchain

import (
	"context"
	"fmt"

	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/go-merkle-tree"
)

// ByteVal represents a byte value from an HashTree
type ByteVal string

// Construct returns self/this
func (s ByteVal) Construct() interface{} {
	return s
}

// HashTree uses an underlying merkle tree for the purpose of
// storing state object hashes. We use the HashTree to track object
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
	store   common.Store
}

// NewHashTree creates a HashTree instance which includes an initialized
// merkle tree for the chainID specified. Merkle tree is initialized with a
// TreeStorage engine allow access to previously stored tree elements.
func NewHashTree(chainID string, store common.Store) *HashTree {
	ht := new(HashTree)
	ht.chainID = chainID
	ht.treeStore = NewTreeStorage(chainID, store)
	config := merkleTree.NewConfig(merkleTree.SHA512Hasher{}, 256, 512, new(ByteVal))
	ht.tree = merkleTree.NewTree(ht.treeStore, config)
	return ht
}

// NewMemHashTree is like NewMemHashTree but it creates a HashTree instance backed by
// a memory storage engine. It also supports setting an initial root hash and node so
// that we can build off from a previous tree state.
func NewMemHashTree(initialRootHash []byte, initialRootNode []byte) *HashTree {
	ht := new(HashTree)
	ht.treeStore = NewMemEngine(initialRootHash, initialRootNode)
	config := merkleTree.NewConfig(merkleTree.SHA512Hasher{}, 256, 512, new(ByteVal))
	ht.tree = merkleTree.NewTree(ht.treeStore, config)
	return ht
}

// Root returns the root of the tree as well as the serialized root node.
func (ht *HashTree) Root() (key []byte, node []byte, err error) {
	key, err = ht.treeStore.LookupRoot(context.TODO())
	if err != nil {
		return nil, nil, err
	}
	node, err = ht.treeStore.LookupNode(context.TODO(), key)
	if err != nil {
		return nil, nil, err
	}
	return key, node, nil
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
