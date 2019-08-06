package burner

import (
	"fmt"
	"sync"
	"time"

	"github.com/ellcrys/elld/params"

	"github.com/ellcrys/elld/config"

	"gopkg.in/oleiade/lane.v1"

	"github.com/ellcrys/elld/accountmgr"

	"github.com/fatih/color"
	"github.com/shopspring/decimal"

	"github.com/ellcrys/elld/util"

	"github.com/thoas/go-funk"

	"github.com/ellcrys/elld/ltcsuite/ltcd/btcjson"

	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/ltcsuite/ltcd/rpcclient"
	"github.com/ellcrys/elld/util/logger"
)

// AccountsUTXOKeeper is responsible for scanning the
// Litecoin blockchain to find and index unspent outputs.
// It also removes outputs that have been spent.
type AccountsUTXOKeeper struct {
	*sync.Mutex
	client        *rpcclient.Client
	netVersion    string
	db            elldb.DB
	log           logger.Logger
	stop          bool
	indexerTicker *time.Ticker
	interrupt     <-chan struct{}
}

// NewBurnerAccountUTXOKeeper creates an instance of AccountsUTXOKeeper.
// The netVersion argument is used to namespace the database such that
// testnet and mainnet utxos are not mixed up.
func NewBurnerAccountUTXOKeeper(log logger.Logger, db elldb.DB, netVersion string,
	interrupt <-chan struct{}) *AccountsUTXOKeeper {
	log = log.Module("BurnerUTXOKeeper")
	return &AccountsUTXOKeeper{
		Mutex:      &sync.Mutex{},
		log:        log,
		netVersion: netVersion,
		db:         db,
		interrupt:  interrupt,
	}
}

// SetClient sets an RPC client to the burner chain server
func (k *AccountsUTXOKeeper) SetClient(client *rpcclient.Client) {
	k.client = client
}

// lastScannedHeight returns the last scanned block
func (k *AccountsUTXOKeeper) lastScannedHeight(address string) int32 {
	key := MakeKeyLastScannedBlock(k.netVersion, address)
	result := k.db.GetFirstOrLast(key, true)
	if result == nil {
		return 0
	}
	height := util.DecodeNumber(result.Value)
	return int32(height)
}

// setLastScannedHeight sets the last scanned block for the given address
func (k *AccountsUTXOKeeper) setLastScannedHeight(db elldb.Tx, address string, height int64) error {
	key := MakeKeyLastScannedBlock(k.netVersion, address)
	kv := elldb.NewKVObject(key, util.EncodeNumber(uint64(height)))
	if err := db.Put([]*elldb.KVObject{kv}); err != nil {
		return err
	}
	return nil
}

// getBestBlockHeight the height of the best block
func (k *AccountsUTXOKeeper) getBestBlockHeight() (int32, error) {
	_, height, err := k.client.GetBestBlock()
	if err != nil {
		return 0, err
	}
	return height, nil
}

