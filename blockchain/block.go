package blockchain

import (
	"fmt"
	"math/big"
	"time"

	"github.com/shopspring/decimal"

	"github.com/ellcrys/elld/crypto"
	p "github.com/ellcrys/elld/params"
	"github.com/ellcrys/elld/types"

	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"
)

// HaveBlock checks whether we have a block matching
// the hash in any of the known chains
func (b *Blockchain) HaveBlock(hash util.Hash) (bool, error) {
	b.chainLock.RLock()
	chains := b.chains
	b.chainLock.RUnlock()
	for _, chain := range chains {
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

// getFeeAllocTx creates an allocation transaction
// with value equal to the sum of all fee of
// all transactions in a given block.
// The transaction will be awarded to the provide
// beneficiary.
func (b *Blockchain) getFeeAllocTx(block *core.Block, beneficiary *crypto.Key) *core.Transaction {

	// calculate total fees
	totalMinerFee := decimal.Zero
	for _, tx := range block.Transactions {
		if tx.Type != core.TxTypeAlloc {
			totalMinerFee = totalMinerFee.Add(tx.GetFee().Decimal())
		}
	}

	// create an alloc transaction
	tx := &core.Transaction{
		Type:         core.TxTypeAlloc,
		Nonce:        0,
		From:         beneficiary.PubKey().Addr(),
		To:           beneficiary.PubKey().Addr(),
		SenderPubKey: util.String(beneficiary.PubKey().Base58()),
		Value:        util.String(totalMinerFee.StringFixed(p.Decimals)),
		Fee:          "0",
		Timestamp:    time.Now().Unix(),
	}
	tx.Hash = tx.ComputeHash()
	sig, _ := core.TxSign(tx, beneficiary.PrivKey().Base58())
	tx.SetSignature(sig)

	return tx
}

// Generate produces a valid block for a target chain. By default
// the main chain is used but a different chain can be passed in
// as a CallOp.
func (b *Blockchain) Generate(params *types.GenerateBlockParams, opts ...types.CallOp) (types.Block, error) {

	var chain types.Chainer
	var block *core.Block

	if params == nil {
		return nil, fmt.Errorf("params is required")
	} else if params.Creator == nil {
		return nil, fmt.Errorf("creator's key is required")
	} else if params.Difficulty == nil || params.Difficulty.Cmp(util.Big0) == 0 {
		return nil, fmt.Errorf("difficulty is required")
	}

	// Determine if an explicit chain is to be used as
	// opposed to the main chain.
	chainerOp := common.GetChainerOp(opts...)
	chain = chainerOp.Chain

	// If an explicit chain has not been set, we use
	// the main chain
	if chain == nil && b.bestChain != nil {
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

	block = &core.Block{
		Header: &core.Header{
			ParentHash:       util.EmptyHash,
			CreatorPubKey:    util.String(params.Creator.PubKey().Base58()),
			Number:           1,
			TransactionsRoot: common.ComputeTxsRoot(params.Transactions),
			Nonce:            params.Nonce,
			Timestamp:        time.Now().Unix(),
			TotalDifficulty:  new(big.Int).SetInt64(0),
		},
	}

	for _, tx := range params.Transactions {
		block.Transactions = append(block.Transactions, tx.(*core.Transaction))
	}

	// override the total difficult if a
	// total difficulty is provided in the given params
	if params.OverrideTotalDifficulty != nil {
		block.Header.SetTotalDifficulty(params.OverrideTotalDifficulty)
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
		block.Header.SetDifficulty(params.Difficulty)
	}

	// select transactions and compute transaction root
	if len(params.Transactions) == 0 {
		selectedTxs, err := b.SelectTransactions(p.MaxBlockTxsSize)
		if err != nil {
			return nil, err
		}
		for _, tx := range selectedTxs {
			block.Transactions = append(block.Transactions, tx.(*core.Transaction))
		}
	}

	// If there are transactions in the block,
	// and AddFeeAlloc is true, we must add a fee allocation
	if len(block.Transactions) > 0 && params.AddFeeAlloc {
		block.Transactions = append(block.Transactions, b.getFeeAllocTx(block, params.Creator))
	}

	// Compute transactions root
	block.Header.SetTransactionsRoot(common.ComputeTxsRoot(block.GetTransactions()))

	// mock execute the transaction and set the new state root
	stateRoot, _, err := b.execBlock(chain, block)
	if err != nil {
		return nil, fmt.Errorf("exec: %s", err)
	}
	block.Header.SetStateRoot(stateRoot)

	// override state root if params include a state root
	if !params.OverrideStateRoot.IsEmpty() {
		block.Header.SetStateRoot(params.OverrideStateRoot)
	}

	// Sign the block using the creators private key
	sig, err := core.BlockSign(block, params.Creator.PrivKey().Base58())
	if err != nil {
		return nil, fmt.Errorf("failed to sign block: %s", err)
	}
	block.Sig = sig

	// Compute hash
	block.Hash = block.ComputeHash()

	return block, nil
}

// GetBlock finds a block in any chain with a matching
// block number and hash.
func (b *Blockchain) GetBlock(number uint64, hash util.Hash) (types.Block, error) {
	b.chainLock.RLock()
	chains := b.chains
	b.chainLock.RUnlock()
	for _, chain := range chains {
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
func (b *Blockchain) getBlockByHash(hash util.Hash, opts ...types.CallOp) (types.Block, error) {
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
func (b *Blockchain) GetBlockByHash(hash util.Hash, opts ...types.CallOp) (types.Block, error) {
	return b.getBlockByHash(hash, opts...)
}
