package miner

import (
	"math/big"
	"time"

	"github.com/fatih/color"

	"github.com/olebedev/emitter"

	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/miner/blakimoto"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/util/logger"
	"github.com/ellcrys/elld/wire"
)

const (
	// EventAborted defines an event about an aborted PoW operation
	EventAborted = "event.aborted"
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
	blockMaker common.BlockMaker

	// event is the engine event emitter
	event *emitter.Emitter

	// blakimoto instance
	blakimoto *blakimoto.Blakimoto

	// stop indicates a request to stop all mining
	stop bool

	// abort forces the current mining operations to stop
	abort chan struct{}

	aborted bool

	// proposedBlock is the block currently being mined
	proposedBlock *wire.Block
}

// New creates and returns a new Miner instance
func New(mineKey *crypto.Key, blockMaker common.Blockchain, event *emitter.Emitter, cfg *config.EngineConfig, log logger.Logger) *Miner {

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
		for event := range m.event.On(common.EventNewBlock) {
			m.handleNewBlockEvt(event.Args[0].(*wire.Block))
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
func (m *Miner) getProposedBlock(txs []*wire.Transaction) (*wire.Block, error) {
	proposedBlock, err := m.blockMaker.Generate(&common.GenerateBlockParams{
		Transactions: txs,
		Creator:      m.minerKey,
		Nonce:        wire.EncodeNonce(1),
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
// a new block being accepted in the main chain. This
// will wire cause the current proposed block to dumped
// and it also emits an EventAborted event
func (m *Miner) handleNewBlockEvt(newBlock *wire.Block) {
	if m.proposedBlock == nil || !m.proposedBlock.Hash.Equal(newBlock.Hash) {
		m.log.Debug("New block found. Proposed blocks has been invalidated", "Number", newBlock.Header.Number)
		go m.event.Emit(EventAborted, m.proposedBlock)
		m.abortCurrent()
	}
}

// ValidateHeader validates a given header according to
// the Ethash specification.
func (m *Miner) ValidateHeader(chain common.ChainReader, header, parent *wire.Header, seal bool) {
	m.blakimoto.VerifyHeader(chain, header, parent, seal)
}

// Mine begins the mining process
func (m *Miner) Mine() {

	m.log.Info("Beginning mining protocol")

	for !m.stop {

		var err error
		m.aborted = false
		m.abort = make(chan struct{})

		// Get a proposed block compatible with the
		// main chain and the current block.
		m.proposedBlock, err = m.getProposedBlock([]*wire.Transaction{
			wire.NewTx(wire.TxTypeAllocCoin, 123, util.String(m.minerKey.Addr()), m.minerKey, "0.1", "0.1", time.Now().Unix()),
		})
		if err != nil {
			m.log.Error("Proposed block is not valid", "Error", err)
			return
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
			return
		}

		if block == nil || m.stop {
			continue
		}

		// Finalize the block. Calculate rewards etc
		block, err = m.blakimoto.Finalize(m.blockMaker, block)
		if err != nil {
			m.log.Error("Block finalization failed", "Err", err)
			return
		}

		// Recompute hash and signature
		block.Hash = block.ComputeHash()
		block.Sig, err = wire.BlockSign(block, m.minerKey.PrivKey().Base58())

		// Attempt to add to the blockchain to the main chain.
		if m.cfg.Miner.Mode != blakimoto.ModeTest {
			_, err = m.blockMaker.ProcessBlock(block)
			if err != nil {
				m.log.Error("Failed to process block", "Err", err.Error())
				return
			}
		}

		m.log.Info(color.GreenString("New block mined"),
			"Number", block.Header.Number,
			"Difficulty", block.Header.Difficulty,
			"Hashrate", m.blakimoto.Hashrate(),
			"PoW Time", time.Since(startTime))

		// in test or fake wait for a second before continues to next block
		// TODO: remove when we are sure duplicate transactions do not exist in
		// the proposed block.
		// if m.cfg.Miner.Mode == blakimoto.ModeFake || m.cfg.Miner.Mode == blakimoto.ModeTest {
		// 	time.Sleep(3 * time.Second)
		// }
	}
}
