package burner

import (
	"fmt"
	"sync"
	"time"

	"github.com/ellcrys/elld/config"
	"github.com/k0kubun/pp"
	"github.com/olebedev/emitter"

	"gopkg.in/oleiade/lane.v1"

	"github.com/ellcrys/elld/accountmgr"

	"github.com/fatih/color"
	"github.com/shopspring/decimal"

	"github.com/ellcrys/elld/util"

	"github.com/thoas/go-funk"

	"github.com/ellcrys/ltcd/btcjson"

	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/util/logger"
)

type indexerResult int

type indexResult struct {
	err    error
	status indexerResult
}

var (
	indexUpToDate         indexerResult = 0x01
	stoppedDueToShutdown  indexerResult = 0x02
	stoppedDueToInterrupt indexerResult = 0x03
)

// UTXOIndexer is responsible for scanning the
// Litecoin blockchain to find and index unspent outputs.
// It also removes outputs that have been spent.
type UTXOIndexer struct {
	*sync.Mutex
	client        RPCClient
	db            elldb.DB
	log           logger.Logger
	stop          bool
	indexerTicker *time.Ticker
	bus           *emitter.Emitter
	interrupt     <-chan struct{}
	cfg           *config.EngineConfig
}

// NewUTXOIndexer creates an instance of AccountsUTXOKeeper.
// The netVersion argument is used to namespace the database such that
// testnet and mainnet utxos are not mixed up.
func NewUTXOIndexer(cfg *config.EngineConfig, log logger.Logger, db elldb.DB,
	netVersion string, bus *emitter.Emitter, interrupt <-chan struct{}) *UTXOIndexer {
	log = log.Module("UTXOIndexer")
	return &UTXOIndexer{
		Mutex:     &sync.Mutex{},
		interrupt: interrupt,
		log:       log,
		bus:       bus,
		db:        db,
		cfg:       cfg,
	}
}

// SetClient sets an RPC client to the burner chain server
func (k *UTXOIndexer) SetClient(client RPCClient) {
	k.client = client
}

// getLastIndexedHeight returns the last scanned block
func (k *UTXOIndexer) getLastIndexedHeight(address string) int32 {
	key := MakeKeyLastScannedBlock(address)
	result := k.db.GetFirstOrLast(key, true)
	if result == nil {
		return 0
	}
	height := util.DecodeNumber(result.Value)
	return int32(height)
}

// setLastScannedHeight sets the last scanned block for the given address
func (k *UTXOIndexer) setLastScannedHeight(db elldb.Tx, address string, height int64) error {
	key := MakeKeyLastScannedBlock(address)
	kv := elldb.NewKVObject(key, util.EncodeNumber(uint64(height)))
	if err := db.Put([]*elldb.KVObject{kv}); err != nil {
		return err
	}
	return nil
}

// getBestBlockHeight the height of the best block
func (k *UTXOIndexer) getBestBlockHeight() (int32, error) {
	_, height, err := k.client.GetBestBlock()
	if err != nil {
		return 0, err
	}
	return height, nil
}

// getBlock returns the block at the given height
func (k *UTXOIndexer) getBlock(height int32) (*btcjson.GetBlockVerboseResult, error) {
	hash, err := k.client.GetBlockHash(int64(height))
	if err != nil {
		return nil, err
	}
	block, err := k.client.GetBlockVerboseTx(hash)
	if err != nil {
		return nil, err
	}
	return block, nil
}

// index finds utxo in the given block that belongs to the address.
// If a output that belongs to the address is found as an input field, it
// is removed from the database of unspent utxo for that address.
// If a utxo is found in the output, it is added as an unspent output.
func (k *UTXOIndexer) index(address string, block *btcjson.GetBlockVerboseResult) error {

	db, err := k.db.NewTx()
	if err != nil {
		return fmt.Errorf("failed to create db transaction")
	}

	for _, tx := range block.RawTx {

		// Find outputs belonging only to the given address and index them
		for _, out := range tx.Vout {

			addresses := out.ScriptPubKey.Addresses

			// Ignore output if it belongs to more than one address or
			// if it does not contain the target address
			if len(address) == 0 || !funk.ContainsString(addresses, address) {
				continue
			}

			key := MakeKeyAddressUTXO(block.Height, address, tx.Txid, out.N)

			// Check whether the UTXO is already indexed
			res := db.GetByPrefix(key)
			if len(res) > 0 {
				k.log.Debug("UTXO already indexed")
				continue
			}

			// Construct document object
			doc := DocUTXO{
				TxID:  tx.Txid,
				Index: out.N,
				Value: out.Value,
			}

			// Store the utxo object
			kvObj := elldb.NewKVObject(key, util.ObjectToBytes(doc))
			if err := db.Put([]*elldb.KVObject{kvObj}); err != nil {
				db.Rollback()
				return fmt.Errorf("failed to store burner utxo object")
			}

			k.log.Debug("UTXO indexed", "Address", address)
		}

		// Find inputs that belong to the target address and also
		// existing in the utxo database. If found, we must delete
		// them from the database as they are now spent
		for _, in := range tx.Vin {
			key := MakeQueryKeyAddressUTXO(address, in.Txid, in.Vout)

			// Check whether the UTXO is already indexed
			res := db.GetByPrefix(key)
			if len(res) == 0 {
				continue
			}

			// At this point the UTXO exists, we need to delete it
			if err := db.DeleteByPrefix(key); err != nil {
				db.Rollback()
				return fmt.Errorf("failed to delete spent utxo")
			}

			k.log.Debug(color.RedString("Spent UTXO has been deleted"))
		}

	}

	// Update last scanned height for this address
	if err := k.setLastScannedHeight(db, address, block.Height); err != nil {
		db.Rollback()
		return err
	}

	return db.Commit()
}

