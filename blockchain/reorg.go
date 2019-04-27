package blockchain

import (
	"fmt"
	"math/big"
	"sort"
	"time"

	"github.com/ellcrys/elld/params"

	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"
	"github.com/syndtr/goleveldb/leveldb"
)

// ReOrgInfo describes a re-organization event
type ReOrgInfo struct {
	MainChainID string `json:"mainChainID" msgpack:"mainChainID"`
	BranchID    string `json:"branchID" msgpack:"branchID"`
	BranchLen   uint64 `json:"branchLen" msgpack:"branchLen"`
	ReOrgLen    uint64 `json:"reOrgLen" msgpack:"reOrgLen"`
	Timestamp   int64  `json:"timestamp" msgpack:"timestamp"`
}

// chooseBestChain returns the chain that is considered the
// legitimate chain. It checks all chains according to the rules
// defined below and return the chain that passes the rule on contested
// by another chain.
//
// The rules (executed in the exact order) :
// 1. The chain with the most difficulty.
// 2. The chain that was received first.
// 3. The chain with the larger pointer
//
// NOTE: This method must be called with chain lock held by the caller.
func (b *Blockchain) chooseBestChain(opts ...types.CallOp) (*Chain, error) {

	var highTDChains = []*Chain{}
	var curHighestTD = new(big.Int).SetInt64(0)
	var txOp = common.GetTxOp(b.db, opts...)
	if txOp.Closed() {
		return nil, leveldb.ErrClosed
	}

	// If a db transaction was not injected,
	// then we must prevent methods that we pass
	// this transaction to from finalising it
	hasInjectTx := common.HasTxOp(opts...)
	if !hasInjectTx {
		txOp.CanFinish = false
	}

	defer func() {
		txOp.SetFinishable(!hasInjectTx).Discard()
	}()

	b.chl.RLock()
	chains := b.chains
	b.chl.RUnlock()

	// If no chain exists on the blockchain, return nil
	if len(chains) == 0 {
		return nil, nil
	}

	// for each known chains, we must find the chain with the largest total
	// difficulty and add to highTDChains. If multiple chains have same
	// difficulty, then that indicates a tie and as such the highTDChains
	// will also include these chains.
	for _, chain := range chains {
		tip, err := chain.Current(txOp)
		if err != nil {
			// A chain with no tip is ignored.
			if err == core.ErrBlockNotFound {
				continue
			}

			return nil, err
		}

		cmpResult := tip.GetTotalDifficulty().Cmp(curHighestTD)
		if cmpResult > 0 {
			curHighestTD = tip.GetTotalDifficulty()
			highTDChains = []*Chain{chain}
		} else if cmpResult == 0 {
			highTDChains = append(highTDChains, chain)
		}
	}

	// When there is no tie for the total difficulty rule,
	// we return the only chain immediately
	if len(highTDChains) == 1 {
		return highTDChains[0], nil
	}

	// At this point there is a tie between two or more most difficult chains.
	// We need to perform tie breaker using rule 2.
	var oldestChains = []*Chain{}
	var curOldestTimestamp int64
	if len(highTDChains) > 1 {
		for _, chain := range highTDChains {
			if curOldestTimestamp == 0 || chain.info.Timestamp < curOldestTimestamp {
				curOldestTimestamp = chain.info.Timestamp
				oldestChains = []*Chain{chain}
			} else if chain.info.Timestamp == curOldestTimestamp {
				oldestChains = append(oldestChains, chain)
			}
		}
	}

	// When we have just one oldest chain, we return it immediately
	if len(oldestChains) == 1 {
		return oldestChains[0], nil
	}

	// If at this point we still have a tie in
	// the list of oldest chains, then we find the chain
	// with the highest pointer address
	var largestPointerAddrs = []*Chain{}
	var curLargestPointerAddress *big.Int
	if len(oldestChains) > 1 {
		for _, chain := range oldestChains {
			if curLargestPointerAddress == nil || util.GetPtrAddr(chain).Cmp(curLargestPointerAddress) > 0 {
				curLargestPointerAddress = util.GetPtrAddr(chain)
				largestPointerAddrs = []*Chain{chain}
			} else if util.GetPtrAddr(chain).Cmp(curLargestPointerAddress) == 0 {
				largestPointerAddrs = append(largestPointerAddrs, chain)
			}
		}
	}

	return largestPointerAddrs[0], nil
}

