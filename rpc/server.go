package rpc

import (
	"fmt"
	"net"
	"net/http"
	"net/rpc"

	"github.com/ellcrys/elld/logic"

	"github.com/ellcrys/elld/util/logger"
)

// Service provides functionalities accessible through JSON-RPC
type Service struct {
	logic *logic.Logic
}

// Result represent a response to a service method call
type Result struct {
	Error   string
	ErrCode int
	Status  int
	Data    map[string]interface{}
}

// Server represents a rpc server
type Server struct {
	addr  string
	log   logger.Logger
	conn  net.Listener
	logic *logic.Logic
}

// NewServer creates a new RPC server
func NewServer(addr string, logic *logic.Logic, log logger.Logger) *Server {
	s := new(Server)
	s.addr = addr
	s.log = log
	s.logic = logic
	return s
}

// Serve starts the server
func (s *Server) Serve() error {

	service := new(Service)
	service.logic = s.logic
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
func (s *Server) Stop() {
	if s != nil && s.conn != nil {
		s.conn.Close()
	}
}