// begin indexes the UTXOs of the given address.
// The 'address' argument is the address whose UTXO are searched and indexed.
// The 'skipToHeight' argument forces the algorithm to ignore blocks below the height.
// The 'reIndex' argument overwrites the last scanned height to zero, forcing a rescan.
func (k *UTXOIndexer) begin(workerID int, address string, skipToHeight int32,
	reIndex bool, interrupt chan struct{}) *indexResult {
begin:

	k.log.Debug("Began account indexation", "Account", address, "WorkerID", workerID)

	result := &indexResult{}

	// Get the height of the best block on the chain
	bestHeight, err := k.getBestBlockHeight()
	if err != nil {
		err = fmt.Errorf("Failed to get best block height of the burner chain: %s", err)
		result.err = err
		return result
	}

	lastScannedHeight := int32(0)

	// Get the last scanned block height only if re-index is not requested.
	if !reIndex {
		lastScannedHeight = k.getLastIndexedHeight(address)
	}

	// If skipToHeight is set and it is greater than the last scanned height,
	// use the skip heigh value as the last scanned height
	if skipToHeight > 0 && skipToHeight > lastScannedHeight {
		lastScannedHeight = skipToHeight
	}

	// If the last scanned block height is at least equal to the
	// current best block height, we should do nothing.
	if lastScannedHeight >= bestHeight {
		k.log.Debug("UTXO database of account is up to date", "Address", address)
		result.status = indexUpToDate
		return result
	}

	// At this point, the last scanned block height is less than the
	// current best block height, it means we need to scan more blocks
	// till we reach the bestHeight.
	height := lastScannedHeight
	for {

		if k.HasStopped() {
			result.status = stoppedDueToShutdown
			return result
		}

		height++

		if util.IsStructChanClosed(interrupt) {
			result.status = stoppedDueToInterrupt
			return result
		}

		k.log.Debug("Current Block", "CurrentHeight", height, "BestHeight", bestHeight,
			"WorkerID", workerID)
		if height > bestHeight {
			break
		}

		if util.IsStructChanClosed(interrupt) {
			result.status = stoppedDueToInterrupt
			return result
		}

		// We need to get the block hash at height
		block, err := k.getBlock(height)
		if err != nil {
			result.err = fmt.Errorf("failed to get block: %s", err)
			return result
		}

		if util.IsStructChanClosed(interrupt) {
			result.status = stoppedDueToInterrupt
			return result
		}

		if err = k.index(address, block); err != nil {
			result.err = fmt.Errorf("failed to index block (%d): %s", block.Height, err)
			return result
		}
	}

	// At this point, the best block height may have increased,
	// so we re-execute operations of the function
	goto begin
}

// getUTXOs returns all UTXOs belonging to an address
func getUTXOs(db elldb.DB, address string) (utxos []*DocUTXO) {
	key := MakeQueryKeyAddressUTXOs(address)
	result := db.GetByPrefix(key)
	for _, o := range result {
		var utxo DocUTXO
		o.Scan(&utxo)
		utxos = append(utxos, &utxo)
	}
	return
}

// balanceOf sums up the total value of all unspent output
func balanceOf(db elldb.DB, address string) string {
	var total = decimal.Zero
	for _, utxo := range getUTXOs(db, address) {
		total = total.Add(decimal.NewFromFloat(utxo.Value))
	}
	return total.String()
}

// HasStopped checks whether the indexer has stopped
func (k *UTXOIndexer) HasStopped() bool {
	k.Lock()
	defer k.Unlock()
	return k.stop
}

