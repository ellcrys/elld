// Package blockchain provides functionalities for
// creating, managing and accessing the blockchain state.
package blockchain

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ellcrys/elld/crypto"
	"github.com/syndtr/goleveldb/leveldb"

	"github.com/gobuffalo/packr"

	"github.com/olebedev/emitter"

	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/util/cache"
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
	// lock is a general purpose lock for store etc
	lock *sync.RWMutex

	// coinbase is the key that identifies this blockchain
	// instance.
	coinbase *crypto.Key

	// genesisBlock is the initial, hardcoded block
	// shared by all clients. It is the root of all chains.
	genesisBlock types.Block

	// processLock is used to lock the main block processing method
	processLock *sync.Mutex

	// cfg is the client configuration
	cfg *config.EngineConfig

	// log is used for logging output
	log logger.Logger

	// db is the the database
	db elldb.DB

	// orphanBlocks stores blocks whose parents are unknown
	orphanBlocks *cache.Cache

	// rejectedBlocks stores collection of blocks that have been deemed invalid.
	// This allows us to quickly learn and discard blocks that are found here.
	rejectedBlocks *cache.Cache

	// txPool contains all transactions awaiting inclusion in a block
	txPool types.TxPool

	// eventEmitter allows the manager to listen to specific
	// events or broadcast events about its state
	eventEmitter *emitter.Emitter

	// chl is a lock for chain events
	chl *sync.RWMutex

	// bestChain is the chain considered to be the true chain.
	// It is protected by lock
	bestChain *Chain

	// chains holds all known chains
	chains map[util.String]*Chain

	// rol is a lock used for re-org events
	rol *sync.RWMutex

	// reOrgActive indicates an ongoing reorganization
	reOrgActive bool
}

// New creates a Blockchain instance.
func New(txPool types.TxPool, cfg *config.EngineConfig, log logger.Logger) *Blockchain {
	bc := new(Blockchain)
	bc.txPool = txPool
	bc.log = log
	bc.cfg = cfg
	bc.lock = &sync.RWMutex{}
	bc.chl = &sync.RWMutex{}
	bc.rol = &sync.RWMutex{}
	bc.processLock = &sync.Mutex{}
	bc.chains = make(map[util.String]*Chain)
	bc.orphanBlocks = cache.NewCache(MaxOrphanBlocksCacheSize)
	bc.rejectedBlocks = cache.NewCache(MaxRejectedBlocksCacheSize)
	bc.eventEmitter = &emitter.Emitter{}
	return bc
}

// SetCoinbase sets the coinbase key that is used to
// identify the current blockchain instance
func (b *Blockchain) SetCoinbase(coinbase *crypto.Key) {
	b.coinbase = coinbase
}

// SetGenesisBlock sets the genesis block
func (b *Blockchain) SetGenesisBlock(block types.Block) {
	b.genesisBlock = block
}

// LoadBlockFromFile loads a block from a file
func LoadBlockFromFile(name string) (types.Block, error) {

	box := packr.NewBox("./data")
	data := box.Bytes(name)
	if len(data) == 0 {
		return nil, fmt.Errorf("block file not found")
	}

	var gBlock core.Block
	if err := json.Unmarshal(data, &gBlock); err != nil {
		return nil, err
	}

	return &gBlock, nil
}

