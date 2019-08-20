package burner

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/util"

	"github.com/ellcrys/elld/params"

	"github.com/shopspring/decimal"

	"github.com/jinzhu/gorm"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/util/logger"
	"github.com/ellcrys/ltcd/btcjson"
	_ "github.com/jinzhu/gorm/dialects/sqlite" // sqlite driver
	"github.com/olebedev/emitter"
)

// minTMgrIdxrStartHeight is the minimum block height
// to start the watcher from. We set this to avoid indexing blocks
// well before the launch of the network and far in the history
// of the burn chain
// TODO: Set to a real value before mainnet
var minTMgrIdxrStartHeight = int64(1174296)

// Ticket describes a block producer ticket
type Ticket struct {
	TxID        string      `gorm:"type:varchar(64);index:tx_id,out_index"`
	OutIndex    uint32      `gorm:"index:out_index"`
	MatureBy    int64       `gorm:"index:mature_by"`
	DecayBy     int64       `gorm:"index:decay_by"`
	Producer    util.String `gorm:"type:varchar(40);index:producer"`
	BlockHeight int64
}

// ticketManagerMeta stores meta information
// about the manager
type ticketManagerMeta struct {
	Key   string `gorm:"unique_index"`
	Value int64
}

// TicketManager manages a database of all known tickets.
// It finds and index tickets from the burner chain and
// provides access to them for other processes.
type TicketManager struct {
	*sync.Mutex
	bus       *emitter.Emitter
	interrupt <-chan struct{}
	cfg       *config.EngineConfig
	client    RPCClient
	db        *gorm.DB
	log       logger.Logger
	ticker    *time.Ticker
	stop      bool
}

// NewTicketManager creates an instance of TicketManager
func NewTicketManager(cfg *config.EngineConfig, client RPCClient,
	interrupt <-chan struct{}) *TicketManager {
	log := cfg.G().Log.Module("TicketManager")
	return &TicketManager{
		Mutex:     &sync.Mutex{},
		client:    client,
		interrupt: interrupt,
		cfg:       cfg,
		log:       log,
		bus:       cfg.G().Bus,
	}
}

// openDB opens the ticket database
func (tm *TicketManager) openDB() error {
	var err error
	tm.db, err = gorm.Open("sqlite3", tm.cfg.GetTicketDBPath())
	if err != nil {
		return fmt.Errorf("failed to connect database")
	}

	tm.db.AutoMigrate(&Ticket{})
	tm.db.AutoMigrate(&ticketManagerMeta{})

	return nil
}

// Stop the block indexer
func (tm *TicketManager) Stop() {
	tm.Lock()
	defer tm.Unlock()

	if tm.stop {
		return
	}

	tm.stop = true
	tm.ticker.Stop()

	if tm.db != nil {
		tm.db.Close()
		tm.db = nil
	}
}

// getBestBlockHeight the height of the best block
func (tm *TicketManager) getBestBlockHeight() (int32, error) {
	_, height, err := tm.client.GetBestBlock()
	if err != nil {
		return 0, err
	}
	return height, nil
}

// getBlock returns the block at the given height
func (tm *TicketManager) getBlock(height int64) (*btcjson.GetBlockVerboseResult, error) {
	hash, err := tm.client.GetBlockHash(height)
	if err != nil {
		return nil, err
	}
	block, err := tm.client.GetBlockVerboseTx(hash)
	if err != nil {
		return nil, err
	}
	return block, nil
}

// getLastIndexedHeight returns the height of the last
// block that was indexed
func (tm *TicketManager) getLastIndexedHeight() int64 {
	var meta ticketManagerMeta
	tm.db.Where(ticketManagerMeta{Key: "last_height"}).
		Select("value").
		Find(&meta)
	return meta.Value
}

