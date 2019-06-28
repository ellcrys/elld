package store

import (
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/syndtr/goleveldb/leveldb"

	"github.com/ellcrys/mother/blockchain/common"
	"github.com/ellcrys/mother/elldb"
	"github.com/ellcrys/mother/types"
	"github.com/ellcrys/mother/types/core"
	"github.com/ellcrys/mother/util"
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

// DB gets the database
func (s *ChainStore) DB() elldb.DB {
	return s.db
}

// hasBlock checks whether a block exists
func (s *ChainStore) hasBlock(number uint64, opts ...types.CallOp) (bool, error) {
	_, err := s.getBlock(number, opts...)
	if err != nil {
		if err != core.ErrBlockNotFound {
			return false, err
		}
		return false, nil
	}
	return true, nil
}

// getBlock gets a block by the block number.
// If number is 0, return the block with the highest block number
func (s *ChainStore) getBlock(number uint64, opts ...types.CallOp) (types.Block, error) {

	var txOp = common.GetTxOp(s.db, opts...)
	if txOp.Closed() {
		return nil, leveldb.ErrClosed
	}

	// Since number is '0', we must fetch the last block
	// which is the block with the highest number and the most recent
	if number == 0 {
		return s.Current(txOp)
	}

	r := txOp.Tx.GetByPrefix(common.MakeKeyBlock(s.chainID.Bytes(), number))
	if len(r) == 0 {
		txOp.Discard()
		return nil, core.ErrBlockNotFound
	}

	var block core.Block
	if err := r[0].Scan(&block); err != nil {
		txOp.Discard()
		return nil, err
	}

	return &block, txOp.Discard()
}

// GetHeader gets the header of the current block in the chain
func (s *ChainStore) GetHeader(number uint64, opts ...types.CallOp) (types.Header, error) {
	var err error
	block, err := s.getBlock(number, opts...)
	if err != nil {
		return nil, err
	}

	return block.GetHeader(), nil
}

// GetHeaderByHash returns the header of a block by searching using its hash
func (s *ChainStore) GetHeaderByHash(hash util.Hash, opts ...types.CallOp) (types.Header, error) {

	block, err := s.GetBlockByHash(hash, opts...)
	if err != nil {
		return nil, err
	}

	return block.GetHeader(), nil
}

// GetBlock fetches a block by its block number.
// If the block number begins with -1, the block with the highest block number is returned.
func (s *ChainStore) GetBlock(number uint64, opts ...types.CallOp) (types.Block, error) {
	var err error

	block, err := s.getBlock(number, opts...)
	if err != nil {
		return nil, err
	}

	return block, nil
}

// PutTransactions creates transaction pointers
// that stores the block numbers where the transactions
// can be found as opposed to storing the entire
// transaction. This saves disk space when considering
// that the block on disk already contains the transaction.
func (s *ChainStore) PutTransactions(txs []types.Transaction, blockNumber uint64, opts ...types.CallOp) error {
	var txOp = common.GetTxOp(s.db, opts...)
	if txOp.Closed() {
		return leveldb.ErrClosed
	}

	var kvObjs = []*elldb.KVObject{}
	for _, tx := range txs {
		txKey := common.MakeKeyTransaction(s.chainID.Bytes(),
			blockNumber, tx.GetHash().Hex())
		txObj := elldb.NewKVObject(txKey, util.EncodeNumber(blockNumber))
		kvObjs = append(kvObjs, txObj)
	}

	if err := txOp.Tx.Put(kvObjs); err != nil {
		txOp.Rollback()
		return err
	}

	return txOp.Commit()
}

// PutMinedBlock stores a brief information about a
// block that was created by the blockchain's coinbase key
func (s *ChainStore) PutMinedBlock(block types.Block, opts ...types.CallOp) error {
	var txOp = common.GetTxOp(s.db, opts...)
	if txOp.Closed() {
		return leveldb.ErrClosed
	}

	mb := &core.MinedBlock{
		Number:        block.GetNumber(),
		CreatorPubKey: block.GetHeader().GetCreatorPubKey(),
		Timestamp:     block.GetHeader().GetTimestamp(),
		TxCount:       uint(len(block.GetTransactions())),
		Hash:          block.GetHash(),
	}

	// Calculate total fees
	totalFee := decimal.New(0, 0)
	for _, tx := range block.GetTransactions() {
		if tx.GetType() == core.TxTypeAlloc {
			continue
		}
		totalFee.Add(tx.GetFee().Decimal())
	}
	mb.TotalFees = totalFee.String()
	value := util.ObjectToBytes(mb)

	key := common.MakeKeyMinedBlock(s.chainID.Bytes(), block.GetNumber())
	obj := elldb.NewKVObject(key, value)
	if err := txOp.Tx.Put([]*elldb.KVObject{obj}); err != nil {
		txOp.Rollback()
		return fmt.Errorf("failed to put mined block index: %s", err)
	}

	return txOp.Commit()
}

// Current gets the current block at the tip of the chain
func (s *ChainStore) Current(opts ...types.CallOp) (types.Block, error) {

	var err error
	var block core.Block
	var r *elldb.KVObject

	var txOp = common.GetTxOp(s.db, opts...)
	if txOp.Closed() {
		return nil, leveldb.ErrClosed
	}

	queryKey := common.MakeQueryKeyBlocks(s.chainID.Bytes())
	txOp.Tx.Iterate(queryKey, false, func(kv *elldb.KVObject) bool {
		r = kv
		return true
	})

	if r == nil {
		txOp.Discard()
		return nil, core.ErrBlockNotFound
	}

	if err = r.Scan(&block); err != nil {
		txOp.Discard()
		return nil, core.ErrDecodeFailed("")
	}

	return &block, txOp.Discard()
}

// GetBlockByHash fetches a block by its block hash.
func (s *ChainStore) GetBlockByHash(hash util.Hash, opts ...types.CallOp) (types.Block, error) {

	var txOp = common.GetTxOp(s.db, opts...)
	if txOp.Closed() {
		return nil, leveldb.ErrClosed
	}

	// First, we must get the block number stored
	// as the value of the block hash key
	queryKey := common.MakeKeyBlockHash(s.chainID.Bytes(), hash.Hex())
	r := txOp.Tx.GetByPrefix(queryKey)
	if len(r) == 0 {
		txOp.Discard()
		return nil, core.ErrBlockNotFound
	}
	blockNum := util.DecodeNumber(r[0].Value)

	return s.GetBlock(blockNum, txOp)
}

// GetBlockByNumberAndHash finds by number and hash
func (s *ChainStore) GetBlockByNumberAndHash(number uint64, hash util.Hash, opts ...types.CallOp) (types.Block, error) {

	// find a block in the chain with a matching number.
	// Expect to find 1 of such block
	block, err := s.getBlock(number, opts...)
	if err != nil {
		return nil, err
	}

	// If the found block does not have a hash that
	// matches the given hash, we conclude that the block
	// was not found.
	if !block.GetHash().Equal(hash) {
		return nil, core.ErrBlockNotFound
	}

	return block, nil
}

// PutBlock adds a block to the store.
// Returns error if a block with same number exists.
func (s *ChainStore) PutBlock(block types.Block, opts ...types.CallOp) error {

	var txOp = common.GetTxOp(s.db, opts...)
	if txOp.Closed() {
		return leveldb.ErrClosed
	}

	if err := s.putBlock(block, txOp); err != nil {
		txOp.Rollback()
		return err
	}

	return txOp.Commit()
}

// putBlock adds a block to the store using the provided transaction object.
// Returns error if a block with same number exists.
func (s *ChainStore) putBlock(block types.Block, opts ...types.CallOp) error {

	var txOp = common.GetTxOp(s.db, opts...)
	if txOp.Closed() {
		return leveldb.ErrClosed
	}

	// check if block already exists. return nil if block exists.
	hasBlock, err := s.hasBlock(block.GetNumber(), &common.OpTx{Tx: txOp.Tx})
	if err != nil {
		txOp.Rollback()
		return fmt.Errorf("failed to check block existence: %s", err)
	} else if hasBlock {
		return txOp.Commit()
	}

	value := util.ObjectToBytes(block)

	// store the block
	key := common.MakeKeyBlock(s.chainID.Bytes(), block.GetNumber())
	blockObj := elldb.NewKVObject(key, value)
	if err := txOp.Tx.Put([]*elldb.KVObject{blockObj}); err != nil {
		txOp.Rollback()
		return fmt.Errorf("failed to put block: %s", err)
	}

	// To allow query using the block hash,
	// we need to add a key constructed with the
	// the block's hash with the value set to the
	// block number
	pointerKey := common.MakeKeyBlockHash(s.chainID.Bytes(), block.GetHash().Hex())
	pointerObj := elldb.NewKVObject(pointerKey, util.EncodeNumber(block.GetNumber()))
	if err := txOp.Tx.Put([]*elldb.KVObject{pointerObj}); err != nil {
		txOp.Rollback()
		return fmt.Errorf("failed to put block number pointer: %s", err)
	}

	return txOp.Commit()
}

// Delete deletes objects
func (s *ChainStore) Delete(key []byte, opts ...types.CallOp) error {

	var txOp = common.GetTxOp(s.db, opts...)
	if txOp.Closed() {
		return leveldb.ErrClosed
	}

	if err := txOp.Tx.DeleteByPrefix(key); err != nil {
		txOp.Rollback()
		return fmt.Errorf("failed to remove object: %s", err)
	}

	return txOp.Commit()
}

// Put stores an object
func (s *ChainStore) put(key []byte, value []byte, opts ...types.CallOp) error {

	var txOp = common.GetTxOp(s.db, opts...)
	if txOp.Closed() {
		return leveldb.ErrClosed
	}

	obj := elldb.NewKVObject(key, value)
	if err := txOp.Tx.Put([]*elldb.KVObject{obj}); err != nil {
		txOp.Rollback()
		return fmt.Errorf("failed to put object: %s", err)
	}

	return txOp.Commit()
}

// get an object by key (and optionally by prefixes)
func (s *ChainStore) get(key []byte, result *[]*elldb.KVObject, opts ...types.CallOp) error {
	var txOp = common.GetTxOp(s.db, opts...)
	if txOp.Closed() {
		return leveldb.ErrClosed
	}

	r := txOp.Tx.GetByPrefix(key)
	if r == nil {
		return txOp.Discard()
	}

	*result = append(*result, r...)
	return txOp.Discard()
}

// GetTransaction gets a transaction (by hash)
// belonging to a chain
func (s *ChainStore) GetTransaction(hash util.Hash, opts ...types.CallOp) (types.Transaction, error) {

	var txOp = common.GetTxOp(s.db, opts...)
	if txOp.Closed() {
		return nil, leveldb.ErrClosed
	}

	var result []*elldb.KVObject
	err := s.get(common.MakeQueryKeyTransaction(s.chainID.Bytes(),
		hash.Hex()), &result, &common.OpTx{Tx: txOp.Tx})
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		txOp.Discard()
		return nil, core.ErrTxNotFound
	}

	// Using the block number stored as the transaction
	// key value, we must fetch the block and find
	// the transaction in its list of transactions
	blockNumber := util.DecodeNumber(result[0].Value)
	block, err := s.getBlock(blockNumber, txOp)
	if err != nil {
		if err == core.ErrBlockNotFound {
			return nil, fmt.Errorf("transaction's block not found")
		}
		return nil, err
	}

	for _, tx := range block.GetTransactions() {
		if tx.GetHash().Equal(hash) {
			return tx, nil
		}
	}

	txOp.Discard()

	return nil, core.ErrTxNotFound
}

