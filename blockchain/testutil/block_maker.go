package testutil

import (
	"math/big"
	"time"

	"github.com/ellcrys/mother/crypto"
	"github.com/ellcrys/mother/types"
	"github.com/ellcrys/mother/types/core"
	"github.com/ellcrys/mother/util"
)

// MakeTestBlock creates a block and adds
// the transactions in the transactions pool
// attached to the blockchain instance
func MakeTestBlock(bc types.Blockchain, chain types.Chainer,
	gp *types.GenerateBlockParams) types.Block {
	// blk, err := bc.Generate(gp, &common.OpChainer{Chain: chain})
	// if err != nil {
	// 	panic(err)
	// }
	// if !gp.NoPoolAdditionInTest && bc.GetTxPool() != nil {
	// 	for _, tx := range blk.GetTransactions() {
	// 		bc.GetTxPool().Put(tx)
	// 	}
	// }
	// return blk
	return nil
}

// MakeBlock creates a block
func MakeBlock(bc types.Blockchain, ch types.Chainer, sender, receiver *crypto.Key) types.Block {
	return MakeTestBlock(bc, ch, &types.GenerateBlockParams{
		Transactions: []types.Transaction{
			core.NewTx(core.TxTypeBalance, 1, sender.Addr(), sender, "0", "2.5",
				time.Now().UnixNano()),
		},
		Creator:           sender,
		Nonce:             util.EncodeNonce(1),
		Difficulty:        new(big.Int).SetInt64(131072),
		OverrideTimestamp: time.Now().Unix(),
		AddFeeAlloc:       true,
	})
}

// MakeBlockWithTx creates a block with only one balance transaction.
// The sender param is used as the transaction sender and receiver.
// The sender nonce must be consistent with the provided chain.
func MakeBlockWithTx(bc types.Blockchain, ch types.Chainer, sender *crypto.Key,
	senderNonce uint64) types.Block {
	return MakeTestBlock(bc, ch, &types.GenerateBlockParams{
		Transactions: []types.Transaction{
			core.NewTx(core.TxTypeBalance, senderNonce, sender.Addr(), sender, "0", "2.5",
				time.Now().UnixNano()),
		},
		Creator:           sender,
		Nonce:             util.EncodeNonce(1),
		Difficulty:        new(big.Int).SetInt64(131072),
		OverrideTimestamp: time.Now().Unix(),
		AddFeeAlloc:       true,
	})
}

// MakeBlockWithTxAndReceiver is like MakeBlockWithTx but it also accepts
// a receiver address.
func MakeBlockWithTxAndReceiver(bc types.Blockchain, ch types.Chainer, sender, receiver *crypto.Key,
	senderNonce uint64) types.Block {
	return MakeTestBlock(bc, ch, &types.GenerateBlockParams{
		Transactions: []types.Transaction{
			core.NewTx(core.TxTypeBalance, senderNonce, receiver.Addr(), sender, "0", "2.5",
				time.Now().UnixNano()),
		},
		Creator:           sender,
		Nonce:             util.EncodeNonce(1),
		Difficulty:        new(big.Int).SetInt64(131072),
		OverrideTimestamp: time.Now().Unix(),
		AddFeeAlloc:       true,
	})
}

// MakeBlockWithTxAndTime creates a block with only one balance transaction.
// It overrides the block time using blockTime.
// The sender nonce must be consistent with the provided chain.
func MakeBlockWithTxAndTime(bc types.Blockchain, ch types.Chainer, sender *crypto.Key,
	senderNonce uint64, blockTime int64) types.Block {
	return MakeTestBlock(bc, ch, &types.GenerateBlockParams{
		Transactions: []types.Transaction{
			core.NewTx(core.TxTypeBalance, senderNonce, sender.Addr(), sender, "0", "2.5",
				time.Now().UnixNano()),
		},
		Creator:           sender,
		Nonce:             util.EncodeNonce(1),
		Difficulty:        new(big.Int).SetInt64(131072),
		OverrideTimestamp: blockTime,
		AddFeeAlloc:       true,
	})
}

// MakeBlockWithTxNotInPool is like MakeBlockWithTx
// but does not add the transactions in the pool
func MakeBlockWithTxNotInPool(bc types.Blockchain, ch types.Chainer,
	sender *crypto.Key) types.Block {
	return MakeTestBlock(bc, ch, &types.GenerateBlockParams{
		Transactions: []types.Transaction{
			core.NewTx(core.TxTypeBalance, 1, sender.Addr(), sender, "1", "2.5",
				time.Now().UnixNano()),
		},
		Creator:              sender,
		Nonce:                util.EncodeNonce(1),
		Difficulty:           new(big.Int).SetInt64(131072),
		AddFeeAlloc:          true,
		NoPoolAdditionInTest: true,
	})
}

// MakeBlockWithParentHash creates a block with one
// balance transaction and a given parent block hash
func MakeBlockWithParentHash(bc types.Blockchain, ch types.Chainer, sender *crypto.Key,
	parentHash util.Hash) types.Block {
	return MakeTestBlock(bc, ch, &types.GenerateBlockParams{
		Transactions: []types.Transaction{
			core.NewTx(core.TxTypeBalance, 1, sender.Addr(), sender, "1", "2.5",
				time.Now().UnixNano()),
		},
		Creator:            sender,
		Nonce:              util.EncodeNonce(1),
		Difficulty:         new(big.Int).SetInt64(131072),
		OverrideParentHash: parentHash,
	})
}

// MakeBlockWithTotalDifficulty creates a block with one
// balance transaction and a given total difficulty
func MakeBlockWithTotalDifficulty(bc types.Blockchain, ch types.Chainer, sender *crypto.Key,
	td *big.Int) types.Block {
	return MakeTestBlock(bc, ch, &types.GenerateBlockParams{
		Transactions: []types.Transaction{
			core.NewTx(core.TxTypeBalance, 1, sender.Addr(), sender, "1", "2.5",
				time.Now().UnixNano()),
		},
		Creator:                 sender,
		Nonce:                   util.EncodeNonce(1),
		Difficulty:              new(big.Int).SetInt64(131072),
		OverrideTotalDifficulty: td,
		AddFeeAlloc:             true,
	})
}

// MakeBlockWithTDAndNonce creates a block with one
// transaction, a given total difficulty and sender tx nonce
func MakeBlockWithTDAndNonce(bc types.Blockchain, ch types.Chainer, sender *crypto.Key,
	senderNonce uint64, td *big.Int) types.Block {
	return MakeTestBlock(bc, ch, &types.GenerateBlockParams{
		Transactions: []types.Transaction{
			core.NewTx(core.TxTypeBalance, senderNonce, sender.Addr(), sender, "1", "2.5",
				time.Now().UnixNano()),
		},
		Creator:                 sender,
		Nonce:                   util.EncodeNonce(1),
		Difficulty:              new(big.Int).SetInt64(131072),
		OverrideTotalDifficulty: td,
		AddFeeAlloc:             true,
	})
}
