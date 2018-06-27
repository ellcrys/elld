package blockchain

import (
	"fmt"
	"sync"

	"github.com/ellcrys/elld/blockchain/types"
	"github.com/ellcrys/elld/configdir"
	"github.com/ellcrys/elld/util/logger"
)

// Blockchain represents the Ellcrys blockchain. It provides
// functionalities for interacting with the underlying database
// and primitives.
type Blockchain struct {
	lock      *sync.Mutex
	cfg       *configdir.Config // Node configuration
	log       logger.Logger     // Logger
	store     types.Store       // The database where block data is stored
	bestChain *Chain            // The chain considered to be the true chain
}

// New creates a Blockchain instance.
func New(cfg *configdir.Config, log logger.Logger) *Blockchain {
	bc := new(Blockchain)
	bc.log = log
	bc.cfg = cfg
	bc.lock = &sync.Mutex{}
	return bc
}

// SetStore sets the store to use
func (b *Blockchain) SetStore(store types.Store) {
	b.store = store
}

// Up opens the database, initializes the store and
// creates the genesis block (if required)
func (b *Blockchain) Up() error {

	if b.store == nil {
		return fmt.Errorf("store not set")
	}

	b.log.Info("Initializing blockchain store")
	if err := b.store.Initialize(); err != nil {
		return err
	}

	b.lock.Lock()
	defer b.lock.Unlock()

	b.bestChain = NewChain(b.store, b.log)
	if err := b.bestChain.init(GenesisBlock); err != nil {
		b.log.Debug("best chain initialization: %s", "Err", err)
		return err
	}

	return nil
}

// IsEndorser takes an address and checks whether it has an active endorser ticket
func (b *Blockchain) IsEndorser(address string) bool {
	return false
}