// makeChainID creates a chain id
// Combination: blake2b(current nano time + initial block hash
// + pointer address of initial block)
func makeChainID(initialBlock types.Block) util.String {
	now := time.Now().Unix()
	blockHash := initialBlock.GetHashAsHex()
	id := fmt.Sprintf("%s %d %d", blockHash, now, util.GetPtrAddr(initialBlock))
	hash := util.BytesToHash(util.Blake2b256([]byte(id)))
	return util.String(hash.HexStr())
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

	// If at this point the genesis block has not
	// been set, we attempt to load it from the
	// genesis.json file.
	if b.genesisBlock == nil {
		b.genesisBlock, err = LoadBlockFromFile("genesis.json")
		if err != nil {
			return err
		}
	}

	// If there are no known chains described in the metadata and none
	// in the cache, then we create a new chain and save it
	if len(chains) == 0 {
		b.log.Info("No branch found. Creating genesis state")

		// Create the genesis chain and the genesis block.
		gBlock := b.genesisBlock
		if gBlock.GetNumber() != 1 {
			return fmt.Errorf("genesis block error: expected block number 1")
		}

		// Create and save the genesis chain
		gChainID := makeChainID(gBlock)
		gChain := NewChain(gChainID, b.db, b.cfg, b.log)
		if err := b.saveChain(gChain, "", 0); err != nil {
			return fmt.Errorf("failed to save genesis chain: %s", err)
		}

		// Process the genesis block.
		if _, err := b.maybeAcceptBlock(gBlock, gChain); err != nil {
			return fmt.Errorf("genesis block error: %s", err)
		}

		b.log.Info("Genesis state created",
			"Hash", gBlock.GetHash().SS(),
			"Difficulty", gBlock.GetHeader().GetDifficulty())
		return nil
	}

	// Load all known chains
	for _, chainInfo := range chains {
		if err := b.loadChain(chainInfo); err != nil {
			return err
		}
	}

	if numChains := len(chains); numChains > 0 {
		b.log.Info("Known branches have been loaded", "NumBranches", numChains)
	}

	// Set the root chain and make it the initial main chain
	b.bestChain = b.getRootChain()

	// Using the best chain rule, we mush select the best chain
	// and set it as the current bestChain.
	err = b.decideBestChain()
	if err != nil {
		return fmt.Errorf("failed to determine best chain: %s", err)
	}

	return nil
}

// getRootChain finds the root chain of the tree.
// This is usually the main chain and it has no branch.
func (b *Blockchain) getRootChain() *Chain {
	b.chl.RLock()
	defer b.chl.RUnlock()
	for _, c := range b.chains {
		if !c.HasParent() {
			return c
		}
	}
	return nil
}

// getBlockByHash finds and returns a block by hash only
// SetEventEmitter sets the event emitter
func (b *Blockchain) SetEventEmitter(ee *emitter.Emitter) {
	b.eventEmitter = ee
}

func (b *Blockchain) getBlockValidator(block types.Block) *BlockValidator {
	v := NewBlockValidator(block, b.txPool, b, b.cfg, b.log)
	v.setContext(types.ContextBlock)
	return v
}

// GetTxPool gets the transaction pool
func (b *Blockchain) GetTxPool() types.TxPool {
	return b.txPool
}

func (b *Blockchain) reOrgIsActive() bool {
	b.rol.RLock()
	defer b.rol.RUnlock()
	return b.reOrgActive
}

func (b *Blockchain) setReOrgStatus(active bool) {
	b.rol.Lock()
	defer b.rol.Unlock()
	b.reOrgActive = active
}

// loadChain finds and load a chain into the chain cache. It
// can be used to find both standalone chain and child chains.
func (b *Blockchain) loadChain(ci *core.ChainInfo) error {

	if ci == nil {
		return fmt.Errorf("chain info is required")
	}

	// Check whether the chain information is a genesis chain.
	// A genesis chain info does not include a parent chain id
	// and a parent block number since it has no parent.
	if ci.GetParentChainID() == "" && ci.GetParentBlockNumber() == 0 {
		b.addChain(NewChainFromChainInfo(ci, b.db, b.cfg, b.log))
		return nil
	}

	// Both parent chain ID and block number must be
	// set. We cannot allow one to only be set.
	if (ci.GetParentChainID() != "" && ci.GetParentBlockNumber() == 0) ||
		(ci.GetParentChainID() == "" && ci.GetParentBlockNumber() != 0) {
		return fmt.Errorf("chain load failed: chain parent chain ID and block are required")
	}

	// construct a new chain
	chain := NewChainFromChainInfo(ci, b.db, b.cfg, b.log)

	// Load the chain's parent chain and block
	_, err := chain.loadParent()
	if err != nil {
		return fmt.Errorf("chain load failed: %s", err)
	}

	// add chain to cache
	b.addChain(chain)

	return nil
}

