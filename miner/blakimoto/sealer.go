// Copyright 2017 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package blakimoto

import (
	crand "crypto/rand"
	"math"
	"math/big"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/util"
)

// Seal implements consensus.Engine, attempting to find a nonce that satisfies
// the block's difficulty requirements.
func (b *Blakimoto) Seal(block types.Block, stop <-chan struct{}) (types.Block, error) {

	// If we're running a fake PoW, simply return a 0 nonce immediately
	if b.config.PowMode == ModeTest {
		header := block.GetHeader()
		header.SetNonce(util.EncodeNonce(0))
		block.SetHeader(header)

		// delay for the specified time
		time.Sleep(b.fakeDelay)

		return block, nil
	}

	// Create a runner and the multiple search threads it directs
	abort := make(chan struct{})
	found := make(chan types.Block)

	b.lock.Lock()
	threads := b.threads
	if b.rand == nil {
		seed, err := crand.Int(crand.Reader, big.NewInt(math.MaxInt64))
		if err != nil {
			b.lock.Unlock()
			return nil, err
		}
		b.rand = rand.New(rand.NewSource(seed.Int64()))
	}
	b.lock.Unlock()
	if threads == 0 {
		threads = runtime.NumCPU()
	}

	if threads < 0 {
		threads = 0 // Allows disabling local mining without extra logic around local/remote
	}
	var pend sync.WaitGroup
	for i := 0; i < threads; i++ {
		pend.Add(1)
		go func(id int, nonce uint64) {
			defer pend.Done()
			b.mine(block, id, nonce, abort, found)
		}(i, uint64(b.rand.Int63()))
	}
	// Wait until sealing is terminated or a nonce is found
	var result types.Block
	select {
	case <-stop:
		// Outside abort, stop all miner threads
		close(abort)
	case result = <-found:
		// One of the threads found a block, abort all others
		close(abort)
	case <-b.update:
		// Thread count was changed on user request, restart
		close(abort)
		pend.Wait()
		return b.Seal(block, stop)
	}
	// Wait for all miners to terminate and return the block
	pend.Wait()
	return result, nil
}

// mine is the actual proof-of-work miner that searches for a nonce starting from
// seed that results in correct final block difficulty.
func (b *Blakimoto) mine(block types.Block, id int, seed uint64,
	abort chan struct{}, found chan types.Block) {

	// Extract some data from the header
	var (
		header = block.GetHeader()
		hash   = header.GetHashNoNonce().Bytes()
		target = new(big.Int).Div(maxUint256, header.GetDifficulty())
	)

	// Start generating random nonces until we abort or find a good one
	var (
		attempts = int64(0)
		nonce    = seed
	)

	b.log.Debug("Started blakimoto search for new nonces",
		"Seed", seed, "WorkerID", id)

search:
	for {
		select {
		case <-abort:
			// Mining terminated, update stats and abort
			b.log.Debug("Blakimoto nonce search aborted",
				"Attempts", nonce-seed, "WorkerID", id)
			b.hashrate.Mark(attempts)
			break search

		default:
			// We don't have to update hash rate on every nonce, so update after after 2^X nonces
			attempts++
			if (attempts % (1 << 15)) == 0 {
				b.hashrate.Mark(attempts)
				attempts = 0
			}

			// Compute the PoW value of this nonce
			result := blakimoto(hash, nonce)
			if new(big.Int).SetBytes(result).Cmp(target) <= 0 {

				// Correct nonce found, create a new header with it
				header = header.Copy()
				header.SetNonce(util.EncodeNonce(nonce))

				// Seal and return a block (if still needed)
				select {
				case found <- block.WithSeal(header):
					b.log.Debug("Blakimoto nonce found and reported", "Attempts", nonce-seed, "Nonce", nonce, "WorkerID", id)
				case <-abort:
					b.log.Debug("Blakimoto nonce found but discarded", "Attempts", nonce-seed, "Nonce", nonce, "WorkerID", id)
				}
				break search
			}
			nonce++
		}
	}
}
