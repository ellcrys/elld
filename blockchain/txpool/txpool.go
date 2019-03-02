package txpool

import (
	"sync"
	"time"

	"github.com/ellcrys/elld/params"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"
)

// TxPool stores transactions.
type TxPool struct {
	sync.RWMutex              // general mutex
	container    *TxContainer // transaction queue
}

// New creates a new instance of TxPool.
// Cap size is the max amount of transactions
// that can be maintained in the pool.
func New(cap int64) *TxPool {
	tp := new(TxPool)
	tp.container = newTxContainer(cap)
	return tp
}

// Remove removes transactions
func (tp *TxPool) Remove(txs ...types.Transaction) {
	tp.Lock()
	defer tp.Unlock()
	tp.container.Remove(txs...)
	tp.clean()
}

// Put adds a transaction
func (tp *TxPool) Put(tx types.Transaction) error {
	tp.Lock()
	defer tp.Unlock()

	if err := tp.addTx(tx); err != nil {
		return err
	}

	tp.clean()

	return nil
}

// isExpired checks whether a transaction has expired
func (tp *TxPool) isExpired(tx types.Transaction) bool {
	expTime := time.Unix(tx.GetTimestamp(), 0).UTC().AddDate(0, 0, params.TxTTL)
	return time.Now().UTC().After(expTime)
}

// clean removes old transactions
func (tp *TxPool) clean() {
	tp.container.IFind(func(tx types.Transaction) bool {
		if tp.isExpired(tx) {
			tp.container.remove(tx)
		}
		return false
	})
}

// addTx adds a transaction to the queue.
// (Not thread-safe)
func (tp *TxPool) addTx(tx types.Transaction) error {

	switch tx.GetType() {
	case core.TxTypeBalance:
	default:
		return core.ErrTxTypeUnknown
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
func (tp *TxPool) Has(tx types.Transaction) bool {
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
	return tp.container.IFind(func(tx types.Transaction) bool {
		return tx.GetFrom().Equal(address) && tx.GetNonce() == nonce
	}) != nil
}

// Container gets the underlying
// transaction container
func (tp *TxPool) Container() types.TxContainer {
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
