package console

import (
	"bytes"
	gojson "encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ellcrys/elld/rpc/jsonrpc"
	json "github.com/gorilla/rpc/v2/json2"
)

// RPCClient provides the ability to
// call methods of an rpc server.
type RPCClient struct {
	Address string
}

// RPCClientError creates an error describing
// an issue with the rpc client.
func RPCClientError(msg string) error {
	return fmt.Errorf("rpc client error: %s", msg)
}

// call invokes a method in the server
func (c *RPCClient) call(method string, params interface{}, authToken string) (*jsonrpc.Response, error) {

	// create the message
	message, err := json.EncodeClientRequest(method, params)
	if err != nil {
		return nil, RPCClientError("failed to create message")
	}

	// create a request, pass the message
	// and send the request
	req, err := http.NewRequest("POST", c.Address, bytes.NewBuffer(message))
	if err != nil {
		return nil, RPCClientError(err.Error())
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authToken)
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return nil, RPCClientError(err.Error())
	}

	// decode the result into a jsonrpc.Response
	var result jsonrpc.Response
	if err := gojson.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, RPCClientError(err.Error())
	}

	return &result, nil
}

// RPCConfig holds information required
// to create an rpc client
type RPCConfig struct {
	Client RPCClient
}

// GetAddr constructs the appropriate
// RPC server address
func makeAddr(addr string, secured bool) string {
	var s = "s"
	if !secured {
		s = ""
	}
	return fmt.Sprintf("http%s://%s", s, addr)
}

// startRPCServer starts the RPC server.
// It is intended to be called in the JS environment.
// It will panic if console is in attach mode.
func (e *Executor) startRPCServer() {

	if e.console.attached {
		panic(e.vm.MakeCustomError("AttachError", "cannot start rpc server in attach mode"))
	}

	if e.rpcServer.IsStarted() {
		return
	}

	go e.rpcServer.Serve()

	go func() {
		time.Sleep(10 * time.Millisecond)
		e.console.Prepare()
	}()
}
