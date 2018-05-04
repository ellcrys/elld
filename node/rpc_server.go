package node

import (
	"fmt"
	"net"
	"net/http"
	"net/rpc"

	"github.com/ellcrys/druid/util/logger"
)

// Service provides functionalities accessible through JSON-RPC
type Service struct {
}

// RPCServer represents a rpc server
type RPCServer struct {
	addr string
	log  logger.Logger
	conn net.Listener
}

// NewRPCServer creates a new server
func NewRPCServer(addr string, log logger.Logger) *RPCServer {
	s := new(RPCServer)
	s.addr = addr
	s.log = log
	return s
}

// Serve starts the server
func (s *RPCServer) Serve() error {

	service := new(Service)
	err := rpc.Register(service)
	if err != nil {
		return fmt.Errorf("failed to start rpc server. %s", err)
	}

	rpc.HandleHTTP()

	s.conn, err = net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("failed to start listening on port %s", s.addr))
	}

	s.log.Info("RPC service started", "Address", s.addr)

	if err = http.Serve(s.conn, nil); err != nil {
		if err != http.ErrServerClosed {
			return fmt.Errorf("failed to serve")
		}
	}

	return nil
}

// Stop stops the server and frees resources
func (s *RPCServer) Stop() {
	if s != nil && s.conn != nil {
		s.conn.Close()
	}
}
