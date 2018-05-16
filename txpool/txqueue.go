package txpool

import (
	"sync"

	"github.com/ellcrys/druid/wire"
)

// TxQueue represents the internal queue used by TxPool.
// It provides Append, Prepend, First, Last, Full, Sort operations. We also
// provide the ability to sort queue by custom function.
// First and Last operations sort the transactions by fees in descending order.
// The queue is synchronized and thread-safe.
type TxQueue struct {
	container        []*wire.Transaction
	cap              int64
	gmx              *sync.RWMutex
	len              int64
	disabledAutoSort bool
}

// NewQueue creates a new queue
func NewQueue(cap int64) *TxQueue {
	q := new(TxQueue)
	q.container = []*wire.Transaction{}
	q.cap = cap
	q.gmx = &sync.RWMutex{}
	return q
}

// NewQueueNoSort creates a new queue with implicit sorting during
// insertion turned off.
func NewQueueNoSort(cap int64) *TxQueue {
	q := new(TxQueue)
	q.container = []*wire.Transaction{}
	q.cap = cap
	q.gmx = &sync.RWMutex{}
	q.disabledAutoSort = true
	return q
}

// Size returns the number of items in the queue
func (q *TxQueue) Size() int64 {
	q.gmx.RLock()
	defer q.gmx.RUnlock()
	return q.len
}

// Full checks if the queue's capacity has been reached
func (q *TxQueue) Full() bool {
	q.gmx.RLock()
	defer q.gmx.RUnlock()
	return q.len >= q.cap
}

// Append adds a transaction to the end of the queue.
// Returns false if queue capacity has been reached
func (q *TxQueue) Append(tx *wire.Transaction) bool {

	if q.Full() {
		return false
	}

	q.gmx.Lock()
	q.container = append(q.container, tx)
	q.len++
	q.gmx.Unlock()

	if !q.disabledAutoSort {
		q.Sort(SortByTxFeeDesc)
	}

	return true
}

// Prepend adds a transaction to the head of the queue.
// Returns false if queue capacity has been reached
func (q *TxQueue) Prepend(tx *wire.Transaction) bool {

	if q.Full() {
		return false
	}

	q.gmx.Lock()
	q.container = append([]*wire.Transaction{tx}, q.container...)
	q.len++
	q.gmx.Unlock()

	if !q.disabledAutoSort {
		q.Sort(SortByTxFeeDesc)
	}

	return true
}

// First returns a single transaction at head.
// Returns nil if queue is empty
func (q *TxQueue) First() *wire.Transaction {

	if q.Size() <= 0 {
		return nil
	}

	q.gmx.Lock()
	defer q.gmx.Unlock()

	tx := q.container[0]
	q.container = q.container[1:]
	q.len--
	return tx
}

// Last returns a single transaction at head.
// Returns nil if queue is empty
func (q *TxQueue) Last() *wire.Transaction {

	if q.Size() <= 0 {
		return nil
	}

	q.gmx.Lock()
	defer q.gmx.Unlock()

	lastIndex := len(q.container) - 1
	tx := q.container[lastIndex]
	q.container = q.container[0:lastIndex]
	q.len--
	return tx
}

// Sort accepts a sort function and passes the container
// to it to be sorted.
func (q *TxQueue) Sort(sf func([]*wire.Transaction)) {
	q.gmx.Lock()
	defer q.gmx.Unlock()
	sf(q.container)
}
