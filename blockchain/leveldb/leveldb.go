package leveldb

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/database"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/wire"
)

// MetadataKey is the key used to store the metadata
const MetadataKey = "meta"

// Store represents a store that implements the Store
// interface meant to be used for persisting and retrieving
// blockchain data.
type Store struct {
	db        database.DB
	namespace string
}

// New creates an instance of the store
func New(db database.DB) (*Store, error) {
	return &Store{
		db: db,
	}, nil
}

// hasBlock checks whether a block exists
func (s *Store) hasBlock(tx database.Tx, chainID string, number uint64) (bool, error) {
	var block wire.Block
	if err := s.getBlock(tx, chainID, number, &block); err != nil {
		if err == common.ErrBlockNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// getBlock gets a block by the block number.
// If number is 0, return the block with the highest block number
func (s *Store) getBlock(tx database.Tx, chainID string, number uint64, result interface{}) error {

	if number == 0 {
		kvO := s.db.GetFirstOrLast(database.MakePrefix([]string{"block", chainID, "number"}), false)
		if kvO == nil {
			return common.ErrBlockNotFound
		}
		return json.Unmarshal(kvO.Value, result)
	}

	key := database.MakeKey([]byte(fmt.Sprintf("%d", number)), []string{"block", chainID, "number"})
	r := tx.GetByPrefix(key)
	if len(r) == 0 {
		return common.ErrBlockNotFound
	}

	return json.Unmarshal(r[0].Value, result)
}

// getBlockNumberByHash gets a block's block number by its hash
func (s *Store) getBlockNumberByHash(tx database.Tx, chainID string, hash string, result *uint64) error {

	key := database.MakeKey([]byte(fmt.Sprintf("%s", hash)), []string{"block", chainID, "hash"})
	r := tx.GetByPrefix(key)
	if len(r) == 0 {
		return common.ErrBlockNotFound
	}

	blockNum, err := strconv.ParseUint(string(r[0].Value), 10, 64)
	if err != nil {
		return fmt.Errorf("failed to convert block number from string to uint64")
	}

	*result = blockNum

	return nil
}

// GetBlockHeader the header of the current block
func (s *Store) GetBlockHeader(chainID string, number uint64, header *wire.Header) error {

	tx, err := s.db.NewTx()
	if err != nil {
		return fmt.Errorf("failed to create transaction")
	}

	var block wire.Block
	if err := s.getBlock(tx, chainID, number, &block); err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	*header = *block.Header

	return nil
}

// GetBlockHeaderByHash returns the header of a block by searching using its hash
func (s *Store) GetBlockHeaderByHash(chainID string, hash string, header *wire.Header) error {

	var block wire.Block
	if err := s.GetBlockByHash(chainID, hash, &block); err != nil {
		return err
	}

	*header = *block.Header
	return nil
}

// GetBlock fetches a block by its block number.
// If the block number begins with -1, the block with the highest block number is returned.
func (s *Store) GetBlock(chainID string, number uint64, result common.Block) error {

	tx, err := s.db.NewTx()
	if err != nil {
		return fmt.Errorf("failed to create transaction")
	}

	if err := s.getBlock(tx, chainID, number, result); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// GetBlockWithTx is like GetBlock but it accepts a transaction object
func (s *Store) GetBlockWithTx(tx database.Tx, chainID string, number uint64, result common.Block) error {
	return s.getBlock(tx, chainID, number, result)
}

// GetBlockByHash fetches a block by its block hash.
func (s *Store) GetBlockByHash(chainID string, hash string, result common.Block) error {

	tx, err := s.db.NewTx()
	if err != nil {
		return fmt.Errorf("failed to create transaction")
	}

	var blockNum uint64
	if err := s.getBlockNumberByHash(tx, chainID, hash, &blockNum); err != nil {
		tx.Rollback()
		return err
	}

	if err := s.getBlock(tx, chainID, blockNum, result); err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()

	return nil
}

// PutBlock adds a block to the store.
// Returns error if a block with same number exists.
func (s *Store) PutBlock(chainID string, block common.Block) error {

	tx, err := s.db.NewTx()
	if err != nil {
		return fmt.Errorf("failed to create transaction")
	}

	if err := s.putBlock(tx, chainID, block); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// putBlock adds a block to the store using the provided transaction object.
// Returns error if a block with same number exists.
func (s *Store) putBlock(tx database.Tx, chainID string, block common.Block) error {

	// check if block already exists. return nil if block exists.
	hasBlock, err := s.hasBlock(tx, chainID, block.GetNumber())
	if err != nil {
		return fmt.Errorf("failed to check block existence: %s", err)
	} else if hasBlock {
		return nil
	}

	value := util.ObjectToBytes(block)

	// store the block with a key format that allows
	// for query using the block number
	key := common.MakeBlockKey(chainID, block.GetNumber())
	blockObj := database.NewKVObject(key, value)
	if err := tx.Put([]*database.KVObject{blockObj}); err != nil {
		return fmt.Errorf("failed to put block: %s", err)
	}

	// also index the block with a hash key to allow query
	// by block hash. But, do not store the full block, just the number.
	key = common.MakeBlockHashKey(chainID, block.GetHash())
	value = []byte(fmt.Sprintf("%d", block.GetNumber()))
	blockObj = database.NewKVObject(key, value)
	if err := tx.Put([]*database.KVObject{blockObj}); err != nil {
		return fmt.Errorf("failed to index block by hash: %s", err)
	}

	return nil
}

// PutBlockWithTx is like PutBlock but accepts a transaction
func (s *Store) PutBlockWithTx(tx database.Tx, chainID string, block common.Block) error {
	return s.putBlock(tx, chainID, block)
}

// Put stores an object
func (s *Store) Put(key []byte, value []byte) error {
	obj := database.NewKVObject(key, value)
	if err := s.db.Put([]*database.KVObject{obj}); err != nil {
		return fmt.Errorf("failed to put object: %s", err)
	}
	return nil
}

// PutWithTx is like Put but accepts a transaction object.
func (s *Store) PutWithTx(tx database.Tx, key []byte, value []byte) error {
	obj := database.NewKVObject(key, value)
	if err := tx.Put([]*database.KVObject{obj}); err != nil {
		return fmt.Errorf("failed to put object: %s", err)
	}
	return nil
}

// Get an object by key (and optionally by prefixes)
func (s *Store) Get(key []byte, result *[]*database.KVObject) {
	r := s.db.GetByPrefix(key)
	if r == nil {
		return
	}
	*result = append(*result, r...)
}

// GetFirstOrLast returns the first or last object matching the key.
// Set first to true to return the first or false for last.
func (s *Store) GetFirstOrLast(first bool, key []byte, result *database.KVObject) {
	r := s.db.GetFirstOrLast(key, first)
	if r == nil {
		return
	}
	*result = *r
}

// NewTx creates and returns a transaction
func (s *Store) NewTx() database.Tx {

	tx, err := s.db.NewTx()
	if err != nil {
		return nil
	}

	return tx
}
