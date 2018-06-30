package blockchain

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/ellcrys/elld/blockchain/types"
	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/util/logger"
	"github.com/ellcrys/elld/wire"
)

// Chain represents a chain of blocks
type Chain struct {
	cfg       *config.EngineConfig
	chainLock *sync.RWMutex
	store     types.Store
	log       logger.Logger
}

// NewChain creates an instance of a chain
func NewChain(store types.Store, cfg *config.EngineConfig, log logger.Logger) (chain *Chain) {
	chain = new(Chain)
	chain.cfg = cfg
	chain.store = store
	chain.chainLock = &sync.RWMutex{}
	chain.log = log
	return
}

// init initializes a new chain. If there is no committed genesis block,
// it adds a genesis block to the store if it does not have one.
func (c *Chain) init(genesisBlockJSON string) error {

	c.chainLock.Lock()
	defer c.chainLock.Unlock()

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

		// 	// ensure it passes validation
		// 	if err := genBlock.Validate(); err != nil {
		// 		return fmt.Errorf("genesis block failed validation: %s", err)
		// 	}

		// 	// verify the block
		// 	if err := wire.BlockVerify(genBlock); err != nil {
		// 		return fmt.Errorf("genesis block signature is not valid: %s", err)
		// 	}
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

func (c *Chain) getCurrentBlockHeader() (*wire.Header, error) {
	var h wire.Header
	if err := c.store.GetBlockHeader(-1, &h); err != nil {
		return nil, err
	}
	return &h, nil
}

// getMatureTickets returns mature ticket transactions in the
// last n blocks.
// This method is safe for concurrent calls.
func (c *Chain) getMatureTickets(nLastBlocks int64) (mTxs []*wire.Transaction, err error) {

	c.chainLock.RLock()
	defer c.chainLock.RUnlock()

	// get the current block header. We need the block height.
	curBlockHeader, err := c.getCurrentBlockHeader()
	if err != nil {
		return nil, err
	}

	var startBlock int64
	var endBlock = int64(curBlockHeader.Number)

	// Set startBlock to 1 if the chain height is less or equal to nLastBlocks specified
	if int64(curBlockHeader.Number) <= nLastBlocks {
		startBlock = 1
	} else if int64(curBlockHeader.Number) > nLastBlocks {
		startBlock = int64(curBlockHeader.Number) - nLastBlocks + 1
	}

	// Find blocks within this range and get the matured endorsers
	for i := startBlock; i <= endBlock; i++ {

		var block wire.Block
		if err := c.store.GetBlock(startBlock, &block); err != nil {
			return nil, err
		}

		// find endorser ticket transactions and collect the
		// matured and non-revoked ones.
		for _, tx := range block.Transactions {
			if tx.Type == wire.TxTypeEndorserTicketCreate &&
				(curBlockHeader.Number-block.Header.Number) >= uint64(c.cfg.Consensus.NumBlocksForTicketMaturity) {
				mTxs = append(mTxs, tx)
			}
		}
	}

	return

}
