package txpool

import (
	"fmt"
	"math/big"
	"sort"
	"sync"

	"github.com/thoas/go-funk"

	"github.com/ellcrys/elld/params"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/util"
	"github.com/shopspring/decimal"
)

var (
	// ErrContainerFull is an error about a full container
	ErrContainerFull = fmt.Errorf("container is full")

	// ErrTxAlreadyAdded is an error about a transaction
	// that is in the pool.
	ErrTxAlreadyAdded = fmt.Errorf("exact transaction already in the pool")
)

// ContainerItem represents the a container
// item. It wraps a transaction and its
// related information
type ContainerItem struct {
	Tx      types.Transaction
	FeeRate util.String
}

// newItem creates a container item
func newItem(tx types.Transaction) *ContainerItem {
	item := &ContainerItem{Tx: tx}
	return item
}

// TxContainer represents the internal container
// used by TxPool. It provides a Put operation
// with automatic sorting by fee rate and nonce.
// The container is thread-safe.
type TxContainer struct {
	container []*ContainerItem // main transaction container (the pool)
	cap       int64            // cap is the amount of transactions in the
	gmx       *sync.RWMutex
	len       int64
	noSorting bool
	index     map[string]interface{}
	byteSize  int64
}

// newTxContainer creates a new container
func newTxContainer(cap int64) *TxContainer {
	q := new(TxContainer)
	q.container = []*ContainerItem{}
	q.cap = cap
	q.gmx = &sync.RWMutex{}
	q.index = map[string]interface{}{}
	return q
}

// NewQueueNoSort creates a new container
// with sorting turned off
func NewQueueNoSort(cap int64) *TxContainer {
	q := new(TxContainer)
	q.container = []*ContainerItem{}
	q.cap = cap
	q.gmx = &sync.RWMutex{}
	q.index = map[string]interface{}{}
	q.noSorting = true
	return q
}

// Size returns the number of items in the container
func (q *TxContainer) Size() int64 {
	q.gmx.RLock()
	defer q.gmx.RUnlock()
	return q.len
}

// ByteSize gets the total byte size of
// all transactions in the container
func (q *TxContainer) ByteSize() int64 {
	return q.byteSize
}

// Full checks if the container's capacity has been reached
func (q *TxContainer) Full() bool {
	q.gmx.RLock()
	defer q.gmx.RUnlock()
	return q.len >= q.cap
}

// Add adds a transaction to the end of the container.
// Returns false if container capacity has been reached.
// It computes the fee rate and sorts the transactions
// after addition.
func (q *TxContainer) Add(tx types.Transaction) bool {

	if q.Full() {
		return false
	}

	item := newItem(tx)

	// Calculate the transaction's fee rate
	// formula: tx fee / size
	txSizeDec := decimal.NewFromBigInt(new(big.Int).SetInt64(tx.GetSizeNoFee()), 0)
	item.FeeRate = util.String(tx.GetFee().Decimal().Div(txSizeDec).StringFixed(params.Decimals))

	q.gmx.Lock()
	q.container = append(q.container, item)
	q.index[tx.GetHash().HexStr()] = struct{}{}
	q.len++
	q.byteSize += tx.GetSizeNoFee()
	q.gmx.Unlock()

	if !q.noSorting {
		q.Sort()
	}

	return true
}

// Has checks whether a transaction is in the container
func (q *TxContainer) Has(tx types.Transaction) bool {
	q.gmx.RLock()
	defer q.gmx.RUnlock()
	return q.index[tx.GetHash().HexStr()] != nil
}

// HasByHash is like Has but accepts a transaction hash
func (q *TxContainer) HasByHash(hash string) bool {
	q.gmx.RLock()
	defer q.gmx.RUnlock()
	return q.index[hash] != nil
}

// First returns a single transaction at head.
// Returns nil if container is empty
func (q *TxContainer) First() types.Transaction {

	if q.Size() <= 0 {
		return nil
	}

	q.gmx.Lock()
	defer q.gmx.Unlock()

	item := q.container[0]
	q.container = q.container[1:]
	delete(q.index, item.Tx.GetHash().HexStr())
	q.byteSize -= item.Tx.GetSizeNoFee()
	q.len--
	return item.Tx
}

// Last returns a single transaction at head.
// Returns nil if container is empty
func (q *TxContainer) Last() types.Transaction {

	if q.Size() <= 0 {
		return nil
	}

	q.gmx.Lock()
	defer q.gmx.Unlock()

	lastIndex := len(q.container) - 1
	item := q.container[lastIndex]
	q.container = q.container[0:lastIndex]
	delete(q.index, item.Tx.GetHash().HexStr())
	q.byteSize -= item.Tx.GetSizeNoFee()
	q.len--
	return item.Tx
}

// Sort sorts the container
func (q *TxContainer) Sort() {
	q.gmx.Lock()
	defer q.gmx.Unlock()
	sort.Slice(q.container, func(i, j int) bool {

		// When transaction i & j belongs to same sender
		// Sort by nonce in ascending order when the nonces are not the same.
		// When they are the same, we sort by the highest fee rate
		if q.container[i].Tx.GetFrom() == q.container[j].Tx.GetFrom() {
			if q.container[i].Tx.GetNonce() < q.container[j].Tx.GetNonce() {
				return true
			}
			if q.container[i].Tx.GetNonce() == q.container[j].Tx.GetNonce() &&
				q.container[i].FeeRate.Decimal().GreaterThan(q.container[j].FeeRate.Decimal()) {
				return true
			}
			return false
		}

		// For other transactions, sort by highest fee rate
		return q.container[i].FeeRate.Decimal().
			GreaterThan(q.container[j].FeeRate.Decimal())
	})
}

// IFind iterates over the transactions
// and passes each to the predicate function.
// When the predicate returns true, it stops
// and returns the last transaction that was
// passed to the predicate.
//
// Do not modify the transaction in the predicate
// as it is a pointer to the transaction in container.
func (q *TxContainer) IFind(predicate func(types.Transaction) bool) types.Transaction {
	q.gmx.Lock()
	defer q.gmx.Unlock()
	for _, item := range q.container {
		if predicate(item.Tx) == true {
			return item.Tx
		}
	}
	return nil
}

// remove removes a transaction.
// Note: Not thread-safe
func (q *TxContainer) remove(txs ...types.Transaction) {
	finalTxs := funk.Filter(q.container, func(o *ContainerItem) bool {
		if funk.Find(txs, func(tx types.Transaction) bool {
			return o.Tx.GetHash().Equal(tx.GetHash())
		}) != nil {
			delete(q.index, o.Tx.GetHash().HexStr())
			q.byteSize -= o.Tx.GetSizeNoFee()
			q.len--
			return false
		}
		return true
	})

	q.container = finalTxs.([]*ContainerItem)
}

// Remove removes a transaction
func (q *TxContainer) Remove(txs ...types.Transaction) {
	q.gmx.Lock()
	defer q.gmx.Unlock()
	q.remove(txs...)
}
