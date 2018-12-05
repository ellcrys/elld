// Package miner provides proof-of-work mining capability
package miner

import (
	"fmt"
	"math/big"
	"runtime"
	"sync"
	"time"

	"github.com/ellcrys/abool"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/miner/blakimoto"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/util/logger"
	"github.com/ethereum/go-ethereum/metrics"
	"github.com/olebedev/emitter"
)

const (
	// EventWorkerFoundBlock indicates that a worker found a block
	EventWorkerFoundBlock = "event.workerFoundBlock"
)

var (
	// maxUint256 is a big integer representing 2^256-1
	maxUint256 = new(big.Int).Exp(big.NewInt(2), big.NewInt(256), big.NewInt(0))
)

// Miner provides proof-of-work computation,
// difficulty calculation and prepares a
// mine block for processing.
type Miner struct {
	sync.RWMutex

	// numThreads is the number of threads
	// running the pow computation
	numThreads int

	// minerKey is the key associated with
	// the loaded account (a.k.a coinbase)
	minerKey *crypto.Key

	// cfg is the miner configuration
	cfg *config.EngineConfig

	// log is the logger for the miner
	log logger.Logger

	// blakimoto instance
	blakimoto *blakimoto.Blakimoto

	// Event emitter
	event *emitter.Emitter

	// iEvent is an event emitter used internally
	iEvent *emitter.Emitter

	// workers holds instances of the PoW computers
	workers []*Worker

	// blockMaker provides functions for creating a block
	blockMaker types.BlockMaker

	// hashrate for tracking average hashrate
	hashrate metrics.Meter

	// processing indicates that a block is being
	// processed for inclusion in a branch
	processing *abool.AtomicBool

	// lastFoundBlockHash is the hash of the last
	// block found by this miner
	lastFoundBlockHash util.Hash

	stopped bool
	done    chan bool
	mining  bool
}

// NewMiner creates a Miner instance
func NewMiner(mineKey *crypto.Key, blockMaker types.BlockMaker,
	event *emitter.Emitter, cfg *config.EngineConfig,
	log logger.Logger) *Miner {
	return &Miner{
		cfg:        cfg,
		log:        log,
		event:      event,
		blockMaker: blockMaker,
		iEvent:     &emitter.Emitter{},
		minerKey:   mineKey,
		blakimoto:  blakimoto.ConfiguredBlakimoto(blakimoto.Mode(cfg.Miner.Mode), log),
		hashrate:   metrics.NewMeter(),
		done:       make(chan bool),
		processing: abool.New(),
	}
}

// Begin starts proof-of-work computation
// and all managing functions
func (m *Miner) Begin() error {

	m.Lock()
	m.done = make(chan bool)

	// If the number of threads haven't been set,
	// set number of threads to the available number of CPUs
	if m.numThreads == 0 {
		m.numThreads = runtime.NumCPU()
	}

	m.Unlock()

	// Handle incoming events
	go m.handleWorkersEvents()

	// start worker
	if err := m.startWorkers(); err != nil {
		return err
	}

	m.Lock()
	m.mining = true
	m.Unlock()

	<-m.done

	return nil
}

// startWorkers generates a block, starts
// and passes the proposed block to the workers.
func (m *Miner) startWorkers() error {

	proposed, err := m.getProposedBlock(nil)
	if err != nil {
		m.log.Error("Proposed block is not valid", "Err", err)
		return err
	}

	// Prepare the proposed block.
	m.blakimoto.Prepare(m.blockMaker.ChainReader(), proposed.GetHeader())

	m.Lock()
	m.workers = []*Worker{}
	m.Unlock()

	for i := 0; i < m.numThreads; i++ {
		m.Lock()
		w := &Worker{
			event:      m.iEvent,
			id:         i,
			log:        m.log,
			blockMaker: m.blockMaker,
			blakimoto:  m.blakimoto,
			hashrate:   m.hashrate,
		}
		m.workers = append(m.workers, w)
		go w.mine(proposed)
		m.Unlock()
	}

	return nil
}

