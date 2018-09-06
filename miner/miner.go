package miner

import (
	"math/big"
	"time"

	"github.com/fatih/color"

	"github.com/olebedev/emitter"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/miner/blakimoto"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/types/core/objects"
	"github.com/ellcrys/elld/util/logger"
)

// Miner provides mining, block header modification and
// validation capabilities with respect to PoW. The miner
// leverages Ethash to performing PoW computation.
type Miner struct {

	// minerKey is the key associated with the loaded account (a.k.a coinbase)
	minerKey *crypto.Key

	// cfg is the miner configuration
	cfg *config.EngineConfig

	// log is the logger for the miner
	log logger.Logger

	// blockMaker provides functions for creating a block
	blockMaker core.BlockMaker

	// event is the engine event emitter
	event *emitter.Emitter

	// blakimoto instance
	blakimoto *blakimoto.Blakimoto

	// stop indicates a request to stop all mining
	stop bool

	// abort forces the current mining operations to stop
	abort chan struct{}

	// isMining indicates whether or not
	// mining is ongoing
	isMining bool

	// aborted indicates whether or not mining has been
	// aborted so we do not attempt to re-abort
	aborted bool

	// proposedBlock is the block currently being mined
	proposedBlock core.Block
}

// New creates and returns a new Miner instance
func New(mineKey *crypto.Key, blockMaker core.BlockMaker, event *emitter.Emitter, cfg *config.EngineConfig, log logger.Logger) *Miner {

	m := &Miner{
		minerKey:   mineKey,
		cfg:        cfg,
		log:        log,
		blockMaker: blockMaker,
		event:      event,
		abort:      make(chan struct{}),
		blakimoto:  blakimoto.ConfiguredBlakimoto(cfg.Miner.Mode, log),
	}

	// Subscribe to the global event emitter to learn
	// about new blocks that may invalidate the currently
	// proposed block
	go func() {
		for event := range m.event.On(core.EventNewBlock) {
			m.handleNewBlockEvt(event.Args[0].(*objects.Block))
		}
	}()

	return m
}

// setFakeDelay sets the delay duration for ModeFake
func (m *Miner) setFakeDelay(d time.Duration) {
	m.blakimoto.SetFakeDelay(d)
}

// getProposedBlock creates a full valid block compatible with the
// main chain.
func (m *Miner) getProposedBlock(txs []core.Transaction) (core.Block, error) {
	proposedBlock, err := m.blockMaker.Generate(&core.GenerateBlockParams{
		Transactions: txs,
		Creator:      m.minerKey,
		Nonce:        core.EncodeNonce(1),
		Difficulty:   new(big.Int).SetInt64(1),
	})
	if err != nil {
		return nil, err
	}
	return proposedBlock, nil
}

// abortCurrent forces the currently running mining threads to
// stop. This will cause a new proposed block to be created.
func (m *Miner) abortCurrent() {
	if m.aborted {
		return
	}
	close(m.abort)
	m.aborted = true
}

// Stop stops the miner completely
func (m *Miner) Stop() {
	m.stop = true
	m.abortCurrent()
}

// handleNewBlockEvt detects and processes event about
// a new block being accepted in a chain. Since the
// miner always mines on the main chain, it will
// will cause the current proposed block to be dumped
// if the new block was appended to the main chain.
// Additionally, it emits a core.EventAborted event.
func (m *Miner) handleNewBlockEvt(newBlock *objects.Block) {
	if m.isMining && (m.proposedBlock == nil ||
		(m.blockMaker.IsMainChain(newBlock.ChainReader) && !m.proposedBlock.GetHash().Equal(newBlock.GetHash()))) {
		m.log.Debug("Aborting on-going miner session. Proposing a new block.", "Number", newBlock.Header.Number)
		go m.event.Emit(core.EventAborted, m.proposedBlock)
		m.abortCurrent()
	}
}

// ValidateHeader validates a given header according to
// the Ethash specification.
func (m *Miner) ValidateHeader(chain core.ChainReader, header, parent *objects.Header, seal bool) {
	m.blakimoto.VerifyHeader(header, parent, seal)
}

// IsMining checks whether or not the miner is actively
// performing PoW operation.
func (m *Miner) IsMining() bool {
	return m.isMining
}

// Mine begins the mining process
func (m *Miner) Mine() {

	m.log.Info("Beginning mining protocol")
	m.stop = false

	for !m.stop {

		var err error
		m.aborted = false
		m.abort = make(chan struct{})
		m.isMining = true

		// Get a proposed block compatible with the
		// main chain and the current block.
		m.proposedBlock, err = m.getProposedBlock(nil)
		if err != nil {
			m.log.Error("Proposed block is not valid", "Error", err)
			break
		}

		// if no transactions in the proposed block,
		// do not mine the block, sleep for a few seconds
		// and continue.
		if len(m.proposedBlock.GetTransactions()) == 0 {
			m.log.Debug("Proposed block has no transactions")
			time.Sleep(3 * time.Second)
			continue
		}

		// Prepare the proposed block. It will calculate
		// the difficulty and update the proposed block difficulty
		// field in its header
		m.blakimoto.Prepare(m.blockMaker.ChainReader(), m.proposedBlock.GetHeader())

		// Begin the PoW computation
		startTime := time.Now()
		block, err := m.blakimoto.Seal(m.proposedBlock, m.abort)
		if err != nil {
			m.log.Error(err.Error())
			break
		}

		// abort due to new winning block being discovered
		// or due to Stop() being called.
		if block == nil || m.stop {
			continue
		}

		// Finalize the block. Calculate rewards etc
		block, err = m.blakimoto.Finalize(m.blockMaker, block)
		if err != nil {
			m.log.Error("Block finalization failed", "Err", err)
			break
		}

		// Recompute hash and signature
		block.SetHash(block.ComputeHash())
		blockSig, err := objects.BlockSign(block, m.minerKey.PrivKey().Base58())
		block.SetSignature(blockSig)

		// Attempt to add to the blockchain to the main chain.
		if m.cfg.Miner.Mode != blakimoto.ModeTest {
			_, err = m.blockMaker.ProcessBlock(block)
			if err != nil {
				m.log.Error("Failed to process block", "Err", err.Error())
				break
			}
		}

		m.log.Info(color.GreenString("New block mined"),
			"Number", block.GetNumber(),
			"Difficulty", block.GetHeader().GetDifficulty(),
			"TotalDifficulty", block.GetHeader().GetTotalDifficulty(),
			"Hashrate", m.blakimoto.Hashrate(),
			"PoW Time", time.Since(startTime))

		// in test or fake wait for a second before continues to next block
		// TODO: remove when we are sure duplicate transactions do not exist in
		// the proposed block.
		if m.cfg.Miner.Mode == blakimoto.ModeTest || m.cfg.Node.Mode == config.ModeDev {
			time.Sleep(3 * time.Second)
		}
	}

	m.isMining = false
}
