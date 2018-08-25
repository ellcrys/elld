package blockchain

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/olebedev/emitter"

	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/blockchain/store"
	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/types/core/objects"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/util/logger"
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
	chains map[util.String]*Chain

	// orphanBlocks stores blocks whose parents are unknown
	orphanBlocks *Cache

	// rejectedBlocks stores collection of blocks that have been deemed invalid.
	// This allows us to quickly learn and discard blocks that are found here.
	rejectedBlocks *Cache

	// txPool contains all transactions awaiting inclusion in a block
	txPool types.TxPool

	// eventEmitter allows the manager to listen to specific
	// events or broadcast events about its state
	eventEmitter *emitter.Emitter

	// reOrgActive indicates an ongoing reorganization
	reOrgActive bool
}

// New creates a Blockchain instance.
func New(txPool types.TxPool, cfg *config.EngineConfig, log logger.Logger) *Blockchain {
	bc := new(Blockchain)
	bc.txPool = txPool
	bc.log = log
	bc.cfg = cfg
	bc.chainLock = &sync.RWMutex{}
	bc.mLock = &sync.Mutex{}
	bc.chains = make(map[util.String]*Chain)
	bc.orphanBlocks = NewCache(MaxOrphanBlocksCacheSize)
	bc.rejectedBlocks = NewCache(MaxRejectedBlocksCacheSize)
	bc.eventEmitter = &emitter.Emitter{}
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
		b.log.Debug("No existing genesis block found. Creating genesis block")

		// Create the genesis chain and the genesis block.
		gBlock := GenesisBlock
		if gBlock.GetNumber() != 1 {
			return fmt.Errorf("genesis block error: expected block number 1")
		}

		// The ID of the genesis chain is the hash of the genesis block hash.
		gChainID := util.ToHex(util.Blake2b256(gBlock.GetHash().Bytes()))
		gChain := NewChain(util.String(gChainID), b.db, b.cfg, b.log)
		// Save the chain the chain (which also adds it to the chain cache)
		if err := b.saveChain(gChain, "", 0); err != nil {
			return fmt.Errorf("failed to save genesis chain: %s", err)
		}

		// Process the genesis block.
		if _, err := b.maybeAcceptBlock(gBlock, gChain); err != nil {
			return fmt.Errorf("genesis block error: %s", err)
		}

		b.log.Debug("Genesis block successfully created", "Hash", gBlock.GetHash().HexStr())
		return nil
	}

	// Load all known chains
	for _, chainInfo := range chains {
		if err := b.loadChain(chainInfo); err != nil {
			return err
		}
	}

	if numChains := len(chains); numChains > 0 {
		b.log.Info("Chain load completed", "NumChains", numChains)
	}

	// Using the best chain rule, we mush select the best chain
	// and set it as the current bestChain.
	err = b.decideBestChain()
	if err != nil {
		return fmt.Errorf("failed to determine best chain: %s", err)
	}

	return nil
}

// SetEventEmitter sets the event emitter
func (b *Blockchain) SetEventEmitter(ee *emitter.Emitter) {
	b.eventEmitter = ee
}

func (b *Blockchain) createBlockValidator(block core.Block) *BlockValidator {
	return NewBlockValidator(block, b.txPool, b, true, b.cfg, b.log)
}

// OrphanBlocks returns a cache reader for orphan blocks
func (b *Blockchain) OrphanBlocks() core.CacheReader {
	return b.orphanBlocks
}

// GetEventEmitter gets the event emitter
func (b *Blockchain) GetEventEmitter() *emitter.Emitter {
	return b.eventEmitter
}

// getChainParentBlock find the parent chain and block
// of a chain using the chain's ChainInfo
func (b *Blockchain) getChainParentBlock(ci *core.ChainInfo) (core.Block, error) {

	r := b.db.GetByPrefix(common.MakeBlockKey(ci.ParentChainID.Bytes(), ci.ParentBlockNumber))
	if len(r) == 0 {
		return nil, core.ErrBlockNotFound
	}

	var pb objects.Block
	if err := util.BytesToObject(r[0].Value, &pb); err != nil {
		return nil, err
	}

	return &pb, nil
}

