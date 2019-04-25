package blockchain

import (
	"fmt"

	"github.com/syndtr/goleveldb/leveldb"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/params"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"

	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/util"
)

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

// processBalanceTx process a TxTypeBalance transaction.
// It takes value from a sender's account and adds to
// a recipient's account. The nonce of the sender
// account is incremented.
//
// The recipient account is searched in the
// given ops which contains other transition objects
// effected by other transactions in same block.
//
// It will create a OpCreateAccount transition
// object if the recipient account does not exist.
func (b *Blockchain) processBalanceTx(tx types.Transaction, ops []common.Transition,
	chain types.Chainer, opts ...types.CallOp) ([]common.Transition, error) {
	var err error
	var txOps []common.Transition
	var senderAcct, recipientAcct types.Account

	// Find the current account object in previous operations
	// passed via ops. If an account has been updated by
	// the processing of other transactions, the new account
	// state must be taken as the truth current state of the account
	for _, prevOp := range ops {
		if opNewBalance, yes := prevOp.(*common.OpNewAccountBalance); yes &&
			opNewBalance.Address() == tx.GetFrom() {
			senderAcct = opNewBalance.Account
		}
		if opNewBalance, yes := prevOp.(*common.OpNewAccountBalance); yes &&
			opNewBalance.Address() == tx.GetTo() {
			recipientAcct = opNewBalance.Account
		}
	}

	// If we did not get the latest account status
	// of the sender from previous operations, we must
	// fetch it from the database.
	if senderAcct == nil {
		senderAcct, err = b.NewWorldReader().GetAccount(chain, tx.GetFrom(), opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to get sender's account: %s", err)
		}
	}

	// If the sender and recipient account
	// are the same, assign the sender account
	// to the recipient account variable.
	if tx.GetFrom().Equal(tx.GetTo()) {
		recipientAcct = senderAcct
	}

	// If we don't know the recipient account yet,
	// we must fetch it from the database or create it
	if recipientAcct == nil {
		recipientAcct, err = b.NewWorldReader().GetAccount(chain, tx.GetTo(), opts...)
		if err != nil {
			if err != core.ErrAccountNotFound {
				return nil, fmt.Errorf("failed to retrieve recipient account: %s", err)
			}
			recipientAcct = &core.Account{
				Type:    core.AccountTypeBalance,
				Address: tx.GetTo(),
				Balance: "0",
			}
			txOps = append(txOps, &common.OpCreateAccount{
				OpBase:  &common.OpBase{Addr: tx.GetTo()},
				Account: recipientAcct,
			})
		}
	}

	// Convert the amount to be sent to decimal
	sendingAmount := tx.GetValue().Decimal()
	fee := tx.GetFee().Decimal()
	deductable := sendingAmount.Add(fee)

	// Ensure the sender's account balance is
	// sufficient for this transaction value + fee
	if senderAcct.GetBalance().Decimal().LessThan(deductable) {
		return nil, fmt.Errorf("insufficient sender account balance")
	}

	// Add an operation to set a new account
	// balance for the sender
	newSenderBal := senderAcct.GetBalance().Decimal().
		Sub(deductable).StringFixed(params.Decimals)
	senderAcct.SetBalance(util.String(newSenderBal))
	txOps = append(txOps, &common.OpNewAccountBalance{
		OpBase:  &common.OpBase{Addr: tx.GetFrom()},
		Account: senderAcct,
	})

	// Add an operation to set a new balance
	// of the recipient
	newRecipientBal := recipientAcct.GetBalance().Decimal().
		Add(sendingAmount).StringFixed(params.Decimals)
	recipientAcct.SetBalance(util.String(newRecipientBal))
	txOps = append(txOps, &common.OpNewAccountBalance{
		OpBase:  &common.OpBase{Addr: tx.GetTo()},
		Account: recipientAcct,
	})

	// increment the sender's nonce
	senderAcct.IncrNonce()

	return txOps, nil
}

