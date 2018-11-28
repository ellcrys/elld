// Package miner provides proof-of-work mining capability
package miner

import (
	crand "crypto/rand"
	"math"
	"math/big"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/miner/blakimoto"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/util/logger"
	"github.com/ethereum/go-ethereum/metrics"
	"github.com/fatih/color"
	"github.com/olebedev/emitter"
)

const (
	// EventProposedBlock represent an event about a
	// block to be mined.
	EventProposedBlock = "event.proposedBlock"

	// EventFoundBlock represents an event about
	// a block with a valid PoW nonce
	EventFoundBlock = "event.foundBlock"
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
	processing bool

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
	}
}

// Begin starts proof-of-work computation
// and all managing functions
func (m *Miner) Begin() error {

	m.Lock()
	m.done = make(chan bool)
	m.Unlock()

	// If the number of threads haven't been set,
	// Set number of threads to the available
	// number of CPUs
	if m.numThreads == 0 {
		m.numThreads = runtime.NumCPU()
	}

	// Handle incoming events
	go m.handleEvents()
	go m.handleInternalEvents()

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
	m.blakimoto.Prepare(m.blockMaker.ChainReader(),
		proposed.GetHeader())

	m.Lock()
	m.workers = []*Worker{}
	m.Unlock()

	for i := 0; i < m.numThreads; i++ {
		w := &Worker{
			event:      m.iEvent,
			id:         i,
			log:        m.log,
			blockMaker: m.blockMaker,
			blakimoto:  m.blakimoto,
			hashrate:   m.hashrate,
		}

		m.Lock()
		m.workers = append(m.workers, w)
		m.Unlock()

		go w.mine(proposed)
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

// handleEvents handles events from
// components and processes outside the miner.
func (m *Miner) handleEvents() {
	go func() {
		for {
			select {
			case evt := <-m.event.Once(core.EventNewBlock):
				m.OnNewBlock(evt.Args[0].(*core.Block), evt.Args[1].(types.ChainReader))
			}
		}
	}()
}

// processBlock computes hash and signature and
// attempts to append the block to a branch.
func (m *Miner) processBlock(fb *FoundBlock) error {

	// Update the block header with the found nonce
	header := fb.block.GetHeader().Copy()
	header.SetNonce(util.EncodeNonce(fb.nonce))
	fb.block = fb.block.ReplaceHeader(header)

	// Compute and set hash and signature
	fb.block.SetHash(fb.block.ComputeHash())
	blockSig, _ := core.BlockSign(fb.block, m.minerKey.PrivKey().Base58())
	fb.block.SetSignature(blockSig)

	m.Lock()
	prev := m.lastFoundBlockHash
	m.lastFoundBlockHash = fb.block.GetHash()
	m.Unlock()

	// Attempt to append the block to a branch
	_, err := m.blockMaker.ProcessBlock(fb.block)
	if err != nil {
		m.log.Warn("Failed to process block", "Err", err.Error())
		m.Lock()
		m.lastFoundBlockHash = prev
		m.Unlock()
		return err
	}

	m.log.Info(color.GreenString("New block mined"),
		"Number", fb.block.GetNumber(),
		"Difficulty", fb.block.GetHeader().GetDifficulty(),
		"TotalDifficulty", fb.block.GetHeader().GetTotalDifficulty(),
		"Hashrate", m.blakimoto.Hashrate(),
		"PoW Time", time.Since(fb.started))

	return nil
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

// onFoundBlock is called when a worker finds a
// valid nonce for the current proposed block.
func (m *Miner) onFoundBlock(fb *FoundBlock) {

	m.Lock()
	if m.processing {
		m.Unlock()
		m.log.Debug("Rejected a block. Currently processing a winner")
		return
	}

	m.processing = true
	m.Unlock()

	// Stop all workers
	m.stopWorkers()

	// Process block
	m.processBlock(fb)

	// wait a second so the next block
	// does not have same time as its parent
	time.Sleep(1 * time.Second)

	m.Lock()
	m.processing = false
	m.Unlock()

	if m.isMining() {
		if err := m.startWorkers(); err != nil {
			m.log.Debug("Unable to start workers", "Err", err.Error())
		}
	}
}

// handleInternalEvents handles events from
// components and processes within the miner.
func (m *Miner) handleInternalEvents() {
	go func() {
		for {
			select {
			case evt := <-m.iEvent.Once(EventFoundBlock, emitter.Sync):
				m.onFoundBlock(evt.Args[0].(*FoundBlock))
			}
		}
	}()
}

// IsMining checks whether is active
func (m *Miner) isMining() bool {
	m.RLock()
	defer m.RUnlock()
	return m.mining
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

// OnNewBlock is called when a new block
// has been appended to the main chain.
// We must restart a new mining round.
func (m *Miner) OnNewBlock(newBlock *core.Block, chain types.ChainReader) {

	// If this block was the one that was previous found
	// by this miner, we do not need to stop the workers
	// as they had been stopped after the block was processed.
	m.Lock()
	if m.lastFoundBlockHash.Equal(newBlock.GetHash()) {
		m.Unlock()
		return
	}
	m.Unlock()

	m.stopWorkers()

	if err := m.startWorkers(); err != nil {
		m.log.Debug("Unable to restart workers", "Err", err.Error())
	}
}

// FoundBlock represents a block with a valid nonce
type FoundBlock struct {
	workerID int
	block    types.Block
	nonce    uint64
	started  time.Time
	finished time.Time
}

// Worker performs proof-of-work computation
type Worker struct {
	sync.RWMutex
	event      *emitter.Emitter
	id         int
	log        logger.Logger
	blakimoto  *blakimoto.Blakimoto
	blockMaker types.BlockMaker
	stop       bool
	hashrate   metrics.Meter
}

// Stop the worker
func (w *Worker) Stop() {
	w.Lock()
	w.stop = true
	w.Unlock()
}

func (w *Worker) isStopped() bool {
	w.RLock()
	defer w.RUnlock()
	return w.stop
}

func (w *Worker) mine(block types.Block) error {

	// Generate random number source
	source, err := crand.Int(crand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		return err
	}

	r := rand.New(rand.NewSource(source.Int64()))
	seed := uint64(r.Int63())
	nonce := seed

	// Extract some data from the header.
	// Compute difficulty target
	var (
		header   = block.GetHeader()
		hash     = header.GetHashNoNonce().Bytes()
		target   = new(big.Int).Div(maxUint256, header.GetDifficulty())
		attempts = int64(0)
	)

	w.log.Debug("Started search for new nonces", "Seed", seed, "WorkerID", w.id)

	now := time.Now()
	for !w.isStopped() {

		// We don't have to update hash rate on every
		// nonce, so update after after 2^X nonces
		attempts++
		if (attempts % (1 << 5)) == 0 {
			w.hashrate.Mark(attempts)
			attempts = 0
		}

		foundBlock := &FoundBlock{
			block:    block,
			workerID: w.id,
			started:  now,
		}

		// Compute the PoW value of this nonce
		result := blakimoto.BlakeHash(hash, nonce)
		if new(big.Int).SetBytes(result).Cmp(target) <= 0 {

			foundBlock.finished = time.Now()
			foundBlock.nonce = nonce

			// Check whether there is a request to stop
			// this current round
			if w.isStopped() {
				w.log.Debug("Nonce found but discarded", "Attempts", nonce-seed,
					"Nonce", nonce,
					"WorkerID", w.id)
				break
			}

			w.log.Debug("Nonce found", "Attempts", nonce-seed, "Nonce", nonce,
				"WorkerID", w.id)

			// Broadcast this block
			w.event.Emit(EventFoundBlock, foundBlock)

			w.Stop()
		}
		nonce++
	}

	w.log.Debug("Miner worker has stopped", "ID", w.id)

	return nil
}
