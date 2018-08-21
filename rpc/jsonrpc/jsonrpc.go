package jsonrpc

import (
	"encoding/json"
	"net/http"

	"github.com/thoas/go-funk"

	"github.com/gorilla/mux"
	"github.com/gorilla/rpc/v2"
	"github.com/gorilla/rpc/v2/json2"
)

const (
	middlewareErrCode = -32000
)

// OnRequestFunc is the type of function to use
// as a callback when new requests are received
type OnRequestFunc func(r *http.Request) error

// Request represent a JSON RPC request
type Request struct {
	JSONRPCVersion string `json:"jsonrpc"`
	Method         string `json:"method"`
	Params         Params `json:"params"`
	ID             int    `json:"id,omitempty"`
}

// IsNotification checks whether the request is a notification
// according to JSON RPC specification
func (r Request) IsNotification() bool {
	return r.ID == 0
}

// Err represents JSON RPC error object
type Err struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Response represents a JSON RPC response
type Response struct {
	JSONRPCVersion string      `json:"jsonrpc"`
	Result         interface{} `json:"result,omitempty"`
	Err            *Err        `json:"error,omitempty"`
	ID             int         `json:"id,omitempty"`
}

// IsError checks whether r is an error response
func (r Response) IsError() bool {
	return r.Err != nil
}

// JSONRPC defines a wrapper over mux json rpc
// that works with RPC functions of type `types.API`
// defined in packages that offer RPC APIs.`
type JSONRPC struct {

	// addr is the listening address
	addr string

	// apiSet is a collection of
	// all known API functions
	apiSet APISet

	// OnRequestFunc accepts a function. It is called
	// each time a request is received. It is a good
	// place to verify authentication. If error
	// is returned, the handler is not called and
	// the error is returned.
	OnRequest OnRequestFunc
}

// Error creates an error response
func Error(code int, message string, data interface{}) *Response {
	return &Response{
		JSONRPCVersion: "2.0",
		Err: &Err{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
}

// Success creates a success response
func Success(result interface{}) *Response {
	return &Response{
		JSONRPCVersion: "2.0",
		Result:         result,
	}
}

// New creates a JSONRPC server
func New(addr string) *JSONRPC {
	rpc := &JSONRPC{
		addr:   addr,
		apiSet: APISet{},
	}
	rpc.MergeAPISet(rpc.APIs())
	return rpc
}

// APIs returns system APIs
func (s *JSONRPC) APIs() APISet {
	return APISet{
		"methods": APIInfo{
			Func: func(Params) *Response {
				return Success(s.Methods())
			},
		},
	}
}

// Methods gets the names of all methods
// in the API set.
func (s *JSONRPC) Methods() []string {
	return funk.Keys(s.apiSet).([]string)
}

// Serve starts the server
func (s *JSONRPC) Serve() {

	r := mux.NewRouter()
	server := rpc.NewServer()
	server.RegisterCodec(json2.NewCodec(), "application/json")
	server.RegisterCodec(json2.NewCodec(), "application/json;charset=UTF-8")
	r.Handle("/rpc", server)

	// Set request handler
	http.ListenAndServe(s.addr, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if s.OnRequest != nil {
			if err := s.OnRequest(r); err != nil {
				json.NewEncoder(w).Encode(Error(middlewareErrCode, err.Error(), nil))
				return
			}
		}

		json.NewEncoder(w).Encode(s.handle(w, r))
	}))
}

// MergeAPISet merges an API set with s current api sets
func (s *JSONRPC) MergeAPISet(apiSets ...APISet) {
	for _, set := range apiSets {
		for k, v := range set {
			s.apiSet[k] = v
		}
	}
}

// AddAPI adds an API to s api set
func (s *JSONRPC) AddAPI(name string, api APIInfo) {
	s.apiSet[name] = api
}

// handle processes incoming requests. It validates
// the request according to JSON RPC specification,
// find the method and passes it off.
func (s *JSONRPC) handle(w http.ResponseWriter, r *http.Request) *Response {

	// attempt to decode the body
	var newReq Request
	if err := json.NewDecoder(r.Body).Decode(&newReq); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return Error(-32700, "Parse error", nil)
	}

	// JSON RPC version must be 2.0
	if newReq.JSONRPCVersion != "2.0" {
		w.WriteHeader(http.StatusBadRequest)
		return Error(-32600, "Invalid Request", nil)
	}

	// Method must be known
	f := s.apiSet.Get(newReq.Method)
	if f == nil {
		w.WriteHeader(http.StatusNotFound)
		return Error(-32601, "Method not found", nil)
	}

	resp := f.Func(newReq.Params)
	if resp == nil {
		w.WriteHeader(http.StatusOK)
		return Success(nil)
	}

	if !resp.IsError() {

		resp.ID = newReq.ID

		// a notification. Send no response.
		if newReq.IsNotification() {
			resp.Result = nil
		}

		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}

	return resp
}
