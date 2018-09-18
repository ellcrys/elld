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

	// When a new block is added to the chain,
	// remove any of its transactions from the pool
	for evt := range tp.event.On(core.EventNewBlock) {
		tp.removeTransactionsInBlock(evt.Args[0].(core.Block))
	}
}

// Put adds a transaction to the transaction pool queue.
// Perform signature validation.
// Timestamp validation.
func (tp *TxPool) Put(tx core.Transaction) error {
	tp.Lock()
	defer tp.Unlock()
	return tp.addTx(tx)
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

	// Emit an event about the accepted
	// transaction so it can be relayed etc.
	<-tp.event.Emit(core.EventNewTransaction, tx)

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

// Select collects transactions from the head
// of the pool up to the specified maxSize.
func (tp *TxPool) Select(maxSize int64) (txs []core.Transaction) {
	tp.Lock()
	defer tp.Unlock()

	curTxSize := int64(0)
	unSelected := []core.Transaction{}
	for tp.Size() > 0 {
		tx := tp.container.First()
		if curTxSize+tx.SizeNoFee() <= maxSize {
			txs = append(txs, tx)
			curTxSize += tx.SizeNoFee()
			continue
		}
		unSelected = append(unSelected, tx)
	}

	// put the unfit transactions back
	for _, tx := range unSelected {
		tp.addTx(tx)
	}

	return
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
