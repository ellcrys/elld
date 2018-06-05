package miner

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ellcrys/elld/util/logger"
	"github.com/ellcrys/elld/wire"
	"github.com/ellcrys/go-ethereum/metrics"
)

var log = logger.NewLogrus()

// SetLogger sets the package default logger
func SetLogger(l logger.Logger) {
	log = l
}

// GetLogger returns the default logger
func GetLogger() logger.Logger {
	return log
}

// Mode defines the type and amount of PoW verification an ethash engine makes.
type Mode uint

const (
	datasetInitBytes   = 1 << 30 // Bytes in dataset at genesis
	datasetGrowthBytes = 1 << 23 // Dataset growth per epoch
	cacheInitBytes     = 1 << 24 // Bytes in cache at genesis
	cacheGrowthBytes   = 1 << 17 // Cache growth per epoch
	epochLength        = 30000   // Blocks per epoch
	mixBytes           = 128     // Width of mix
	hashBytes          = 64      // Hash length in bytes
	hashWords          = 16      // Number of 32 bit ints in a hash
	datasetParents     = 256     // Number of parents of each dataset element
	cacheRounds        = 3       // Number of rounds in cache production
	loopAccesses       = 64      // Number of accesses in hashimoto loop
)

const (
	ModeNormal Mode = iota
	ModeTest
	ModeFake
	ModeFullFake
)

var (
	errInvalidDifficulty      = errors.New("non-positive difficulty")
	errInvalidMixDigest       = errors.New("invalid mix digest")
	errInvalidPoW             = errors.New("invalid proof-of-work")
	errNonPositiveBlockNumber = errors.New("non Positive Block Number")
)

var (
	// maxUint256 is a big integer representing 2^256-1
	maxUint256 = new(big.Int).Exp(big.NewInt(2), big.NewInt(256), big.NewInt(0))

	// algorithmRevision is the data structure version used for file naming.
	algorithmRevision = 23

	// dumpMagic is a dataset dump header to sanity check a data dump.
	dumpMagic = []uint32{0xbaddcafe, 0xfee1dead}
)

// Config are the configuration parameters of the ethash.
type Config struct {
	NumCPU         int
	CacheDir       string
	CachesInMem    int
	CachesOnDisk   int
	DatasetDir     string
	DatasetsInMem  int
	DatasetsOnDisk int
	PowMode        Mode
}

// MiningResult represents hashimoto params and output that satisfies the target
type MiningResult struct {
	digest []byte
	result []byte
	nonce  uint64
}

// Miner defines the entire mining functionality proving a memory hard
// mining algorithm. It is heavily based on Ethash.
type Miner struct {
	numActiveMiners int64 // Keeps count of the number of active miners
	stopped         bool  // Flag to stop all miners

	config   Config
	caches   *lru // In memory caches to avoid regenerating too often
	datasets *lru // In memory datasets to avoid regenerating too often

	hashrate metrics.Meter // Meter tracking the average hashrate

	lock sync.Mutex // Ensures thread safety for the in-memory caches and mining fields
}

// New creates a new miner object
func New(config Config) *Miner {

	if config.CachesInMem <= 0 {
		log.Debug("One ethash cache must always be in memory", "requested", config.CachesInMem)
		config.CachesInMem = 1
	}
	if config.CacheDir != "" && config.CachesOnDisk > 0 {
		log.Debug("Disk storage enabled for ethash caches", "dir", config.CacheDir, "count", config.CachesOnDisk)
	}
	if config.DatasetDir != "" && config.DatasetsOnDisk > 0 {
		log.Debug("Disk storage enabled for ethash DAGs", "dir", config.DatasetDir, "count", config.DatasetsOnDisk)
	}

	m := Miner{
		config:   config,
		caches:   newlru("cache", config.CachesInMem, newCache),
		datasets: newlru("dataset", config.DatasetsInMem, newDataset),
		hashrate: metrics.NewMeter(),
	}

	return &m
}

func (m *Miner) isStopped() bool {
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.stopped
}

// Begin starts the miners on separate threads
func (m *Miner) Begin(header *wire.Header) (*MiningResult, error) {

	numCPU := runtime.NumCPU()
	nThreads := m.config.NumCPU

	if nThreads <= 0 {
		nThreads = 1
	} else if nThreads > numCPU {
		nThreads = numCPU
	}

	runtime.GOMAXPROCS(nThreads)

	miningRes := make(chan *MiningResult)
	for i := 0; i < nThreads; i++ {
		go m.mine(i, header, miningRes)
		atomic.AddInt64(&m.numActiveMiners, 1)
	}

	// send nil to mining result if miner is stopped
	go func() {
		for !m.isStopped() {
		}
		miningRes <- nil
	}()

	var err error
	res := <-miningRes

	if m.isStopped() {
		err = fmt.Errorf("miner was stopped abruptly")
	}

	return res, err
}

