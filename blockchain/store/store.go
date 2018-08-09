package store

import (
	"fmt"

	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/wire"
)

// MetadataKey is the key used to store the metadata
const MetadataKey = "meta"

// ChainStore represents a store that implements the ChainStore
// interface meant to be used for persisting and retrieving
// objects for a given chain.
type ChainStore struct {
	db        elldb.DB
	namespace string
	chainID   util.String
}

// New creates an instance of the store
func New(db elldb.DB, chainID util.String) *ChainStore {
	return &ChainStore{
		db:      db,
		chainID: util.String(chainID),
	}
}

// hasBlock checks whether a block exists
func (s *ChainStore) hasBlock(number uint64, opts ...common.CallOp) (bool, error) {
	_, err := s.getBlock(number, opts...)
	if err != nil {
		if err != common.ErrBlockNotFound {
			return false, err
		}
		return false, nil
	}
	return true, nil
}

// getBlock gets a block by the block number.
// If number is 0, return the block with the highest block number
func (s *ChainStore) getBlock(number uint64, opts ...common.CallOp) (*wire.Block, error) {

	var txOp = common.GetTxOp(s.db, opts...)

	// Since number is '0', we must fetch the last block
	// which is the block with the highest number and the most recent
	if number == 0 {
		return s.Current(txOp)
	}

	r := txOp.Tx.GetByPrefix(common.MakeBlockKey(s.chainID.Bytes(), number))
	if len(r) == 0 {
		txOp.Rollback()
		return nil, common.ErrBlockNotFound
	}

	var block wire.Block
	if err := r[0].Scan(&block); err != nil {
		txOp.Rollback()
		return nil, err
	}

	txOp.Commit()

	return &block, nil
}

// GetHeader gets the header of the current block in the chain
func (s *ChainStore) GetHeader(number uint64, opts ...common.CallOp) (*wire.Header, error) {
	var err error

	block, err := s.getBlock(number, opts...)
	if err != nil {
		return nil, err
	}

	return block.Header, nil
}

// GetHeaderByHash returns the header of a block by searching using its hash
func (s *ChainStore) GetHeaderByHash(hash util.Hash, opts ...common.CallOp) (*wire.Header, error) {

	block, err := s.GetBlockByHash(hash, opts...)
	if err != nil {
		return nil, err
	}

	return block.Header, nil
}

// GetBlock fetches a block by its block number.
// If the block number begins with -1, the block with the highest block number is returned.
func (s *ChainStore) GetBlock(number uint64, opts ...common.CallOp) (*wire.Block, error) {
	var err error

	block, err := s.getBlock(number, opts...)
	if err != nil {
		return nil, err
	}

	return block, nil
}

// PutTransactions stores a collection of transactions
func (s *ChainStore) PutTransactions(txs []*wire.Transaction, opts ...common.CallOp) error {
	var txOp = common.GetTxOp(s.db, opts...)

	for i, tx := range txs {
		txKey := common.MakeTxKey(s.chainID.Bytes(), tx.Hash.Bytes())
		if err := txOp.Tx.Put([]*elldb.KVObject{elldb.NewKVObject(txKey, util.ObjectToBytes(tx))}); err != nil {
			txOp.Rollback()
			return fmt.Errorf("index %d: %s", i, err)
		}
	}

	return txOp.Commit()
}

// Current gets the current block at the tip of the chain
func (s *ChainStore) Current(opts ...common.CallOp) (*wire.Block, error) {

	var err error
	var block wire.Block
	var highestBlockNum uint64
	var r *elldb.KVObject
	var txOp = common.GetTxOp(s.db, opts...)

	// iterate over the blocks in the chain and locate the highest block

	txOp.Tx.Iterate(common.MakeBlocksQueryKey(s.chainID.Bytes()), true, func(kv *elldb.KVObject) bool {
		var bn = common.DecodeBlockNumber(kv.Key)
		if bn > highestBlockNum {
			highestBlockNum = bn
			r = kv
		}
		return false
	})

	if r == nil {
		txOp.Rollback()
		return nil, common.ErrBlockNotFound
	}

	if err = r.Scan(&block); err != nil {
		txOp.Rollback()
		return nil, common.ErrDecodeFailed("")
	}

	txOp.Commit()

	return &block, nil
}

// GetBlockByHash fetches a block by its block hash.
func (s *ChainStore) GetBlockByHash(hash util.Hash, opts ...common.CallOp) (*wire.Block, error) {

	var err error
	var txOp = common.GetTxOp(s.db, opts...)

	// iterate over the blocks in the chain and locate the block
	// matching the specified hash
	var block wire.Block
	var found = false
	txOp.Tx.Iterate(common.MakeBlocksQueryKey(s.chainID.Bytes()), true, func(kv *elldb.KVObject) bool {
		if err = kv.Scan(&block); err != nil {
			return true
		}
		found = block.Hash.Equal(hash)
		return found
	})
	if err != nil {
		txOp.Rollback()
		return nil, err
	}

	if !found {
		txOp.Rollback()
		return nil, common.ErrBlockNotFound
	}

	txOp.Commit()

	return &block, nil
}

