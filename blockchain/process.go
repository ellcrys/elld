package blockchain

import (
	"bytes"
	"fmt"

	"github.com/ellcrys/elld/elldb"

	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/wire"
	"github.com/shopspring/decimal"
)

// validateBlock handles block validation. A block that successfully
// passes this validation is considered safe to add to the chain.
func (b *Blockchain) validateBlock(block *wire.Block) error {

	blockValidator := NewBlockValidator(block, b.txPool, b, true)
	if errs := blockValidator.Validate(); len(errs) > 0 {
		return errs[0]
	}

	// TODO: move this to the block validator
	// validate the transaction root
	if !block.Header.TransactionsRoot.Equal(common.ComputeTxsRoot(block.Transactions)) {
		return fmt.Errorf("failed transaction root check")
	}

	return nil
}

// addOp adds a transition operation object to the list of
// operations (ops). It a similar transition to op already exists,
// it will replaced by the new op.
// @ops 	The current list of operations to add to.
// @op 		The operation to be added
// @returns	A new slice of operations with op included
func addOp(ops []common.Transition, op common.Transition) []common.Transition {
	var newOps []common.Transition
	for _, _op := range ops {
		if !_op.Equal(op) {
			newOps = append(newOps, _op)
		}
	}
	return append(newOps, op)
}

