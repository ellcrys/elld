package txpool

import (
	"fmt"
	"sync"

	"github.com/ellcrys/elld/wire"
)

// TxPool defines a structure and functionalities of a transaction pool
// which is responsible for collecting, validating and providing processed
// transactions for block inclusion and propagation.
type TxPool struct {
	gmx            *sync.Mutex                      // general mutex
	queue          *TxQueue                         // transaction queue
	queueMap       map[string]struct{}              // maps transactions present in queue by their hash
	beforeAppendCB func(tx *wire.Transaction) error // called each time a transaction is queued
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

	if tp.Has(tx) {
		return fmt.Errorf("exact transaction already in pool")
	}

	switch tx.Type {
	case wire.TxTypeBalance:
		return tp.addTx(tx)
	default:
		return wire.ErrTxTypeUnknown
	}
}

func (tp *TxPool) addTx(tx *wire.Transaction) error {

	tp.gmx.Lock()
	tp.queueMap[tx.ID()] = struct{}{}
	if tp.beforeAppendCB != nil {
		if err := tp.beforeAppendCB(tx); err != nil {
			tp.gmx.Unlock()
			return err
		}
	}

	if !tp.queue.Append(tx) {
		tp.gmx.Unlock()
		return ErrQueueFull
	}

	tp.gmx.Unlock()

	return nil
}

// BeforeAppend sets the callback to be called before a transaction is added to the queue
func (tp *TxPool) BeforeAppend(f func(tx *wire.Transaction) error) {
	tp.beforeAppendCB = f
}

// Has checks whether a transaction has been queued
func (tp *TxPool) Has(tx *wire.Transaction) bool {
	tp.gmx.Lock()
	defer tp.gmx.Unlock()
	_, has := tp.queueMap[tx.ID()]
	return has
}
