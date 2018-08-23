package blockchain

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/wire"
)

// HaveBlock checks whether we have a block matching
// the hash in any of the known chains
func (b *Blockchain) HaveBlock(hash util.Hash) (bool, error) {
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

// IsKnownBlock checks whether a block with the has exists
// in at least one of all block chains and caches (e.g orphan)
func (b *Blockchain) IsKnownBlock(hash util.Hash) (bool, string, error) {
	b.chainLock.RLock()
	defer b.chainLock.RUnlock()
	var have bool
	var reason string

	have, err := b.HaveBlock(hash)
	if err != nil {
		return false, "", err
	} else if have {
		reason = "chain"
	}

	if !have {
		if have = b.isOrphanBlock(hash); have {
			reason = "orphan cache"
		}
	}

	return have, reason, nil
}

// Generate produces a valid block for a target chain. By default
// the main chain is used but a different chain can be passed in
// as a CallOp.
func (b *Blockchain) Generate(params *core.GenerateBlockParams, opts ...core.CallOp) (core.Block, error) {

	var chain *Chain
	var block *wire.Block

	if params == nil {
		return nil, fmt.Errorf("params is required")
	} else if len(params.Transactions) == 0 {
		return nil, fmt.Errorf("at least one transaction is required")
	} else if params.Creator == nil {
		return nil, fmt.Errorf("creator's key is required")
	} else if params.Difficulty == nil || params.Difficulty.Cmp(util.Big0) == 0 {
		return nil, fmt.Errorf("difficulty is required")
	}

	// Determine if an explicit chain is to be used as
	// opposed to the main chain.
	for _, opt := range opts {
		if _opt, ok := opt.(ChainOp); ok {
			chain = _opt.Chain
			break
		}
	}
	// If an explicit chain has not been set, we use
	// the main chain
	if chain == nil {
		chain = b.bestChain
	}

	// At this point, if we still don't have a target
	// chain, we return with an error
	if chain == nil {
		return nil, fmt.Errorf("target chain not set")
	}

	// Set chain tip number. Override it
	// if set in params.
	// Note: Only use in tests
	chainTipNumber := uint64(0)
	if params.OverrideChainTip > 0 {
		chainTipNumber = params.OverrideChainTip
	}

	// Get the latest block header
	chainTip, err := chain.GetBlock(chainTipNumber)
	if err != nil {
		if err != core.ErrBlockNotFound {
			return nil, err
		}
	}

	block = &wire.Block{
		Header: &wire.Header{
			ParentHash:       util.EmptyHash,
			CreatorPubKey:    util.String(params.Creator.PubKey().Base58()),
			Number:           1,
			TransactionsRoot: common.ComputeTxsRoot(params.Transactions),
			Nonce:            params.Nonce,
			Timestamp:        time.Now().Unix(),
			TotalDifficulty:  new(big.Int).SetInt64(0),
		},
		ChainReader: chain.ChainReader(),
	}

	for _, tx := range params.Transactions {
		block.Transactions = append(block.Transactions, tx.(*wire.Transaction))
	}

	// override the total difficult if a
	// total difficulty is provided in the given params
	if params.OverrideTotalDifficulty != nil {
		block.Header.TotalDifficulty = params.OverrideTotalDifficulty
	}

	// override the block's timestamp if a timestamp is
	// provided in the given param.
	if params.OverrideTimestamp > 0 {
		block.Header.SetTimestamp(params.OverrideTimestamp)
	}

	// If the chain has no tip block but it has a parent,
	// then we set the block's parent hash to the parent's hash
	// and set the block number to BlockNumber(parent) + 1
	if chainTip == nil && chain.GetParentBlock() != nil {
		block.Header.SetParentHash(chain.GetParentBlock().GetHash())
		block.Header.SetNumber(chain.GetParentBlock().GetNumber() + 1)
	}

	// If a the chain tip exists, we set the block's parent
	// hash to the tip's hash and set the block number to
	// BlockNumber(parent) + 1
	if chainTip != nil {
		block.Header.SetParentHash(chainTip.GetHash())
		block.Header.SetNumber(chainTip.GetNumber() + 1)
	}

	// override parent hash with the parent hash provided in
	// in the params.
	if !params.OverrideParentHash.IsEmpty() {
		block.Header.SetParentHash(params.OverrideParentHash)
	}

	// Override difficulty if provided in params
	if params.Difficulty != nil {
		block.Header.Difficulty = params.Difficulty
	}

	// mock execute the transaction and set the new state root
	root, _, err := b.execBlock(chain, block)
	if err != nil {
		return nil, fmt.Errorf("exec: %s", err)
	}
	block.Header.SetStateRoot(root)

	// override state root if params include a state root
	if !params.OverrideStateRoot.IsEmpty() {
		block.Header.SetStateRoot(params.OverrideStateRoot)
	}

	// Compute hash
	block.Hash = block.ComputeHash()

	// Sign the block using the creators private key
	sig, err := wire.BlockSign(block, params.Creator.PrivKey().Base58())
	if err != nil {
		return nil, fmt.Errorf("failed to sign block: %s", err)
	}

	block.Sig = sig

	// Finally, validate the block to ensure it meets every
	// requirement for a valid block.
	bv := NewBlockValidator(block, b.txPool, b, true, b.cfg, b.log)
	if errs := bv.Validate(); len(errs) > 0 {
		return nil, fmt.Errorf("failed final validation: %s", errs[0])
	}

	return block, nil
}

// GetBlock finds a block in any chain with a matching
// block number and hash.
func (b *Blockchain) GetBlock(number uint64, hash util.Hash) (core.Block, error) {
	b.chainLock.RLock()
	defer b.chainLock.RUnlock()
	for _, chain := range b.chains {
		block, err := chain.getBlockByNumberAndHash(number, hash)
		if err != nil {
			if err != core.ErrBlockNotFound {
				return nil, err
			}
			continue
		}
		return block, nil
	}
	return nil, core.ErrBlockNotFound
}

// getBlockByHash finds a block in any chain with a matching hash.
func (b *Blockchain) getBlockByHash(hash util.Hash, opts ...core.CallOp) (core.Block, error) {
	b.chainLock.RLock()
	defer b.chainLock.RUnlock()
	for _, chain := range b.chains {
		block, err := chain.getBlockByHash(hash, opts...)
		if err != nil {
			if err != core.ErrBlockNotFound {
				return nil, err
			}
			continue
		}
		return block, nil
	}
	return nil, core.ErrBlockNotFound
}

// GetBlockByHash finds a block in any chain with a matching hash.
func (b *Blockchain) GetBlockByHash(hash util.Hash) (core.Block, error) {
	return b.getBlockByHash(hash)
}