// CreateAccount creates an account on a target block
func (s *ChainStore) CreateAccount(targetBlockNum uint64, account types.Account, opts ...types.CallOp) error {
	key := common.MakeKeyAccount(targetBlockNum, s.chainID.Bytes(), account.GetAddress().Bytes())
	return s.put(key, util.ObjectToBytes(account), opts...)
}

// GetAccount gets an account
func (s *ChainStore) GetAccount(address util.String, opts ...types.CallOp) (types.Account, error) {

	var r *elldb.KVObject

	var txOp = common.GetTxOp(s.db, opts...)
	if txOp.Closed() {
		return nil, leveldb.ErrClosed
	}

	queryKey := common.MakeQueryKeyAccount(s.chainID.Bytes(), address.Bytes())
	var blockRangeOp = common.GetBlockQueryRangeOp(opts...)
	txOp.Tx.Iterate(queryKey, false, func(kv *elldb.KVObject) bool {
		var bn = util.DecodeNumber(kv.Key)

		// Check block range constraint.
		// if the block number is less that the minimum
		// block number specified in the block range, skip to next.
		// Likewise, if the block number of the key is greater than
		// the maximum block number specified in the block range, skip object.
		if (blockRangeOp.Min > 0 && bn < blockRangeOp.Min) || blockRangeOp.Max > 0 && bn > blockRangeOp.Max {
			return false
		}

		r = kv

		return true
	})

	if r == nil {
		txOp.Discard()
		return nil, core.ErrAccountNotFound
	}

	var account core.Account
	r.Scan(&account)

	return &account, txOp.Discard()
}