// processBalanceTx processes a TxTypeBalance transaction. It returns series of transition
// operations that must be applied to effect changes the world state and transaction accounts.
// It accepts the following args:
//
// @tx: 	The transaction
// @ops: 	The list of recent operations generated from other transactions of same block as tx.
//			We use ops to check the latest and uncommitted  operations of an account derived from other transactions.
// @returns	A slice of transitions to be applied to the chain state or error if something bad happened.
func (b *Blockchain) processBalanceTx(tx *wire.Transaction, ops []common.Transition, chain common.Chainer, opts ...common.CallOp) ([]common.Transition, error) {
	var err error
	var txOps []common.Transition
	var senderAcct, recipientAcct *wire.Account
	var senderAcctBalance = decimal.Zero
	var recipientAcctBalance = decimal.Zero

	// first, we check if we can determine the balances of the sender and recipient accounts
	// from OpNewAccountBalance operations by previous transactions. If an account was
	// updated by a previous transaction, the new balance will be found in the ops list.
	for _, prevOp := range ops {
		// check for balance change for the sender
		if opNewBalance, yes := prevOp.(*common.OpNewAccountBalance); yes && opNewBalance.Address() == tx.From {
			senderAcctBalance, _ = util.StrToDecimal(opNewBalance.Account.Balance)
		}
		// check for balance change for the recipient
		if opNewBalance, yes := prevOp.(*common.OpNewAccountBalance); yes && opNewBalance.Address() == tx.To {
			recipientAcctBalance, _ = util.StrToDecimal(opNewBalance.Account.Balance)
		}
	}

	// find the sender account. Return error if sender account
	// does not exist. This should never happen here as the caller must
	// have validated all transactions in the containing block.
	senderAcct, err = b.getAccount(chain, tx.From, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to get sender's account: %s", err)
	}

	// if we were unable to learn about the sender's latest balance from the ops list
	// as a result of previous transactions in same block, then we use the current account balance.
	if senderAcctBalance.Equals(decimal.Zero) {
		senderAcctBalance, _ = util.StrToDecimal(senderAcct.Balance)
	}

	// find the account of the recipient. If the recipient account does not
	// exists, then we must create a OpCreateAccount transition to instruct the creation of a new account.
	recipientAcct, err = b.getAccount(chain, tx.To, opts...)
	if err != nil {
		if err != common.ErrAccountNotFound {
			return nil, fmt.Errorf("failed to retrieve recipient account: %s", err)
		}
		txOps = append(txOps, &common.OpCreateAccount{
			OpBase: &common.OpBase{Addr: tx.To},
			Account: &wire.Account{
				Type:    wire.AccountTypeBalance,
				Address: tx.To,
				Balance: "0",
			},
		})
		recipientAcct = &wire.Account{
			Type:    wire.AccountTypeBalance,
			Address: tx.To,
			Balance: "0",
		}
	}

	// if we are unable to learn about the recipient's latest balance from the ops list as
	// then we can use the balance of the recipient account
	if recipientAcctBalance.Equals(decimal.Zero) {
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
	senderAcct.Balance = senderAcctBalance.Sub(sendingAmount).String()
	txOps = append(txOps, &common.OpNewAccountBalance{
		OpBase:  &common.OpBase{Addr: tx.From},
		Account: senderAcct,
	})

	// add an operation to set a new balance of the recipient
	recipientAcct.Balance = recipientAcctBalance.Add(sendingAmount).String()
	txOps = append(txOps, &common.OpNewAccountBalance{
		OpBase:  &common.OpBase{Addr: tx.To},
		Account: recipientAcct,
	})

	return txOps, nil
}

// processAllocCoinTx process a TxTypeAllocCoin which request the creation of a new
// account as well as allocation of some value to that account.
// @tx: 	The transaction
// @ops: 	The list of recent operations generated from other transactions of same block as tx.
//			We use ops to check the latest and uncommitted  operations of an account derived from other transactions.
// @returns	A slice of transitions to be applied to the chain state or error if something bad happened.
func (b *Blockchain) processAllocCoinTx(tx *wire.Transaction, ops []common.Transition, chain common.Chainer, opts ...common.CallOp) ([]common.Transition, error) {
	var err error
	var txOps []common.Transition
	var recipientAcct *wire.Account
	var recipientAcctBalance = decimal.Zero

	// first, we check if we can determine the balances of the recipient account
	// from OpNewAccountBalance operations by previous transactions. If an account was
	// updated by a previous transaction, the new balance will be found in the ops list.
	for _, prevOp := range ops {
		if opNewBalance, yes := prevOp.(*common.OpNewAccountBalance); yes && opNewBalance.Address() == tx.To {
			recipientAcctBalance, _ = util.StrToDecimal(opNewBalance.Account.Balance)
		}
	}

	// find the account of the recipient. If the account does not exists,
	// initialize a new account object for the recipient
	recipientAcct, err = b.getAccount(chain, tx.To, opts...)
	if err != nil {
		if err != common.ErrAccountNotFound {
			return nil, fmt.Errorf("failed to retrieve recipient account: %s", err)
		}
		recipientAcct = &wire.Account{
			Type:    wire.AccountTypeBalance,
			Address: tx.To,
			Balance: "0",
		}
	}

	// If the recipients account balance is zero and unchanged
	// in previous ops execution, we set it to the value of the current,
	// committed account balance
	if recipientAcctBalance.Equals(decimal.Zero) {
		recipientAcctBalance, _ = util.StrToDecimal(recipientAcct.Balance)
	}

	// Update the recipients account balance to be the
	// sum of current balance and the new allocation
	recipientAcct.Balance = recipientAcctBalance.Add(util.StrToDec(tx.Value)).StringFixed(16)

	// construct an OpNewAccountBalance transition object
	// and set the account to the updated recipient.
	txOps = append(txOps, &common.OpNewAccountBalance{
		OpBase:  &common.OpBase{Addr: tx.To},
		Account: recipientAcct,
	})

	return txOps, nil
}

// opsToKVObjects takes a slice of operations and apply them to the provided chain
func (b *Blockchain) opsToStateObjects(block *wire.Block, chain common.Chainer, ops []common.Transition) ([]*common.StateObject, error) {

	stateObjs := []*common.StateObject{}

	for _, op := range ops {
		switch _op := op.(type) {

		case *common.OpCreateAccount:
			stateObjs = append(stateObjs, &common.StateObject{
				TreeKey: common.MakeTreeKey(block.GetNumber(), common.ObjectTypeAccount),
				Key:     common.MakeAccountKey(block.GetNumber(), chain.GetID(), _op.Address()),
				Value:   util.ObjectToBytes(_op.Account),
			})

		case *common.OpNewAccountBalance:
			stateObjs = append(stateObjs, &common.StateObject{
				TreeKey: common.MakeTreeKey(block.GetNumber(), common.ObjectTypeAccount),
				Key:     common.MakeAccountKey(block.GetNumber(), chain.GetID(), _op.Address()),
				Value:   util.ObjectToBytes(_op.Account),
			})

		default:
			return nil, fmt.Errorf("unknown transition sub-type")
		}
	}

	return stateObjs, nil
}

// processTransactions computes the operations that must be applied to the
// hash tree and world state.
func (b *Blockchain) processTransactions(txs []*wire.Transaction, chain common.Chainer, opts ...common.CallOp) ([]common.Transition, error) {

	var ops []common.Transition

	// here we will process each transaction and attempt
	// to decide what should happen to the chain state by
	// producing transition objects.
	for i, tx := range txs {
		var err error
		var newOps []common.Transition

		switch tx.Type {
		case wire.TxTypeBalance:
			newOps, err = b.processBalanceTx(tx, ops, chain, opts...)
		case wire.TxTypeAllocCoin:
			newOps, err = b.processAllocCoinTx(tx, ops, chain, opts...)
		}

		if err != nil {
			return nil, fmt.Errorf("index{%d}: %s", i, err)
		}

		for _, op := range newOps {
			ops = addOp(ops, op)
		}
	}

	return ops, nil
}

// maybeAcceptBlock attempts to determine the suitable chain for the
// provided block, execute the block's transactions to derive the
// new set of state objects. The state objects and transactions are
// then stored to the store.
//
// If a chain is passed as parameter, no attempt to determine the chain
// is taken. Instead, the block will be processed for inclusion in the
// passed chain. This should only be used for the genesis block.
//
// NOTE: This method must be called with chain lock held by the caller.
func (b *Blockchain) maybeAcceptBlock(block *wire.Block, chain *Chain) (*Chain, error) {

	var parentBlock *wire.Block
	var chainTip *wire.Header
	var err error

	if chain == nil {
		// find the chain where the parent of the block exists on. If a chain is not found,
		// then the block is considered an orphan. If the chain is found but the block at the tip
		// is has the same or a greater block number compared to the new block, it is considered a stale block.
		parentBlock, chain, chainTip, err = b.findChainByBlockHash(block.Header.ParentHash.HexStr())
		if err != nil {
			if err != common.ErrBlockNotFound {
				return nil, err
			}
			b.log.Debug("Block not compatible with any chain", "Err", err.Error())

		} else if chain == nil {
			b.addOrphanBlock(block)
			return nil, nil

		} else if block.Header.Number < chainTip.Number {
			// This is a much older stale block. We only support stale blocks of same height
			// as the current block on the chain.
			// TODO: This should be a fork too; New chain should be created
			b.addRejectedBlock(block)
			return nil, common.ErrVeryStaleBlock

		} else if block.GetNumber()-chainTip.Number > 1 {
			b.addRejectedBlock(block)
			return nil, common.ErrBlockFailedValidation
		}
	}

	tx, _ := chain.store.NewTx()
	txOp := common.TxOp{Tx: tx, CanFinish: false}

	// If the block number is the same as the chainTip, then this
	// is a fork and as such creates a new chain.
	if chainTip != nil && block.GetNumber() == chainTip.Number {
		// create the new chain, set its root to the parent of the forked block
		if chain, err = b.newChain(tx, block, parentBlock, chain); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create subtree out of stale block")
		}
	}

	// Execute block to derive the state objects and the resulting
	// state root should the state object be applied to the blockchain state tree.
	newStateRoot, stateObjs, err := b.execBlock(chain, block, txOp)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// Compare the state root in the block header with the root obtained
	// from the mock execution of the block.
	if !block.Header.StateRoot.Equal(newStateRoot) {
		tx.Rollback()
		return nil, common.ErrBlockStateRootInvalid
	}

	// Next we need to update the blockchain objects in the store
	// as described by the state objects
	var batchObjs []*elldb.KVObject
	for _, so := range stateObjs {
		batchObjs = append(batchObjs, elldb.NewKVObject(so.Key, so.Value))
	}
	if err := txOp.Tx.Put(batchObjs); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to add state object to store: %s", err)
	}

	// At this point, the block is good to go. We add it to the chain
	if err := chain.append(block, txOp); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to add block: %s", err)
	}

	// Index the transactions so they can be queried directly
	if err := chain.PutTransactions(block.Transactions, txOp); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("put transaction failed: %s", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("commit error: %s", err)
	}

	// If the chain is new, add it to the chain cache
	// if it has not been added.
	b.addChain(chain)

	return chain, nil
}