// Stop the indexer.
func (k *UTXOIndexer) Stop() {
	k.Lock()
	defer k.Unlock()

	if k.stop {
		return
	}

	k.stop = true

	if k.indexerTicker != nil {
		k.indexerTicker.Stop()
		k.indexerTicker = nil
	}
}

// deleteIndexFrom deletes indexed UTXO associated with blocks
// with height greater than or equal to the given height
func (k *UTXOIndexer) deleteIndexFrom(height uint64) error {
	return k.db.TruncateWithFunc(MakeQueryKeyUTXO(), true, func(kv *elldb.KVObject) bool {
		h := util.DecodeNumber(kv.Key)
		return h >= height
	})
}

// Begin initiates the scanning and indexing process.
func (k *UTXOIndexer) Begin(am *accountmgr.AccountManager,
	numWorkers int, skipToHeight int32, reIndex bool, focusAddr string) error {

	// Define an interrupt channel that will be passed to
	// indexer workers so they can be interrupted whenever
	// an invalid block is discovered.
	workersInterrupt := make(chan struct{})

	// Start a goroutine that listens for general application
	// interrupt signal. If interrupted, it stops interrupts
	// indexer workers and stops the entire indexer.
	go func() {
		<-k.interrupt
		close(workersInterrupt)
		k.Stop()
		k.log.Info("UTXO keeper has been interrupted has stopped")
	}()

	dur := time.Duration(k.cfg.Node.UTXOKeeperIndexInterval) * time.Second
	k.indexerTicker = time.NewTicker(dur)

	// Start a goroutine that listens for worker invalid
	// block event; It reacts by interrupting the workers
	// and deleting UTXO found on the invalid block and
	// its descendants.
	go func() {
		for evt := range k.bus.On(EventInvalidLocalBlock) {
			pp.Println("LO AND BEHOLD", evt)
			invalidHeight := evt.Args[0].(int64)
			if !util.IsStructChanClosed(workersInterrupt) {
				close(workersInterrupt)
				k.deleteIndexFrom(uint64(invalidHeight))
			}
		}
	}()

	// Here we run a goroutine that waits for ticks from
	// the indexer ticker. On each tick, it attempts to index
	// the UTXOs of all burner accounts by spreading the work
	// between index workers.
	go func() {
		for range k.indexerTicker.C {

			// Reinitialize the worker interrupt chan only if
			// it was previously closed.
			if util.IsStructChanClosed(workersInterrupt) {
				workersInterrupt = make(chan struct{})
			}

			// Get the burner accounts.
			// We only need accounts compatible with the current
			// network configuration. For example, if we are running
			// a testnet burner chain, we fetch accounts for the testnet.
			accounts, err := am.ListBurnerAccounts()
			if err != nil {
				return
			}
			isMainnet := config.IsBurnerMainnet()
			accounts = funk.Filter(accounts, func(a *accountmgr.StoredAccount) bool {
				r := a.GetMeta().Get("testnet")
				return r != nil && r.(bool) == !isMainnet
			}).([]*accountmgr.StoredAccount)

			// If the caller set an address to focus on, we need to ensure
			// the address belongs to one of the accounts we fetched above.
			if focusAddr != "" {
				found := funk.Find(accounts, func(a *accountmgr.StoredAccount) bool {
					return a.Address == focusAddr
				})
				if found != nil {
					accounts = []*accountmgr.StoredAccount{found.(*accountmgr.StoredAccount)}
				} else {
					k.log.Error("Cannot focus on an unknown account", "FocusAddress", focusAddr)
					return
				}
			}

			k.log.Debug("Starting burner accounts UTXO indexation",
				"NumWorkers", numWorkers,
				"NumAccounts", len(accounts))

			// Here, we define a queue and populate it with the accounts
			// that we are interested in indexing their UTXOs. This queue
			// is the work queue consumed by the indexer workers.
			queue := *lane.NewQueue()
			for _, a := range accounts {
				queue.Append(a)
			}

			// Start indexer workers to process the queue.
			// Each worker will continuously fetch work from
			// the queue until it is emptied.
			numWorkers = 1
			wg := &sync.WaitGroup{}
			for i := 0; i < numWorkers; i++ {
				wg.Add(1)
				go func(id int) {
					for {
						acct := queue.Shift()
						if acct == nil {
							break
						}
						addr := acct.(*accountmgr.StoredAccount).Address
						if r := k.begin(id, addr, skipToHeight, reIndex, workersInterrupt); r.err != nil {
							k.log.Error("Failed to complete UTXO indexation",
								"Account", addr, "Err", r.err)
						}
					}
					wg.Done()
				}(i + 1)
			}

			// Wait for all workers to finish.
			wg.Wait()
		}
	}()

	return nil
}