// SetNumThreads sets the number of threads
// performing PoW computation
func (m *Miner) SetNumThreads(n int) {
	m.Lock()
	m.numThreads = n
	m.Unlock()

	if m.isMining() {
		m.stopWorkers()

		if err := m.startWorkers(); err != nil {
			m.log.Debug("Unable to restart workers", "Err", err.Error())
		}
	}
}

// RestartWorkers restarts workers. Any previous task
// is immediately dropped and a new block is proposed and worked on
func (m *Miner) RestartWorkers() error {
	m.stopWorkers()
	if !m.isMining() {
		return fmt.Errorf("miner has stopped")
	}
	return m.startWorkers()
}

// processBlock computes hash and signature and
// attempts to append the block to a branch.
func (m *Miner) processBlock(fb *FoundBlock) error {

	// Update the block header with the found nonce
	header := fb.Block.GetHeader().Copy()
	header.SetNonce(util.EncodeNonce(fb.Nonce))
	fb.Block = fb.Block.ReplaceHeader(header)

	// Compute and set block hash and signature
	fb.Block.SetHash(fb.Block.ComputeHash())
	blockSig, _ := core.BlockSign(fb.Block, m.minerKey.PrivKey().Base58())
	fb.Block.SetSignature(blockSig)

	errCh := make(chan error)
	m.log.Debug("Emitting Event")
	m.event.Emit(core.EventFoundBlock, fb, errCh)
	m.log.Debug("Emitted Event")
	r := <-errCh
	m.log.Debug("Waiting for Event Error")

	return r
}

// onFoundBlock is called when a worker finds a
// valid nonce for the current proposed block.
func (m *Miner) onFoundBlock(fb *FoundBlock) {

	if !m.processing.SetToIf(false, true) {
		m.log.Debug("Rejected a block. Currently processing a winner")
		return
	}

	// Stop all workers who are currently
	// trying to solve PoW for the current round
	// that has just been solved.
	m.stopWorkers()

	m.log.Debug("About to Process Found Block")

	// Attempt to process the block.
	// If it failed, restart the workers
	if err := m.processBlock(fb); err != nil {
		m.processing.SetTo(false)
		m.RestartWorkers()
		m.log.Debug("Block Processing Failed")
		return
	}
	m.log.Debug("Finished Processing Found Block")

	m.processing.SetTo(false)

	if m.isMining() {
		m.log.Debug("Restarting Mining")

		// If the new block timestamp is the same as the current
		// time, we should wait for 1 second so we don't allow
		// the workers create and work on an invalid block that
		// share same time as their parent.
		// This condition is common when difficulty is extremely low.
		if fb.Block.GetHeader().GetTimestamp() == time.Now().Unix() {
			time.Sleep(1 * time.Second)
		}

		if err := m.startWorkers(); err != nil {
			m.log.Debug("Unable to start workers", "Err", err.Error())
		}
		m.log.Debug("Restarted Mining")
	}
}

// handleWorkersEvents handles events from
// components and processes within the miner.
func (m *Miner) handleWorkersEvents() {
	for {
		select {
		case evt := <-m.iEvent.Once(EventWorkerFoundBlock, emitter.Sync):
			m.onFoundBlock(evt.Args[0].(*FoundBlock))
		}
	}
}

// getProposedBlock creates a full valid block
// compatible with the main chain.
func (m *Miner) getProposedBlock(txs []types.Transaction) (types.Block, error) {
	proposedBlock, err := m.blockMaker.Generate(&types.GenerateBlockParams{
		Transactions: txs,
		Creator:      m.minerKey,
		Nonce:        util.EncodeNonce(1),
		Difficulty:   new(big.Int).SetInt64(1),
		AddFeeAlloc:  true,
	})
	if err != nil {
		return nil, err
	}
	return proposedBlock, nil
}

// IsMining checks whether is active
func (m *Miner) isMining() bool {
	m.RLock()
	defer m.RUnlock()
	return m.mining
}

// Stop the miner
func (m *Miner) Stop() {
	if !m.isMining() {
		return
	}

	m.Lock()
	close(m.done)
	m.mining = false
	m.stopped = true
	m.Unlock()

	m.stopWorkers()
}

// stopWorkers stops the workers
func (m *Miner) stopWorkers() {
	m.RLock()
	workers := m.workers
	m.RUnlock()
	for _, w := range workers {
		w.Stop()
	}
}
