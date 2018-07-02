package leveldb

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ellcrys/elld/blockchain/types"
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
func (s *Store) hasBlock(number uint64) (bool, error) {
	var block wire.Block
	if err := s.getBlock(number, &block); err != nil {
		if err == types.ErrBlockNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// getBlock gets a block by the block number.
// If number is 0, return the current block.
func (s *Store) getBlock(number uint64, result interface{}) error {

	if number == 0 {
		var meta types.Meta
		if err := s.GetMetadata(&meta); err != nil {
			return fmt.Errorf("failed to get meta: %s", err)
		}
		number = meta.CurrentBlockNumber
	}

	key := database.MakeKey([]byte(fmt.Sprintf("%d", number)), []string{"block", "number"})
	r := s.db.GetByPrefix(key)
	if len(r) == 0 {
		return types.ErrBlockNotFound
	}

	return json.Unmarshal(r[0].Value, result)
}

// getBlockNumberByHash gets a block's block number by its hash
func (s *Store) getBlockNumberByHash(hash string, result *uint64) error {

	key := database.MakeKey([]byte(fmt.Sprintf("%s", hash)), []string{"block", "hash"})
	r := s.db.GetByPrefix(key)
	if len(r) == 0 {
		return types.ErrBlockNotFound
	}

	blockNum, err := strconv.ParseUint(string(r[0].Value), 10, 64)
	if err != nil {
		return fmt.Errorf("failed to convert block number from string to uint64")
	}

	*result = blockNum

	return nil
}

// GetBlockHeader the header of the current block
func (s *Store) GetBlockHeader(number uint64, header *wire.Header) error {
	var block wire.Block
	if err := s.getBlock(number, &block); err != nil {
		return err
	}

	*header = *block.Header
	return nil
}

// GetBlock fetches a block by its block number.
// If the block number begins with -1, the block with the highest block number is returned.
func (s *Store) GetBlock(number uint64, result types.Block) error {
	return s.getBlock(number, result)
}

// GetBlockByHash fetches a block by its block hash.
func (s *Store) GetBlockByHash(hash string, result types.Block) error {
	var blockNum uint64
	if err := s.getBlockNumberByHash(hash, &blockNum); err != nil {
		return err
	}

	return s.GetBlock(blockNum, result)
}

// PutBlock adds a block to the store.
// Returns error if a block with same number exists.
func (s *Store) PutBlock(block types.Block) error {

	// check if block already exists.
	// return nil if block exists.
	hasBlock, err := s.hasBlock(block.GetNumber())
	if err != nil {
		return fmt.Errorf("failed to check block existence: %s", err)
	} else if hasBlock {
		return nil
	}

	value := util.ObjectToBytes(block)

	// create a transaction
	tx, err := s.db.NewTx()
	if err != nil {
		return fmt.Errorf("failed to create transaction")
	}

	// store the block with a key format that allows
	// for query using the block number
	numKey := []byte(fmt.Sprintf("%d", block.GetNumber()))
	blockObj := database.NewKVObject(numKey, value, "block", "number")
	if err := tx.Put([]*database.KVObject{blockObj}); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to put block: %s", err)
	}

	// also index the block with a hash key to allow query
	// by block hash. But, do not store the full block, just the number.
	hashKey := []byte(fmt.Sprintf("%s", block.GetHash()))
	value = []byte(fmt.Sprintf("%d", block.GetNumber()))
	blockObj = database.NewKVObject(hashKey, value, "block", "hash")
	if err := tx.Put([]*database.KVObject{blockObj}); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to index block by hash: %s", err)
	}

	// get current metadata
	var meta = &types.Meta{}
	if err := s.getMetadata(tx, meta); err != nil {
		if err != types.ErrMetadataNotFound {
			tx.Rollback()
			return fmt.Errorf("failed to get metadata: %s", err)
		}
	}

	// update block number and save the updated meta
	meta.CurrentBlockNumber = block.GetNumber()
	if err := s.updateMetadata(tx, meta); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update metadata: %s", err)
	}

	return tx.Commit()
}

func (s *Store) getMetadata(db database.Tx, result *types.Meta) error {
	objs := db.GetByPrefix([]byte(MetadataKey))
	if len(objs) == 0 {
		return types.ErrMetadataNotFound
	}
	return json.Unmarshal(objs[0].Value, result)
}

func (s *Store) updateMetadata(db database.Tx, meta *types.Meta) error {
	value := util.ObjectToBytes(meta)
	obj := database.NewKVObject([]byte(MetadataKey), value)
	return db.Put([]*database.KVObject{obj})
}

// GetMetadata gets the metadata and copies it to result
func (s *Store) GetMetadata(result *types.Meta) error {

	tx, err := s.db.NewTx()
	if err != nil {
		return err
	}

	if err := s.getMetadata(tx, result); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// UpdateMetadata updates the meta data
func (s *Store) UpdateMetadata(meta *types.Meta) error {

	tx, err := s.db.NewTx()
	if err != nil {
		return err
	}

	if err := s.updateMetadata(tx, meta); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
