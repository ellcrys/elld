package database

import (
	"fmt"
	path "path/filepath"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

const dbfile = "data%s.db"

// LevelDB provides local data storage and functionalities for various purpose.
// It implements DB interface
type LevelDB struct {
	cfgDir string
	ldb    *leveldb.DB
}

// NewLevelDB creates a new instance of LevelDB
func NewLevelDB(cfgDir string) *LevelDB {
	db := new(LevelDB)
	db.cfgDir = cfgDir
	return db
}

// Open opens the database.
// namespace is used as a suffix on the database name
func (db *LevelDB) Open(namespace string) error {
	if namespace != "" {
		namespace = "_" + namespace
	}
	ldb, err := leveldb.OpenFile(path.Join(db.cfgDir, fmt.Sprintf(dbfile, namespace)), nil)
	if err != nil {
		return fmt.Errorf("failed to create database. %s", err)
	}
	db.ldb = ldb
	return nil
}

// Close closes the database
func (db *LevelDB) Close() error {
	if db.ldb != nil {
		return db.ldb.Close()
	}
	return nil
}

// WriteBatch writes many objects in one request.
func (db *LevelDB) WriteBatch(objs []*KVObject) error {

	batch := new(leveldb.Batch)
	for _, o := range objs {
		batch.Put(o.GetKey(), o.Value)
	}

	return db.ldb.Write(batch, nil)
}

// GetByPrefix returns keys matching a prefix. Their key and value are returned
func (db *LevelDB) GetByPrefix(prefix []byte) []*KVObject {
	var result []*KVObject
	var key, val []byte
	iter := db.ldb.NewIterator(util.BytesPrefix(prefix), nil)
	for iter.Next() {
		key = append(key, iter.Key()...)
		val = append(val, iter.Value()...)
		result = append(result, FromKeyValue(key, val))
		key, val = []byte{}, []byte{}
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
