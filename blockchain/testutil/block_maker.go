package testutil

import (
	"math/big"
	"time"

	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/types/core/objects"
	"github.com/ellcrys/elld/util"
)

// MakeTestBlock creates a block and adds
// the transactions in the transactions pool
// attached to the blockchain instance
func MakeTestBlock(bc core.Blockchain, chain core.Chainer, gp *core.GenerateBlockParams) core.Block {
	blk, err := bc.Generate(gp, &common.ChainerOp{Chain: chain})
	if err != nil {
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
func MakeBlock(bc core.Blockchain, ch core.Chainer, sender, receiver *crypto.Key) core.Block {
	return MakeTestBlock(bc, ch, &core.GenerateBlockParams{
		Transactions: []core.Transaction{
			objects.NewTx(objects.TxTypeBalance, 1, util.String(sender.Addr()), sender, "0", "2.5", time.Now().UnixNano()),
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
func MakeBlockWithOnlyAllocTx(bc core.Blockchain, ch core.Chainer, sender, receiver *crypto.Key) core.Block {
	return MakeTestBlock(bc, ch, &core.GenerateBlockParams{
		Transactions: []core.Transaction{
			objects.NewTx(objects.TxTypeAlloc, 1, util.String(sender.Addr()), sender, "0", "2.5", time.Now().UnixNano()),
		},
		Creator:           sender,
		Nonce:             util.EncodeNonce(1),
		Difficulty:        new(big.Int).SetInt64(131072),
		OverrideTimestamp: time.Now().Unix(),
	})
}

// MakeBlockWithNoTx creates a block with no
// transaction in it.
func MakeBlockWithNoTx(bc core.Blockchain, ch core.Chainer, sender, receiver *crypto.Key) core.Block {
	return MakeTestBlock(bc, ch, &core.GenerateBlockParams{
		Transactions:      []core.Transaction{},
		Creator:           sender,
		Nonce:             util.EncodeNonce(1),
		Difficulty:        new(big.Int).SetInt64(131072),
		OverrideTimestamp: time.Now().Unix(),
	})
}

// MakeBlockWithSingleTx creates a block with
// only one balance transaction. Sender nonce
// is required
func MakeBlockWithSingleTx(bc core.Blockchain, ch core.Chainer, sender, receiver *crypto.Key, senderNonce uint64) core.Block {
	return MakeTestBlock(bc, ch, &core.GenerateBlockParams{
		Transactions: []core.Transaction{
			objects.NewTx(objects.TxTypeBalance, senderNonce, util.String(sender.Addr()), sender, "0", "2.5", time.Now().UnixNano()),
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
func MakeBlockWithBalanceTx(bc core.Blockchain, ch core.Chainer, sender, receiver *crypto.Key) core.Block {
	return MakeTestBlock(bc, ch, &core.GenerateBlockParams{
		Transactions: []core.Transaction{
			objects.NewTx(objects.TxTypeBalance, 1, util.String(receiver.Addr()), sender, "1", "2.5", time.Now().UnixNano()),
		},
		Creator:     sender,
		Nonce:       util.EncodeNonce(1),
		Difficulty:  new(big.Int).SetInt64(131072),
		AddFeeAlloc: true,
	})
}

// MakeBlockWithParentHash creates a block with one
// balance transaction and a given parent block hash
func MakeBlockWithParentHash(bc core.Blockchain, ch core.Chainer, sender, receiver *crypto.Key, parentHash util.Hash) core.Block {
	return MakeTestBlock(bc, ch, &core.GenerateBlockParams{
		Transactions: []core.Transaction{
			objects.NewTx(objects.TxTypeBalance, 1, util.String(receiver.Addr()), sender, "1", "2.5", time.Now().UnixNano()),
		},
		Creator:            sender,
		Nonce:              util.EncodeNonce(1),
		Difficulty:         new(big.Int).SetInt64(131072),
		OverrideParentHash: parentHash,
	})
}

// MakeBlockWithTotalDifficulty creates a block with one
// balance transaction and a given total difficulty
func MakeBlockWithTotalDifficulty(bc core.Blockchain, ch core.Chainer, sender, receiver *crypto.Key, td *big.Int) core.Block {
	return MakeTestBlock(bc, ch, &core.GenerateBlockParams{
		Transactions: []core.Transaction{
			objects.NewTx(objects.TxTypeBalance, 1, util.String(receiver.Addr()), sender, "1", "2.5", time.Now().UnixNano()),
		},
		Creator:                 sender,
		Nonce:                   util.EncodeNonce(1),
		Difficulty:              new(big.Int).SetInt64(131072),
		OverrideTotalDifficulty: td,
		AddFeeAlloc:             true,
	})
}
