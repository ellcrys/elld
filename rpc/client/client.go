package client

//go:generate mockgen -destination=../mocks/mock_client.go -package=mocks github.com/ellcrys/partnertracker/rpcclient Client

import (
	"bytes"
	encJson "encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/rpc/v2/json"
)

const RequestTimeout = time.Duration(15 * time.Second)

// IClient represents a JSON-RPC client
type IClient interface {
	Call(method string, params interface{}) (interface{}, error)
	New(opts *Options) IClient
	GetOptions() *Options
}

// Client provides the ability create and
// send requests to a JSON-RPC 2.0 service
type Client struct {
	c    *http.Client
	opts *Options
}

// Options describes the options used to
// configure the client
type Options struct {
	Host  string
	Port  int
	HTTPS bool
}

// URL returns a fully formed url to
// use for making requests
func (o *Options) URL() string {
	return "http://" + net.JoinHostPort(o.Host, strconv.Itoa(o.Port))
}

// Error represents a custom JSON-RPC error
type Error struct {
	Data map[string]interface{} `json:"data"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("%v", e)
}

// NewClient creates an instance of Client
func NewClient(opts *Options) *Client {

	if opts == nil {
		opts = &Options{}
	}

	if opts.Host == "" {
		panic("options.host is required")
	}

	if opts.Port == 0 {
		opts.Port = 8999
	}

	return &Client{
		c:    new(http.Client),
		opts: opts,
	}
}

// GetOptions returns the client's option
func (c *Client) GetOptions() *Options {
	return c.opts
}

// Call calls a method on the RPC service.
func (c *Client) Call(method string, params interface{}) (interface{}, error) {

	if c.c == nil {
		return nil, fmt.Errorf("http client and options not set")
	}

	var request = map[string]interface{}{
		"method":  method,
		"params":  params,
		"id":      uint64(rand.Int63()),
		"jsonrpc": "2.0",
	}

	msg, err := encJson.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.opts.URL(), bytes.NewBuffer(msg))
	if err != nil {
		return nil, err
	}

	c.c.Timeout = RequestTimeout
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call method: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("request unsuccessful. Status code: %d. Body: %s",
			resp.StatusCode, string(body))
	}

	var m interface{}
	err = json.DecodeClientResponse(resp.Body, &m)
	if err != nil {
		if e, ok := err.(*json.Error); ok {
			return nil, &Error{Data: e.Data.(map[string]interface{})}
		}
		return nil, err
	}

	return m, nil
}