// processAllocCoinTx process a TxTypeAllocCoin. It
// allocates value set in a transaction to specific
// account.
//
// The recipient account is searched in the
// given ops which contains other transition objects
// effected by other transactions in same block.
//
// It will create a OpCreateAccount transition
// object if the account does not exist.
func (b *Blockchain) processAllocCoinTx(tx types.Transaction, ops []common.Transition,
	chain types.Chainer,
	opts ...types.CallOp) ([]common.Transition, error) {
	var err error
	var txOps []common.Transition
	var recipientAcct types.Account

	// Find the current account object in previous operations
	// passed via ops. If an account has been updated by
	// the processing of other transactions, the new account
	// state must be taken as the truth current state of the account
	for _, prevOp := range ops {
		if opNewBalance, yes := prevOp.(*common.OpNewAccountBalance); yes &&
			opNewBalance.Address() == tx.GetTo() {
			recipientAcct = opNewBalance.Account
		}
	}

	// If we did not get the latest account status
	// from previous operations, we must fetch it
	// from the database. If the account does not exists,
	// initialize a new account object for the recipient
	if recipientAcct == nil {
		recipientAcct, err = b.NewWorldReader().GetAccount(chain, tx.GetTo(), opts...)
		if err != nil {
			if err != core.ErrAccountNotFound {
				return nil, fmt.Errorf("failed to retrieve recipient account: %s", err)
			}
			recipientAcct = &core.Account{
				Type:    core.AccountTypeBalance,
				Address: tx.GetTo(),
				Balance: "0",
			}
		}
	}

	// Update the recipients account balance to be the
	// sum of current balance and the new allocation
	newBal := recipientAcct.GetBalance().Decimal().
		Add(tx.GetValue().Decimal()).StringFixed(params.Decimals)
	recipientAcct.SetBalance(util.String(newBal))

	// construct an OpNewAccountBalance transition object
	// and set the account to the updated recipient.
	txOps = append(txOps, &common.OpNewAccountBalance{
		OpBase:  &common.OpBase{Addr: tx.GetTo()},
		Account: recipientAcct,
	})

	return txOps, nil
}

// opsToKVObjects takes a slice of operations
// and apply them to the provided chain
func (b *Blockchain) opsToStateObjects(block types.Block, chain types.Chainer,
	ops []common.Transition) ([]*common.StateObject, error) {

	stateObjs := []*common.StateObject{}

	for _, op := range ops {
		switch _op := op.(type) {

		case *common.OpCreateAccount:
			stateObjs = append(stateObjs, &common.StateObject{
				Key: common.MakeKeyAccount(block.GetNumber(), chain.GetID().Bytes(),
					_op.Address().Bytes()),
				Value: util.ObjectToBytes(_op.Account),
			})

		case *common.OpNewAccountBalance:
			stateObjs = append(stateObjs, &common.StateObject{
				Key: common.MakeKeyAccount(block.GetNumber(),
					chain.GetID().Bytes(), _op.Address().Bytes()),
				Value: util.ObjectToBytes(_op.Account),
			})

		default:
			return nil, fmt.Errorf("unknown transition sub-type")
		}
	}

	return stateObjs, nil
}

