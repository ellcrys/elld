package console

import (
	"bytes"
	gojson "encoding/json"
	"fmt"
	"net/http"

	"github.com/ellcrys/elld/rpc/jsonrpc"
	json "github.com/gorilla/rpc/v2/json2"
)

// RPCClient provides the ability to
// call methods of an rpc server.
type RPCClient string

// RPCClientError creates an error describing
// an issue with the rpc client.
func RPCClientError(msg string) error {
	return fmt.Errorf("rpc client error: %s", msg)
}

// call invokes an method in the server
func (c *RPCClient) call(method string, params jsonrpc.Params) (*jsonrpc.Response, error) {

	// create the message
	message, err := json.EncodeClientRequest(method, params)
	if err != nil {
		return nil, RPCClientError("failed to create message")
	}

	// create a request, pass the message
	// and send the request
	req, err := http.NewRequest("POST", string(*c), bytes.NewBuffer(message))
	if err != nil {
		return nil, RPCClientError(err.Error())
	}

	req.Header.Set("Content-Type", "application/json")
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
	Client  RPCClient
	Secured bool
}

// GetAddr constructs the appropriate
// RPC server address
func (c *RPCConfig) GetAddr() string {
	var s = "s"
	if !c.Secured {
		s = ""
	}
	return fmt.Sprintf("http%s://%s", s, c.Client)
}