// deleteTicketFromHeight deletes tickets with height greater or
// equal to the given height. It will reset the 'last_height'
// to the height before the given invalid height.
func (tm *TicketManager) deleteTicketFromHeight(height uint64) error {
	db := tm.db.Begin()
	if err := tm.db.Where("block_height >= ?", height).Delete(&Ticket{}).Error; err != nil {
		db.Rollback()
		return err
	}

	err := db.Where(ticketManagerMeta{Key: "last_height"}).
		Delete(&ticketManagerMeta{}).Error
	if err != nil {
		db.Rollback()
		return err
	}

	meta := &ticketManagerMeta{Key: "last_height", Value: int64(height - uint64(1))}
	if err = db.Create(meta).Error; err != nil {
		db.Rollback()
		return err
	}

	return db.Commit().Error
}

// Begin ticket discovery
func (tm *TicketManager) Begin() error {

	// discoverInterrupt is used to stop the discovery process
	// when an invalid block or block re-org is discovered.
	discoverInterrupt := make(chan struct{})

	// Start a goroutine that listens for general application
	// interrupt signal. If interrupted, it stops manager.
	go func() {
		<-tm.interrupt
		close(discoverInterrupt)
		tm.Stop()
		tm.log.Info("Ticket manager has been interrupted and has stopped")
	}()

	// Open the database
	if err := tm.openDB(); err != nil {
		tm.log.Error(err.Error())
		return err
	}

	tm.log.Debug("Ticket manager has started")

	dur := time.Duration(tm.cfg.Node.BurnerTicketIndexInterval) * time.Second
	tm.ticker = time.NewTicker(dur)

	// Start a goroutine that listens for invalid block / re-org event;
	// It reacts by interrupting the discoverer and deleting
	// tickets found on the invalid block up to the latest/heights block.
	go func() {
		for evt := range tm.bus.On(EventInvalidLocalBlock) {

			if util.IsStructChanClosed(discoverInterrupt) {
				return
			}

			invalidHeight := evt.Args[0].(int64)

			tm.log.Debug("Invalid local block detected. Deleting any invalidated tickets...",
				"Height", invalidHeight)

			// We can only react if the discovery process have not been
			// interrupted. If it isn't, we close the discovery interrupt
			// channel stops it and then delete tickets from the invalid
			// block and its descendants. After we are done, we reinitialize
			// the discover interrupt channel.
			close(discoverInterrupt)
			if err := tm.deleteTicketFromHeight(uint64(invalidHeight)); err != nil {
				tm.log.Error("failed to delete invalid tickets")
			}
			discoverInterrupt = make(chan struct{})
			tm.log.Debug("Invalid ticket removal process completed")
		}
	}()

	// Here we run a goroutine that waits for ticks from
	// the ticker. On each tick, it attempts to find new
	// blocks and index any ticket in them
	go func() {
		for range tm.ticker.C {

			// Reinitialize the discover interrupt channel
			// only if it was previously closed.
			if util.IsStructChanClosed(discoverInterrupt) {
				discoverInterrupt = make(chan struct{})
			}

			tm.discover(discoverInterrupt)
		}
	}()

	return nil
}

