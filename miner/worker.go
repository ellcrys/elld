package miner

import (
	crand "crypto/rand"
	"math"
	"math/big"
	"math/rand"
	"sync"
	"time"

	"github.com/ellcrys/elld/miner/blakimoto"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/util/logger"
	"github.com/ethereum/go-ethereum/metrics"
	"github.com/olebedev/emitter"
)

// FoundBlock represents a block with a valid nonce
type FoundBlock struct {
	WorkerID int
	Block    types.Block
	Nonce    uint64
	Started  time.Time
	Finished time.Time
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
	stopped := w.stop
	w.RUnlock()
	return stopped
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
			Block:    block,
			WorkerID: w.id,
			Started:  now,
		}

		// Compute the PoW value of this nonce
		result := blakimoto.BlakeHash(hash, nonce)
		if new(big.Int).SetBytes(result).Cmp(target) <= 0 {

			foundBlock.Finished = time.Now()
			foundBlock.Nonce = nonce

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
			go w.event.Emit(EventWorkerFoundBlock, foundBlock)

			w.Stop()
		}
		nonce++
	}

	w.log.Debug("Miner worker has stopped", "ID", w.id)

	return nil
}
