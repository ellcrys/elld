package blockchain

import (
	"fmt"
	"sync"
	"time"

	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/database"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/util/logger"
	"github.com/ellcrys/elld/wire"
)

// FuncOpts represents options to be passed to a function
type FuncOpts interface{}

// DBTxOpt is an option describing a transaction to be used by a function
type DBTxOpt struct {
	Tx          database.Tx
	CanFinalize bool
}

const (
	// MainChainID is the unique ID of the main chain
	MainChainID = "main"

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

	// store is the the database where block data and other meta data are stored
	store common.Store

	// bestChain is the chain considered to be the true chain.
	// It is protected by lock
	bestChain *Chain

	// chains holds all known chains
	chains []*Chain

	// orphanBlocks stores blocks whose parents are unknown
	orphanBlocks *Cache

	// rejectedBlocks stores collection of blocks that have been deemed invalid.
	// This allows us to quickly learn and discard blocks that are found here.
	rejectedBlocks *Cache
}

// New creates a Blockchain instance.
func New(cfg *config.EngineConfig, log logger.Logger) *Blockchain {
	bc := new(Blockchain)
	bc.log = log
	bc.cfg = cfg
	bc.chainLock = &sync.RWMutex{}
	bc.mLock = &sync.Mutex{}
	bc.orphanBlocks = NewCache(MaxOrphanBlocksCacheSize)
	bc.rejectedBlocks = NewCache(MaxRejectedBlocksCacheSize)
	return bc
}

// SetStore sets the store to use
func (b *Blockchain) SetStore(store common.Store) {
	b.store = store
}

// Up opens the database, initializes the store and
// creates the genesis block (if required)
func (b *Blockchain) Up() error {

	var err error

	if b.store == nil {
		return fmt.Errorf("store not set")
	}

	// get the blockchain main metadata
	meta := b.GetMeta()
	if meta == nil {
		meta = &common.BlockchainMeta{}
	}

	// If there are no known chains described in the metadata and none
	// in the cache, then we create a new chain.
	if len(meta.Chains) == 0 && len(b.chains) == 0 {

		b.log.Debug("No existing chain found. Creating genesis chain")

		// create the new chain
		b.bestChain, err = NewChain(MainChainID, b.store, b.cfg, b.log)
		if err != nil {
			return fmt.Errorf("failed to create new chain: %s", err)
		}

		// initialize the chain with the genesis block
		if err := b.bestChain.init(GenesisBlock); err != nil {
			return fmt.Errorf("failed to initialize new chain: %s", err)
		}

		// save the chain in the meta data
		meta.Chains = append(meta.Chains, &common.ChainInfo{
			ID:           b.bestChain.id,
			ParentNumber: 0,
		})
		if err := b.updateMeta(meta); err != nil {
			return fmt.Errorf("failed to save metadata: %s", err)
		}

		b.addChain(b.bestChain)
		return nil
	}

	// At this point, some chains already exists, so we must create
	// chain objects representing these chains.
	for _, c := range meta.Chains {
		chain, err := NewChain(c.ID, b.store, b.cfg, b.log)
		if err != nil {
			return fmt.Errorf(fmt.Sprintf("failed to load chain {%s}: %s", chain.id, err))
		}
		b.addChain(chain)
	}

	// Using the best chain rule, we mush select the best chain
	// and set it as the current bestChain.
	b.bestChain, err = b.chooseBestChain()
	if err != nil {
		return fmt.Errorf("failed to determine best chain: %s", err)
	}

	return nil
}

// hasChain checks whether a chain exists.
func (b *Blockchain) hasChain(chain *Chain) bool {
	b.chainLock.Lock()
	defer b.chainLock.Unlock()

	for _, c := range b.chains {
		if chain.id == c.id {
			return true
		}
	}

	return false
}