// ProcessTransactions computes the state transition operations
// for each transactions that must be applied to the state tree
// and world state
func (b *Blockchain) ProcessTransactions(txs []types.Transaction, chain types.Chainer,
	opts ...types.CallOp) ([]common.Transition, error) {

	var ops = common.GetTransitions(opts...)
	for i, tx := range txs {
		var err error
		var newOps []common.Transition

		switch tx.GetType() {
		case core.TxTypeBalance:
			newOps, err = b.processBalanceTx(tx, ops, chain, opts...)
		case core.TxTypeAlloc:
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
// new set of state core. The state objects and transactions are
// then stored to the store.
//
// If a chain is passed as parameter, no attempt to determine the chain
// is taken. Instead, the block will be processed for inclusion in the
// passed chain. This should only be used for the genesis block.
//
// NOTE: This method must be called with chain lock held by the caller.
func (b *Blockchain) maybeAcceptBlock(block types.Block, chain *Chain,
	opts ...types.CallOp) (*Chain, error) {

	var err error
	var parentBlock types.Block
	var chainTip types.Header
	var createNewChain bool
	var bValidator = b.getBlockValidator(block)

	// Add any assigned validation contexts
	for _, ctx := range block.GetValidationContexts() {
		bValidator.setContext(ctx)
	}

	// Sanity check. This should have been done by the caller
	if errs := bValidator.CheckFields(); len(errs) > 0 {
		return nil, errs[0]
	}

	// Skip trying to determine what chain the block
	// belongs to if a chain was explicitly provided
	if chain != nil {
		goto process
	}

	// We need to find the chain in which the block's
	// parent belongs to. This chain may be the main cain or
	// a side chain (branch). We also need the tip of this chain.
	parentBlock, chain, chainTip, err = b.findChainByBlockHash(block.GetHeader().
		GetParentHash(), opts...)

	// If the block's parent does not belong to
	// any known chain. This is a orphan block
	if err != nil {
		if err != core.ErrBlockNotFound {
			return nil, err
		}
		b.log.Debug("Block is not compatible with any chain",
			"BlockNo", block.GetNumber(), "Err", err.Error())
	}

	// Since we are unable to find a chain for this block,
	// we will add it to the orphan cache until a
	// time when its parent is unknown/processed.
	if chain == nil {
		b.addOrphanBlock(block)

		// Emitting core.EventOrphanBlock will cause
		// the block manager to request the parent block
		// from the originating peer.
		go b.eventEmitter.Emit(core.EventOrphanBlock, block)
		return nil, nil
	}

	// Ensure the block is not older than its parent.
	// If so, we must reject such a block
	if block.GetHeader().GetTimestamp() < parentBlock.GetHeader().GetTimestamp() {
		b.log.Info("Block's timestamp must be greater than its parent's",
			"BlockNo", block.GetNumber(),
			"BlockTime", block.GetHeader().GetTimestamp(),
			"ParentBlockTime", parentBlock.GetHeader().GetTimestamp())
		b.addRejectedBlock(block)
		return nil, fmt.Errorf("block timestamp must be greater than its parent's")
	}

	// Since this block is of a lower height than
	// the current block in the chain, it should
	// result in new chain.
	if block.GetHeader().GetNumber() < chainTip.GetNumber() {
		createNewChain = true
		b.log.Info("Stale block found. New branch required",
			"BlockNo", block.GetNumber(),
			"BestChainHeight", chainTip.GetNumber())
		goto process
	}

	// Here, the block height and the chain height
	// are the same.A new chain must be created
	if block.GetNumber() == chainTip.GetNumber() {
		createNewChain = true
		b.log.Info("Block with same height exists. New branch required.",
			"BlockNo", block.GetNumber(),
			"ChainHeight", chainTip.GetNumber())
	}

process:
	// Verify that the block's PoW for non-genesis blocks is valid.
	// Only do this in production or development mode
	if (b.cfg.Node.Mode != config.ModeTest) && block.GetNumber() > 1 {
		if errs := bValidator.CheckPoW(opts...); len(errs) > 0 {
			b.log.Debug("Block PoW is invalid", "BlockNo", block.GetNumber(), "Err", errs[0])
			return nil, errs[0]
		}
	}

	txOp := common.GetTxOp(chain.store.DB(), opts...)
	if txOp.Closed() {
		return nil, leveldb.ErrClosed
	}

	// If a db transaction was not injected,
	// then we must prevent methods that we pass
	// this transaction to from finalising it
	// (commit/rollback)
	hasInjectTx := common.HasTxOp(opts...)
	if !hasInjectTx {
		txOp.CanFinish = false
	}

	// create the new chain, set its root to
	// the parent of the forked block
	if createNewChain {
		if chain, err = b.newChain(txOp.Tx, block, parentBlock, chain); err != nil {
			txOp.SetFinishable(!hasInjectTx).Rollback()
			return nil, fmt.Errorf("failed to create subtree out of stale block: %s", err)
		}

		b.log.Debug("New chain created",
			"ChainID", chain.GetID(),
			"BlockNo", block.GetNumber(),
			"ParentBlockNo", parentBlock.GetNumber())
	}

	if chain.HasParent(txOp) {
		// Update the validator context to ContextBranch
		// since we intend to add the block to a branch.
		bValidator.setContext(types.ContextBranch)
	}

	// validate transactions in the block
	chainOp := &common.OpChainer{Chain: chain}
	if errs := bValidator.CheckTransactions(txOp, chainOp); len(errs) > 0 {
		txOp.SetFinishable(!hasInjectTx).Rollback()
		return nil, errs[0]
	}

	var batchObjs []*elldb.KVObject
	var stateObjs []*common.StateObject
	var newStateRoot util.Hash

	// Do not perform state transition or
	// validate state root for blocks belonging to
	// branches and when OpAllowExec is set to false.
	// When OpAllowExec is set to true, state transition
	// will occur.
	// Note: OpAllowExec is used in tests for
	// mocking branch blocks with valid state.
	if chain.HasParent() && !common.ExecAllowed(opts...) {
		goto commit
	}

	// Execute block to derive the state objects and
	// the expected state root when the state
	// objects are applied to the current blockchain state.
	newStateRoot, stateObjs, err = b.execBlock(chain, block, txOp)
	if err != nil {
		txOp.SetFinishable(!hasInjectTx).Rollback()
		b.log.Error("Block execution failed", "BlockNo", block.GetNumber(), "Err", err)
		return nil, err
	}

	// Compare the state root in the block header with
	// the root obtained from the mock execution of the block.
	if !block.GetHeader().GetStateRoot().Equal(newStateRoot) {
		txOp.SetFinishable(!hasInjectTx).Rollback()
		b.log.Error("Compute state root and block state root do not match",
			"BlockNo", block.GetNumber(),
			"BlockStateRoot", block.GetHeader().GetStateRoot().HexStr(),
			"ComputedStateRoot", newStateRoot.HexStr())
		return nil, core.ErrBlockStateRootInvalid
	}

	// We need to update the world state using the latest
	// state objects derived from executing the block
	for _, so := range stateObjs {
		batchObjs = append(batchObjs, elldb.NewKVObject(so.Key, so.Value))
	}
	if err := txOp.Tx.Put(batchObjs); err != nil {
		txOp.SetFinishable(!hasInjectTx).Rollback()
		return nil, fmt.Errorf("failed to add state object to store: %s", err)
	}

	// Make transactions queryable by indexing them
	if err := chain.PutTransactions(block.GetTransactions(),
		block.GetNumber(), txOp); err != nil {
		txOp.SetFinishable(!hasInjectTx).Rollback()
		return nil, fmt.Errorf("put transaction failed: %s", err)
	}

	// To allow the client operator check for blocks they
	// have mined, add a mined block index for this block
	// only when the block was signed by the coinbase key
	if block.GetHeader().GetCreatorPubKey().
		Equal(util.String(b.coinbase.PubKey().Base58())) {
		if err := chain.PutMinedBlock(block, txOp); err != nil {
			return nil, err
		}
	}

commit:
	// At this point, the block is good to go.
	// We add it to the chain
	if err := chain.append(block, txOp); err != nil {
		txOp.SetFinishable(!hasInjectTx).Rollback()
		return nil, fmt.Errorf("failed to add block: %s", err)
	}

	// Commit the db transaction
	if err := txOp.SetFinishable(!hasInjectTx).Commit(); err != nil {
		txOp.SetFinishable(!hasInjectTx).Rollback()
		return nil, fmt.Errorf("commit error: %s", err)
	}

	// Cache the chain if it has
	// not been seen before
	b.addChain(chain)

	// Decide which chain/branch is the best/main.
	// This could potentially cause a reorganization.
	// We will skip this step if a reorganization is ongoing
	if !b.reOrgIsActive() {
		var err error
		if !hasInjectTx {
			txOp = nil
		}

		err = b.decideBestChain(txOp)
		if err != nil {
			b.log.Error("Failed to decide best chain", "Err", err)
			return nil, fmt.Errorf("failed to choose best chain: %s", err)
		}
	}

	// When the chain is the best chain, emit a new block
	// event for other processes to act on the new block
	if b.GetBestChain().GetID().Equal(chain.GetID()) {
		go b.eventEmitter.Emit(core.EventNewBlock, block, chain.ChainReader())
	}

	return chain, nil
}

// ProcessBlock takes a block, performs initial validations
// and attempts to add it to the tip of one of the known
// chains (main chain or forked chain). It returns a chain reader
// pointing to the chain in which the block was added to.
func (b *Blockchain) ProcessBlock(block types.Block,
	opts ...types.CallOp) (types.ChainReaderFactory, error) {

	b.processLock.Lock()
	defer b.processLock.Unlock()

	b.log.Debug("Processing block", "BlockNo", block.GetNumber(),
		"Hash", block.GetHash().SS())

	// If ever we forgot to set the transaction pool,
	// the client should be forced to exit.
	if b.txPool == nil {
		panic("initialization error: transaction pool not set")
	}

	// Validate the block fields.
	bValidator := b.getBlockValidator(block)

	// add any assigned validation contexts
	for _, ctx := range block.GetValidationContexts() {
		bValidator.setContext(ctx)
	}

	if errs := bValidator.CheckFields(); len(errs) > 0 {
		return nil, errs[0]
	}

	// Validate allocations. We need to know whether
	// the allocations in this block are as expected.
	if errs := bValidator.CheckAllocs(); len(errs) > 0 {
		return nil, errs[0]
	}

	// Check whether the block has been previously rejected
	if b.isRejected(block) {
		b.log.Debug("Block had already been rejected", "BlockNo", block.GetNumber())
		return nil, core.ErrBlockRejected
	}

	// Check whether the block has previously been detected as an orphan.
	// We do not need to go re-process this block if it is an orphan.
	if b.isOrphanBlock(block.GetHash()) {
		b.log.Debug("Block is already known a an orphan", "BlockNo", block.GetNumber())
		return nil, core.ErrOrphanBlock
	}

	// Check if the block exists in any known chain
	exists, err := b.HaveBlock(block.GetHash())
	if err != nil {
		return nil, fmt.Errorf("failed to check block existence: %s", err)
	} else if exists {
		b.log.Debug("Block already exists", "BlockNo", block.GetNumber())
		return nil, core.ErrBlockExists
	}

	// Attempt to add the block to a chain
	chain, err := b.maybeAcceptBlock(block, nil, opts...)
	if err != nil {
		return nil, err
	}

	// process any remaining orphan blocks
	// that may depend on this newly accepted block
	b.processOrphanBlocks(block.GetHash().HexStr())

	return chain.ChainReader(), nil
}

// execBlock execute the transactions of the blocks to
// output the resulting state objects and state root.
func (b *Blockchain) execBlock(chain types.Chainer,
	block types.Block, opts ...types.CallOp) (root util.Hash,
	stateObjs []*common.StateObject, err error) {

	// Process the transactions to produce a series of transitions
	// that must be applied to the blockchain state.
	ops, err := b.ProcessTransactions(block.GetTransactions(), chain, opts...)
	if err != nil {
		return util.EmptyHash, nil, fmt.Errorf("transaction error: %s", err)
	}

	// Create state objects from the transition
	// core. State objects when written to the
	// blockchain state (store and tree) change
	// the values of data.
	stateObjs, err = b.opsToStateObjects(block, chain, ops)
	if err != nil {
		return util.EmptyHash, nil, err
	}

	// Get a new state tree. The tree is
	// seeded with the state root of the parent block
	tree, err := chain.NewStateTree(opts...)
	if err != nil {
		return util.EmptyHash, nil,
			fmt.Errorf("failed to create new state tree: %s", err)
	}

	// Add the state value into the tree.
	for _, so := range stateObjs {
		tree.Add(common.TreeItem(so.Value))
	}

	// Build the tree and compute new state root
	if err = tree.Build(); err != nil {
		return util.EmptyHash, nil, err
	}

	root = tree.Root()
	return
}

// processOrphanBlocks finds orphan blocks
// in the cache and attempts to re-process
// the blocks that are parented by the latestBlockHash.
//
// This method is not protected by any lock. It must be called with
// the chain lock held.
func (b *Blockchain) processOrphanBlocks(latestBlockHash string) error {

	// Add the passed block hash to this internal slice. This is
	// the slice we will use to perform repetitive orphan processing
	// without needing to recursively call this method at the end.
	var parentHashes = []string{latestBlockHash}

	// As long as we have parent hashes, We will continuously pick a
	// parent hash and try to find an orphan block that
	// references the parent hash.
	for len(parentHashes) > 0 {
		// Pick the next parent hash and remove it from the slice
		curParentHash := parentHashes[0]
		parentHashes[0] = ""
		parentHashes = parentHashes[1:]

		// Retrieve the keys of blocks in the orphan cache.
		// Go through them and attempt to append them to a chain
		// using maybeAcceptBlock.
		orphansKey := b.orphanBlocks.Keys()
		for i := 0; i < len(orphansKey); i++ {

			oBKey := orphansKey[i]

			// Find an orphan block with a parent hash that
			// is same has the latestBlockHash
			orphanBlock := b.orphanBlocks.Peek(oBKey).(types.Block)
			if orphanBlock.GetHeader().GetParentHash().HexStr() != curParentHash {
				continue
			}

			// Remove from the orphan from the cache
			b.orphanBlocks.Remove(orphanBlock.GetHashAsHex())

			// Re-attempt to process the block
			if _, err := b.maybeAcceptBlock(orphanBlock, nil); err != nil {
				return err
			}

			parentHashes = append(parentHashes, orphanBlock.GetHash().HexStr())
		}
	}

	return nil
}