// loadChain finds and load a chain into the chain cache. It
// can be used to find both standalone chain and child chains.
func (b *Blockchain) loadChain(ci *core.ChainInfo) error {

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
			if err == core.ErrBlockNotFound {
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
func (b *Blockchain) GetBestChain() core.Chainer {
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

	return h.GetNumber() >= b.cfg.Chain.TargetHybridModeBlock, nil
}

// findChainByBlockHash finds the chain where the given block
// hash exists. It returns the block, the chain, the header of
// highest block in the chain.
//
// NOTE: This method must be called with chain lock held by the caller.
func (b *Blockchain) findChainByBlockHash(hash util.Hash, opts ...core.CallOp) (block core.Block, chain *Chain, chainTipHeader core.Header, err error) {

	for _, chain := range b.chains {

		// Find the block by its hash. If we don't
		// find the block in this chain, we continue to the
		// next chain.
		block, err := chain.getBlockByHash(hash, opts...)
		if err != nil {
			if err != core.ErrBlockNotFound {
				return nil, nil, nil, err
			}
			continue
		}

		// At the point, we have found chain the block belongs to.
		// Next we get the header of the block at the tip of the chain.
		chainTipHeader, err := chain.Current(opts...)
		if err != nil {
			if err != core.ErrBlockNotFound {
				return nil, nil, nil, err
			}
		}

		return block, chain, chainTipHeader, nil
	}

	return nil, nil, nil, nil
}

// chooseBestChain returns the chain that is considered the
// legitimate chain. It checks all chains according to the rules
// defined below and return the chain that passes the rule on contested
// by another chain.
//
// The rules (executed in the exact order) :
// 1. The chain with the most difficulty wins.
// 2. The chain that was received first.
// 3. The chain with the larger pointer
//
// NOTE: This method must be called with chain lock held by the caller.
func (b *Blockchain) chooseBestChain() (*Chain, error) {

	var highTDChains = []*Chain{}
	var curHighestTD = new(big.Int).SetInt64(0)

	// If no chain exists on the blockchain, return nil
	if len(b.chains) == 0 {
		return nil, nil
	}

	// for each known chains, we must find the chain with the largest total
	// difficulty and add to highTDChains. If multiple chains have same
	// difficulty, then that indicates a tie and as such the highTDChains
	// will also include these chains.
	for _, chain := range b.chains {
		tip, err := chain.Current()
		if err != nil {
			// A chain with no tip is ignored.
			if err == core.ErrBlockNotFound {
				continue
			}
			return nil, err
		}
		cmpResult := tip.GetTotalDifficulty().Cmp(curHighestTD)
		if cmpResult > 0 {
			curHighestTD = tip.GetTotalDifficulty()
			highTDChains = []*Chain{chain}
		} else if cmpResult == 0 {
			highTDChains = append(highTDChains, chain)
		}
	}

	// When there is no tie for the total difficulty rule,
	// we return the only chain immediately
	if len(highTDChains) == 1 {
		return highTDChains[0], nil
	}

	// At this point there is a tie between two or more most difficult chains.
	// We need to perform tie breaker using rule 2.
	var oldestChains = []*Chain{}
	var curOldestTimestamp int64
	if len(highTDChains) > 1 {
		for _, chain := range highTDChains {
			if curOldestTimestamp == 0 || chain.info.Timestamp < curOldestTimestamp {
				curOldestTimestamp = chain.info.Timestamp
				oldestChains = []*Chain{chain}
			} else if chain.info.Timestamp == curOldestTimestamp {
				oldestChains = append(oldestChains, chain)
			}
		}
	}

	// When we have just one oldest chain, we return it immediately
	if len(oldestChains) == 1 {
		return oldestChains[0], nil
	}

	// If at this point we still have a tie in
	// the list of oldest chains, then we find the chain
	// with the highest pointer address
	var largestPointerAddrs = []*Chain{}
	var curLargestPointerAddress *big.Int
	if len(oldestChains) > 1 {
		for _, chain := range oldestChains {
			if curLargestPointerAddress == nil || util.GetPtrAddr(chain).Cmp(curLargestPointerAddress) > 0 {
				curLargestPointerAddress = util.GetPtrAddr(chain)
				largestPointerAddrs = []*Chain{chain}
			} else if util.GetPtrAddr(chain).Cmp(curLargestPointerAddress) == 0 {
				largestPointerAddrs = append(largestPointerAddrs, chain)
			}
		}
	}

	return largestPointerAddrs[0], nil
}

func (b *Blockchain) addRejectedBlock(block core.Block) {
	b.chainLock.Lock()
	defer b.chainLock.Unlock()
	b.rejectedBlocks.Add(block.GetHash().HexStr(), struct{}{})
}

func (b *Blockchain) isRejected(block core.Block) bool {
	b.chainLock.RLock()
	defer b.chainLock.RUnlock()
	return b.rejectedBlocks.Has(block.GetHash().HexStr())
}

// addOrphanBlock adds a block to the collection of orphaned blocks.
func (b *Blockchain) addOrphanBlock(block core.Block) {
	b.chainLock.Lock()
	defer b.chainLock.Unlock()
	// Insert the block to the cache with a 1 hour expiration
	b.orphanBlocks.AddWithExp(block.GetHash().HexStr(), block, time.Now().Add(time.Hour))
	b.log.Debug("Added block to orphan cache", "BlockNo", block.GetNumber(), "CacheSize", b.orphanBlocks.Len())
}

// isOrphanBlock checks whether a block is present in the collection of orphaned blocks.
func (b *Blockchain) isOrphanBlock(blockHash util.Hash) bool {
	b.chainLock.RLock()
	defer b.chainLock.RUnlock()
	return b.orphanBlocks.Get(blockHash.HexStr()) != nil
}

// NewChainFromChainInfo creates an instance of a Chain given a NewChainFromChainInfo
func (b *Blockchain) NewChainFromChainInfo(ci *core.ChainInfo) *Chain {
	ch := NewChain(ci.ID, b.db, b.cfg, b.log)
	ch.info.ParentChainID = ci.ParentChainID
	ch.info.ParentBlockNumber = ci.ParentBlockNumber
	return ch
}

// findChainInfo finds information about chain
// TODO: search cache first before database
func (b *Blockchain) findChainInfo(chainID util.String) (*core.ChainInfo, error) {

	var chainInfo core.ChainInfo
	var chainKey = common.MakeChainKey(chainID.Bytes())

	result := b.db.GetByPrefix(chainKey)
	if len(result) == 0 {
		return nil, core.ErrChainNotFound
	}

	if err := result[0].Scan(&chainInfo); err != nil {
		return nil, err
	}

	return &chainInfo, nil
}

// IsMainChain checks whether cr is the main chain
func (b *Blockchain) IsMainChain(cr core.ChainReader) bool {
	b.chainLock.RLock()
	defer b.chainLock.RUnlock()
	return b.bestChain.GetID() == cr.GetID()
}

// saveChain store a record about this new chain on the database.
// It will also cache the chain in memory that future query will be faster.
func (b *Blockchain) saveChain(chain *Chain, parentChainID util.String, parentBlockNumber uint64, opts ...core.CallOp) error {

	var err error
	var txOp = common.GetTxOp(b.db, opts...)

	chain.info = &core.ChainInfo{
		ID:                chain.GetID(),
		ParentBlockNumber: parentBlockNumber,
		ParentChainID:     parentChainID,
		Timestamp:         chain.info.Timestamp,
	}

	chainKey := common.MakeChainKey(chain.GetID().Bytes())
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
func (b *Blockchain) getChains() (chainsInfo []*core.ChainInfo, err error) {
	chainsKey := common.MakeChainsQueryKey()
	result := b.db.GetByPrefix(chainsKey)
	for _, r := range result {
		var ci core.ChainInfo
		if err = r.Scan(&ci); err != nil {
			return nil, err
		}
		chainsInfo = append(chainsInfo, &ci)
	}
	return
}

// getChainsByParent gets chains by their parent block number
func (b *Blockchain) getChainsByParent(blockNumber uint64) (chains []*Chain) {
	b.chainLock.RLock()
	defer b.chainLock.RUnlock()
	for _, ch := range b.chains {
		if ch.HasParent() && ch.info.ParentBlockNumber == blockNumber {
			chains = append(chains, ch)
		}
	}
	return
}

// hasChain checks whether a chain exists.
func (b *Blockchain) hasChain(chain *Chain) bool {
	b.chainLock.RLock()
	defer b.chainLock.RUnlock()

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
func (b *Blockchain) newChain(tx elldb.Tx, initialBlock core.Block, parentBlock core.Block, parentChain *Chain) (*Chain, error) {

	// The block and its parent must be provided.
	// They must also be related through the
	// initialBlock referencing the parent block's hash.
	if initialBlock == nil {
		return nil, fmt.Errorf("initial block cannot be nil")
	}
	if parentBlock == nil {
		return nil, fmt.Errorf("initial block parent cannot be nil")
	}
	if !initialBlock.GetHeader().GetParentHash().Equal(parentBlock.GetHash()) {
		return nil, fmt.Errorf("initial block and parent are not related")
	}
	if parentChain == nil {
		return nil, fmt.Errorf("parent chain cannot be nil")
	}

	// Create a new chain. Construct and assign a
	// deterministic id to it. This is the blake2b
	// 256 hash of the initial block hash.
	chainID := util.ToHex(util.Blake2b256(append([]byte{}, initialBlock.GetHash().Bytes()...)))
	chain := NewChain(util.String(chainID), b.db, b.cfg, b.log)

	// Set the parent block and parent chain on the
	// new chain.
	chain.parentBlock = parentBlock

	// store a record of this chain in the store
	b.saveChain(chain, parentChain.GetID(), parentBlock.GetNumber(), &common.TxOp{Tx: tx, CanFinish: false})

	return chain, nil
}

// GetTransaction finds a transaction in the main chain and returns it
func (b *Blockchain) GetTransaction(hash util.Hash) (core.Transaction, error) {
	b.chainLock.RLock()
	defer b.chainLock.RUnlock()

	if b.bestChain == nil {
		return nil, core.ErrBestChainUnknown
	}

	tx := b.bestChain.GetTransaction(hash)
	if tx == nil {
		return nil, core.ErrTxNotFound
	}

	return tx, nil
}

// ChainReader creates a chain reader for best/main chain
func (b *Blockchain) ChainReader() core.ChainReader {
	return store.NewChainReader(b.bestChain.store, b.bestChain.id)
}

// GetChainReaderByHash returns a chain reader to a chain
// where a block with the given hash exists
func (b *Blockchain) GetChainReaderByHash(hash util.Hash) core.ChainReader {
	b.chainLock.RLock()
	defer b.chainLock.RUnlock()
	_, chain, _, _ := b.findChainByBlockHash(hash)
	if chain == nil {
		return nil
	}
	return chain.ChainReader()
}

// GetChainsReader gets chain reader for all known chains
func (b *Blockchain) GetChainsReader() (readers []core.ChainReader) {
	b.chainLock.Lock()
	defer b.chainLock.Unlock()
	for _, c := range b.chains {
		readers = append(readers, store.NewChainReader(c.store, c.id))
	}
	return
}
