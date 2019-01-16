package jsonrpc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/ncodes/authtoken"

	"github.com/gorilla/mux"
	"github.com/gorilla/rpc/v2"
	"github.com/gorilla/rpc/v2/json2"
)

const (
	middlewareErrCode = -32000
	serverErrCode     = -32001
)

// MethodInfo describe an RPC method info
type MethodInfo struct {
	Name        string `json:"name"`
	Namespace   string `json:"-"`
	Description string `json:"description"`
	Private     bool   `json:"private"`
}

// OnRequestFunc is the type of function to use
// as a callback when new requests are received
type OnRequestFunc func(r *http.Request) error

// Request represent a JSON RPC request
type Request struct {
	JSONRPCVersion string      `json:"jsonrpc"`
	Method         string      `json:"method"`
	Params         interface{} `json:"params"`
	ID             uint64      `json:"id,omitempty"`
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
	Result         interface{} `json:"result"`
	Err            *Err        `json:"error,omitempty"`
	ID             uint64      `json:"id,omitempty"`
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

	// sessionKey is used to validate JWT tokens
	sessionKey string

	// disableAuth when set to true causes
	// authorization check to be skipped (not recommended)
	disableAuth bool

	// handlerConfigured lets us know when the
	// handle has been configured
	handlerConfigured bool

	// done is used to wait for the server to
	// shutdown
	done chan bool
}

// MakeSessionToken creates a session token for
// RPC requests
func MakeSessionToken(username, secretKey string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(1 * time.Hour),
	})
	tokenStr, _ := token.SignedString([]byte(secretKey))
	return tokenStr
}

// VerifySessionToken verifies that the given token was created
// using the given secretKey
func VerifySessionToken(tokenStr, secretKey string) error {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		return err
	}
	if token.Valid {
		return nil
	}
	return fmt.Errorf("invalid token")
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
func New(addr string, sessionKey string, disableAuth bool) *JSONRPC {
	rpc := &JSONRPC{
		addr:        addr,
		apiSet:      APISet{},
		sessionKey:  sessionKey,
		disableAuth: disableAuth,
		done:        make(chan bool),
	}
	rpc.MergeAPISet(rpc.APIs())
	return rpc
}

// APIs returns system APIs
func (s *JSONRPC) APIs() APISet {
	return APISet{
		"methods": APIInfo{
			Description: "List RPC methods",
			Namespace:   "rpc",
			Func: func(interface{}) *Response {
				return Success(s.Methods())
			},
		},
	}
}

// Methods gets the names of all methods
// in the API set.
func (s *JSONRPC) Methods() (methodsInfo []MethodInfo) {
	for name, d := range s.apiSet {
		methodsInfo = append(methodsInfo, MethodInfo{
			Name:        name,
			Description: d.Description,
			Namespace:   d.Namespace,
			Private:     d.Private,
		})
	}
	return
}

// Serve starts the server
func (s *JSONRPC) Serve() {

	s.done = make(chan bool)

	r := mux.NewRouter()
	server := rpc.NewServer()
	server.RegisterCodec(json2.NewCodec(), "application/json")
	server.RegisterCodec(json2.NewCodec(), "application/json;charset=UTF-8")
	r.Handle("/rpc", server)

	srv := &http.Server{Addr: s.addr}
	s.registerHandler()
	go srv.ListenAndServe()

	<-s.done
	srv.Shutdown(context.Background())
}

func (s *JSONRPC) registerHandler() {
	if s.handlerConfigured {
		return
	}
	http.HandleFunc("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.OnRequest != nil {
			if err := s.OnRequest(r); err != nil {
				json.NewEncoder(w).Encode(Error(middlewareErrCode, err.Error(), nil))
				return
			}
		}
		json.NewEncoder(w).Encode(s.handle(w, r))
	}))
	s.handlerConfigured = true
}

// Stop stops the RPC server
func (s *JSONRPC) Stop() {
	close(s.done)
}

// MergeAPISet merges an API set with s current api sets
func (s *JSONRPC) MergeAPISet(apiSets ...APISet) {
	for _, set := range apiSets {
		for k, v := range set {
			s.apiSet[v.Namespace+"_"+k] = v
		}
	}
}

// AddAPI adds an API to s api set
func (s *JSONRPC) AddAPI(name string, api APIInfo) {
	s.apiSet[api.Namespace+"_"+name] = api
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

	// If the method request is a private
	// method, we must authenticate the provided bearer
	// token
	if !s.disableAuth && f.Private {
		authToken, err := authtoken.FromRequest(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return Error(-32600, fmt.Sprintf("Invalid Request: %s", err.Error()), nil)
		}
		if err = VerifySessionToken(authToken, s.sessionKey); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return Error(-32600, fmt.Sprintf("Authorization Error: session token is not valid"), nil)
		}
	}

	var resp *Response

	defer func() {
		if rcv, ok := recover().(error); ok {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(Error(serverErrCode, rcv.Error(), nil))
		}
	}()

	resp = f.Func(newReq.Params)
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