// decideBestChain determines and sets the current
// best chain based on the split resolution rules.
// It will keep start chain re-organization operation
// of the main chain if it finds the proposed chain
// and the main chain to be different.
func (b *Blockchain) decideBestChain(opts ...types.CallOp) error {

	// Do nothing if we are not ready to decide what
	// chain is the best at this time. (For Integration Test)
	if !b.canDecideBestChain() {
		return nil
	}

	txOp := common.GetTxOp(b.db, opts...)
	if txOp.Closed() {
		return leveldb.ErrClosed
	}

	// If a db transaction was not injected, then we must prevent
	// methods that we pass this transaction to from finalising it
	hasInjectTx := common.HasTxOp(opts...)
	if !hasInjectTx {
		txOp.CanFinish = false
	}

start:
	// Determine which chain is the best branch
	proposedChain, err := b.chooseBestChain(txOp)
	if err != nil {
		txOp.SetFinishable(!hasInjectTx).Rollback()
		b.log.Error("Unable to determine best chain", "Err", err.Error())
		return err
	}

	// At this point, we were just not able to choose a best chain.
	// This will be unlikely and only possible in mocked tests scenerio
	if proposedChain == nil {
		txOp.SetFinishable(!hasInjectTx).Rollback()
		b.log.Debug("Unable to choose best chain: no chain was proposed")
		return fmt.Errorf("unable to choose best chain: no chain was proposed")
	}

	mainChain := b.GetBestChain().(*Chain)

	// Do nothing if the current best chain and the proposed
	// best chain are the same.
	if mainChain != nil && mainChain.GetID().Equal(proposedChain.GetID()) {
		txOp.SetFinishable(!hasInjectTx).Rollback()
		return nil
	}

	// If the proposed chain is not a child of the main chain.
	// We need to re-organise the proposed chain's parent with
	// proposed chain and rerun the best block selection logic.
	// Example: Chain structure that will trigger this:
	// [1]-[2]-[3] 				  - Root/Main chain
	//      '--[3]-[4]-[5] 		  - Proposed chain parent
	//          '--[4]-[5]-[6]    - Proposed chain
	// After re-org:
	//      '--[3]-[4]-[5]-[6]    - Proposed chain parent
	// -------------------------------------------------------
	proposedChainParent := proposedChain.GetParent()
	if mainChain != nil && proposedChainParent != nil && !proposedChainParent.GetID().
		Equal(mainChain.GetID()) {
		b.log.Info("Superior grand child detected. Re-organizing child's parent",
			"GrandChildChainID", proposedChain.GetID().SS(),
			"ParentChainID", proposedChainParent.GetID().SS())

		b.setReOrgStatus(true)
		_, err := b.reOrg(proposedChainParent, proposedChain, txOp)
		if err != nil {
			txOp.SetFinishable(!hasInjectTx).Rollback()
			b.setReOrgStatus(false)
			b.log.Error("Re-organization has failed", "Err", err.Error())
			return fmt.Errorf("Reorganization has failed: %s", err)
		}
		b.setReOrgStatus(false)

		// Go back to the top to re-determine the new proposed chain naturally.
		goto start
	}

	// At this point, if the main chain is already known and the
	// proposed chain is not the same as it, we need to re-organise
	// the main chain with proposed chain.
	// Example: Chain structure that will trigger this:
	// [1]-[2]-[3] 				  - Root/Main chain
	//      '--[3]-[4]-[5] 		  - Proposed chain
	// After re-org:
	// [1]-[2]-[3]-[4]-[5]-[6]    - Root/Main chain
	// -------------------------------------------------------
	if mainChain != nil && mainChain.GetID() != proposedChain.GetID() {

		b.log.Info("Superior chain detected. Re-organizing...",
			"CurBestChainID", mainChain.GetID().SS(),
			"ProposedChainID", proposedChain.GetID().SS())

		b.setReOrgStatus(true)

		_, err := b.reOrg(mainChain, proposedChain, txOp)
		if err != nil {
			txOp.SetFinishable(!hasInjectTx).Rollback()
			b.setReOrgStatus(false)
			b.log.Error("Re-organization has failed", "Err", err.Error())
			return fmt.Errorf("Reorganization error: %s", err)
		}

		b.SetBestChain(mainChain)
		b.setReOrgStatus(false)

		b.log.Info("Reorganization completed",
			"MainChainID", mainChain.GetID().SS(),
			"ProposedChainID", proposedChain.GetID().SS())
	}

	// If at this point, a main chain has not been determined,
	// we will consider the proposed chain to be the new main chain.
	// This should only happen when no pre-existing chain has been created
	if mainChain == nil {
		b.SetBestChain(proposedChain)
		b.log.Info("Best chain set", "CurBestChainID", b.bestChain.GetID().SS())
	}

	return txOp.SetFinishable(!hasInjectTx).Commit()
}