// addChain adds a new chain to the list of chains and saves
// a reference in the meta data. It returns an error if the chain already exists.
func (b *Blockchain) addChain(chain *Chain) error {

	if b.hasChain(chain) {
		return common.ErrChainAlreadyKnown
	}

	b.chainLock.Lock()
	b.chains = append(b.chains, chain)
	b.chainLock.Unlock()

	return nil
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

// HaveBlock checks whether we have a block in the
// main chain or other chains.
func (b *Blockchain) HaveBlock(hash string) (bool, error) {
	b.chainLock.RLock()
	defer b.chainLock.RUnlock()
	for _, chain := range b.chains {
		has, err := chain.hasBlock(hash)
		if err != nil {
			return false, err
		}
		if has {
			return true, err
		}
	}
	return false, nil
}

// findBlockChainByHash finds the chain where the block with the hash
// provided hash exist on. It also returns the header of highest block of the chain.
func (b *Blockchain) findBlockChainByHash(hash string) (block *wire.Block, chain *Chain, chainTipHeader *wire.Header, err error) {
	b.chainLock.RLock()
	defer b.chainLock.RUnlock()

	for _, chain := range b.chains {
		block, err := chain.getBlockByHash(hash)
		if err != nil {
			if err != common.ErrBlockNotFound {
				return nil, nil, nil, err
			}
			continue
		}

		// get the header of the highest block
		chainTipHeader, err := chain.getTipHeader()
		if err != nil {
			return nil, nil, nil, err
		}

		return block, chain, chainTipHeader, nil
	}

	return nil, nil, nil, nil
}

// chooseBestChain returns the chain that is considered the
// legitimate chain. For now, we will stick to the longest
// chain being the best.
// TODO: For hybrid mode, the longest chain with the most amount of endorser ticket value
// should be the main chain. If we are implementing GHOST protocol, then this becomes more complicated.
func (b *Blockchain) chooseBestChain() (*Chain, error) {
	b.chainLock.RLock()
	defer b.chainLock.RUnlock()

	var curBest *Chain
	var curHeight uint64
	for _, chain := range b.chains {
		header, err := chain.getTipHeader()
		if err != nil {
			if err == common.ErrBlockNotFound {
				continue
			}
			return nil, err
		}
		if header.Number > curHeight {
			curHeight = header.Number
			curBest = chain
		}
	}
	return curBest, nil
}

func (b *Blockchain) addRejectedBlock(block *wire.Block) {
	b.chainLock.Lock()
	defer b.chainLock.Unlock()
	b.rejectedBlocks.Add(block.GetHash(), struct{}{})
}

func (b *Blockchain) isRejected(block *wire.Block) bool {
	b.chainLock.RLock()
	defer b.chainLock.RUnlock()
	return b.rejectedBlocks.Has(block.GetHash())
}

// addOrphanBlock adds a block to the collection of orphaned blocks.
func (b *Blockchain) addOrphanBlock(block common.Block) {
	b.chainLock.Lock()
	defer b.chainLock.Unlock()
	// Insert the block to the cache with a 1 hour expiration
	b.orphanBlocks.AddWithExp(block.GetHash(), block, time.Now().Add(time.Hour))
}

// isOrphanBlock checks whether a block is present in the collection of orphaned blocks.
func (b *Blockchain) isOrphanBlock(blockHash string) bool {
	b.chainLock.Lock()
	defer b.chainLock.Unlock()
	return b.orphanBlocks.Get(blockHash) != nil
}

// newChain creates a new chain which represents a fork.
// staleBlock is the block that caused the need for a new chain and
// staleBlockParent is the parent of the stale block.
func (b *Blockchain) newChain(staleBlock, staleBlockParent *wire.Block) (*Chain, error) {

	// stale block and its parent must be provided. They must also
	// be related through the stableBlock referencing the parent block's hash.
	if staleBlock == nil {
		return nil, fmt.Errorf("stale block cannot be nil")
	} else if staleBlockParent == nil {
		return nil, fmt.Errorf("stale block parent cannot be nil")
	} else if staleBlock.Header.ParentHash != staleBlockParent.Hash {
		return nil, fmt.Errorf("stale block and parent are not related")
	}

	tx := b.store.NewTx()

	// create a new chain. Assign a unique and random id to it
	chain, err := NewChainWithTx(tx, util.RandString(32), b.store, b.cfg, b.log)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	chain.setParentBlock(staleBlockParent)

	// Stale blocks in a new tree are not a source of truth about the world state since
	// they only serve to increase network security. We do not need to waste storage by
	// keeping the transactions.
	staleBlock.Transactions = nil

	// add the stale block to the new chain.
	if err := chain.appendBlockWithTx(tx, staleBlock); err != nil {
		tx.Rollback()
		return nil, err
	}

	// update the blockchain-wide chain record to include
	// this new chain. This allows us to be able to quickly learn
	// about all the known chains
	b.chainLock.Lock()
	bMeta := b.GetMeta()
	bMeta.Chains = append(bMeta.Chains, &common.ChainInfo{ID: chain.id, ParentNumber: staleBlockParent.GetNumber()})
	if err := b.updateMetaWithTx(tx, bMeta); err != nil {
		tx.Rollback()
		b.chainLock.Unlock()
		return nil, err
	}
	b.chainLock.Unlock()

	// add the tree in the chain cache
	if err := b.addChain(chain); err != nil {
		tx.Rollback()
		return nil, err
	}

	return chain, tx.Commit()
}
