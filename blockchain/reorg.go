package blockchain

import (
	"fmt"
	"math/big"
	"sort"
	"time"

	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"
)

// ReOrgInfo describes a re-organization event
type ReOrgInfo struct {
	MainChainID  string `json:"mainChainID" msgpack:"mainChainID"`
	SideChainID  string `json:"sideChainID" msgpack:"sideChainID"`
	SideChainLen uint64 `json:"sideChainLen" msgpack:"sideChainLen"`
	ReOrgLen     uint64 `json:"reOrgLen" msgpack:"reOrgLen"`
	Timestamp    int64  `json:"timestamp" msgpack:"timestamp"`
}

// chooseBestChain returns the chain that is considered the
// legitimate chain. It checks all chains according to the rules
// defined below and return the chain that passes the rule on contested
// by another chain.
//
// The rules (executed in the exact order) :
// 1. The chain with the most difficulty wins.
// 2. The chain that was received first.
// 3. The chain with the larger pointer
//
// NOTE: This method must be called with chain lock held by the caller.
func (b *Blockchain) chooseBestChain() (*Chain, error) {

	var highTDChains = []*Chain{}
	var curHighestTD = new(big.Int).SetInt64(0)

	// If no chain exists on the blockchain, return nil
	if len(b.chains) == 0 {
		return nil, nil
	}

	// for each known chains, we must find the chain with the largest total
	// difficulty and add to highTDChains. If multiple chains have same
	// difficulty, then that indicates a tie and as such the highTDChains
	// will also include these chains.
	for _, chain := range b.chains {
		tip, err := chain.Current()
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

// decideBestChain determines and sets the current best chain
// based on the split resolution rules.
func (b *Blockchain) decideBestChain() error {
	b.chainLock.RLock()
	defer b.chainLock.RUnlock()

	proposedBestChain, err := b.chooseBestChain()
	if err != nil {
		b.log.Error("Unable to determine best chain", "Err", err.Error())
		return err
	}

	// At this point, we were just not able to choose a best chain.
	// This will be unlikely and only possible in tests
	if proposedBestChain == nil {
		b.log.Debug("Unable to choose best chain")
		return fmt.Errorf("unable to choose best chain")
	}

	// If the current best chain and the new best chain
	// are not the same. Then we must reorganize
	if b.bestChain != nil && b.bestChain.GetID() != proposedBestChain.GetID() {
		b.log.Info("New best chain detected. Re-organizing.",
			"CurBestChainID", b.bestChain.GetID().SS(), "ProposedChainID", proposedBestChain.GetID().SS())
		b.bestChain, err = b.reOrg(proposedBestChain)
		if err != nil {
			return fmt.Errorf("Reorganization error: %s", err)
		}

		b.log.Info("Reorganization completed", "ChainID", proposedBestChain.GetID().SS())
	}

	// When no best chain has been set, set
	// the best chain to the proposed best chain
	if b.bestChain == nil {
		b.bestChain = proposedBestChain
		b.log.Info("Best chain set", "CurBestChainID", b.bestChain.GetID().SS())
	}

	return nil
}

// recordReOrg stores a record of a reorganization
// NOTE: This method must be called with write chain lock held by the caller.
func (b *Blockchain) recordReOrg(timestamp int64, sidechain *Chain, opts ...core.CallOp) error {

	var txOp = common.GetTxOp(b.db, opts...)
	txOp.CanFinish = false

	var reOrgInfo = &ReOrgInfo{
		MainChainID: b.bestChain.id.String(),
		SideChainID: sidechain.id.String(),
		Timestamp:   timestamp,
	}

	mainTip, err := b.bestChain.Current(txOp)
	if err != nil {
		txOp.AllowFinish().Rollback()
		return err
	}

	sideTip, err := sidechain.Current(txOp)
	if err != nil {
		txOp.AllowFinish().Rollback()
		return err
	}

	reOrgInfo.SideChainLen = sideTip.GetNumber() - sidechain.parentBlock.GetNumber()
	reOrgInfo.ReOrgLen = mainTip.GetNumber() - sidechain.parentBlock.GetNumber()

	key := common.MakeReOrgKey(timestamp)
	err = txOp.Tx.Put([]*elldb.KVObject{elldb.NewKVObject(key, util.ObjectToBytes(reOrgInfo))})
	if err != nil {
		txOp.AllowFinish().Rollback()
		return err
	}

	return txOp.AllowFinish().Commit()
}

// getReOrgs fetches information about all reorganizations
func (b *Blockchain) getReOrgs(opts ...core.CallOp) []*ReOrgInfo {
	b.chainLock.RLock()
	defer b.chainLock.RUnlock()

	var reOrgs = []*ReOrgInfo{}
	key := common.MakeReOrgQueryKey()
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

// reOrg overwrites the main chain with the blocks of
// the sidechain beginning from sidechain parent block + 1.
// Returns the re-organized chain or error.
//
// NOTE: This method must be called with write chain lock held by the caller.
func (b *Blockchain) reOrg(sidechain *Chain) (*Chain, error) {

	now := time.Now()

	// indicate the commencement of a re-org
	b.reOrgActive = true
	defer func() {
		b.reOrgActive = false
	}()

	tx, _ := b.db.NewTx()
	txOp := &common.TxOp{Tx: tx, CanFinish: false}

	// get the tip of the current best chain
	tip, err := b.bestChain.Current(txOp)
	if err != nil {
		txOp.AllowFinish().Rollback()
		return nil, fmt.Errorf("failed to get best chain tip: %s", err)
	}

	// get the side chain tip
	sideTip, err := sidechain.Current(txOp)
	if err != nil {
		txOp.AllowFinish().Rollback()
		return nil, fmt.Errorf("failed to get side chain tip: %s", err)
	}

	// get the parent block of the side chain
	parentBlock := sidechain.GetParentBlock()
	if parentBlock == nil {
		txOp.AllowFinish().Rollback()
		return nil, fmt.Errorf("parent block not set on sidechain")
	}

	// delete blocks from the current best chain,
	// starting from side chain parent block + 1
	nextBlockNumber := parentBlock.GetNumber() + 1
	for nextBlockNumber <= tip.GetNumber() {
		b.bestChain.removeBlock(nextBlockNumber, txOp)
		nextBlockNumber++
	}

	// At this point the blocks that are not in the
	// side chain have been removed from the main chain.
	// Now, we will re-process the blocks in the sidechain
	// targeted for addition in the current best chain
	nextBlockNumber = parentBlock.GetNumber() + 1
	for nextBlockNumber <= sideTip.GetNumber() {

		// get the side chain block
		proposedBlock, err := sidechain.GetBlock(nextBlockNumber, txOp)
		if err != nil {
			txOp.AllowFinish().Rollback()
			return nil, fmt.Errorf("failed to get proposed block: %s", err)
		}

		// attempt to process and append to
		// the current main chain
		if _, err := b.maybeAcceptBlock(proposedBlock, b.bestChain, txOp); err != nil {
			txOp.AllowFinish().Rollback()
			return nil, fmt.Errorf("proposed block was not accepted: %s", err)
		}

		nextBlockNumber++
	}

	// store a record of this re-org
	if err := b.recordReOrg(now.UnixNano(), sidechain, txOp); err != nil {
		return nil, fmt.Errorf("failed to store re-org record")
	}

	if err := txOp.AllowFinish().Commit(); err != nil {
		b.reOrgActive = false
		txOp.AllowFinish().Rollback()
		return nil, fmt.Errorf("failed to commit: %s", err)
	}

	b.reOrgActive = false

	return b.bestChain, nil
}
