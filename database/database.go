package database

import (
	"fmt"
	path "path/filepath"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

var dbfile = "data_%s.db"

// GeneralDB provides local data storage and access for various purpose.
// It implements DB interface
type GeneralDB struct {
	cfgDir    string
	ldb       *leveldb.DB
	addrStore AddrStore
}

// NewGeneralDB creates a new instance of GeneralDB
func NewGeneralDB(cfgDir string) DB {
	db := new(GeneralDB)
	db.cfgDir = cfgDir
	db.addrStore = NewAddressStore(db)
	return db
}

// Open opens the database.
// namespace is used as a suffix on the database name
func (db *GeneralDB) Open(namespace string) error {
	ldb, err := leveldb.OpenFile(path.Join(db.cfgDir, fmt.Sprintf(dbfile, namespace)), nil)
	if err != nil {
		return fmt.Errorf("failed to create database. %s", err)
	}
	db.ldb = ldb
	return nil
}

// Close closes the database
func (db *GeneralDB) Close() error {
	if db.ldb != nil {
		return db.ldb.Close()
	}
	return nil
}

// WriteBatch writes a slice of byte slices to the database.
// Length of key and value must be equal
func (db *GeneralDB) WriteBatch(key [][]byte, value [][]byte) error {

	if len(key) != len(value) {
		return fmt.Errorf("key and value slices must have equal length")
	}

	batch := new(leveldb.Batch)
	for i, k := range key {
		batch.Put(k, value[i])
	}

	return db.ldb.Write(batch, nil)
}

// GetByPrefix returns keys matching a prefix. Their key and value are returned
func (db *GeneralDB) GetByPrefix(prefix []byte) (keys [][]byte, values [][]byte) {
	iter := db.ldb.NewIterator(util.BytesPrefix(prefix), nil)
	for iter.Next() {
		key := iter.Key()
		val := iter.Value()

		_key := make([]byte, len(key))
		copy(_key, key)
		keys = append(keys, _key)

		_val := make([]byte, len(val))
		copy(_val, val)
		values = append(values, _val)
	}
	iter.Release()
	return
}

// DeleteByPrefix deletes items with the matching prefix
func (db *GeneralDB) DeleteByPrefix(prefix []byte) error {
	tx, err := db.ldb.OpenTransaction()
	if err != nil {
		return err
	}
	iter := tx.NewIterator(util.BytesPrefix(prefix), nil)
	for iter.Next() {
		key := iter.Key()
		if err := tx.Delete(key, nil); err != nil {
			tx.Discard()
			return err
		}
	}
	iter.Release()
	err = tx.Commit()
	if err != nil {
		return err
	}

	iter.Error()
	if err != nil {
		return err
	}

	return nil
}

// Address returns the address store logic
func (db *GeneralDB) Address() AddrStore {
	return db.addrStore
}
