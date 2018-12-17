package node

import (
	"fmt"
	"time"

	"gopkg.in/oleiade/lane.v1"

	"github.com/ellcrys/elld/blockchain"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util/logger"
	"github.com/olebedev/emitter"
)

// TxManager is responsible for processing
// incoming transactions, constructing new
// transactions and more.
type TxManager struct {

	// engine is the node's instance
	engine *Node

	// evt is the global event emitter
	evt *emitter.Emitter

	// log is the logger used by this module
	log logger.Logger

	// bChain is the blockchain manager
	bChain types.Blockchain

	// txBroadcastQueue store transactions to broadcast
	txBroadcastQueue *lane.Deque
}

// NewTxManager creates a new transaction manager
func NewTxManager(n *Node) *TxManager {
	return &TxManager{
		engine:           n,
		log:              n.log,
		evt:              n.event,
		bChain:           n.bChain,
		txBroadcastQueue: lane.NewDeque(),
	}
}

// Manage incoming transaction related events
func (tm *TxManager) Manage() {

	go func() {
		ticker := time.NewTicker(3 * time.Second)
		for {
			select {
			case <-ticker.C:
				go tm.broadcastTx()
			}
		}
	}()

	go func() {
		for evt := range tm.evt.Once(core.EventTransactionReceived) {
			tm.AddTx(evt.Args[0].(*core.Transaction))
		}
	}()
}

// AddTx adds a transaction to the pool
func (tm *TxManager) AddTx(tx types.Transaction) error {

	// TxTypeAlloc transactions are not allowed
	if tx.GetType() == core.TxTypeAlloc {
		err := fmt.Errorf("allocation transaction type is not allowed")
		go tm.evt.Emit(core.EventTransactionInvalid, tx, err)
		return err
	}

	// We need to validate the transaction, returning
	// the first error we find.
	txValidator := blockchain.NewTxValidator(tx, tm.engine.txsPool, tm.bChain)
	if errs := txValidator.Validate(); len(errs) > 0 {
		go tm.evt.Emit(core.EventTransactionInvalid, tx, errs[0])
		return errs[0]
	}

	// Next we attempt to add the transaction
	// to the transactions pool.
	if err := tm.engine.GetTxPool().Put(tx); err != nil {
		go tm.evt.Emit(core.EventTransactionInvalid, tx, err)
		return err
	}

	// Since we successfully added the transaction
	// to the pool, we need to add it to the
	// broadcast queue so it will be broadcast to peers
	tm.txBroadcastQueue.Append(tx)

	go tm.evt.Emit(core.EventTransactionPooled, tx)

	return nil
}

// broadcastTx broadcast a transaction
func (tm *TxManager) broadcastTx() error {

	tx := tm.txBroadcastQueue.Shift()
	if tx == nil {
		return nil
	}

	return tm.engine.gossipMgr.BroadcastTx(tx.(types.Transaction),
		tm.engine.PM().GetAcquaintedPeers())
}
