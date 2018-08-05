package blockchain

import (
	"fmt"
	"sync"
	"time"

	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/blockchain/store"
	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/txpool"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/util/logger"
	"github.com/ellcrys/elld/wire"
)

const (
	// MaxOrphanBlocksCacheSize is the number of blocks we can keep in the orphan block cache
	MaxOrphanBlocksCacheSize = 500

	// MaxRejectedBlocksCacheSize is the number of blocks we can keep in the rejected block cache
	MaxRejectedBlocksCacheSize = 100
)

// ChainOp defines a method option for passing a chain object
type ChainOp struct {
	Chain *Chain
}

// GetName returns the name of the op
func (t ChainOp) GetName() string {
	return "ChainOp"
}

// Blockchain represents the Ellcrys blockchain. It provides
// functionalities for interacting with the underlying database
// and primitives.
type Blockchain struct {

	// chainLock is a general purpose chainLock for store, bestChain, chains etc
	chainLock *sync.RWMutex

	// mLock is used to lock methods that should be called completely atomically
	mLock *sync.Mutex

	// cfg is the client configuration
	cfg *config.EngineConfig

	// log is used for logging output
	log logger.Logger

	// store is the the database where block data and other meta data are stored
	store store.Storer

	// bestChain is the chain considered to be the true chain.
	// It is protected by lock
	bestChain *Chain

	// chains holds all known chains
	chains map[string]*Chain

	// orphanBlocks stores blocks whose parents are unknown
	orphanBlocks *Cache

	// rejectedBlocks stores collection of blocks that have been deemed invalid.
	// This allows us to quickly learn and discard blocks that are found here.
	rejectedBlocks *Cache

	// txPool contains all transactions awaiting inclusion in a block
	txPool *txpool.TxPool
}

// New creates a Blockchain instance.
func New(txPool *txpool.TxPool, cfg *config.EngineConfig, log logger.Logger) *Blockchain {
	bc := new(Blockchain)
	bc.txPool = txPool
	bc.log = log
	bc.cfg = cfg
	bc.chainLock = &sync.RWMutex{}
	bc.mLock = &sync.Mutex{}
	bc.chains = make(map[string]*Chain)
	bc.orphanBlocks = NewCache(MaxOrphanBlocksCacheSize)
	bc.rejectedBlocks = NewCache(MaxRejectedBlocksCacheSize)
	return bc
}

// Up opens the database, initializes the store and
// creates the genesis block (if required)
func (b *Blockchain) Up() error {

	var err error

	// We cannot boot up the blockchain manager if a store.Storer
	// implementation has not been set.
	if b.store == nil {
		return fmt.Errorf("store has not been initialized")
	}

	// Get known chains
	chains, err := b.getChains()
	if err != nil {
		return err
	}

	// If there are no known chains described in the metadata and none
	// in the cache, then we create a new chain and save it
	if len(chains) == 0 {
		b.log.Debug("No existing chain found. Creating genesis chain")

		// Create the genesis chain and the genesis block.
		gBlock := GenesisBlock
		if gBlock.GetNumber() != 1 {
			return fmt.Errorf("genesis block error: expected block number 1")
		}

		// The ID of the genesis chain is the hash of the genesis block hash.
		gChainID := util.ToHex(util.Blake2b256(gBlock.Hash.Bytes()))
		gChain := NewChain(gChainID, b.store, b.cfg, b.log)

		// Save the chain the chain (which also adds it to the chain cache)
		if err := b.saveChain(gChain, "", 0); err != nil {
			return fmt.Errorf("failed to save genesis chain: %s", err)
		}

		// Process the genesis block.
		if _, err := b.maybeAcceptBlock(gBlock, gChain); err != nil {
			return fmt.Errorf("genesis block error: %s", err)
		}

		b.bestChain = gChain
	}

	// Load all known chains
	for _, chainInfo := range chains {
		if err := b.loadChain(chainInfo); err != nil {
			return err
		}
	}

	// Using the best chain rule, we mush select the best chain
	// and set it as the current bestChain.
	b.bestChain, err = b.chooseBestChain()
	if err != nil {
		return fmt.Errorf("failed to determine best chain: %s", err)
	}

	return nil
}

// loadChain finds and load a chain into the chain cache. It
// can be used to find both standalone chain and child chains.
func (b *Blockchain) loadChain(chainInfo *common.ChainInfo) error {

	chain := NewChain(chainInfo.ID, b.store, b.cfg, b.log)

	// Parent chain id and parent block number
	// are required to find a parent block.
	if (len(chainInfo.ParentChainID) != 0 && chainInfo.ParentBlockNumber == 0) ||
		(len(chainInfo.ParentChainID) == 0 && chainInfo.ParentBlockNumber != 0) {
		return fmt.Errorf("chain load failed: parent chain id and parent block id are both required")
	}

	b.chainLock.Lock()

	// For a chain with a parent chain and block.
	// We can attempt to find the parent block
	if len(chainInfo.ParentChainID) > 0 && chainInfo.ParentBlockNumber != 0 {

		parentBlock, err := b.store.GetBlock(chainInfo.ParentChainID, chainInfo.ParentBlockNumber)
		if err != nil {
			b.chainLock.Unlock()
			if err == common.ErrBlockNotFound {
				return fmt.Errorf(fmt.Sprintf("chain load failed: parent block {%d} of chain {%s} not found", chainInfo.ParentBlockNumber, chain.GetID()))
			}
			return err
		}

		// set the parent block and chain info
		chain.parentBlock = parentBlock
		chain.info = chainInfo
	}

	b.chainLock.Unlock()
	return b.addChain(chain)
}

