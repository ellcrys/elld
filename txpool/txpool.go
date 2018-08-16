package txpool

import (
	"fmt"
	"sync"

	"github.com/olebedev/emitter"

	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/wire"
)

// TxPool defines a structure and functionalities of a transaction pool
// which is responsible for collecting, validating and providing processed
// transactions for block inclusion and propagation.
type TxPool struct {
	gmx      *sync.Mutex         // general mutex
	queue    *TxQueue            // transaction queue
	queueMap map[string]struct{} // maps transactions present in queue by their hash
	event    *emitter.Emitter    // event emitter
}

// NewTxPool creates a new instance of TxPool
// Cap size is the max amount of transactions we can have in the
// queue at any given time. If it is full, transactions will be dropped.
func NewTxPool(cap int64) *TxPool {
	tp := new(TxPool)
	tp.queue = NewQueue(cap)
	tp.gmx = &sync.Mutex{}
	tp.queueMap = make(map[string]struct{})
	tp.event = &emitter.Emitter{}
	return tp
}

// SetEventEmitter sets the event emitter
func (tp *TxPool) SetEventEmitter(ee *emitter.Emitter) {
	tp.event = ee
}

// Put adds a transaction to the transaction pool queue.
// Perform signature validation.
// Timestamp validation.
func (tp *TxPool) Put(tx core.Transaction) error {

	if tp.queue.Full() {
		return fmt.Errorf("capacity reached")
	}

	if tp.Has(tx) {
		return fmt.Errorf("exact transaction already in pool")
	}

	switch tx.GetType() {
	case wire.TxTypeBalance:
		return tp.addTx(tx)
	default:
		return wire.ErrTxTypeUnknown
	}
}

// addTx adds a transaction to the queue
// and sends out core.EventNewTransaction event.
func (tp *TxPool) addTx(tx core.Transaction) error {
	tp.gmx.Lock()
	defer tp.gmx.Unlock()

	tp.queueMap[tx.ID()] = struct{}{}
	if !tp.queue.Append(tx) {
		tp.gmx.Unlock()
		return ErrQueueFull
	}

	<-tp.event.Emit(core.EventNewTransaction, tx)

	return nil
}

// Has checks whether a transaction has been queued
func (tp *TxPool) Has(tx core.Transaction) bool {
	tp.gmx.Lock()
	defer tp.gmx.Unlock()
	_, has := tp.queueMap[tx.ID()]
	return has
}
