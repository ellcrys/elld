package common

import (
	"bytes"

	"github.com/cbergoon/merkletree"
	"github.com/ellcrys/elld/util"
	"golang.org/x/crypto/blake2b"
)

// TreeItem represents an item to be added to the tree
type TreeItem []byte

// CalculateHash returns the blake2b-256 hash of the item
func (i TreeItem) CalculateHash() []byte {
	h, _ := blake2b.New256(nil)
	h.Write(i)
	return h.Sum(nil)
}

// Equals checks whether the item equal another item
func (i TreeItem) Equals(other merkletree.Content) bool {
	return bytes.Equal(i, other.(TreeItem))
}

// Tree provides merkle tree functionality with
// the ability to persist the tree content to a storage.
type Tree struct {
	// tree is the internal tree
	tree *merkletree.MerkleTree

	// items represents the items to be included in the tree
	items []merkletree.Content
}

// NewTree creates an instance of Tree and accepts storage
// module that implements store.Storer.
func NewTree() *Tree {
	return &Tree{}
}

// Add adds an item to collection of items to be used to build the tree
func (t *Tree) Add(item merkletree.Content) {
	t.items = append(t.items, item)
}

// GetItems returns the tree items
// Needs tests
func (t *Tree) GetItems() []merkletree.Content {
	return t.items
}

// Build the tree from a slice of item
func (t *Tree) Build() error {

	// create a new tree
	newTree, err := merkletree.NewTree(t.items)
	if err != nil {
		return err
	}

	t.tree = newTree
	return nil
}

// Root returns the root of the tree
func (t *Tree) Root() util.Hash {
	if t.tree == nil {
		return util.EmptyHash
	}
	return util.BytesToHash(t.tree.MerkleRoot())
}
