// Package blockchain provides functionalities for
// creating, managing and accessing the blockchain state.
package blockchain

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/ellcrys/elld/crypto"

	"github.com/gobuffalo/packr"

	"github.com/olebedev/emitter"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util/cache"
	"github.com/ellcrys/elld/util/logger"
)

const (
	// MaxOrphanBlocksCacheSize is the number of blocks we can keep in the orphan block cache
	MaxOrphanBlocksCacheSize = 500

	// MaxRejectedBlocksCacheSize is the number of blocks we can keep in the rejected block cache
	MaxRejectedBlocksCacheSize = 100
)

var (
	// GenesisBlockFileName is the name of the file that contains the genesis block
	GenesisBlockFileName = "genesis.json"
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
}

// New creates a Blockchain instance.
func New(txPool types.TxPool, cfg *config.EngineConfig, log logger.Logger) *Blockchain {
	bc := new(Blockchain)
	bc.txPool = txPool
	bc.log = log
	bc.lock = &sync.RWMutex{}
	bc.cfg = cfg
	bc.processLock = &sync.Mutex{}
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

// Up opens the database, initializes the store and
// creates the genesis block (if required)
func (b *Blockchain) Up() error {
	return nil
}

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

// SetDB sets the database to use
func (b *Blockchain) SetDB(db elldb.DB) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.db = db
}
