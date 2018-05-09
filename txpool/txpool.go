package txpool

import (
	"fmt"
	"sync"
	"time"

	"github.com/ellcrys/druid/wire"
)

// TxPool defines a structure and functionalities of a transaction pool
// which is responsible for collecting, validating and providing processed
// transactions for block inclusion and propagation.
type TxPool struct {
	gmx        *sync.Mutex                      // general mutex
	queue      *TxQueue                         // transaction queue
	queueMap   map[string]struct{}              // maps transactions present in queue by their hash
	onQueuedCB func(tx *wire.Transaction) error // called each time a transaction is queued
}

// NewTxPool creates a new instance of TxPool
// Cap size is the max amount of transactions we can have in the
// queue at any given time. If it is full, transactions will be dropped.
func NewTxPool(cap int64) *TxPool {
	tp := new(TxPool)
	tp.queue = NewQueue(cap)
	tp.gmx = &sync.Mutex{}
	tp.queueMap = make(map[string]struct{})
	return tp
}

// Put adds a transaction to the transaction pool queue.
// Perform signature validation.
// Timestamp validation.
func (tp *TxPool) Put(tx *wire.Transaction) error {

	if tp.queue.Full() {
		return fmt.Errorf("capacity reached")
	}

	if err := wire.TxVerify(tx); err != nil {
		return err
	}

	if tp.Has(tx) {
		return fmt.Errorf("exact transaction already in pool")
	}

	if time.Unix(tx.Timestamp, 0).After(time.Now()) {
		return fmt.Errorf("future time not allowed")
	}

	switch tx.Type {
	case wire.TxTypeRepoCreate:
		tp.addTx(tx)
	}

	return nil
}

func (tp *TxPool) addTx(tx *wire.Transaction) bool {

	tp.gmx.Lock()

	if !tp.queue.Append(tx) {
		tp.gmx.Unlock()
		return false
	}

	tp.queueMap[string(tx.Hash())] = struct{}{}

	if tp.onQueuedCB != nil {
		tp.gmx.Unlock()
		tp.onQueuedCB(tx)
		return true
	}

	tp.gmx.Unlock()
	return true
}

// OnQueued sets the callback to be called each time a transaction has been
// validated and queued.
func (tp *TxPool) OnQueued(f func(tx *wire.Transaction) error) {
	tp.onQueuedCB = f
}

// Has checks whether a transaction has been queued
func (tp *TxPool) Has(tx *wire.Transaction) bool {
	tp.gmx.Lock()
	defer tp.gmx.Unlock()
	_, has := tp.queueMap[string(tx.Hash())]
	return has
}
