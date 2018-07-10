package blockchain

import (
	"fmt"

	"github.com/ellcrys/elld/blockchain/types"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/wire"
	"github.com/shopspring/decimal"
)

// validateBlock handles block validation. A block that successfully
// passes this validation is considered safe to add to the chain.
func (b *Blockchain) validateBlock(block *wire.Block) error {
	if err := block.Validate(); err != nil {
		return fmt.Errorf("failed block validation: %s", err)
	}

	if err := wire.BlockVerify(block); err != nil {
		return fmt.Errorf("failed block signature verification: %s", err)
	}
	return nil
}

// addOp adds a transition operation object to the list of
// operations (ops). It a similar transition to op already exists,
// it will replaced by the new op.
// @ops 	The current list of operations to add to.
// @op 		The operation to be added
// @returns	A new slice of operations with op included
func addOp(ops []types.Transition, op types.Transition) []types.Transition {
	var newOps []types.Transition
	for _, _op := range ops {
		if !_op.Equal(op) {
			newOps = append(newOps, _op)
		}
	}
	return append(newOps, op)
}

// processBalanceTx processes a transaction. It returns series of transition
// operations that must be applied to effect the transaction.
// It accepts the following args:
//
// @tx: 	The transaction
// @ops: 	The list of current operations generated from other transactions of same block as tx.
//			We use ops to check the latest proposed operation of an account initiated by other transactions.
// @returns	A slice of transitions to be applied to the chain state or error if something bad happened.
func (b *Blockchain) processBalanceTx(tx *wire.Transaction, ops []types.Transition) ([]types.Transition, error) {

	var err error
	var txOps []types.Transition
	var senderAcctBalance = decimal.Zero
	var recipientAcctBalance = decimal.Zero

	// first, we check if we can determine the balances of the sender and recipient accounts
	// from OpNewAccountBalance operations by previous transactions. Because, if an account was
	// updated by an previous transaction, the new balance will be found in the ops list.
	for _, prevOp := range ops {
		// check for balance change for the sender
		if opNewBalance, yes := prevOp.(*types.OpNewAccountBalance); yes && opNewBalance.Address() == tx.From {
			senderAcctBalance, _ = util.StrToDecimal(opNewBalance.Amount)
		}
		// check for balance change for the recipient
		if opNewBalance, yes := prevOp.(*types.OpNewAccountBalance); yes && opNewBalance.Address() == tx.To {
			recipientAcctBalance, _ = util.StrToDecimal(opNewBalance.Amount)
		}
	}

	// if we are unable to learn about the sender's latest balance in the ops list,
	// then we can fetch the account
	if senderAcctBalance.Equals(decimal.Zero) {
		// find the sender account. Return error if sender account
		// does not exist. This should never happen here as the caller must
		// have validated all transactions in the containing block.
		senderAcct, err := b.GetAccount(tx.From)
		if err != nil {
			return nil, fmt.Errorf("failed to get sender's account: %s", err)
		}
		senderAcctBalance, _ = util.StrToDecimal(senderAcct.Balance)
	}

	// if we are unable to learn about the recipient's latest balance in the ops list,
	// then we can fetch the account
	if recipientAcctBalance.Equals(decimal.Zero) {
		// find the account of the recipient. If the recipient account does not
		// exists, then we must create a OpCreateAccount transition for the address
		recipientAcct, err := b.GetAccount(tx.To)
		if err != nil {
			if err != ErrAccountNotFound {
				return nil, fmt.Errorf("failed to retrieve recipient account: %s", err)
			}
			txOps = append(txOps, &types.OpCreateAccount{
				OpBase: &types.OpBase{Addr: tx.To},
			})
		}
		recipientAcctBalance, _ = util.StrToDecimal(recipientAcct.Balance)
	}

	// convert the amount to be sent to decimal
	sendingAmount, err := decimal.NewFromString(tx.Value)
	if err != nil {
		return nil, fmt.Errorf("sending amount error: %s", err)
	}

	// ensure the sender's account balance is sufficient for this transaction
	if senderAcctBalance.LessThan(sendingAmount) {
		return nil, fmt.Errorf("insufficient sender account balance")
	}

	// add an operation to set a new account balance for the sender
	txOps = append(txOps, &types.OpNewAccountBalance{
		OpBase: &types.OpBase{Addr: tx.From},
		Amount: senderAcctBalance.Sub(sendingAmount).String(),
	})

	// add an operation to set a new balance of the recipient
	txOps = append(txOps, &types.OpNewAccountBalance{
		OpBase: &types.OpBase{Addr: tx.To},
		Amount: recipientAcctBalance.Add(sendingAmount).String(),
	})

	return txOps, nil
}

// processTransactions executes all transactions in the block and verifies that
// the state of the state root after execution matches the one set in the block header.
//
// Expects the caller to ensure that all transactions are valid.
func (b *Blockchain) processTransactions(txs []*wire.Transaction) error {

	var ops []types.Transition

	// here we will process each transaction and attempt
	// to decide what should happen to the chain state by
	// producing transition objects.
	for _, tx := range txs {
		switch tx.Type {

		case wire.TxTypeBalance:
			newOps, err := b.processBalanceTx(tx, ops)
			if err != nil {
				return err
			}
			for _, op := range newOps {
				ops = addOp(ops, op)
			}
		}
	}

	return nil
}

// ProcessBlock takes a block and attempts to add it to the
// tip of the blockchain.
func (b *Blockchain) ProcessBlock(block *wire.Block) error {
	b.mLock.Lock()
	defer b.mLock.Unlock()

	// validate the content and format of the block as well as the signature.
	// if err := b.validateBlock(block); err != nil {
	// 	return nil
	// }

	blockHash := block.ComputeHash()

	b.log.Debug("Processing block", "Hash", block.ComputeHash())

	// do not process if it is a known orphan
	if _, exist := b.orphanBlocks[blockHash]; exist {
		return types.ErrOrphanBlock
	}

	// check if any of the chain have a block with the matching hash
	exists, err := b.HasBlock(blockHash)
	if err != nil {
		return fmt.Errorf("failed to check block existence: %s", err)
	}
	if exists {
		b.log.Debug("Block already exists", "Hash", blockHash)
		return types.ErrBlockExists
	}

	// find the chain to which this block can be appended to. We do this by finding the chain
	// where the most recent block is the parent of the block in question.
	// If we are unable to find a chain, the block is considered an orphan block.
	chain, err := b.findChainByTipHash(block.Header.ParentHash)
	if err != nil {
		return fmt.Errorf("failed to find chain: %s", err)
	} else if chain == nil {
		b.orphanBlocks[blockHash] = block
		return b.processOrphanBlocks()
	}

	// TODO: Perform checkpoint check

	// execute the block transactions and verify that the state root matches
	// the state root value set in the block header
	if err := b.processTransactions(block.Transactions); err != nil {
		return fmt.Errorf("failed to process transactions: rejected: %s", err)
	}

	return nil
}

// processOrphanBlocks
func (b *Blockchain) processOrphanBlocks() error {
	return nil
}
