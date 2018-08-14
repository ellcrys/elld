package blockchain

import (
	"fmt"
	"time"

	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/wire"
)

// HaveBlock checks whether we have a block in the
// main chain or other chains.
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
func (b *Blockchain) Generate(params *common.GenerateBlockParams, opts ...common.CallOp) (*wire.Block, error) {

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
	} else if params.MixHash.IsEmpty() {
		return nil, fmt.Errorf("mix hash is required")
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

	// Get the latest block header
	chainTip, err := chain.GetBlock(0)
	if err != nil {
		if err != common.ErrBlockNotFound {
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
			MixHash:          params.MixHash,
			Difficulty:       params.Difficulty,
			Timestamp:        time.Now().Unix(),
		},
		Transactions: params.Transactions,
	}

	// override the block's timestamp if a timestamp is
	// provided in the given param.
	if params.OverrideTimestamp > 0 {
		block.Header.Timestamp = params.OverrideTimestamp
	}

	// If the chain has no tip block but it has a parent,
	// then we set the block's parent hash to the parent's hash
	// and set the block number to BlockNumber(parent) + 1
	if chainTip == nil && chain.GetParentBlock() != nil {
		block.Header.ParentHash = chain.GetParentBlock().Hash
		block.Header.Number = chain.GetParentBlock().Header.Number + 1
	}

	// If a the chain tip exists, we set the block's parent
	// hash to the tip's hash and set the block number to
	// BlockNumber(parent) + 1
	if chainTip != nil {
		block.Header.ParentHash = chainTip.Hash
		block.Header.Number = chainTip.Header.Number + 1
	}

	// override parent hash with the parent hash provided in
	// in the params.
	if !params.OverrideParentHash.IsEmpty() {
		block.Header.ParentHash = params.OverrideParentHash
	}

	// mock execute the transaction and set the new state root
	root, _, err := b.execBlock(chain, block)
	if err != nil {
		return nil, fmt.Errorf("exec: %s", err)
	}
	block.Header.StateRoot = root

	// override state root if params include a state root
	if !params.OverrideStateRoot.IsEmpty() {
		block.Header.StateRoot = params.OverrideStateRoot
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
