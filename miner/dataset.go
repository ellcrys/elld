package miner

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	mmap "github.com/edsrzf/mmap-go"
)

// dataset wraps an ethash dataset with some metadata to allow easier concurrent use.
type dataset struct {
	epoch   uint64    // Epoch for which this cache is relevant
	dump    *os.File  // File descriptor of the memory mapped cache
	mmap    mmap.MMap // Memory map itself to unmap before releasing
	dataset []uint32  // The actual cache data content
	once    sync.Once // Ensures the cache is generated only once
}

// newDataset creates a new ethash mining dataset and returns it as a plain Go
// interface to be usable in an LRU cache.
func newDataset(epoch uint64) interface{} {
	return &dataset{epoch: epoch}
}

func (d *dataset) generate(dir string, limit int, test bool) {
	d.once.Do(func() {
		csize := cacheSize(d.epoch*epochLength + 1)
		dsize := datasetSize(d.epoch*epochLength + 1)
		seed := seedHash(d.epoch*epochLength + 1)
		if test {
			csize = 1024
			dsize = 32 * 1024
		}
		// If we don't store anything on disk, generate and return
		if dir == "" {
			cache := make([]uint32, csize/4)
			generateCache(cache, d.epoch, seed)

			d.dataset = make([]uint32, dsize/4)
			generateDataset(d.dataset, d.epoch, cache)
		}
		// Disk storage is needed, this will get fancy
		var endian string
		if !isLittleEndian() {
			endian = ".be"
		}
		path := filepath.Join(dir, fmt.Sprintf("full-R%d-%x%s", algorithmRevision, seed[:8], endian))

		// We're about to mmap the file, ensure that the mapping is cleaned up when the
		// cache becomes unused.
		runtime.SetFinalizer(d, (*dataset).finalizer)

		log.Debug("Attempting to load dataset", "Path", path)

		// Try to load the file from disk and memory map it
		var err error
		d.dump, d.mmap, d.dataset, err = memoryMap(path)
		if err == nil {
			log.Debug("Successfully loaded dataset from disk")
			return
		}

		log.Debug("No previous dataset file exist. Creating a new one")

		// No previous dataset available, create a new dataset file to fill
		cache := make([]uint32, csize/4)
		generateCache(cache, d.epoch, seed)

		d.dump, d.mmap, d.dataset, err = memoryMapAndGenerate(path, dsize, func(buffer []uint32) { generateDataset(buffer, d.epoch, cache) })
		if err != nil {
			log.Debug("Failed to generate mapped ethash dataset", "err", err)

			d.dataset = make([]uint32, dsize/2)
			generateDataset(d.dataset, d.epoch, cache)
		}

		log.Debug("Dataset successfully created")

		// Iterate over all previous instances and delete old ones
		for ep := int(d.epoch) - limit; ep >= 0; ep-- {
			seed := seedHash(uint64(ep)*epochLength + 1)
			path := filepath.Join(dir, fmt.Sprintf("full-R%d-%x%s", algorithmRevision, seed[:8], endian))
			os.Remove(path)
		}
	})
}

// finalizer closes any file handlers and memory maps open.
func (d *dataset) finalizer() {
	if d.mmap != nil {
		d.mmap.Unmap()
		d.dump.Close()
		d.mmap, d.dump = nil, nil
	}
}
