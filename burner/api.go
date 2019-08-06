package burner

import (
	"github.com/ellcrys/elld/rpc/jsonrpc"
	"github.com/ellcrys/elld/types"
)

// API exposes JSON-RPC methods that perform
// coin burning operations
type API struct {
}

// NewAPI creates an instance of API
func NewAPI() *API {
	return &API{}
}

func (api *API) apiBurn(interface{}) *jsonrpc.Response {
	return nil
}

// APIs returns all API handlers
func (api *API) APIs() jsonrpc.APISet {
	return map[string]jsonrpc.APIInfo{
		// namespace: "personal"
		"burn": {
			Namespace:   types.NamespaceBurner,
			Description: "List all accounts",
			Func:        api.apiBurn,
		},
	}
}