// ProcessBlock takes a block and attempts to add it to the
// tip of one of the known chains (main chain or forked chain). It returns
// the chain where the block was appended to.
func (b *Blockchain) ProcessBlock(block *wire.Block) (common.Chainer, error) {
	b.mLock.Lock()
	defer b.mLock.Unlock()

	b.log.Debug("Processing block", "Hash", block.Hash)

	// If ever we forgot to set the transaction pool,
	// the client should be forced to exit.
	if b.txPool == nil {
		panic("initialization error: transaction pool not set")
	}

	// validate the block. Although, we expect the block to have been
	// validated on the gossip level before getting here
	if err := b.validateBlock(block); err != nil {
		return nil, err
	}
	// if the block has been previously rejected, return err
	if b.isRejected(block) {
		return nil, common.ErrBlockRejected
	}

	// check if the block has previously been detected as an orphan.
	// We do not need to go re-process this block if it is an orphan.
	if b.isOrphanBlock(block.Hash.HexStr()) {
		return nil, common.ErrOrphanBlock
	}

	// check if the block exists in any known chain
	exists, err := b.HaveBlock(block.Hash.HexStr())
	if err != nil {
		return nil, fmt.Errorf("failed to check block existence: %s", err)
	}
	if exists {
		b.log.Debug("Block already exists", "Hash", block.Hash)
		return nil, common.ErrBlockExists
	}

	// attempt to add the block to a chain
	chain, err := b.maybeAcceptBlock(block, nil)
	if err != nil {
		return nil, err
	}

	// process any remaining orphan blocks
	b.processOrphanBlocks(block.Hash.HexStr())

	return chain, nil
}

