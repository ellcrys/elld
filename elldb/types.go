package elldb

import (
	"bytes"

	"github.com/ellcrys/mother/util"
)

const (
	// KeyPrefixSeparator is used to separate prefix and key
	KeyPrefixSeparator = "@@"
	prefixSeparator    = ":"
)

// KVObject represents an item in the elldb
type KVObject struct {
	Key    []byte `json:"key"`
	Value  []byte `json:"value"`
	Prefix []byte `json:"prefix"`
}

// IsEmpty checks whether the object is empty
func (kv *KVObject) IsEmpty() bool { // TODO: test
	return len(kv.Key) == 0 && len(kv.Value) == 0
}

// Scan marshals the value into dest
func (kv *KVObject) Scan(dest interface{}) error {
	return util.BytesToObject(kv.Value, &dest)
}

// MakePrefix creates a prefix string
func MakePrefix(prefixes ...[]byte) (result []byte) {
	return bytes.Join(prefixes, []byte(prefixSeparator))
}

// MakeKey construct a key from the key and prefixes
func MakeKey(key []byte, prefixes ...[]byte) []byte {
	var prefix = MakePrefix(prefixes...)
	var sep = []byte(KeyPrefixSeparator)
	if len(key) == 0 || len(prefix) == 0 {
		sep = []byte{}
	}
	return append(prefix, append(sep, key...)...)
}

// GetKey creates and returns the key
func (kv *KVObject) GetKey() []byte {
	return MakeKey(kv.Key, kv.Prefix)
}

// Equal performs equality check with another KVObject
func (kv *KVObject) Equal(other *KVObject) bool {
	return bytes.Equal(kv.Key, other.Key) && bytes.Equal(kv.Value, other.Value)
}

// NewKVObject creates a key value object.
// The prefixes provided is joined together and prepended to the key before insertion.
func NewKVObject(key, value []byte, prefixes ...[]byte) *KVObject {
	return &KVObject{Key: key, Value: value, Prefix: MakePrefix(prefixes...)}
}

// FromKeyValue takes a key and creates a KVObject
func FromKeyValue(key []byte, value []byte) *KVObject {

	var k, p []byte

	// break down the key to determine the prefix and the original key.
	parts := bytes.SplitN(key, []byte(KeyPrefixSeparator), 2)

	// If there are more than 2 parts, it is an invalid key.
	// If there are only two parts, then the 0 index is the prefix
	// while the 1 index is the key.
	// It there are only one part, the 0 part is considered the key.
	partsLen := len(parts)
	if partsLen > 2 {
		panic("invalid key format: " + string(key))
	} else if partsLen == 2 {
		k = parts[1]
		p = parts[0]
	} else if partsLen == 1 {
		k = parts[0]
	}

	return &KVObject{
		Key:    k,
		Value:  value,
		Prefix: p,
	}
}

// Tx represents a database transaction instance
type Tx interface {

	// Put puts one or more objects
	Put([]*KVObject) error

	// GetByPrefix gets objects by prefix
	GetByPrefix([]byte) (result []*KVObject)

	// Iterate finds a set of objects by prefix and passes them to iterFunc
	// for further processing. If iterFunc returns true, the iterator is immediately released.
	// If first is set to true, it begins from the first item, otherwise, the last
	Iterate(prefix []byte, first bool, iterFunc func(kv *KVObject) bool)

	// DeleteByPrefix deletes one or many records by prefix
	DeleteByPrefix([]byte) error

	// Commit commits the transaction
	Commit() error

	// Rollback roles back the transaction
	Rollback()

	// Discard the transaction. Do not call
	// functions in the transaction after this.
	Discard()
}

// DB describes the database access, model and available functionalities
type DB interface {

	// Open opens the database
	Open(namespace string) error

	// Close closes the database
	Close() error

	// Put writes many objects to the database in one atomic request
	Put([]*KVObject) error

	// GetByPrefix returns valyes matching a prefix
	GetByPrefix([]byte) (result []*KVObject)

	// GetFirstOrLast returns one value matching a prefix.
	// Set first to return the first value we find or false if the last.
	GetFirstOrLast(prefix []byte, first bool) *KVObject

	// Iterate finds a set of objects by prefix and passes them ro iterFunc
	// for further processing. If iterFunc returns true, the iterator is immediately released.
	// If first is set to true, it begins from the first item, otherwise, the last
	Iterate(prefix []byte, first bool, iterFunc func(kv *KVObject) bool) error

	// DeleteByPrefix deletes one or many records by prefix
	DeleteByPrefix([]byte) error

	// Truncate removes all items
	Truncate() error

	// NewTx creates a transaction
	NewTx() (Tx, error)
}

// TxCreator defines an interface for creating database transaction
type TxCreator interface {
	// NewTx creates a transaction
	NewTx() (Tx, error)
}
