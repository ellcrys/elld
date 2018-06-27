package blockchain

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/ellcrys/elld/blockchain/types"
	"github.com/ellcrys/elld/util/logger"
	"github.com/ellcrys/elld/wire"
)

// Chain represents a chain of blocks
type Chain struct {
	chainLock *sync.RWMutex
	store     types.Store
	log       logger.Logger
}

// NewChain creates an instance of a chain
func NewChain(store types.Store, log logger.Logger) (chain *Chain) {
	chain = new(Chain)
	chain.store = store
	chain.chainLock = &sync.RWMutex{}
	chain.log = log
	return
}

// init initializes a new chain. If there is no committed genesis block,
// it adds a genesis block to the store if it does not have one.
func (c *Chain) init(genesisBlockJSON string) error {

	err := c.store.GetBlock(1, &wire.Block{})
	if err != nil && err != types.ErrBlockNotFound {
		return fmt.Errorf("failed to check genesis block existence: %s", err)
	}

	var genBlock = &wire.Block{}

	// If genesis block does not exists, we must
	// create it from the content of GenesisBlock
	if err == types.ErrBlockNotFound {
		// Unmarshal the genesis block JSON data to wire.Block
		if err = json.Unmarshal([]byte(genesisBlockJSON), genBlock); err != nil {
			return fmt.Errorf("failed to unmarshal genesis block data: %s", err)
		}

		// ensure it passes validation
		if err := genBlock.Validate(); err != nil {
			return fmt.Errorf("genesis block failed validation: %s", err)
		}

		// verify the block
		if err := wire.BlockVerify(genBlock); err != nil {
			return fmt.Errorf("genesis block signature is not valid: %s", err)
		}
	}

	// Save the genesis block if has been set.
	if genBlock.Sig != "" {
		if err := c.store.PutBlock(genBlock); err != nil {
			return fmt.Errorf("failed to commit genesis block to store. %s", err)
		}
	}

	// genesisKey := crypto.NewKeyFromIntSeed(127465328937663)
	// fmt.Println(genesisKey.PubKey().Base58())

	return nil
}