// execBlock execute the transactions of the blocks to
// output the resulting state objects and state root.
func (b *Blockchain) execBlock(chain common.Chainer, block *wire.Block, opts ...common.CallOp) (root util.Hash, stateObjs []*common.StateObject, err error) {

	// Process the transactions to produce a series of transitions
	// that must be applied to the blockchain state.
	ops, err := b.processTransactions(block.Transactions, chain, opts...)
	if err != nil {
		return util.EmptyHash, nil, fmt.Errorf("transaction error: %s", err)
	}

	// Create state objects from the transition objects. State objects when written
	// to the blockchain state (store and tree) change the values of data.
	stateObjs, err = b.opsToStateObjects(block, chain, ops)
	if err != nil {
		return util.EmptyHash, nil, err
	}

	// Get a new state tree. The tree is seeded with the state root of the parent block
	tree, err := chain.NewStateTree(false, opts...)
	if err != nil {
		return util.EmptyHash, nil, fmt.Errorf("failed to create new state tree: %s", err)
	}

	// Add the state objects into the tree. Contantenate the key and value into a TreeItem
	for _, so := range stateObjs {
		tree.Add(common.TreeItem(bytes.Join([][]byte{so.TreeKey, so.Value}, nil)))
	}

	// Build the tree and compute new state root
	if err = tree.Build(); err != nil {
		return util.EmptyHash, nil, err
	}

	root = tree.Root()
	return
}

// processOrphanBlocks finds orphan blocks in the cache and attempts
// to re-process the blocks that are parented by the latestBlockHash.
//
// This method is not protected by any lock. It must be called with
// the chain lock held.
func (b *Blockchain) processOrphanBlocks(latestBlockHash string) error {

	// Add the passed block hash to this internal slice. This is
	// the slice we will use to perform repetitive orphan processing
	// without needing to recursively call this method at the end.
	var parentHashes = []string{latestBlockHash}

	// As long as we have parent hashes, We will continuously pick a
	// parent hash and try to find an orphan block that references the parent hash.
	for len(parentHashes) > 0 {
		// pick the next parent hash and remove it from the slice
		curParentHash := parentHashes[0]
		parentHashes[0] = ""
		parentHashes = parentHashes[1:]

		// Retrieve the keys of blocks in the orphan cache.
		// Go through them and attempt to append them to a chain
		// using maybeAcceptBlock.
		orphansKey := b.orphanBlocks.Keys()
		for i := 0; i < len(orphansKey); i++ {

			oBKey := orphansKey[i]

			// find an orphan block with a parent hash that
			// is same has the latestBlockHash
			orphanBlock := b.orphanBlocks.Peek(oBKey).(*wire.Block)
			if orphanBlock.Header.ParentHash.HexStr() != curParentHash {
				continue
			}

			// remove from the orphan from the cache
			b.orphanBlocks.Remove(orphanBlock.HashToHex())

			// re-attempt to process the block
			if _, err := b.maybeAcceptBlock(orphanBlock, nil); err != nil {
				return err
			}

			parentHashes = append(parentHashes, orphanBlock.Hash.HexStr())
		}
	}

	return nil
}
