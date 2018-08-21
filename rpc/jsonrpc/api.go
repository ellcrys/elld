package jsonrpc

import (
	"github.com/mitchellh/mapstructure"
)

// Params represent JSON API parameters
type Params map[string]interface{}

// Scan attempts to convert the params to a struct or map type
func (p *Params) Scan(dest interface{}) error {
	return mapstructure.Decode(p, &dest)
}

// APIInfo defines a standard API function type
// and other parameters.
type APIInfo struct {

	// Func is the API function to be execute.
	Func func(params Params) Response

	// Private indicates a requirement for a private, authenticated
	// user session before this API function is executed.
	Private bool
}

// APISet defines a collection of APIs
type APISet map[string]APIInfo

// Get gets an API function by name.
func (a APISet) Get(name string) *APIInfo {
	if api, ok := a[name]; ok {
		return &api
	}
	return nil
}

// API defines an interface for providing and
// accessing API functions. Packages that offer
// services accessed via RPC or any service-oriented
// interface must implement it.
type API interface {
	APIs() APISet
}
