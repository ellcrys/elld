package blockchain

import (
	"fmt"
	"sync"

	"github.com/ellcrys/elld/blockchain/types"
	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/util/logger"
	"github.com/ellcrys/elld/wire"
)

// MainChainID is the unique ID of the main chain
const MainChainID = "main"

// BlockchainMetaName is the name of the meta information for the entire blockchain
const BlockchainMetaName = "blockchain_meta"

// Blockchain represents the Ellcrys blockchain. It provides
// functionalities for interacting with the underlying database
// and primitives.
type Blockchain struct {

	// lock is a general purpose lock for store, bestChain, chains etc
	lock *sync.RWMutex

	// mLock is used to lock methods that should be called completely atomically
	mLock *sync.Mutex

	// cfg is the client configuration
	cfg *config.EngineConfig

	// log is used for logging output
	log logger.Logger

	// store is the the database where block data and other meta data are stored
	store types.Store

	// bestChain is the chain considered to be the true chain.
	// It is protected by lock
	bestChain *Chain

	// chains holds all known chains
	chains []*Chain

	// orphanBlocks stores blocks whose parents are unknown
	orphanBlocks map[string]*wire.Block
}

// New creates a Blockchain instance.
func New(cfg *config.EngineConfig, log logger.Logger) *Blockchain {
	bc := new(Blockchain)
	bc.log = log
	bc.cfg = cfg
	bc.lock = &sync.RWMutex{}
	return bc
}

// SetStore sets the store to use
func (b *Blockchain) SetStore(store types.Store) {
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
	blockchainMeta, err := b.GetMeta()
	if err != nil {
		if err != types.ErrMetadataNotFound {
			return fmt.Errorf("failed to get blockchain meta: %s", err)
		}
	}

	knownChains := blockchainMeta.Chains

	// since no chain exists, we must create a chain and consider it the best chain.
	// Also add the new chain to the list of chains.
	if len(knownChains) == 0 {
		b.log.Debug("No existing chain found. Creating new chain")

		b.bestChain, err = NewChain(MainChainID, b.store, b.cfg, b.log)
		if err != nil {
			return fmt.Errorf("failed to create new chain: %s", err)
		}
		if err := b.bestChain.init(GenesisBlock); err != nil {
			return fmt.Errorf("failed to initialize new chain: %s", err)
		}

		b.addChain(b.bestChain)
		return nil
	}

	// at this point, some chains already exists, so we must create chain objects representing
	// these chains.
	for _, chainID := range knownChains {
		chain, err := NewChain(chainID, b.store, b.cfg, b.log)
		if err != nil {
			return fmt.Errorf(fmt.Sprintf("failed to load chain {%s}: %s", chain.id, err))
		}
		b.addChain(chain)
	}

	// using the best chain rule, we mush select the best chain
	// and set it as the current bestChain.
	b.bestChain, err = b.chooseBestChain()
	if err != nil {
		return fmt.Errorf("failed to determine best chain: %s", err)
	}

	return nil
}

// hasChain checks whether a chain exists.
func (b *Blockchain) hasChain(chain *Chain) bool {
	b.lock.Lock()
	defer b.lock.Unlock()

	for _, c := range b.chains {
		if chain.id == c.id {
			return true
		}
	}

	return false
}

// addChain adds a new chain to the list of chains.
// It returns an error if the chain already exists
func (b *Blockchain) addChain(chain *Chain) error {

	if b.hasChain(chain) {
		return types.ErrChainAlreadyKnown
	}

	b.lock.Lock()
	b.chains = append(b.chains, chain)
	b.lock.Unlock()

	return nil
}

// IsEndorser takes an address and checks whether it has an active endorser ticket
func (b *Blockchain) IsEndorser(address string) bool {
	return false
}

// GetMeta gets the information about the blockchain
func (b *Blockchain) GetMeta() (*types.BlockchainMeta, error) {
	var err error
	var meta types.BlockchainMeta
	err = b.store.GetMetadata(BlockchainMetaName, &meta)
	return &meta, err
}

// UpdateMeta updates the meta information of the blockchain
func (b *Blockchain) UpdateMeta(meta *types.BlockchainMeta) error {
	return b.store.UpdateMetadata(BlockchainMetaName, meta)
}

// HybridMode checks whether the blockchain is a point where hybrid consensus
// can be utilized. Hybrid consensus mode allows consensus and blocks processed differently
// from standard block processing model. This mode is activated when we reach a target block height.
func (b *Blockchain) HybridMode() (bool, error) {
	b.lock.RLock()
	defer b.lock.RUnlock()

	h, err := b.bestChain.getCurrentBlockHeader()
	if err != nil {
		return false, err
	}

	return h.Number >= b.cfg.Chain.TargetHybridModeBlock, nil
}

// HasBlock checks whether we have a block in the
// main chain or other chains.
func (b *Blockchain) HasBlock(hash string) (bool, error) {
	b.lock.RLock()
	defer b.lock.RUnlock()
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

// findChainByTipHash finds the chain whether the last block hash is the
// same as the @hash arg provided.
func (b *Blockchain) findChainByTipHash(hash string) (*Chain, error) {
	b.lock.RLock()
	defer b.lock.RUnlock()

	for _, chain := range b.chains {
		header, err := chain.getCurrentBlockHeader()
		if err != nil {
			if err == types.ErrBlockNotFound {
				continue
			}
			return nil, err
		}
		if header.ComputeHash() == hash {
			return chain, nil
		}
	}
	return nil, nil
}

// chooseBestChain returns the chain that is considered the
// legitimate chain. For now, we will stick to the longest
// chain being the best.
// TODO: For hybrid mode, the longest chain with the most amount of endorser ticket value
// should be the main chain. If we are implementing GHOST protocol, then this becomes more complicated.
func (b *Blockchain) chooseBestChain() (*Chain, error) {
	b.lock.RLock()
	defer b.lock.RUnlock()

	var curBest *Chain
	var curHeight uint64
	for _, chain := range b.chains {
		header, err := chain.getCurrentBlockHeader()
		if err != nil {
			if err == types.ErrBlockNotFound {
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
