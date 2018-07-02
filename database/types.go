package database

import (
	"fmt"
	"strings"
)

// PrefixSeparator is the separator that separates multiple prefixes
const PrefixSeparator = "~"

// KeyPrefixSeparator is the separator that separates the prefix and the key
const KeyPrefixSeparator = ":"

// KVObject represents an item in the database
type KVObject struct {
	Key      []byte
	Value    []byte
	Prefixes []string
}

// MakePrefix creates a prefix string
func MakePrefix(prefixes []string) []byte {
	return []byte(strings.Join(prefixes, PrefixSeparator))
}

// MakeKey construct a key from the object key and a slice of prefixes
func MakeKey(key []byte, prefixes []string) []byte {
	kpSep := KeyPrefixSeparator

	var prefix []byte
	if len(prefixes) > 0 {
		prefix = MakePrefix(prefixes)
	} else {
		kpSep = ""
	}

	return []byte(fmt.Sprintf("%s%s%s", prefix, kpSep, key))
}

// GetKey creates and returns the object key which is combined with the prefixes
func (o *KVObject) GetKey() []byte {
	return MakeKey(o.Key, o.Prefixes)
}

// NewKVObject creates a key value object.
// The prefixes provided is joined together and prepended to the key before insertion.
func NewKVObject(key, value []byte, prefixes ...string) *KVObject {
	return &KVObject{Key: key, Value: value, Prefixes: prefixes}
}

// FromKeyValue takes a key and creates a KVObject
func FromKeyValue(key []byte, value []byte) *KVObject {
	var prefixes []string
	var _key string

	// break down the key to determine the prefixes and the original key
	parts := strings.Split(string(key), KeyPrefixSeparator)
	if len(parts) > 2 {
		panic("invalid key format")
	} else if len(parts) == 2 {
		prefixes = strings.Split(parts[0], PrefixSeparator)
		_key = parts[1]
	} else {
		_key = parts[0]
	}

	return &KVObject{
		Key:      []byte(_key),
		Value:    value,
		Prefixes: prefixes,
	}
}

// Tx represents a database transaction instance
type Tx interface {

	// Put puts one or more objects
	Put([]*KVObject) error

	// GetByPrefix gets objects by prefix
	GetByPrefix([]byte) (result []*KVObject)

	// Commit commits the transaction
	Commit() error

	// Rollback roles back the transaction
	Rollback()
}

// DB describes the database access, model and available functionalities
type DB interface {

	// Open opens the database
	Open(namespace string) error

	// Close closes the database
	Close() error

	// Put writes many objects to the database in one atomic request
	Put([]*KVObject) error

	// GetByPrefix returns keys matching a prefix
	GetByPrefix([]byte) (result []*KVObject)

	// DeleteByPrefix deletes one or many records by prefix
	DeleteByPrefix([]byte) error

	// NewTx creates a transaction
	NewTx() (Tx, error)
}