// SetStore sets the store to use
func (b *Blockchain) SetStore(store store.Storer) {
	b.store = store
}

func (b *Blockchain) removeChain(chain *Chain) {
	b.chainLock.Lock()
	defer b.chainLock.Unlock()
	if _, ok := b.chains[chain.GetID()]; ok {
		delete(b.chains, chain.GetID())
	}
	return
}

// HybridMode checks whether the blockchain is a point where hybrid consensus
// can be utilized. Hybrid consensus mode allows consensus and blocks processed differently
// from standard block processing model. This mode is activated when we reach a target block height.
func (b *Blockchain) HybridMode() (bool, error) {
	b.chainLock.RLock()
	defer b.chainLock.RUnlock()

	h, err := b.bestChain.getTipHeader()
	if err != nil {
		return false, err
	}

	return h.Number >= b.cfg.Chain.TargetHybridModeBlock, nil
}

// findBlockChainByHash finds the chain where the block with the hash
// provided hash exist on. It also returns the header of highest block of the chain.
func (b *Blockchain) findBlockChainByHash(hash string) (block *wire.Block, chain *Chain, chainTipHeader *wire.Header, err error) {
	b.chainLock.RLock()
	defer b.chainLock.RUnlock()

	for _, chain := range b.chains {

		// Find the block by its hash. If we don't
		// find the block in this chain, we continue to the
		// next chain.
		block, err := chain.getBlockByHash(hash)
		if err != nil {
			if err != common.ErrBlockNotFound {
				return nil, nil, nil, err
			}
			continue
		}

		// At the point, we have found chain the block belongs to.
		// Next we get the header of the block at the tip of the chain.
		chainTipHeader, err := chain.getTipHeader()
		if err != nil {
			if err != common.ErrBlockNotFound {
				return nil, nil, nil, err
			}
		}

		return block, chain, chainTipHeader, nil
	}

	return nil, nil, nil, nil
}

// chooseBestChain returns the chain that is considered the
// legitimate chain. The longest chain is considered the best chain.
func (b *Blockchain) chooseBestChain() (*Chain, error) {
	b.chainLock.RLock()
	defer b.chainLock.RUnlock()

	var bestChains []*Chain
	var curHeight uint64

	// If no chain exists on the blockchain, return nil
	if len(b.chains) == 0 {
		return nil, nil
	}

	// for each known chains, we must find the longest chain and
	// add to bestChain. If two chains are of same height, then that indicates
	// a tie and as such the bestChain will include these chains.
	for _, chain := range b.chains {
		height, err := chain.height()
		if err != nil {
			return nil, err
		}
		if height > curHeight {
			curHeight = height
			bestChains = []*Chain{chain}
		} else if height == curHeight {
			bestChains = append(bestChains, chain)
		}
	}

	// When there is a definite best chain, we return it immediately
	if len(bestChains) == 1 {
		return bestChains[0], nil
	}

	// TODO: At this point there is a tie between two or more chains.
	// We need to perform tie breaker algorithms.

	return bestChains[0], nil
}

func (b *Blockchain) addRejectedBlock(block *wire.Block) {
	b.chainLock.Lock()
	defer b.chainLock.Unlock()
	b.rejectedBlocks.Add(block.GetHash().HexStr(), struct{}{})
}

func (b *Blockchain) isRejected(block *wire.Block) bool {
	b.chainLock.RLock()
	defer b.chainLock.RUnlock()
	return b.rejectedBlocks.Has(block.GetHash().HexStr())
}

// addOrphanBlock adds a block to the collection of orphaned blocks.
func (b *Blockchain) addOrphanBlock(block common.Block) {
	b.chainLock.Lock()
	defer b.chainLock.Unlock()
	// Insert the block to the cache with a 1 hour expiration
	b.orphanBlocks.AddWithExp(block.GetHash().HexStr(), block, time.Now().Add(time.Hour))
}

// isOrphanBlock checks whether a block is present in the collection of orphaned blocks.
func (b *Blockchain) isOrphanBlock(blockHash string) bool {
	b.chainLock.RLock()
	defer b.chainLock.RUnlock()
	return b.orphanBlocks.Get(blockHash) != nil
}