// GetAccounts gets all accounts
func (s *ChainStore) GetAccounts(opts ...types.CallOp) ([]types.Account, error) {

	var accounts []types.Account
	var index = map[util.String]struct{}{}
	var txOp = common.GetTxOp(s.db, opts...)
	if txOp.Closed() {
		return nil, leveldb.ErrClosed
	}

	queryKey := common.MakeQueryKeyAccounts(s.chainID.Bytes())
	var blockRangeOp = common.GetBlockQueryRangeOp(opts...)
	txOp.Tx.Iterate(queryKey, false, func(kv *elldb.KVObject) bool {
		var bn = util.DecodeNumber(kv.Key)

		// Check block range constraint.
		// if the block number is less that the minimum
		// block number specified in the block range, skip to next.
		// Likewise, if the block number of the key is greater than
		// the maximum block number specified in the block range, skip object.
		if (blockRangeOp.Min > 0 && bn < blockRangeOp.Min) || blockRangeOp.Max > 0 && bn > blockRangeOp.Max {
			return false
		}

		var account core.Account
		kv.Scan(&account)
		if _, has := index[account.GetAddress()]; !has {
			accounts = append(accounts, &account)
			index[account.GetAddress()] = struct{}{}
		}

		return false
	})

	return accounts, txOp.Discard()
}

// NewTx creates and returns a transaction
func (s *ChainStore) NewTx() (elldb.Tx, error) {
	return s.db.NewTx()
}
