package common

import (
	"time"

	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/types/core"
)

// OrphanBlock represents an orphan block
type OrphanBlock struct {
	Block      core.Block
	Expiration time.Time
}

// TxOp defines a method option for passing external transactions
type TxOp struct {
	Tx        elldb.Tx
	CanFinish bool
	finished  bool
}

// Closed gets the status of the transaction
func (t *TxOp) Closed() bool {
	return t.finished
}

// GetName returns the name of the op
func (t *TxOp) GetName() string {
	return "TxOp"
}

// Commit commits the transaction if it has not been done before.
// It ignores the call if CanFinish is false.
func (t *TxOp) Commit() error {
	if !t.CanFinish || t.finished {
		return nil
	}
	if err := t.Tx.Commit(); err != nil {
		return err
	}
	t.finished = true
	return nil
}

// Rollback rolls back the transaction if it has not been done before.
// It ignores the call if CanFinish is false.
func (t *TxOp) Rollback() error {
	if !t.CanFinish || t.finished {
		return nil
	}
	t.Tx.Rollback()
	t.finished = true
	return nil
}

// Object represents an object that can be converted to JSON encoded byte slice
type Object interface {
	JSON() ([]byte, error)
}

// StateObject describes an object to be stored in a elldb.StateObject.
// Usually created after processing a Transition object.
type StateObject struct {

	// Key represents the key to use
	// to persist the object to database
	Key []byte

	// TreeKey represents the key to use
	// to add a record of this object in
	// a merkle tree
	TreeKey []byte

	// Value is the content of this state
	// object. It is written to the database
	// and the tree
	Value []byte
}
