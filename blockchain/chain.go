package blockchain

import (
	"fmt"
	"sync"
	"time"

	"github.com/syndtr/goleveldb/leveldb"

	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"

	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/blockchain/store"
	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/util/logger"
)

// Chain represents a chain of blocks
// Implements types.Chainer
type Chain struct {

	// id represents the identifier of this chain
	id util.String

	// parentBlock represents the block from which this chain is formed.
	// A chain that is not a subtree of another chain will have this set to nil.
	parentBlock types.Block

	// parentChain is the parent chain from which this
	// chain is rooted on
	parentChain *Chain

	// info holds information about the chain
	info *core.ChainInfo

	// cfg includes configuration parameters of the client
	cfg *config.EngineConfig

	// chainLock is used to synchronize access to fields that are
	// accessed concurrently
	chainLock *sync.RWMutex

	// store provides functionalities for storing objects
	store types.ChainStorer

	// log is used for logging
	log logger.Logger
}

// NewChain creates an instance of a chain. It will create metadata object for the
// chain if not exists. It will return error if it is unable to do so.
func NewChain(id util.String, db elldb.DB,
	cfg *config.EngineConfig, log logger.Logger) *Chain {
	chain := new(Chain)
	chain.id = id
	chain.cfg = cfg
	chain.store = store.New(db, chain.id)
	chain.chainLock = &sync.RWMutex{}
	chain.log = log
	chain.parentChain = nil
	chain.info = &core.ChainInfo{
		ID:        id,
		Timestamp: time.Now().UnixNano(),
	}
	return chain
}

// NewChainFromChainInfo creates a
// chain with a given chain info
func NewChainFromChainInfo(ci *core.ChainInfo, db elldb.DB,
	cfg *config.EngineConfig, log logger.Logger) *Chain {
	ch := NewChain(ci.ID, db, cfg, log)
	ch.info = ci
	return ch
}

// GetStore gets the store
func (c *Chain) GetStore() types.ChainStorer {
	return c.store
}

// GetID returns the id of the chain
func (c *Chain) GetID() util.String {
	return c.id
}

// ChainReader gets a chain reader for this chain
func (c *Chain) ChainReader() types.ChainReaderFactory {
	return NewChainReader(c)
}

// GetParentBlock gets the chain's parent block if it has one
func (c *Chain) GetParentBlock() types.Block {
	return c.parentBlock
}

// GetInfo gets the chain information
func (c *Chain) GetInfo() types.ChainInfo {
	return c.info
}

// GetParent gets an instance of this chain's parent
func (c *Chain) GetParent(opts ...types.CallOp) *Chain {
	if c.info == nil || c.info.ParentChainID == "" {
		return nil
	}
	c.chainLock.Lock()
	c.loadParent(opts...)
	c.chainLock.Unlock()
	return c.parentChain
}

// HasParent checks whether the chain has a parent
func (c *Chain) HasParent(opts ...types.CallOp) bool {
	return c.GetParent(opts...) != nil
}

