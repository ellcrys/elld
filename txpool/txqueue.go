package txpool

import (
	"fmt"
	"sync"

	"github.com/ellcrys/elld/types/core"
)

var (
	// ErrQueueFull is an error about a full queue
	ErrQueueFull = fmt.Errorf("queue is full")

	// ErrTxAlreadyAdded is an error about a transaction
	// that is in the pool.
	ErrTxAlreadyAdded = fmt.Errorf("exact transaction already in the pool")
)

// TxQueue represents the internal queue used by TxPool.
// It provides Append, Prepend, First, Last, Full, Sort operations. We also
// provide the ability to sort queue by custom function.
// First and Last operations sort the transactions by fees in descending order.
// The queue is synchronized and thread-safe.
type TxQueue struct {
	container        []core.Transaction // main transaction container (the pool)
	cap              int64              // cap is the amount of transactions in the
	gmx              *sync.RWMutex
	len              int64
	disabledAutoSort bool
	index            map[string]interface{}
}

// newQueue creates a new queue
func newQueue(cap int64) *TxQueue {
	q := new(TxQueue)
	q.container = []core.Transaction{}
	q.cap = cap
	q.gmx = &sync.RWMutex{}
	q.index = map[string]interface{}{}
	return q
}

// NewQueueNoSort creates a new queue with implicit sorting during
// insertion turned off.
func NewQueueNoSort(cap int64) *TxQueue {
	q := new(TxQueue)
	q.container = []core.Transaction{}
	q.cap = cap
	q.gmx = &sync.RWMutex{}
	q.index = map[string]interface{}{}
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
func (q *TxQueue) Append(tx core.Transaction) bool {

	if q.Full() {
		return false
	}

	q.gmx.Lock()
	q.container = append(q.container, tx)
	q.index[tx.GetHash().HexStr()] = struct{}{}
	q.len++
	q.gmx.Unlock()

	if !q.disabledAutoSort {
		q.Sort(SortByTxFeeDesc)
	}

	return true
}

// Prepend adds a transaction to the head of the queue.
// Returns false if queue capacity has been reached
func (q *TxQueue) Prepend(tx core.Transaction) bool {

	if q.Full() {
		return false
	}

	q.gmx.Lock()
	q.container = append([]core.Transaction{tx}, q.container...)
	q.index[tx.GetHash().HexStr()] = struct{}{}
	q.len++
	q.gmx.Unlock()

	if !q.disabledAutoSort {
		q.Sort(SortByTxFeeDesc)
	}

	return true
}

// Has checks whether a transaction is in the queue
func (q *TxQueue) Has(tx core.Transaction) bool {
	q.gmx.RLock()
	defer q.gmx.RUnlock()
	return q.index[tx.GetHash().HexStr()] != nil
}

// First returns a single transaction at head.
// Returns nil if queue is empty
func (q *TxQueue) First() core.Transaction {

	if q.Size() <= 0 {
		return nil
	}

	q.gmx.Lock()
	defer q.gmx.Unlock()

	tx := q.container[0]
	q.container = q.container[1:]
	delete(q.index, tx.GetHash().HexStr())
	q.len--
	return tx
}

// Last returns a single transaction at head.
// Returns nil if queue is empty
func (q *TxQueue) Last() core.Transaction {

	if q.Size() <= 0 {
		return nil
	}

	q.gmx.Lock()
	defer q.gmx.Unlock()

	lastIndex := len(q.container) - 1
	tx := q.container[lastIndex]
	q.container = q.container[0:lastIndex]
	delete(q.index, tx.GetHash().HexStr())
	q.len--
	return tx
}

// Sort accepts a sort function and passes the container
// to it to be sorted.
func (q *TxQueue) Sort(sf func([]core.Transaction)) {
	q.gmx.Lock()
	defer q.gmx.Unlock()
	sf(q.container)
}

// IFind iterates over the transactions
// and passes each to the predicate function.
// When the predicate returns true, it stops
// and returns the last transaction that was
// passed to the predicate.
//
// Do not modify the transaction in the predicate
// as it is a pointer to the transaction in queue.
func (q *TxQueue) IFind(predicate func(core.Transaction) bool) core.Transaction {
	q.gmx.Lock()
	defer q.gmx.Unlock()
	for _, tx := range q.container {
		if predicate(tx) == true {
			return tx
		}
	}
	return nil
}
