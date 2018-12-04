package common

import (
	"fmt"
	"time"

	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"
)

// OrphanBlock represents an orphan block
type OrphanBlock struct {
	Block      core.Block
	Expiration time.Time
}

// OpTx is used to pass transactions to methods
type OpTx struct {
	Tx        elldb.Tx
	CanFinish bool
	finished  bool
}

// Closed gets the status of the transaction
func (t *OpTx) Closed() bool {
	return t.finished
}

// GetName returns the name of the op
func (t *OpTx) GetName() string {
	return "OpTx"
}

// Discard the transaction. Do not call
// functions in the transaction after this.
func (t *OpTx) Discard() error {
	if !t.CanFinish || t.finished {
		return nil
	}
	t.Tx.Discard()
	t.finished = true
	return nil
}

// Commit commits the transaction if it has not been done before.
// It ignores the call if CanFinish is false.
func (t *OpTx) Commit() error {
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
func (t *OpTx) Rollback() error {
	if !t.CanFinish || t.finished {
		return nil
	}
	t.Tx.Rollback()
	t.finished = true
	return nil
}

// Finishable makes the transaction finishable
func (t *OpTx) Finishable() *OpTx {
	t.CanFinish = true
	return t
}

// SetFinishable sets whether the transaction
// can be finished
func (t *OpTx) SetFinishable(finish bool) *OpTx {
	t.CanFinish = finish
	return t
}

// OpBlockQueryRange defines the minimum and maximum
// block number of objects to access.
type OpBlockQueryRange struct {
	Min uint64
	Max uint64
}

// GetName returns the name of the op
func (o *OpBlockQueryRange) GetName() string {
	return "OpQueryBlockRange"
}

// OpTransitions defines a CallOp for
// passing transition objects
type OpTransitions []Transition

// GetName implements core.CallOp. Allows transitions
func (t *OpTransitions) GetName() string {
	return "OpTransitions"
}

// OpAllowExec defines a CallOp that
// indicates whether to execute something
type OpAllowExec bool

// GetName implements core.CallOp. Allows transitions
func (t OpAllowExec) GetName() string {
	return "OpAllowExec"
}

// OpChainer defines a CallOp for
// passing a chain
type OpChainer struct {
	Chain types.Chainer
	name  string
}

// GetName implements core.CallOp. Allows transitions
func (t *OpChainer) GetName() string {
	return fmt.Sprintf("<chainerOp %s>", t.name)
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

	// Value is the content of this state
	// object. It is written to the database
	// and the tree
	Value []byte
}
