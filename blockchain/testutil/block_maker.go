package testutil

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"
)

// MakeTestBlock creates a block and adds
// the transactions in the transactions pool
// attached to the blockchain instance
func MakeTestBlock(bc types.Blockchain, chain types.Chainer, gp *types.GenerateBlockParams) types.Block {
	blk, err := bc.Generate(gp, &common.ChainerOp{Chain: chain})
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	if !gp.NoPoolAdditionInTest && bc.GetTxPool() != nil {
		for _, tx := range blk.GetTransactions() {
			bc.GetTxPool().PutSilently(tx)
		}
	}
	return blk
}

// MakeBlock creates a block
func MakeBlock(bc types.Blockchain, ch types.Chainer, sender, receiver *crypto.Key) types.Block {
	return MakeTestBlock(bc, ch, &types.GenerateBlockParams{
		Transactions: []types.Transaction{
			core.NewTx(core.TxTypeBalance, 1, sender.Addr(), sender, "0", "2.5", time.Now().UnixNano()),
		},
		Creator:           sender,
		Nonce:             util.EncodeNonce(1),
		Difficulty:        new(big.Int).SetInt64(131072),
		OverrideTimestamp: time.Now().Unix(),
		AddFeeAlloc:       true,
	})
}

// MakeBlockWithOnlyAllocTx creates a block with
// only one allocation transaction
func MakeBlockWithOnlyAllocTx(bc types.Blockchain, ch types.Chainer, sender, receiver *crypto.Key) types.Block {
	return MakeTestBlock(bc, ch, &types.GenerateBlockParams{
		Transactions: []types.Transaction{
			core.NewTx(core.TxTypeAlloc, 1, sender.Addr(), sender, "0", "2.5", time.Now().UnixNano()),
		},
		Creator:           sender,
		Nonce:             util.EncodeNonce(1),
		Difficulty:        new(big.Int).SetInt64(131072),
		OverrideTimestamp: time.Now().Unix(),
	})
}

// MakeBlockWithNoTx creates a block with no
// transaction in it.
func MakeBlockWithNoTx(bc types.Blockchain, ch types.Chainer, sender, receiver *crypto.Key) types.Block {
	return MakeTestBlock(bc, ch, &types.GenerateBlockParams{
		Transactions:      []types.Transaction{},
		Creator:           sender,
		Nonce:             util.EncodeNonce(1),
		Difficulty:        new(big.Int).SetInt64(131072),
		OverrideTimestamp: time.Now().Unix(),
	})
}

// MakeBlockWithSingleTx creates a block with
// only one balance transaction. Sender nonce
// is required
func MakeBlockWithSingleTx(bc types.Blockchain, ch types.Chainer, sender, receiver *crypto.Key, senderNonce uint64) types.Block {
	return MakeTestBlock(bc, ch, &types.GenerateBlockParams{
		Transactions: []types.Transaction{
			core.NewTx(core.TxTypeBalance, senderNonce, sender.Addr(), sender, "0", "2.5", time.Now().UnixNano()),
		},
		Creator:           sender,
		Nonce:             util.EncodeNonce(1),
		Difficulty:        new(big.Int).SetInt64(131072),
		OverrideTimestamp: time.Now().Unix(),
		AddFeeAlloc:       true,
	})
}

// MakeBlockWithBalanceTx is like MakeBlockWithSingleTx
// but does not require a sender nonce
func MakeBlockWithBalanceTx(bc types.Blockchain, ch types.Chainer, sender, receiver *crypto.Key) types.Block {
	return MakeTestBlock(bc, ch, &types.GenerateBlockParams{
		Transactions: []types.Transaction{
			core.NewTx(core.TxTypeBalance, 1, receiver.Addr(), sender, "1", "2.5", time.Now().UnixNano()),
		},
		Creator:     sender,
		Nonce:       util.EncodeNonce(1),
		Difficulty:  new(big.Int).SetInt64(131072),
		AddFeeAlloc: true,
	})
}

// MakeBlockWithParentHash creates a block with one
// balance transaction and a given parent block hash
func MakeBlockWithParentHash(bc types.Blockchain, ch types.Chainer, sender, receiver *crypto.Key, parentHash util.Hash) types.Block {
	return MakeTestBlock(bc, ch, &types.GenerateBlockParams{
		Transactions: []types.Transaction{
			core.NewTx(core.TxTypeBalance, 1, receiver.Addr(), sender, "1", "2.5", time.Now().UnixNano()),
		},
		Creator:            sender,
		Nonce:              util.EncodeNonce(1),
		Difficulty:         new(big.Int).SetInt64(131072),
		OverrideParentHash: parentHash,
	})
}

// MakeBlockWithTotalDifficulty creates a block with one
// balance transaction and a given total difficulty
func MakeBlockWithTotalDifficulty(bc types.Blockchain, ch types.Chainer, sender, receiver *crypto.Key, td *big.Int) types.Block {
	return MakeTestBlock(bc, ch, &types.GenerateBlockParams{
		Transactions: []types.Transaction{
			core.NewTx(core.TxTypeBalance, 1, receiver.Addr(), sender, "1", "2.5", time.Now().UnixNano()),
		},
		Creator:                 sender,
		Nonce:                   util.EncodeNonce(1),
		Difficulty:              new(big.Int).SetInt64(131072),
		OverrideTotalDifficulty: td,
		AddFeeAlloc:             true,
	})
}
