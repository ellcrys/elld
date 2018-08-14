package rpc

import (
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"path/filepath"

	"github.com/ellcrys/elld/types"

	"github.com/ellcrys/elld/accountmgr"

	"github.com/ellcrys/elld/config"

	"github.com/ellcrys/elld/util/logger"
)

// Service provides functionalities accessible through JSON-RPC
type Service struct {
	accountMgr types.APISet
	engine     types.APISet
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
	addr      string
	cfg       *config.EngineConfig
	log       logger.Logger
	conn      net.Listener
	engineAPI types.APISet
}

// NewServer creates a new RPC server
func NewServer(addr string, engineAPI types.APISet, cfg *config.EngineConfig, log logger.Logger) *Server {
	s := new(Server)
	s.addr = addr
	s.log = log
	s.cfg = cfg
	s.engineAPI = engineAPI
	return s
}

// NewService creates a Service instance
func NewService(engine types.APISet, cfg *config.EngineConfig) *Service {
	return &Service{
		accountMgr: accountmgr.New(filepath.Join(cfg.ConfigDir(), config.AccountDirName)).APIs(),
		engine:     engine,
	}
}

// Serve starts the server
func (s *Server) Serve() error {

	service := NewService(s.engineAPI, s.cfg)
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