// recordReOrg stores a record of a reorganization
// NOTE: This method must be called with write chain lock held by the caller.
func (b *Blockchain) recordReOrg(timestamp int64, branch *Chain, opts ...types.CallOp) error {

	var txOp = common.GetTxOp(b.db, opts...)
	if txOp.Closed() {
		return leveldb.ErrClosed
	}

	// If a db transaction was not injected,
	// then we must prevent methods that we pass
	// this transaction to from finishing it
	// (commit/rollback)
	hasInjectTx := common.HasTxOp(opts...)
	if !hasInjectTx {
		txOp.CanFinish = false
	}

	b.chl.RLock()
	mainChain := b.bestChain
	b.chl.RUnlock()

	var reOrgInfo = &ReOrgInfo{
		MainChainID: mainChain.id.String(),
		BranchID:    branch.id.String(),
		Timestamp:   timestamp,
	}

	mainTip, err := mainChain.Current(txOp)
	if err != nil {
		txOp.SetFinishable(!hasInjectTx).Rollback()
		return err
	}

	sideTip, err := branch.Current(txOp)
	if err != nil {
		txOp.SetFinishable(!hasInjectTx).Rollback()
		return err
	}

	reOrgInfo.BranchLen = sideTip.GetNumber() - branch.parentBlock.GetNumber()
	reOrgInfo.ReOrgLen = mainTip.GetNumber() - branch.parentBlock.GetNumber()

	key := common.MakeKeyReOrg(timestamp)
	err = txOp.Tx.Put([]*elldb.KVObject{elldb.NewKVObject(key, util.ObjectToBytes(reOrgInfo))})
	if err != nil {
		txOp.SetFinishable(!hasInjectTx).Rollback()
		return err
	}

	return txOp.SetFinishable(!hasInjectTx).Commit()
}

// getReOrgs fetches information about all reorganizations
func (b *Blockchain) getReOrgs(opts ...types.CallOp) []*ReOrgInfo {
	var reOrgs = []*ReOrgInfo{}
	key := common.MakeQueryKeyReOrg()
	result := b.db.GetByPrefix(key)
	for _, r := range result {
		var reOrg ReOrgInfo
		r.Scan(&reOrg)
		reOrgs = append(reOrgs, &reOrg)
	}

	// sort by timestamp
	sort.Slice(reOrgs, func(i, j int) bool {
		return reOrgs[i].Timestamp > reOrgs[j].Timestamp
	})

	return reOrgs
}

