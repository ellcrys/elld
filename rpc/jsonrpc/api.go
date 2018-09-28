package jsonrpc

import (
	"strings"

	"github.com/ellcrys/elld/util"
)

// Params represent JSON API parameters
type Params map[string]interface{}

// Scan attempts to convert the params to a struct or map type
func (p *Params) Scan(dest interface{}) error {
	return util.MapDecode(p, &dest)
}

// APIInfo defines a standard API function type
// and other parameters.
type APIInfo struct {

	// Func is the API function to be execute.
	Func func(params interface{}) *Response

	// Private indicates a requirement for a private, authenticated
	// user session before this API function is executed.
	Private bool

	// Namespace is the JS namespace where the method should reside
	Namespace string

	// Description describes the API
	Description string
}

// APISet defines a collection of APIs
type APISet map[string]APIInfo

// Get gets an API function by name
// and namespace
func (a APISet) Get(name string) *APIInfo {
	nameParts := strings.Split(name, "_")

	if len(nameParts) != 2 {
		return nil
	}

	if api, ok := a[nameParts[1]]; ok && api.Namespace == nameParts[0] {
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