// discover takes the height of a given block and
// checks whether there is a ticket in to index
func (tm *TicketManager) discover(discoverInt chan struct{}) *indexResult {

	var result = &indexResult{}

	// Get the height of the best block on the upstream chain
	bestHeight, err := tm.getBestBlockHeight()
	if err != nil {
		result.err = fmt.Errorf("Failed to get best block height of the burner chain: %s", err)
		return result
	}

	// Get the height of the block on the burner chain
	// whose block was last checked for tickets
	lastIndexedHeight := tm.getLastIndexedHeight()

	// Reset to the minimum start block if it is greater
	// than the last indexed height
	if minTMgrIdxrStartHeight > 0 && minTMgrIdxrStartHeight > lastIndexedHeight {
		lastIndexedHeight = minTMgrIdxrStartHeight
	}

	// When the last indexed height is at least up to the
	// best block height of the burner chain, we return.
	if lastIndexedHeight >= int64(bestHeight) {
		result.status = indexUpToDate
		tm.log.Debug("Ticket index is up to date")
		return result
	}

	cursor := lastIndexedHeight + 1

loop:

	// Reinitialize the discover interrupt channel
	// only if it was previously closed.
	if util.IsStructChanClosed(discoverInt) {
		result.status = stoppedDueToInterrupt
		return result
	}

	// Get the block at cursor height
	block, err := tm.getBlock(cursor)
	if err != nil {
		if strings.Index(err.Error(), "Block number out of range") != -1 {
			result.status = blockIndexOutOfRange
			return result
		}
		tm.log.Error("Failed to fetch block header", "Err", err.Error(), "Height", cursor)
		result.err = err
		return result
	}

	tm.log.Debug("Checking a block for tickets", "Height", cursor)

	db := tm.db.Begin()

	// Go through the transaction to find valid OP_RETURN outputs
	for _, tx := range block.RawTx {
		for _, out := range tx.Vout {

			// Validate the output. Return the OP_RETURN data
			// only if the output is valid.
			data, err := isValidOpReturn(out)
			if err != nil {
				continue
			}

			var producerAddr [20]byte
			copy(producerAddr[:], data[:])

			ticket := &Ticket{
				TxID:        tx.Txid,
				OutIndex:    out.N,
				BlockHeight: block.Height,
				Producer:    crypto.RIPEMD160ToAddr(producerAddr),
			}

			// Set maturity and decay heights
			ticket.MatureBy = params.TicketMaturityDur + block.Height
			ticket.DecayBy = ticket.MatureBy + params.TicketDecayDur

			// Create and populate the ticket
			if err := db.Where(Ticket{TxID: tx.Txid, OutIndex: out.N}).
				FirstOrCreate(ticket).Error; err != nil {
				result.err = fmt.Errorf("failed to store ticket: %s", err)
				db.Rollback()
				return result
			}

			tm.log.Debug("Found a ticket",
				"TxID", ticket.TxID,
				"OutIndex", ticket.OutIndex)
		}
	}

	// Delete the previous 'last_height' value to reset it
	if err := db.Where(ticketManagerMeta{Key: "last_height"}).
		Delete(&ticketManagerMeta{}).Error; err != nil {
		result.err = fmt.Errorf("failed to delete previous 'last_height' value: %s", err)
		db.Rollback()
		return result
	}

	// Reset 'last_height' to the current cursor value
	if err := db.Create(ticketManagerMeta{Key: "last_height", Value: cursor}).Error; err != nil {
		result.err = fmt.Errorf("failed to update 'last_height' value: %s", err)
		db.Rollback()
		return result
	}

	// Reinitialize the discover interrupt channel
	// only if it was previously closed.
	if util.IsStructChanClosed(discoverInt) {
		db.Rollback()
		result.status = stoppedDueToInterrupt
		return result
	}

	if err := db.Commit().Error; err != nil {
		result.err = fmt.Errorf("failed to commit tickets: %s", err)
		db.Rollback()
		return result
	}

	cursor++

	goto loop
}

// isValidOpReturn checks whether an output is a valid op_return
// output that can represent a block producer ticket.
func isValidOpReturn(out btcjson.Vout) ([]byte, error) {
	if !strings.HasPrefix(strings.ToLower(out.ScriptPubKey.Asm), "op_return") {
		return nil, fmt.Errorf("missing op_return keyword")
	}

	asmParts := strings.Split(out.ScriptPubKey.Asm, " ")
	if len(asmParts) != 2 {
		return nil, fmt.Errorf("invalid Asm format")
	}

	asmDec, err := hex.DecodeString(asmParts[1])
	if err != nil {
		return nil, fmt.Errorf("unable to decoded from hex")
	}

	// We expect a 22 bytes (0-2 = prefix, 2-22 = producer address) value
	if len(asmDec) != 22 {
		return nil, fmt.Errorf("value must be 22 byte in size")
	} else if !bytes.Equal(asmDec[:2], ticketOpReturnPrefix) {
		return nil, fmt.Errorf("unexpected prefix")
	}

	// We expect value to be >= minimum ticket cost
	val := decimal.NewFromFloat(out.Value)
	if val.LessThan(params.MinimumBurnAmt) {
		return nil, fmt.Errorf("output value is insufficient")
	}

	return asmDec[2:], nil
}
