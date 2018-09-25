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
	container    *TxContainer     // transaction queue
	event        *emitter.Emitter // event emitter
}

// New creates a new instance of TxPool
// Cap size is the max amount of transactions we can have in the
// queue at any given time. If it is full, transactions will be dropped.
func New(cap int64) *TxPool {
	tp := new(TxPool)
	tp.container = newTxContainer(cap)
	tp.event = &emitter.Emitter{}
	return tp
}

// SetEventEmitter sets the event emitter
func (tp *TxPool) SetEventEmitter(ee *emitter.Emitter) {
	tp.event = ee
	go tp.handleNewBlockEvent()
}

func (tp *TxPool) removeTransactionsInBlock(block core.Block) {
	tp.container.Remove(block.GetTransactions()...)
}

// handleNewBlockEvent removes transactions from
// the pool if they are contained in the broadcast
// block that was appended to the chain
func (tp *TxPool) handleNewBlockEvent() {
	for {
		select {
		case evt := <-tp.event.Once(core.EventNewBlock):
			tp.removeTransactionsInBlock(evt.Args[0].(core.Block))
		}
	}
}

// Put adds a transaction
func (tp *TxPool) Put(tx core.Transaction) error {
	tp.Lock()
	defer tp.Unlock()

	if err := tp.addTx(tx); err != nil {
		return err
	}

	// Emit an event about the accepted transaction
	tp.event.Emit(core.EventNewTransaction, tx)

	return nil
}

// PutSilently is like Put but it does not
// emit an event on success.
func (tp *TxPool) PutSilently(tx core.Transaction) error {
	tp.Lock()
	defer tp.Unlock()

	if err := tp.addTx(tx); err != nil {
		return err
	}

	return nil
}

// addTx adds a transaction to the queue
// and sends out core.EventNewTransaction event.
// (Not thread-safe)
func (tp *TxPool) addTx(tx core.Transaction) error {

	switch tx.GetType() {
	case objects.TxTypeBalance:
	default:
		return objects.ErrTxTypeUnknown
	}

	// Ensure the transaction does not
	// already exist in the queue
	if tp.container.Has(tx) {
		return ErrTxAlreadyAdded
	}

	// Append the the transaction to the
	// the queue. This will cause the pool
	// to be re-sorted
	if !tp.container.Add(tx) {
		return ErrContainerFull
	}

	return nil
}

// Has checks whether a transaction is in the pool
func (tp *TxPool) Has(tx core.Transaction) bool {
	return tp.container.Has(tx)
}

// HasByHash is like Has but accepts a hash
func (tp *TxPool) HasByHash(hash string) bool {
	return tp.container.HasByHash(hash)
}

// SenderHasTxWithSameNonce checks whether a transaction
// with a matching sender address and nonce exist in
// the pool
func (tp *TxPool) SenderHasTxWithSameNonce(address util.String, nonce uint64) bool {
	return tp.container.IFind(func(tx core.Transaction) bool {
		return tx.GetFrom().Equal(address) && tx.GetNonce() == nonce
	}) != nil
}

// Container gets the underlying
// transaction container
func (tp *TxPool) Container() core.TxContainer {
	return tp.container
}

// ByteSize gets the total byte size of
// all transactions in the pool
func (tp *TxPool) ByteSize() int64 {
	return tp.container.ByteSize()
}

// Size gets the total number of transactions
// in the pool
func (tp *TxPool) Size() int64 {
	return tp.container.Size()
}