// PutBlock adds a block to the store.
// Returns error if a block with same number exists.
func (s *ChainStore) PutBlock(block *wire.Block, opts ...common.CallOp) error {
	var txOp = common.GetTxOp(s.db, opts...)

	if err := s.putBlock(block, txOp); err != nil {
		txOp.Rollback()
		return err
	}

	txOp.Commit()

	return nil
}

// putBlock adds a block to the store using the provided transaction object.
// Returns error if a block with same number exists.
func (s *ChainStore) putBlock(block *wire.Block, opts ...common.CallOp) error {

	var txOp = common.GetTxOp(s.db, opts...)

	// check if block already exists. return nil if block exists.
	hasBlock, err := s.hasBlock(block.GetNumber(), common.TxOp{Tx: txOp.Tx})
	if err != nil {
		txOp.Rollback()
		return fmt.Errorf("failed to check block existence: %s", err)
	} else if hasBlock {
		txOp.Commit()
		return nil
	}

	value := util.ObjectToBytes(block)

	// store the block with a key format that allows
	// for query using the block number
	key := common.MakeBlockKey(s.chainID.Bytes(), block.GetNumber())
	blockObj := elldb.NewKVObject(key, value)
	if err := txOp.Tx.Put([]*elldb.KVObject{blockObj}); err != nil {
		txOp.Rollback()
		return fmt.Errorf("failed to put block: %s", err)
	}

	txOp.Commit()

	return nil
}

// Put stores an object
func (s *ChainStore) put(key []byte, value []byte, opts ...common.CallOp) error {
	var txOp = common.GetTxOp(s.db, opts...)

	obj := elldb.NewKVObject(key, value)
	if err := txOp.Tx.Put([]*elldb.KVObject{obj}); err != nil {
		txOp.Rollback()
		return fmt.Errorf("failed to put object: %s", err)
	}

	txOp.Commit()

	return nil
}

// Get an object by key (and optionally by prefixes)
func (s *ChainStore) get(key []byte, result *[]*elldb.KVObject, opts ...common.CallOp) {
	var txOp = common.GetTxOp(s.db, opts...)

	r := txOp.Tx.GetByPrefix(key)
	if r == nil {
		txOp.Commit()
		return
	}

	*result = append(*result, r...)
	txOp.Commit()
}

// GetTransaction gets a transaction (by hash) belonging to a chain
func (s *ChainStore) GetTransaction(hash util.Hash, opts ...common.CallOp) *wire.Transaction {
	var result []*elldb.KVObject
	var tx wire.Transaction
	var txOp = common.GetTxOp(s.db, opts...)

	s.get(common.MakeTxKey(s.chainID.Bytes(), hash.Bytes()), &result, txOp)
	if len(result) == 0 {
		txOp.Rollback()
		return nil
	}

	txOp.Commit()

	result[0].Scan(&tx)
	return &tx
}

// CreateAccount creates an account on a target block
func (s *ChainStore) CreateAccount(targetBlockNum uint64, account *wire.Account, opts ...common.CallOp) error {
	key := common.MakeAccountKey(targetBlockNum, s.chainID.Bytes(), account.Address.Bytes())
	return s.put(key, util.ObjectToBytes(account), opts...)
}

// GetAccount fetches the account with highest block number prefix.
func (s *ChainStore) GetAccount(address util.String, opts ...common.CallOp) (*wire.Account, error) {

	var key = common.QueryAccountKey(s.chainID.Bytes(), address.Bytes())
	var highestBlockNum uint64
	var r *elldb.KVObject

	var txOp = common.GetTxOp(s.db, opts...)
	txOp.Tx.Iterate(key, false, func(kv *elldb.KVObject) bool {
		var bn = common.DecodeBlockNumber(kv.Key)
		if bn > highestBlockNum {
			highestBlockNum = bn
			r = kv
		}
		return false
	})

	if r == nil {
		txOp.Rollback()
		return nil, common.ErrAccountNotFound
	}

	var account wire.Account
	if err := r.Scan(&account); err != nil {
		txOp.Rollback()
		return nil, err
	}

	txOp.Commit()

	return &account, nil
}

// NewTx creates and returns a transaction
func (s *ChainStore) NewTx() (elldb.Tx, error) {
	return s.db.NewTx()
}
