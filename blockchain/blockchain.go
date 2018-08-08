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

	// db is the the database
	db elldb.DB

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

	// We cannot boot up the blockchain manager if a common.DB
	// implementation has not been set.
	if b.db == nil {
		return fmt.Errorf("db has not been initialized")
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
		gChain := NewChain(gChainID, b.db, b.cfg, b.log)
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

// getChainParentBlock find the parent chain and block
// of a chain using the chain's ChainInfo
func (b *Blockchain) getChainParentBlock(ci *common.ChainInfo) (*wire.Block, error) {

	r := b.db.GetByPrefix(common.MakeBlockKey(ci.ParentChainID, ci.ParentBlockNumber))
	if len(r) == 0 {
		return nil, common.ErrBlockNotFound
	}

	var pb wire.Block
	if err := util.BytesToObject(r[0].Value, &pb); err != nil {
		return nil, err
	}

	return &pb, nil
}

// loadChain finds and load a chain into the chain cache. It
// can be used to find both standalone chain and child chains.
func (b *Blockchain) loadChain(ci *common.ChainInfo) error {

	chain := NewChain(ci.ID, b.db, b.cfg, b.log)

	// Parent chain id and parent block number
	// are required to find a parent block.
	if (len(ci.ParentChainID) != 0 && ci.ParentBlockNumber == 0) ||
		(len(ci.ParentChainID) == 0 && ci.ParentBlockNumber != 0) {
		return fmt.Errorf("chain load failed: parent chain id and parent block id are both required")
	}

	b.chainLock.Lock()

	// For a chain with a parent chain and block.
	// We can attempt to find the parent block
	if len(ci.ParentChainID) > 0 && ci.ParentBlockNumber != 0 {
		parentBlock, err := b.getChainParentBlock(ci)
		if err != nil {
			b.chainLock.Unlock()
			if err == common.ErrBlockNotFound {
				return fmt.Errorf(fmt.Sprintf("chain load failed: parent block {%d} of chain {%s} not found", ci.ParentBlockNumber, chain.GetID()))
			}
			return err
		}

		// set the parent block and chain info
		chain.parentBlock = parentBlock
		chain.info = ci
	}

	b.chainLock.Unlock()
	return b.addChain(chain)
}

// GetBestChain gets the chain that is currently considered the main chain
func (b *Blockchain) GetBestChain() common.Chainer {
	return b.bestChain
}

// SetDB sets the database to use
func (b *Blockchain) SetDB(db elldb.DB) {
	b.db = db
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

	h, err := b.bestChain.Current()
	if err != nil {
		return false, err
	}

	return h.Number >= b.cfg.Chain.TargetHybridModeBlock, nil
}

// findChainByBlockHash finds the chain where the block with the hash
// provided hash exist on. It also returns the header of highest block of the chain.
func (b *Blockchain) findChainByBlockHash(hash string) (block *wire.Block, chain *Chain, chainTipHeader *wire.Header, err error) {
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
		chainTipHeader, err := chain.Current()
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
func (b *Blockchain) addOrphanBlock(block *wire.Block) {
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

	var chainInfo common.ChainInfo
	var chainKey = common.MakeChainKey(chainID)

	result := b.db.GetByPrefix(chainKey)
	if len(result) == 0 {
		return nil, common.ErrChainNotFound
	}

	if err := result[0].Scan(&chainInfo); err != nil {
		return nil, err
	}

	return &chainInfo, nil
}

// saveChain store a record about this new chain on the database.
// It will also cache the chain in memory that future query will be faster.
func (b *Blockchain) saveChain(chain *Chain, parentChainID string, parentBlockNumber uint64, opts ...common.CallOp) error {

	var err error
	var txOp = common.GetTxOp(b.db, opts...)

	chain.info = &common.ChainInfo{
		ID:                chain.GetID(),
		ParentBlockNumber: parentBlockNumber,
		ParentChainID:     parentChainID,
	}

	chainKey := common.MakeChainKey(chain.GetID())
	err = txOp.Tx.Put([]*elldb.KVObject{elldb.NewKVObject(chainKey, util.ObjectToBytes(chain.info))})
	if err != nil {
		txOp.Rollback()
		return err
	}

	if err = txOp.Commit(); err != nil {
		return err
	}

	return b.addChain(chain)
}

// getChains gets all known chains
func (b *Blockchain) getChains() (chainsInfo []*common.ChainInfo, err error) {
	chainsKey := common.MakeChainsQueryKey()
	result := b.db.GetByPrefix(chainsKey)
	for _, r := range result {
		var ci common.ChainInfo
		if err = r.Scan(&ci); err != nil {
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
//
// 'initialBlock' i block that will be added to this chain.
// It can be the genesis block or forked block.
//
// 'parentBlock' is the block that is the parent of the
// initialBlock. The parentBlock will be a block in another
// chain if this chain is being created win response
// to a fork.
//
// While parentChain is the chain on which the parent block
// belongs to.
func (b *Blockchain) newChain(tx elldb.Tx, initialBlock *wire.Block, parentBlock *wire.Block, parentChain *Chain) (*Chain, error) {

	// The block and its parent must be provided.
	// They must also be related through the
	// initialBlock referencing the parent block's hash.
	if initialBlock == nil {
		return nil, fmt.Errorf("initial block cannot be nil")
	}
	if parentBlock == nil {
		return nil, fmt.Errorf("initial block parent cannot be nil")
	}
	if initialBlock.Header.ParentHash != parentBlock.Hash {
		return nil, fmt.Errorf("initial block and parent are not related")
	}
	if parentChain == nil {
		return nil, fmt.Errorf("parent chain cannot be nil")
	}

	// Create a new chain. Construct and assign a
	// deterministic id to it. This is the blake2b
	// 256 hash of the initial block hash.
	chainID := util.ToHex(util.Blake2b256(append([]byte{}, initialBlock.Hash.Bytes()...)))
	chain := NewChain(chainID, b.db, b.cfg, b.log)

	// Set the parent block and parent chain on the
	// new chain.
	chain.parentBlock = parentBlock

	// store a record of this chain in the store
	b.saveChain(chain, parentChain.GetID(), parentBlock.GetNumber(), common.TxOp{Tx: tx, CanFinish: false})

	return chain, nil
}

// GetTransaction finds a transaction in the main chain and returns it
func (b *Blockchain) GetTransaction(hash string) (*wire.Transaction, error) {
	b.chainLock.RLock()
	defer b.chainLock.RUnlock()

	if b.bestChain == nil {
		return nil, common.ErrBestChainUnknown
	}

	tx := b.bestChain.GetTransaction(hash)
	if tx == nil {
		return nil, common.ErrTxNotFound
	}

	return tx, nil
}

// ChainReader creates a chain reader for best/main chain
func (b *Blockchain) ChainReader() common.ChainReader {
	return store.NewChainReader(b.bestChain.store, b.bestChain.id)
}

// GetChainsReader gets chain reader for all known chains
func (b *Blockchain) GetChainsReader() (readers []common.ChainReader) {
	b.chainLock.Lock()
	defer b.chainLock.Unlock()
	for _, c := range b.chains {
		readers = append(readers, store.NewChainReader(c.store, c.id))
	}
	return
}
