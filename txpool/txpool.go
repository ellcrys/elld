package txpool

import (
	"sync"

	"github.com/olebedev/emitter"

	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/types/core/objects"
	"github.com/ellcrys/elld/util"
)

// TxPool defines a structure and functionalities of a transaction pool
// which is responsible for collecting, validating and providing processed
// transactions for block inclusion and propagation.
type TxPool struct {
	sync.RWMutex                  // general mutex
	queue        *TxContainer     // transaction queue
	event        *emitter.Emitter // event emitter
}

// New creates a new instance of TxPool
// Cap size is the max amount of transactions we can have in the
// queue at any given time. If it is full, transactions will be dropped.
func New(cap int64) *TxPool {
	tp := new(TxPool)
	tp.queue = newQueue(cap)
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
	switch tx.GetType() {
	case objects.TxTypeBalance:
		return tp.addTx(tx)
	default:
		return objects.ErrTxTypeUnknown
	}
}

// addTx adds a transaction to the queue
// and sends out core.EventNewTransaction event.
// (Not thread-safe)
func (tp *TxPool) addTx(tx core.Transaction) error {

	// Ensure the transaction does not
	// already exist in the queue
	if tp.queue.Has(tx) {
		return ErrTxAlreadyAdded
	}

	// Append the the transaction to the
	// the queue. This will cause the pool
	// to be re-sorted
	if !tp.queue.Add(tx) {
		return ErrContainerFull
	}

	// Emit an event about the accepted
	// transaction so it can be relayed etc.
	<-tp.event.Emit(core.EventNewTransaction, tx)

	return nil
}

// Has checks whether a transaction is in the pool
func (tp *TxPool) Has(tx core.Transaction) bool {
	return tp.queue.Has(tx)
}

// SenderHasTxWithSameNonce checks whether a transaction
// with a matching sender address and nonce exist in
// the pool
func (tp *TxPool) SenderHasTxWithSameNonce(address util.String, nonce uint64) bool {
	return tp.queue.IFind(func(tx core.Transaction) bool {
		return tx.GetFrom().Equal(address) && tx.GetNonce() == nonce
	}) != nil
}
