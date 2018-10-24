package common

import (
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"
)

// Transition represents a mutation to the state of a chain
type Transition interface {

	// Address returns the address this transition is to affect
	Address() util.String

	// Equal is used to check whether a transition operation t is similar
	Equal(t Transition) bool
}

// OpBase includes common methods and fields for a transition object
type OpBase struct {
	Addr util.String
}

// Address returns the address to be acted on
func (o *OpBase) Address() util.String {
	return o.Addr
}

// Equal checks whether a Transition t is equal to o
func (o *OpBase) Equal(t Transition) bool {
	return o.Address().Equal(t.Address())
}

// OpCreateAccount describes a transition to create an account
type OpCreateAccount struct {
	*OpBase
	Account core.Account
}

// Equal checks whether a Transition t is equal to o
func (o *OpCreateAccount) Equal(t Transition) bool {
	if _t, yes := t.(*OpCreateAccount); yes && o.Address() == _t.Address() {
		return true
	}
	return false
}

// OpNewAccountBalance represents a transition to a new account balance
type OpNewAccountBalance struct {
	*OpBase
	Account core.Account
}

// Equal checks whether a Transition t is equal to o
func (o *OpNewAccountBalance) Equal(t Transition) bool {
	if _t, yes := t.(*OpNewAccountBalance); yes && o.Address() == _t.Address() {
		return true
	}
	return false
}