// GetBlock fetches a block by its number
func (c *Chain) GetBlock(number uint64, opts ...types.CallOp) (types.Block, error) {
	b, err := c.store.GetBlock(number, opts...)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// save saves a chain to the store
func (c *Chain) save(opts ...types.CallOp) error {
	var err error
	var txOp = common.GetTxOp(c.store, opts...)
	if txOp.Closed() {
		return leveldb.ErrClosed
	}

	chainKey := common.MakeKeyChain(c.GetID().Bytes())
	err = txOp.Tx.Put([]*elldb.KVObject{elldb.NewKVObject(chainKey, util.ObjectToBytes(c.info))})
	if err != nil {
		txOp.Rollback()
		return err
	}

	return txOp.Commit()
}

// loadParent fetches the parent chain and parent
// block from the database and caches them. It
// will return the cache parent chain if it is set
// by previous calls. It will return an error it
// failed to find the parent chain or block.
//
// NOTE: should be called with chainLock held
func (c *Chain) loadParent(opts ...types.CallOp) (*Chain, error) {

	// Get the cached version if
	// we already saved it
	if c.parentChain != nil {
		return c.parentChain, nil
	}

	// When chain has no parent,
	// return double nil values
	if c.info.ParentChainID == "" {
		return nil, nil
	}

	var txOp = common.GetTxOp(c.store, opts...)
	if txOp.Closed() {
		return nil, leveldb.ErrClosed
	}

	// Fetch the chain info of the parent.
	// If not found return ErrChainParentNotFound
	chainKey := common.MakeKeyChain(c.info.ParentChainID.Bytes())
	result := txOp.Tx.GetByPrefix(chainKey)
	if len(result) == 0 {
		txOp.Rollback()
		return nil, core.ErrChainParentNotFound
	}

	// Decode the chain info into a
	// ChainInfo object
	var parentInfo core.ChainInfo
	result[0].Scan(&parentInfo)

	// Construct the parent chain
	// and cache it
	pChain := &Chain{
		id:          parentInfo.ID,
		info:        &parentInfo,
		cfg:         c.cfg,
		log:         c.log,
		parentChain: nil,
		parentBlock: nil,
		store:       store.New(c.store.DB(), parentInfo.ID),
		chainLock:   &sync.RWMutex{},
	}

	// Get parent block from the parent chain
	// and cache it
	parentBlock, err := pChain.GetBlock(c.info.ParentBlockNumber, txOp)
	if err != nil {
		txOp.Rollback()
		if err != core.ErrBlockNotFound {
			return nil, err
		}
		return nil, core.ErrChainParentBlockNotFound
	}

	// cache the parent block and chain
	// to make future calls faster
	c.parentBlock = parentBlock
	c.parentChain = pChain

	return pChain, txOp.Commit()
}

// Current returns the header of the highest block on this c
func (c *Chain) Current(opts ...types.CallOp) (types.Header, error) {
	h, err := c.store.GetHeader(0, opts...)
	if err != nil {
		return nil, err
	}

	return h, nil
}

// height returns the height of this chain. The height can
// be deduced by fetching the number of the most recent block
// added to the chain.
func (c *Chain) height(opts ...types.CallOp) (uint64, error) {
	tip, err := c.Current(opts...)
	if err != nil {
		if err != core.ErrBlockNotFound {
			return 0, err
		}
		return 0, nil
	}
	return tip.GetNumber(), nil
}

// hasBlock checks if a block with the provided hash exists on this chain
func (c *Chain) hasBlock(hash util.Hash) (bool, error) {
	h, err := c.store.GetHeaderByHash(hash)
	if err != nil {
		if err != core.ErrBlockNotFound {
			return false, err
		}
	}

	return h != nil, nil
}

// getBlockHeaderByHash returns the header of a block that matches the hash on this chain
func (c *Chain) getBlockHeaderByHash(hash util.Hash) (types.Header, error) {
	h, err := c.store.GetHeaderByHash(hash)
	if err != nil {
		return nil, err
	}

	return h, nil
}

// GetRoot finds the block on the main chain
// from which this chain or its parents/ancestors
// originate from.
//
// Example:
// [1]-[2]-[3]-[4]-[5]  Main
//      |__[3]-[4]		Chain B
//          |__[4]		Chain C
// In the example above, the Chain B is the first generation
// to Chain C. The root parent block of Chain C is [2].
func (c *Chain) GetRoot() types.Block {

	// If this chain has no parent, it is the main chain, and
	// there for, has no root.
	if c.GetParent() == nil {
		return nil
	}

	// Set the current chain to c. Traverse the parent of c till
	// we find the chain that has no parent.
	curChain := c
	for {
		if parent := curChain.GetParent(); parent.GetParent() == nil {
			return curChain.GetParentBlock()
		}
		curChain = curChain.GetParent()
	}
}

// getBlockByHash fetches a block by hash
func (c *Chain) getBlockByHash(hash util.Hash, opts ...types.CallOp) (types.Block, error) {
	return c.store.GetBlockByHash(hash, opts...)
}

// getBlockByNumberAndHash fetches a block by number and hash
func (c *Chain) getBlockByNumberAndHash(number uint64, hash util.Hash) (types.Block, error) {
	return c.store.GetBlockByNumberAndHash(number, hash)
}

// CreateAccount creates an account on a target block
func (c *Chain) CreateAccount(targetBlockNum uint64, account types.Account, opts ...types.CallOp) error {
	return c.store.CreateAccount(targetBlockNum, account, opts...)
}

// GetAccount gets an account
func (c *Chain) GetAccount(address util.String, opts ...types.CallOp) (types.Account, error) {
	return c.store.GetAccount(address, opts...)
}

// GetAccounts gets all accounts
func (c *Chain) GetAccounts(opts ...types.CallOp) ([]types.Account, error) {
	return c.store.GetAccounts()
}

// append adds a block to the tail of the chain. It returns
// error if the previous block hash in the header is not the hash
// of the current block and if the difference between the chain tip
// and the candidate block number is not 1. If there is no block on
// the chain yet, then we assume this to be the first block of a fork.
//
// The caller is expected to validate the block before call.
func (c *Chain) append(candidate types.Block, opts ...types.CallOp) error {

	var err error
	var txOp = common.GetTxOp(c.store, opts...)
	if txOp.Closed() {
		return leveldb.ErrClosed
	}

	// Get the current block at the tip of the chain.
	// Continue if no error or no block currently exist on the chain.
	chainTip, err := c.store.Current(&common.OpTx{Tx: txOp.Tx})
	if err != nil {
		if err != core.ErrBlockNotFound {
			txOp.Rollback()
			return err
		}
	}

	// If the difference between the tip's block number and the new block number
	// is not 1, then this new block will not satisfy the serial numbering of blocks.
	if chainTip != nil && (candidate.GetNumber()-chainTip.GetNumber()) != 1 {
		txOp.Rollback()
		return fmt.Errorf(fmt.Sprintf("unable to append: candidate block number {%d} "+
			"is not the expected block number {expected=%d}", candidate.GetNumber(), chainTip.GetNumber()+1))
	}

	// If we found the current chainTip and its hash does not correspond with the
	// hash of the block we are trying to append, then we return an error.
	if chainTip != nil && !chainTip.GetHash().Equal(candidate.GetHeader().GetParentHash()) {
		txOp.Rollback()
		return fmt.Errorf("unable to append block: parent hash does not match the hash of the current block")
	}

	return c.store.PutBlock(candidate, txOp)
}

// NewStateTree creates a new tree seeded with the
// state root of the chain's tip block. For chains
// with no block (new chains), the state root of
// their parent block is used.
func (c *Chain) NewStateTree(opts ...types.CallOp) (types.Tree, error) {

	var prevRoot util.Hash

	// Get the state root of the block at the tip.
	// If no block was found, it means the chain is empty.
	// In this case, if the chain has a parent block,
	// we use the parent block stateRoot.
	tipHeader, err := c.Current(opts...)
	if err != nil {
		if err != core.ErrBlockNotFound {
			return nil, err
		}
		if c.parentBlock != nil {
			prevRoot = c.parentBlock.GetHeader().GetStateRoot()
		}
	} else {
		prevRoot = tipHeader.GetStateRoot()
		if err != nil {
			return nil, fmt.Errorf("failed to decode chain tip state root")
		}
	}

	// Create the new tree and seed it by adding
	// state root of the previous block. No need
	// to do this if at this point we  have not
	// determined the previous state root.
	tree := common.NewTree()
	if !prevRoot.IsEmpty() {
		tree.Add(common.TreeItem(prevRoot.Bytes()))
	}

	return tree, nil
}

// PutTransactions stores a collection of transactions in the chain
func (c *Chain) PutTransactions(txs []types.Transaction, blockNumber uint64, opts ...types.CallOp) error {
	return c.store.PutTransactions(txs, blockNumber, opts...)
}

// GetTransaction gets a transaction by hash
func (c *Chain) GetTransaction(hash util.Hash, opts ...types.CallOp) (types.Transaction, error) {
	tx, err := c.store.GetTransaction(hash, opts...)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (c *Chain) String() string {
	parent := ""
	if p := c.GetParent(); p != nil {
		parent = p.GetID().String()
	}
	return fmt.Sprintf("<chain id=%s parent=%s>", c.id, parent)
}

// PutMinedBlock records a block mined by the block creator
func (c *Chain) PutMinedBlock(block types.Block, opts ...types.CallOp) error {
	return c.store.PutMinedBlock(block, opts...)
}

// GetMinedBlocks fetches mined blocks. It allows the
// query to be adjusted using values in args.
// - args.Limit forces only a limited number of
//   results to be returned.
// - args.CreatorPubKey filters out results that do
//   not match a given public key.
// - args.LastHash allows only records after the given
//   hash to be returned. Useful for pagination.
// It returns a slice of mined blocks and also a boolean
// that indicates whether there are more records.
func (c *Chain) GetMinedBlocks(args *core.ArgGetMinedBlock, opts ...types.CallOp) ([]*core.MinedBlock, bool, error) {

	txOp := common.GetTxOp(c.store.DB(), opts...)
	if txOp.Closed() {
		return nil, false, leveldb.ErrClosed
	}

	if args.Limit == 0 {
		args.Limit = 25
	}

	var result []*core.MinedBlock

	var err error
	var hasMore bool
	var skip = false

	// When last hash is set, we need to indicate
	// that we want to skip all records until we
	// find a record matching the `last hash`
	if args.LastHash != "" {
		skip = true
	}

	key := common.MakeQueryKeyMinedBlocks(c.id.Bytes())
	txOp.Tx.Iterate(key, false, func(kv *elldb.KVObject) bool {

		// If we exceed the limit, we are certain that
		// that there is at least one object left to
		// be read so we set hasMore to true, and return
		// with the result. We also reduced the result
		// up to the limit specified
		if len(result) > args.Limit {
			result = result[:args.Limit]
			hasMore = true
			return true
		}

		// Convert the object to core.MinedBlock
		var res core.MinedBlock
		if err = kv.Scan(&res); err != nil {
			return true
		}

		// When LastHash is set and the current result
		// hash matches, we need to stop skipping and start
		// collecting results, starting from the next object
		if args.LastHash != "" && res.Hash.HexStr() == args.LastHash {
			skip = false
			return false
		}

		// If skip is enabled, we immediately move to
		// the next object
		if skip {
			return false
		}

		// When CreatorPubKey is specified in the arguments,
		// we have to ignore this object if its creator public
		// key does not match.
		if args.CreatorPubKey != "" &&
			res.CreatorPubKey.String() != args.CreatorPubKey {
			return false
		}

		result = append(result, &res)

		return false
	})

	return result, hasMore, txOp.Finishable().Discard()
}

// removeBlock deletes a block and all objects
// associated with it such as transactions, accounts etc.
// It returns the deleted block.
func (c *Chain) removeBlock(number uint64, opts ...types.CallOp) (types.Block, error) {

	var err error
	txOp := common.GetTxOp(c.store.DB(), opts...)
	if txOp.Closed() {
		return nil, leveldb.ErrClosed
	}
	txOp.CanFinish = false

	// Get the block.
	// Returns ErrBlockNotFound if block does not exist
	block, err := c.store.GetBlock(number, txOp)
	if err != nil {
		if len(opts) == 0 {
			txOp.Finishable().Rollback()
		}
		return nil, err
	}

	// Delete the block
	blockKey := common.MakeKeyBlock(c.id.Bytes(), number)
	if err = c.store.Delete(blockKey, txOp); err != nil {
		if len(opts) == 0 {
			txOp.Finishable().Rollback()
		}
		return nil, fmt.Errorf("failed to delete block: %s", err)
	}

	// Delete the block's hash pointer
	pointerKey := common.MakeKeyBlockHash(c.id.Bytes(), block.GetHash().Hex())
	if err = c.store.Delete(pointerKey, txOp); err != nil {
		if len(opts) == 0 {
			txOp.Finishable().Rollback()
		}
		return nil, fmt.Errorf("failed to delete block's hash pointer: %s", err)
	}

	// Delete the mined block object
	minedBlockKey := common.MakeKeyMinedBlock(c.id.Bytes(), number)
	if err = c.store.Delete(minedBlockKey, txOp); err != nil {
		if len(opts) == 0 {
			txOp.Finishable().Rollback()
		}
		return nil, fmt.Errorf("failed to delete mined block record: %s", err)
	}

	// Find accounts associated to with the block and delete them
	err = nil
	accountsKey := common.MakeQueryKeyAccounts(c.id.Bytes())
	txOp.Tx.Iterate(accountsKey, false, func(kv *elldb.KVObject) bool {
		var bn = util.DecodeNumber(kv.Key)
		if bn == number {
			if err = txOp.Tx.DeleteByPrefix(kv.GetKey()); err != nil {
				return true
			}
		}
		return false
	})
	if err != nil {
		if len(opts) == 0 {
			txOp.Finishable().Rollback()
		}
		return nil, fmt.Errorf("failed to delete accounts: %s", err)
	}

	// Find indexed transactions associated with this block and delete them
	err = nil
	txsKey := common.MakeQueryKeyTransactions(c.id.Bytes())
	txOp.Tx.Iterate(txsKey, false, func(kv *elldb.KVObject) bool {
		var bn = util.DecodeNumber(kv.Key)
		if bn == number {
			if err = txOp.Tx.DeleteByPrefix(kv.GetKey()); err != nil {
				return true
			}
		}
		return false
	})
	if err != nil {
		if len(opts) == 0 {
			txOp.Finishable().Rollback()
		}
		return nil, fmt.Errorf("failed to delete transactions: %s", err)
	}

	if len(opts) == 0 {
		return block, txOp.Finishable().Commit()
	}

	return block, nil
}

// ChainReader provides read-only access to
// objects belonging to a single chain.
type ChainReader struct {
	ch *Chain
}

// NewChainReader creates a ChainReader object
func NewChainReader(ch *Chain) *ChainReader {
	return &ChainReader{
		ch: ch,
	}
}

// GetID gets the chain ID
func (r *ChainReader) GetID() util.String {
	return r.ch.GetID()
}

// GetParent returns a chain reader to the parent chain.
// Returns nil if chain has no parent.
func (r *ChainReader) GetParent() types.ChainReaderFactory {
	if ch := r.ch.GetParent(); ch != nil {
		return ch.ChainReader()
	}
	return nil
}

// GetParentBlock returns the parent block
func (r *ChainReader) GetParentBlock() types.Block {
	return r.ch.GetParentBlock()
}

// GetRoot fetches the root block of this chain. If the chain
// has more than one parents/ancestors, it will traverse
// the parents to return the root parent block.
func (r *ChainReader) GetRoot() types.Block {
	return r.ch.GetRoot()
}

// GetBlock finds and returns a block associated with chainID.
// When 0 is passed, it should return the block with the highest number
func (r *ChainReader) GetBlock(number uint64, opts ...types.CallOp) (types.Block, error) {
	return r.ch.GetBlock(number, opts...)
}

// GetBlockByHash finds and returns a block associated with chainID.
func (r *ChainReader) GetBlockByHash(hash util.Hash, opts ...types.CallOp) (types.Block, error) {
	return r.ch.GetStore().GetBlockByHash(hash, opts...)
}

// GetHeader gets the header of a block.
// When 0 is passed, it should return the header of the block with the highest number
func (r *ChainReader) GetHeader(number uint64, opts ...types.CallOp) (types.Header, error) {
	return r.ch.GetStore().GetHeader(number, opts...)
}

// GetHeaderByHash finds and returns the header of a block matching hash
func (r *ChainReader) GetHeaderByHash(hash util.Hash, opts ...types.CallOp) (types.Header, error) {
	return r.ch.GetStore().GetHeaderByHash(hash, opts...)
}

// Current gets the current block at the tip of the chain
func (r *ChainReader) Current(opts ...types.CallOp) (types.Block, error) {
	return r.ch.GetStore().Current(opts...)
}