// GetBestChain gets the chain that is currently considered the main chain
func (b *Blockchain) GetBestChain() types.Chainer {
	b.chl.RLock()
	defer b.chl.RUnlock()
	return b.bestChain
}

// SetBestChain sets the current main chain
func (b *Blockchain) SetBestChain(c *Chain) {
	b.chl.Lock()
	defer b.chl.Unlock()
	b.bestChain = c
}

// SetDB sets the database to use
func (b *Blockchain) SetDB(db elldb.DB) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.db = db
}

// GetDB gets the database
func (b *Blockchain) GetDB() elldb.DB {
	b.lock.RLock()
	defer b.lock.RUnlock()
	return b.db
}

func (b *Blockchain) removeChain(chain *Chain) {
	b.chl.Lock()
	defer b.chl.Unlock()
	if _, ok := b.chains[chain.GetID()]; ok {
		delete(b.chains, chain.GetID())
	}
	return
}

// findChainByBlockHash finds the chain where the given block
// hash exists. It returns the block, the chain, the header of
// highest block in the chain.
func (b *Blockchain) findChainByBlockHash(hash util.Hash,
	opts ...types.CallOp) (block types.Block, chain *Chain,
	chainTipHeader types.Header, err error) {

	b.chl.RLock()
	defer b.chl.RUnlock()
	chains := b.chains

	for _, chain := range chains {

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

// addRejectedBlock adds the block to the rejection cache.
func (b *Blockchain) addRejectedBlock(block types.Block) {
	b.rejectedBlocks.Add(block.GetHash().HexStr(), struct{}{})
}

// isRejected checks whether a block exists in the
// rejection cache
func (b *Blockchain) isRejected(block types.Block) bool {
	return b.rejectedBlocks.Has(block.GetHash().HexStr())
}

// addOrphanBlock adds a block to the collection of
// orphaned blocks.
func (b *Blockchain) addOrphanBlock(block types.Block) {
	// Insert the block to the cache with a 1 hour expiration
	b.orphanBlocks.AddWithExp(block.GetHash().HexStr(), block,
		time.Now().Add(time.Hour))

	b.log.Debug("Added block to orphan cache",
		"BlockNo", block.GetNumber(),
		"CacheSize", b.orphanBlocks.Len())
}

// isOrphanBlock checks whether a block is present in the
// collection of orphaned blocks.
func (b *Blockchain) isOrphanBlock(blockHash util.Hash) bool {
	return b.orphanBlocks.Get(blockHash.HexStr()) != nil
}

// NewChainFromChainInfo creates an instance of a Chain given a NewChainFromChainInfo
func (b *Blockchain) NewChainFromChainInfo(ci types.ChainInfo) *Chain {
	ch := NewChain(ci.GetID(), b.db, b.cfg, b.log)
	ch.info.ParentChainID = ci.GetParentChainID()
	ch.info.ParentBlockNumber = ci.GetParentBlockNumber()
	return ch
}

// findChainInfo finds information about chain
func (b *Blockchain) findChainInfo(chainID util.String) (*core.ChainInfo, error) {

	b.chl.RLock()
	chains := b.chains
	b.chl.RUnlock()

	// check whether the chain exists in the cache
	if chain, ok := chains[chainID]; ok {
		return chain.info, nil
	}

	var chainInfo core.ChainInfo

	// At this point, we did not find the chain in
	// the cache. We search the database instead.
	var chainKey = common.MakeKeyChain(chainID.Bytes())
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
func (b *Blockchain) IsMainChain(cr types.ChainReaderFactory) bool {
	b.chl.RLock()
	isMain := b.bestChain.GetID() == cr.GetID()
	b.chl.RUnlock()
	return isMain
}

// saveChain persist a given chain to the database.
// It will also cache the chain to support faster querying.
func (b *Blockchain) saveChain(chain *Chain, parentChainID util.String,
	parentBlockNumber uint64, opts ...types.CallOp) error {

	var txOp = common.GetTxOp(b.db, opts...)
	if txOp.Closed() {
		return leveldb.ErrClosed
	}

	chain.info = &core.ChainInfo{
		ID:                chain.GetID(),
		ParentBlockNumber: parentBlockNumber,
		ParentChainID:     parentChainID,
		Timestamp:         chain.info.Timestamp,
	}

	if err := chain.save(txOp); err != nil {
		return txOp.Rollback()
	}

	b.addChain(chain)

	return nil
}

// getChains gets all known chains from storage
func (b *Blockchain) getChains() (chainsInfo []*core.ChainInfo, err error) {
	chainsKey := common.MakeQueryKeyChains()
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

// hasChain checks whether a chain exists.
func (b *Blockchain) hasChain(chain *Chain) bool {
	b.chl.RLock()
	defer b.chl.RUnlock()
	_, ok := b.chains[chain.GetID()]
	return ok
}

// addChain adds a new chain to the list of chains.
func (b *Blockchain) addChain(chain *Chain) {
	b.chl.Lock()
	defer b.chl.Unlock()
	_, ok := b.chains[chain.GetID()]
	if !ok {
		b.chains[chain.GetID()] = chain
	}
	return
}

// newChain creates a new chain which may represent a fork.
//
// 'initialBlock' i block that will be added to this chain.
// It can be the genesis block or branch block.
//
// 'parentBlock' is the block that is the parent of the
// initialBlock. The parentBlock will be a block in another
// chain if this chain is being created in response
// to a new branch.
//
// While parentChain is the chain on which the parent block
// belongs to.
func (b *Blockchain) newChain(tx elldb.Tx, initialBlock types.Block,
	parentBlock types.Block, parentChain *Chain) (*Chain, error) {

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

	// Create a new chain.
	chainID := makeChainID(initialBlock)
	chain := NewChain(util.String(chainID), b.db, b.cfg, b.log)

	// Set the parent block and parent chain on the new chain.
	chain.parentBlock = parentBlock
	chain.parentChain = parentChain

	// keep a record of this chain in the store
	b.saveChain(chain, parentChain.GetID(), parentBlock.GetNumber(),
		&common.OpTx{Tx: tx, CanFinish: false})

	return chain, nil
}

// GetTransaction finds a transaction in the main chain and returns it
func (b *Blockchain) GetTransaction(hash util.Hash,
	opts ...types.CallOp) (types.Transaction, error) {

	b.chl.RLock()
	defer b.chl.RUnlock()
	var mainChain = b.bestChain

	if mainChain == nil {
		return nil, core.ErrBestChainUnknown
	}

	tx, err := mainChain.GetTransaction(hash, opts...)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

// ChainReader creates a chain reader to read the main chain
func (b *Blockchain) ChainReader() types.ChainReaderFactory {
	b.chl.RLock()
	defer b.chl.RUnlock()
	return NewChainReader(b.bestChain)
}

// GetChainReaderByHash returns a chain reader to a chain
// where a block with the given hash exists
func (b *Blockchain) GetChainReaderByHash(hash util.Hash) types.ChainReaderFactory {
	_, chain, _, _ := b.findChainByBlockHash(hash)
	if chain == nil {
		return nil
	}
	return chain.ChainReader()
}

// GetChainsReader gets chain reader for all known chains
func (b *Blockchain) GetChainsReader() (readers []types.ChainReaderFactory) {
	b.chl.RLock()
	defer b.chl.RUnlock()
	chains := b.chains
	for _, c := range chains {
		readers = append(readers, NewChainReader(c))
	}
	return
}

// GetLocators fetches a list of block hashes used to
// compare and sync the local chain with a remote chain.
// We collect the most recent 10 block hashes and
// then exponentially fetch more hashes until there are
// no more blocks.
// The genesis block must be added as the last hash
// if not already included.
func (b *Blockchain) GetLocators() ([]util.Hash, error) {

	b.chl.RLock()
	mainChain := b.bestChain
	b.chl.RUnlock()

	if mainChain == nil {
		return nil, core.ErrBestChainUnknown
	}

	curBlockHeader, err := mainChain.Current()
	if err != nil {
		return nil, fmt.Errorf("failed to get current block header: %s", err)
	}

	locators := []util.Hash{}
	topHeight := curBlockHeader.GetNumber()

	// first, we get the hashes of the last 10 blocks
	// using step of 1 backwards. After the last 10 blocks
	// are fetched, the step
	step := uint64(1)
	for i := topHeight; i > 0; i -= step {
		block, err := mainChain.GetBlock(i)
		if err != nil {
			if err != core.ErrBlockNotFound {
				return nil, err
			}
			break
		}
		locators = append(locators, block.GetHash())
		if len(locators) >= 10 {
			step *= 2
		}
	}

	// Add the genesis block hash as the final hash
	// if it isn't already the last hash.
	// We need to add the genesis block in case it got
	// missed during the above exponential selection.
	if !locators[len(locators)-1].Equal(b.genesisBlock.GetHash()) {
		locators = append(locators, b.genesisBlock.GetHash())
	}

	return locators, nil
}

// OrphanBlocks returns a cache reader for orphan blocks
func (b *Blockchain) OrphanBlocks() types.CacheReader {
	return b.orphanBlocks
}

// GetEventEmitter gets the event emitter
func (b *Blockchain) GetEventEmitter() *emitter.Emitter {
	return b.eventEmitter
}

// selectTransactions collects transactions from the head
// of the pool up to the specified maxSize.
func (b *Blockchain) selectTransactions(maxSize int64) (selectedTxs []types.Transaction,
	err error) {

	totalSelectedTxsSize := int64(0)
	cache := []types.Transaction{}
	nonces := make(map[util.String]uint64)
	for b.txPool.Size() > 0 {

		// Get a transaction from the top of
		// the pool
		tx := b.txPool.Container().First()

		// Check whether the addition of this
		// transaction will push us over the
		// size limit
		if totalSelectedTxsSize+tx.GetSizeNoFee() > maxSize {
			cache = append(cache, tx)

			// And also, if the amount of space left for new
			// transactions is less that the minimum
			// transaction size, then we exit immediately
			if maxSize-totalSelectedTxsSize < 230 {
				break
			}
			continue
		}

		// Check the current nonce value from
		// the cache and ensure the transaction's
		// nonce matches the expected/next nonce value.
		if nonce, ok := nonces[tx.GetFrom()]; ok {
			if (nonce + 1) != tx.GetNonce() {
				cache = append(cache, tx)
				continue
			}
			nonces[tx.GetFrom()] = tx.GetNonce()
		} else {
			// At this point, the nonce was not cached,
			// so we need to fetch it from the database,
			// add to the cache and ensure the
			// transaction's nonce matches the expected/next value
			nonces[tx.GetFrom()], err = b.GetAccountNonce(tx.GetFrom())
			if err != nil {
				return nil, err
			}
			if (nonces[tx.GetFrom()] + 1) != tx.GetNonce() {
				cache = append(cache, tx)
				continue
			}
			nonces[tx.GetFrom()] = tx.GetNonce()
		}

		// Add the transaction to the
		// selected tx slice and update the
		// total selected transactions size
		selectedTxs = append(selectedTxs, tx)
		totalSelectedTxsSize += tx.GetSizeNoFee()

		// Add the transaction back the cache
		// so it can be put back in the pool.
		cache = append(cache, tx)
	}

	// put the cached transactions back to the pool
	for _, tx := range cache {
		_ = b.txPool.Put(tx)
	}

	return
}
