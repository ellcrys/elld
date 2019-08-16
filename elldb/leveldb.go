// Package elldb provides LevelDB database utility.
package elldb

import (
	"fmt"
	path "path/filepath"
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

const dbfile = "data%s.db"

// LevelDB provides local data storage and functionalities for various purpose.
// It implements DB interface
type LevelDB struct {
	sync.Mutex
	dataDir string
	ldb     *leveldb.DB
}

// NewDB creates a new instance of the Ellcrys DB
func NewDB(dataDir string) *LevelDB {
	db := new(LevelDB)
	db.dataDir = dataDir
	return db
}

// Open opens the database.
// namespace is used as a suffix on the database name
func (db *LevelDB) Open(namespace string) error {

	if namespace != "" {
		namespace = "_" + namespace
	}

	o := &opt.Options{
		Filter: filter.NewBloomFilter(20),
	}

	file := path.Join(db.dataDir, fmt.Sprintf(dbfile, namespace))
	ldb, err := leveldb.OpenFile(file, o)

	// If database file is corrupted, attempt to recover
	if _, corrupted := err.(*errors.ErrCorrupted); corrupted {
		ldb, err = leveldb.RecoverFile(file, nil)
	}

	if err != nil {
		return fmt.Errorf("failed to create database. %s", err)
	}

	db.ldb = ldb
	return nil
}

// Close closes the database
func (db *LevelDB) Close() error {
	db.Lock()
	defer db.Unlock()

	if db.ldb != nil {
		return db.ldb.Close()
	}

	db.ldb = nil

	return nil
}

// Put writes many objects in one request.
func (db *LevelDB) Put(objs []*KVObject) error {

	batch := new(leveldb.Batch)
	for _, o := range objs {
		batch.Put(o.GetKey(), o.Value)
	}

	return db.ldb.Write(batch, nil)
}

// GetByPrefix returns keys matching a prefix. Their key and value are returned
func (db *LevelDB) GetByPrefix(prefix []byte) []*KVObject {
	iter := db.ldb.NewIterator(util.BytesPrefix(prefix), nil)
	return getByPrefix(iter)
}

// Iterate finds a set of objects and
// passes them to iterFunc for further processing.
// If iterFunc returns true, the iteration is discontinued.
// If first is set to true, iteration begins from the
// first item, or the last if set to false
func (db *LevelDB) Iterate(prefix []byte, first bool, iterFunc func(kv *KVObject) bool) error {
	tx, err := db.ldb.OpenTransaction()
	if err != nil {
		return err
	}
	defer tx.Commit()
	db.iterate(tx, prefix, first, iterFunc)
	return nil
}

func (db *LevelDB) iterate(ldb *leveldb.Transaction, prefix []byte, first bool, iterFunc func(kv *KVObject) bool) {
	iter := ldb.NewIterator(util.BytesPrefix(prefix), nil)
	iterate(iter, prefix, first, iterFunc)
}

// GetFirstOrLast returns one value matching a prefix.
// Set first to return the first value we find or false if the last.
func (db *LevelDB) GetFirstOrLast(prefix []byte, first bool) *KVObject {
	var result *KVObject
	var key, val []byte
	iter := db.ldb.NewIterator(util.BytesPrefix(prefix), nil)

	var f = iter.First
	if !first {
		f = iter.Last
	}

	for f() {
		key = append(key, iter.Key()...)
		val = append(val, iter.Value()...)
		result = FromKeyValue(key, val)
		key, val = []byte{}, []byte{}
		break
	}

	iter.Release()
	return result
}

// DeleteByPrefix deletes items with the matching prefix
func (db *LevelDB) DeleteByPrefix(prefix []byte) error {
	tx, err := db.ldb.OpenTransaction()
	if err != nil {
		return err
	}

	if err := deleteByPrefix(tx, prefix); err != nil {
		tx.Discard()
		return err
	}

	return tx.Commit()
}

// Truncate deletes all items
func (db *LevelDB) Truncate() error {

	tx, err := db.ldb.OpenTransaction()
	if err != nil {
		return err
	}
	defer tx.Commit()

	db.iterate(tx, nil, true, func(kv *KVObject) bool {
		if _err := tx.Delete(kv.GetKey(), nil); _err != nil {
			err = _err
			return true
		}
		return false
	})

	return err
}

// TruncateWithFunc deletes items that predicate returns true for.
// It accepts a 'prefix' to narrow the search space matching keys.
// The 'first' argument determines whether to search from top or bottom.
func (db *LevelDB) TruncateWithFunc(prefix []byte, first bool,
	predicate func(kv *KVObject) bool) error {

	if predicate == nil {
		return fmt.Errorf("predicate function is required")
	}

	tx, err := db.ldb.OpenTransaction()
	if err != nil {
		return err
	}
	defer tx.Commit()

	db.iterate(tx, prefix, first, func(kv *KVObject) bool {
		if !predicate(kv) {
			return false
		}
		if _err := tx.Delete(kv.GetKey(), nil); _err != nil {
			err = _err
			return true
		}
		return false
	})

	return err
}

// NewTx creates a new transaction
func (db *LevelDB) NewTx() (Tx, error) {
	tx, err := db.ldb.OpenTransaction()
	if err != nil {
		return nil, err
	}
	return &Transaction{
		ldb: tx,
	}, nil
}

// Transaction defines interface for working with a database transaction
type Transaction struct {
	sync.Mutex
	ldb *leveldb.Transaction
}

// Put adds a key and value
func (tx *Transaction) Put(objs []*KVObject) error {
	batch := new(leveldb.Batch)
	for _, obj := range objs {
		batch.Put(obj.GetKey(), obj.Value)
	}
	return tx.ldb.Write(batch, nil)
}

// GetByPrefix get objects by prefix
func (tx *Transaction) GetByPrefix(prefix []byte) []*KVObject {
	iter := tx.ldb.NewIterator(util.BytesPrefix(prefix), nil)
	return getByPrefix(iter)
}

// Iterate finds a set of objects by prefix and passes them ro iterFunc
// for further processing. If iterFunc returns true, the iterator is immediately released.
// If first is set to true, it begins from the first item, otherwise, the last
func (tx *Transaction) Iterate(prefix []byte, first bool, iterFunc func(kv *KVObject) bool) {
	iter := tx.ldb.NewIterator(util.BytesPrefix(prefix), nil)
	iterate(iter, prefix, first, iterFunc)
}

// Commit the transaction
func (tx *Transaction) Commit() error {
	tx.Lock()
	defer tx.Unlock()
	return tx.ldb.Commit()
}

// Discard the transaction
func (tx *Transaction) Discard() {
	tx.Lock()
	defer tx.Unlock()
	tx.ldb.Discard()
}

// Rollback discards the transaction
func (tx *Transaction) Rollback() {
	tx.Lock()
	defer tx.Unlock()
	tx.ldb.Discard()
}

// DeleteByPrefix deletes items with the matching prefix
func (tx *Transaction) DeleteByPrefix(prefix []byte) error {
	return deleteByPrefix(tx.ldb, prefix)
}

func getByPrefix(iter iterator.Iterator) []*KVObject {
	var result []*KVObject
	var key, val []byte
	for iter.Next() {
		key = append(key, iter.Key()...)
		val = append(val, iter.Value()...)
		result = append(result, FromKeyValue(key, val))
		key, val = []byte{}, []byte{}
	}
	iter.Release()
	return result
}

func iterate(iter iterator.Iterator, prefix []byte, first bool, iterFunc func(kv *KVObject) bool) {
	var key, val []byte

	var f, f2 = iter.First, iter.Next
	if !first {
		f, f2 = iter.Last, iter.Prev
	}

	for f() {
		key = append(key, iter.Key()...)
		val = append(val, iter.Value()...)
		if iterFunc(FromKeyValue(key, val)) {
			break
		}
		key, val = []byte{}, []byte{}
		f = f2
	}

	iter.Release()
}

func deleteByPrefix(tx *leveldb.Transaction, prefix []byte) error {
	iter := tx.NewIterator(util.BytesPrefix(prefix), nil)
	for iter.Next() {
		key := iter.Key()
		if err := tx.Delete(key, nil); err != nil {
			return err
		}
	}
	iter.Release()
	return iter.Error()
}
