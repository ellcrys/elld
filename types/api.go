package types

import (
	"fmt"
)

// APIFunc defines a standard API function type.
// It takes an arbitrary amount of arguments and
// returns a result and an error.
type APIFunc func(args ...interface{}) (interface{}, error)

// APISet defines a collection of APIs
type APISet map[string]APIFunc

// MustGet gets an API function by name.
// Panics if none is found.
func (a APISet) MustGet(name string) APIFunc {
	if api, ok := a[name]; ok {
		return api
	}
	panic(fmt.Errorf("api with name <%s> not in set", name))
}

// API defines an interface for providing and
// accessing API functions. Packages that offer
// services accessed via RPC or any service-oriented
// interface must implement it.
type API interface {
	APIs() APISet
}