// getBlock returns the block at the given height
func (k *AccountsUTXOKeeper) getBlock(height int32) (*btcjson.GetBlockVerboseResult, error) {
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
func (k *AccountsUTXOKeeper) index(address string, block *btcjson.GetBlockVerboseResult) error {

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

			key := MakeKeyAddressUTXO(block.Height, k.netVersion, address, tx.Txid, out.N)

			// Check whether the UTXO is already indexed
			res := db.GetByPrefix(key)
			if len(res) > 0 {
				k.log.Debug("UTXO already indexed")
				continue
			}

			// Construct document object
			doc := DocUTXO{
				TxHash:      tx.Txid,
				Index:       out.N,
				PkScriptStr: out.ScriptPubKey.Hex,
				Value:       out.Value,
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
			key := MakeQueryKeyAddressUTXO(k.netVersion, address, in.Txid, in.Vout)

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

// begin indexes the utxo of the given burner address.
// address argument is the address whose utxos are searched and indexed.
// skipToHeight forces the algorithm to ignore blocks below the height.
// reIndex overwrite the last scanned height to zero, causing a rescan.
func (k *AccountsUTXOKeeper) begin(workerID int, address string, skipToHeight int32, reIndex bool) error {
begin:

	k.log.Debug("Beginning account indexation", "Account", address, "WorkerID", workerID)

	// Get the height of the best block on the burner chain
	bestHeight, err := k.getBestBlockHeight()
	if err != nil {
		err = fmt.Errorf("Failed to get best block height of the burner chain: %s", err)
		k.log.Error(err.Error())
		return err
	}

	lastScannedHeight := int32(0)

	// Get the last scanned block height only if re-index is not requested.
	if !reIndex {
		lastScannedHeight = k.lastScannedHeight(address)
	}

	// If skipToHeight is set and it is greater than the last scanned height,
	// use the skip heigh value as the last scanned height
	if skipToHeight > 0 && skipToHeight > lastScannedHeight {
		lastScannedHeight = skipToHeight
	}

	// If the last scanned block height is at least equal to the
	// current best block height, we should do nothing.
	if lastScannedHeight >= bestHeight {
		k.log.Debug("UTXO database is up to date")
		return nil
	}

	// At this point, the last scanned block heigh is less than the
	// current best block height, it means we need to scan more blocks
	// till we reach the bestHeight.
	height := lastScannedHeight
	for {

		if k.HasStopped() {
			return nil
		}

		height++
		k.log.Debug("Current Block", "CurrentHeight", height, "BestHeight", bestHeight,
			"WorkerID", workerID)
		if height > bestHeight {
			break
		}

		// We need to get the block hash at height
		block, err := k.getBlock(height)
		if err != nil {
			return fmt.Errorf("failed to get block: %s", err)
		}

		if err = k.index(address, block); err != nil {
			return fmt.Errorf("failed to index block (%d): %s", block.Height, err)
		}
	}

	// At this point, the best block height may have increased,
	// so we re-execute operations of the function
	goto begin
}

// balanceOf sums up the total value of all unspent output
func (k *AccountsUTXOKeeper) balanceOf(address string) float64 {
	key := MakeQueryKeyAddressUTXOs(k.netVersion, address)
	result := k.db.GetByPrefix(key)
	var total = decimal.Zero
	for _, o := range result {
		var utxo DocUTXO
		o.Scan(&utxo)
		total = total.Add(decimal.NewFromFloat(utxo.Value))
	}
	totalF, _ := total.Float64()
	return totalF
}

// HasStopped checks whether the keeper has stopped
func (k *AccountsUTXOKeeper) HasStopped() bool {
	k.Lock()
	defer k.Unlock()
	return k.stop
}

// Stop indexing UTXOs
func (k *AccountsUTXOKeeper) Stop() {
	k.Lock()
	defer k.Unlock()
	if k.stop {
		return
	}
	k.stop = true
	k.indexerTicker.Stop()
}

// Begin initiates the scanning and indexing process.
// Must be called in a goroutine.
func (k *AccountsUTXOKeeper) Begin(am *accountmgr.AccountManager,
	numWorkers int, skipToHeight int32, reIndex bool) error {

	go func() {
		<-k.interrupt
		k.Stop()
		k.log.Info("UTXO keeper has been interrupted and is stopping")
	}()

	k.indexerTicker = time.NewTicker(params.BurnerUTXOIndexerIntDur)
	// for range k.indexerTicker.C {
	// 	//
	// }

	// Get the burner accounts
	accounts, err := am.ListBurnerAccounts()
	if err != nil {
		return fmt.Errorf("failed to get burner accounts: %s", err)
	}

	// Only work with accounts compatible with either the mainnet or testnet
	isMainnet := config.IsBurnerMainnet()
	accounts = funk.Filter(accounts, func(a *accountmgr.StoredAccount) bool {
		r := a.GetMeta().Get("testnet")
		return r != nil && r.(bool) == !isMainnet
	}).([]*accountmgr.StoredAccount)

	k.log.Debug("Starting burner accounts UTXO indexation",
		"NumWorkers", numWorkers,
		"NumAccounts", len(accounts))

	// Added the burner accounts to a queue
	queue := *lane.NewQueue()
	for _, a := range accounts {
		queue.Append(a)
	}

	// Start indexer workers to process the queue
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
				if err := k.begin(id, addr, skipToHeight, reIndex); err != nil {
					k.log.Error("Failed to begin UTXO indexation", "Account", addr)
				}
			}
			wg.Done()
		}(i + 1)
	}

	wg.Wait()
	k.log.Info("UTXO keeper has stopped", "NumAccounts", len(accounts))

	return nil
}
