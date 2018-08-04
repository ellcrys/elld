package leveldb

import (
	"encoding/json"
	"fmt"

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
	_, err := s.getBlock(tx, chainID, number)
	if err != nil {
		if err == common.ErrBlockNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// getBlock gets a block by the block number.
// If number is 0, return the block with the highest block number
func (s *Store) getBlock(tx database.Tx, chainID string, number uint64) (*wire.Block, error) {

	// Since number is '0', we must fetch the last block
	// which is the block with the highest number and the most recent
	if number == 0 {
		kvO := s.db.GetFirstOrLast(common.MakeBlocksQueryKey(chainID), false)
		if kvO == nil {
			return nil, common.ErrBlockNotFound
		}
		var block wire.Block
		if err := json.Unmarshal(kvO.Value, &block); err != nil {
			return nil, err
		}
		return &block, nil
	}

	r := tx.GetByPrefix(common.MakeBlockKey(chainID, number))
	if len(r) == 0 {
		return nil, common.ErrBlockNotFound
	}

	var block wire.Block
	if err := json.Unmarshal(r[0].Value, &block); err != nil {
		return nil, err
	}
	return &block, nil
}

// GetBlockHeader the header of the current block
func (s *Store) GetBlockHeader(chainID string, number uint64, opts ...common.CallOp) (*wire.Header, error) {

	var err error
	var txOp = common.GetTxOp(s.db, opts...)

	block, err := s.getBlock(txOp.Tx, chainID, number)
	if err != nil {
		if txOp.CanFinish {
			txOp.Tx.Rollback()
		}
		return nil, err
	}

	if txOp.CanFinish {
		return block.Header, txOp.Tx.Commit()
	}

	return block.Header, nil
}

// GetBlockHeaderByHash returns the header of a block by searching using its hash
func (s *Store) GetBlockHeaderByHash(chainID string, hash string) (*wire.Header, error) {

	block, err := s.GetBlockByHash(chainID, hash)
	if err != nil {
		return nil, err
	}

	return block.Header, nil
}

// GetBlock fetches a block by its block number.
// If the block number begins with -1, the block with the highest block number is returned.
func (s *Store) GetBlock(chainID string, number uint64, opts ...common.CallOp) (*wire.Block, error) {

	var err error
	var txOp = common.GetTxOp(s.db, opts...)

	block, err := s.getBlock(txOp.Tx, chainID, number)
	if err != nil {
		if txOp.CanFinish {
			txOp.Tx.Rollback()
		}
		return nil, err
	}

	if txOp.CanFinish {
		return block, txOp.Tx.Commit()
	}

	return block, nil
}

// GetBlockByHash fetches a block by its block hash.
func (s *Store) GetBlockByHash(chainID string, hash string, opts ...common.CallOp) (*wire.Block, error) {

	var err error
	var txOp = common.GetTxOp(s.db, opts...)

	// iterate over the blocks in the chain and locate the block
	// matching the specified hash
	var block wire.Block
	var found = false
	txOp.Tx.Iterate(common.MakeBlocksQueryKey(chainID), true, func(kv *database.KVObject) bool {
		if err = util.BytesToObject(kv.Value, &block); err != nil {
			return true
		}
		found = block.Hash.HexStr() == hash
		return found
	})
	if err != nil {
		if txOp.CanFinish {
			txOp.Tx.Rollback()
		}
		return nil, err
	}

	if !found {
		if txOp.CanFinish {
			txOp.Tx.Rollback()
		}
		return nil, common.ErrBlockNotFound
	}

	if txOp.CanFinish {
		return &block, txOp.Tx.Commit()
	}

	return &block, nil
}

// PutBlock adds a block to the store.
// Returns error if a block with same number exists.
func (s *Store) PutBlock(chainID string, block *wire.Block, opts ...common.CallOp) error {

	var txOp = common.GetTxOp(s.db, opts...)

	if err := s.putBlock(txOp.Tx, chainID, block); err != nil {
		if txOp.CanFinish {
			txOp.Tx.Rollback()
		}
		return err
	}

	if txOp.CanFinish {
		return txOp.Tx.Commit()
	}

	return nil
}

// putBlock adds a block to the store using the provided transaction object.
// Returns error if a block with same number exists.
func (s *Store) putBlock(tx database.Tx, chainID string, block *wire.Block) error {

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

	return nil
}

// Put stores an object
func (s *Store) Put(key []byte, value []byte, opts ...common.CallOp) error {

	var txOp = common.GetTxOp(s.db, opts...)

	obj := database.NewKVObject(key, value)
	if err := txOp.Tx.Put([]*database.KVObject{obj}); err != nil {
		if txOp.CanFinish {
			txOp.Tx.Rollback()
		}
		return fmt.Errorf("failed to put object: %s", err)
	}

	if txOp.CanFinish {
		return txOp.Tx.Commit()
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
func (s *Store) NewTx() (database.Tx, error) {

	tx, err := s.db.NewTx()
	if err != nil {
		return nil, err
	}

	return tx, nil
}