// reOrg overwrites the main chain with blocks of
// branch. The blocks after the branch's parent/root
// blocks are deleted from the main branch and replaced
// with the blocks of the branch.
// Returns the re-organized chain or error.
//
// NOTE: This method must be called with write chain lock held by the caller.
func (b *Blockchain) reOrg(mainChain, proposedBranch *Chain, opts ...types.CallOp) (*Chain, error) {

	now := time.Now()
	txOp := common.GetTxOp(b.db, opts...)

	// If a db transaction was not injected,
	// then we must prevent methods that we pass
	// this transaction to from finalising it
	hasInjectTx := common.HasTxOp(opts...)
	if !hasInjectTx {
		txOp.CanFinish = false
	}

	// Get the tip of the current best branch
	tip, err := mainChain.Current(txOp)
	if err != nil {
		txOp.SetFinishable(!hasInjectTx).Rollback()
		return nil, fmt.Errorf("failed to get best chain tip: %s", err)
	}

	// Get the tip block of the proposed branch
	sideTip, err := proposedBranch.Current(txOp)
	if err != nil {
		txOp.SetFinishable(!hasInjectTx).Rollback()
		return nil, fmt.Errorf("failed to get branch chain tip: %s", err)
	}

	// Get the parent block of the proposed branch
	parentBlock := proposedBranch.GetParentBlock()
	if parentBlock == nil {
		txOp.SetFinishable(!hasInjectTx).Rollback()
		return nil, fmt.Errorf("parent block not set on branch")
	}

	// We need to check whether the parent block from
	// which this branch originated from exists on the
	// main chain
	ok, err := mainChain.hasBlock(parentBlock.GetHash(), txOp)
	if err != nil {
		txOp.SetFinishable(!hasInjectTx).Rollback()
		return nil, fmt.Errorf("failed to check the existence of branch parent block in main chain")
	} else if !ok {
		txOp.SetFinishable(!hasInjectTx).Rollback()
		return nil, params.ErrBranchParentNotInMainChain
	}

	// Delete blocks of the current best chain,
	// starting from proposed branch parent block + 1.
	nextBlockNumber := parentBlock.GetNumber() + 1
	for nextBlockNumber <= tip.GetNumber() {
		_, err := mainChain.removeBlock(nextBlockNumber, txOp)
		if err != nil {
			return nil, fmt.Errorf("failed to delete block from current chain: %s", err)
		}
		nextBlockNumber++
	}

	// At this point, the blocks that are not in the
	// proposed branch have been removed from the main chain.
	// Now, we will re-process the blocks in the proposed branch
	// in attempt to add them to the main chain.
	nextBlockNumber = parentBlock.GetNumber() + 1
	for nextBlockNumber <= sideTip.GetNumber() {

		// Get a block from the proposed branch at height parent_block + 1
		proposedBlock, err := proposedBranch.GetBlock(nextBlockNumber, txOp)
		if err != nil {
			txOp.SetFinishable(!hasInjectTx).Rollback()
			return nil, fmt.Errorf("failed to get proposed block: %s", err)
		}

		// Attempt to process and append to the current main chain
		if _, err := b.maybeAcceptBlock(proposedBlock, mainChain, txOp); err != nil {
			txOp.SetFinishable(!hasInjectTx).Rollback()
			return nil, fmt.Errorf("proposed block was not accepted: %s", err)
		}

		// Move to the next block in the chain (if any)
		nextBlockNumber++
	}

	// Store a record of this re-org
	if err := b.recordReOrg(now.Unix(), proposedBranch, txOp); err != nil {
		txOp.SetFinishable(!hasInjectTx).Rollback()
		return nil, fmt.Errorf("failed to store re-org record")
	}

	// Commit the re-org changes
	if err := txOp.SetFinishable(!hasInjectTx).Commit(); err != nil {
		txOp.SetFinishable(!hasInjectTx).Rollback()
		b.reOrgActive = false
		return nil, fmt.Errorf("failed to commit: %s", err)
	}

	return mainChain, nil
}
