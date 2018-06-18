package blockchain

import (
	"fmt"

	"github.com/ellcrys/elld/blockchain/types"
	"github.com/ellcrys/elld/configdir"
	"github.com/ellcrys/elld/util/logger"
)

// Blockchain represents the Ellcrys blockchain. It provides
// functionalities for interacting with the underlying database
// and primitives.
type Blockchain struct {
	cfg   *configdir.Config // Node configuration
	log   logger.Logger     // Logger
	store types.Store
}

// New creates a Blockchain instance.
func New(cfg *configdir.Config, log logger.Logger) *Blockchain {
	bc := new(Blockchain)
	bc.log = log
	bc.cfg = cfg
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

	b.createGenesisBlock()

	return nil
}

func (b *Blockchain) createGenesisBlock() error {
	return nil
}