// findChainInfo finds information about chain
func (b *Blockchain) findChainInfo(chainID string) (*common.ChainInfo, error) {

	var result elldb.KVObject
	var chainInfo common.ChainInfo
	var chainKey = common.MakeChainKey(chainID)

	b.store.GetFirstOrLast(true, chainKey, &result)
	if len(result.Value) == 0 {
		return nil, common.ErrChainNotFound
	}
	if err := util.BytesToObject(result.Value, &chainInfo); err != nil {
		return nil, err
	}

	return &chainInfo, nil
}

// saveChain store a record about this new chain on the database.
// It will also cache the chain in memory that future query will be faster.
func (b *Blockchain) saveChain(chain *Chain, parentChainID string, parentBlockNumber uint64, opts ...common.CallOp) error {

	var err error
	var txOp = common.GetTxOp(b.store, opts...)

	chain.info = &common.ChainInfo{
		ID:                chain.GetID(),
		ParentBlockNumber: parentBlockNumber,
		ParentChainID:     parentChainID,
	}

	chainKey := common.MakeChainKey(chain.GetID())
	err = txOp.Tx.Put([]*elldb.KVObject{elldb.NewKVObject(chainKey, util.ObjectToBytes(chain.info))})
	if err != nil {
		if txOp.CanFinish {
			txOp.Tx.Rollback()
		}
		return err
	}

	if txOp.CanFinish {
		if err = txOp.Tx.Commit(); err != nil {
			return err
		}
	}

	return b.addChain(chain)
}

// getChains return all known chains
func (b *Blockchain) getChains() (chainsInfo []*common.ChainInfo, err error) {
	var result []*elldb.KVObject
	chainsKey := common.MakeChainsQueryKey()
	b.store.Get(chainsKey, &result)
	for _, r := range result {
		var ci common.ChainInfo
		if err = util.BytesToObject(r.Value, &ci); err != nil {
			return nil, err
		}
		chainsInfo = append(chainsInfo, &ci)
	}
	return
}

// hasChain checks whether a chain exists.
func (b *Blockchain) hasChain(chain *Chain) bool {
	b.chainLock.Lock()
	defer b.chainLock.Unlock()

	_, ok := b.chains[chain.GetID()]
	return ok
}

// addChain adds a new chain to the list of chains.
func (b *Blockchain) addChain(chain *Chain) error {
	b.chainLock.Lock()
	defer b.chainLock.Unlock()
	b.chains[chain.GetID()] = chain
	return nil
}

// newChain creates a new chain which may represent a fork.
// initialBlock is the block that will be added to this chain (a genesis block).
// parentBlock is the block that is the parent of the initialBlock. The parentBlock
// will be a block in another chain if this chain represents a fork. While
// parentChain is the chain on which the parent block sits in.
func (b *Blockchain) newChain(tx elldb.Tx, initialBlock *wire.Block, parentBlock *wire.Block, parentChain *Chain) (*Chain, error) {

	// The block and its parent must be provided. They must also
	// be related through the initialBlock referencing the parent block's hash.
	if initialBlock == nil {
		return nil, fmt.Errorf("initial block cannot be nil")
	}
	if parentBlock == nil {
		return nil, fmt.Errorf("initial block parent cannot be nil")
	}
	if initialBlock.Header.ParentHash != parentBlock.Hash {
		return nil, fmt.Errorf("initial block and parent are not related")
	}

	// create a new chain. Assign a unique and random id to it
	// set the parent block and chain on the new chain.
	// Construct a deterministic chain id which is the hash
	// of the initial block's hash
	chainID := util.ToHex(util.Blake2b256([]byte(initialBlock.Hash.Bytes())))
	chain := NewChain(chainID, b.store, b.cfg, b.log)
	chain.parentBlock = parentBlock

	// append the initial block to the new chain.
	if err := chain.append(initialBlock, common.TxOp{Tx: tx, CanFinish: false}); err != nil {
		return nil, err
	}

	var parentChainID string
	if parentChain != nil {
		parentChainID = parentChain.GetID()
	}

	// store a record of this chain in the store
	b.saveChain(chain, parentChainID, parentBlock.GetNumber(), common.TxOp{Tx: tx, CanFinish: false})

	return chain, nil
}

// GetTransaction finds a transaction in the main chain and returns it
func (b *Blockchain) GetTransaction(hash string) (*wire.Transaction, error) {
	b.chainLock.RLock()
	defer b.chainLock.RUnlock()

	if b.bestChain == nil {
		return nil, common.ErrBestChainUnknown
	}

	// Construct transaction key and find it on the main chain
	var txKey = common.MakeTxKey(b.bestChain.GetID(), hash)
	var result []*elldb.KVObject
	b.bestChain.store.Get(txKey, &result)

	if len(result) == 0 {
		return nil, common.ErrTxNotFound
	}
	var tx wire.Transaction
	if err := util.BytesToObject(result[0].Value, &tx); err != nil {
		return nil, err
	}

	return &tx, nil
}

// Reader creates a chain reader for best/main chain
func (b *Blockchain) Reader() store.Reader {
	if b.bestChain == nil {
		return nil
	}
	return store.NewChainReader(b.store, b.bestChain.id)
}