// mine is the entry point to the mining operation.
// It performs the proof-of-work computation using the provided header.
// Results are sent to channel MiningResult.
func (m *Miner) mine(minerID int, header *wire.Header, miningRes chan *MiningResult) {

	blockNumber := header.Number
	epoch := blockNumber / epochLength
	currentI, _ := m.datasets.get(uint64(epoch))
	current := currentI.(*dataset)

	// Wait for generation to finish if need be.
	// cache and Dag file
	current.generate(m.config.DatasetDir, m.config.DatasetsOnDisk, m.config.PowMode == ModeTest)

	var (
		Mhash              = header.HashNoNonce().Bytes()
		blockDifficulty, _ = new(big.Int).SetString(header.Difficulty, 10)
		Mtarget            = new(big.Int).Div(maxUint256, blockDifficulty)
		Mdataset           = current
	)

	// random seed using the current timestamp
	seed := uint64(time.Now().UTC().UnixNano())

	// Start generating random nonces until we abort or find a good one
	var (
		attempts = int64(0)
		nonce    = seed
	)

	log.Debug("Miner has started", "ID", minerID, "Seed", nonce)

	for !m.isStopped() {
		// We don't have to update hash rate on every nonce, so update after after 2^X nonces
		attempts++
		if (attempts % (1 << 15)) == 0 {
			m.hashrate.Mark(attempts)
			attempts = 0
		}

		// Compute the PoW value of this nonce
		digest, result := hashimotoFull(Mdataset.dataset, Mhash, nonce)

		if new(big.Int).SetBytes(result).Cmp(Mtarget) <= 0 {

			// Correct nonce found, create a new header with it
			log.Debug("Ethash nonce found", "MinerID", minerID, "Attempts", nonce-seed, "Nonce", nonce)

			result := &MiningResult{
				digest: digest,
				result: result,
				nonce:  nonce,
			}

			miningRes <- result

			break
		}
		nonce++
	}

	// Datasets are unmapped in a finalizer. Ensure that the dataset stays live
	// during sealing so it's not unmapped while being read.
	runtime.KeepAlive(Mdataset)

}

// func (ethash *Miner) Mine(block *ellBlock.Block, minerID int) (string, string, uint64, error) {

// stop decrements all miner waitgroup which forces the miner to halt
func (m *Miner) stop() {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.stopped = true
}

// cache tries to retrieve a verification cache for the specified block number
// by first checking against a list of in-memory caches, then against caches
// stored on disk, and finally generating one if none can be found.
func (m *Miner) cache(block uint64) *cache {
	epoch := block / epochLength
	currentI, futureI := m.caches.get(epoch)
	current := currentI.(*cache)

	// Wait for generation finish.
	current.generate(m.config.CacheDir, m.config.CachesOnDisk, m.config.PowMode == ModeTest)

	// If we need a new future cache, now's a good time to regenerate it.
	if futureI != nil {
		future := futureI.(*cache)
		go future.generate(m.config.CacheDir, m.config.CachesOnDisk, m.config.PowMode == ModeTest)
	}
	return current
}

// dataset tries to retrieve a mining dataset for the specified block number
// by first checking against a list of in-memory datasets, then against DAGs
// stored on disk, and finally generating one if none can be found.
func (m *Miner) dataset(block uint64) *dataset {
	epoch := block / epochLength
	currentI, futureI := m.datasets.get(epoch)
	current := currentI.(*dataset)

	// Wait for generation finish.
	current.generate(m.config.DatasetDir, m.config.DatasetsOnDisk, m.config.PowMode == ModeTest)

	// If we need a new future dataset, now's a good time to regenerate it.
	if futureI != nil {
		future := futureI.(*dataset)
		go future.generate(m.config.DatasetDir, m.config.DatasetsOnDisk, m.config.PowMode == ModeTest)
	}

	return current
}

// Threads returns the number of mining threads currently enabled. This doesn't
// necessarily mean that mining is running!
func (m *Miner) Threads() int64 {
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.numActiveMiners
}

// Hashrate implements PoW, returning the measured rate of the search invocations
// per second over the last minute.
func (m *Miner) Hashrate() float64 {
	return m.hashrate.Rate1()
}

// VerifyPoW checks whether the given header satisfies
// the PoW difficulty requirements.
func (m *Miner) VerifyPoW(header *wire.Header) error {

	// Ensure that we have a valid difficulty for the block
	blockDifficulty, _ := new(big.Int).SetString(header.Difficulty, 10)
	if blockDifficulty.Sign() <= 0 {
		return errInvalidDifficulty
	}

	// Recompute the digest and PoW value and verify against the header
	number := header.Number

	//block number must be a positive number
	if number <= 0 {
		return errNonPositiveBlockNumber
	}

	size := datasetSize(number)
	if m.config.PowMode == ModeTest {
		size = 32 * 1024
	}

	cache := m.cache(number)

	// get Digest and result for POW verification
	digest, result := hashimotoLight(size, cache.cache, header.HashNoNonce().Bytes(), header.Nonce)

	// Caches are unmapped in a finalizer. Ensure that the cache stays live
	// until after the call to hashimotoLight so it's not unmapped while being used.
	runtime.KeepAlive(cache)

	// check if the mix digest is equivalent to the block Mix Digest
	if !bytes.Equal(digest, header.MixHash) {
		return errInvalidMixDigest
	}

	target := new(big.Int).Div(maxUint256, blockDifficulty)
	if new(big.Int).SetBytes(result).Cmp(target) > 0 {
		return errInvalidPoW
	}

	return nil
}
