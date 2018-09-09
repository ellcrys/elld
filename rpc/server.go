package rpc

import (
	"github.com/ellcrys/elld/elldb"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/rpc/jsonrpc"
	"github.com/ellcrys/elld/util/logger"
)

// Result represent a response to a service method call
type Result struct {
	Error   string
	ErrCode int
	Status  int
	Data    map[string]interface{}
}

// Server represents a rpc server
type Server struct {

	// db is the raw database
	db elldb.DB

	// addr is the address to bind the server to
	addr string

	// cfg is the engine config
	cfg *config.EngineConfig

	// log is the logger
	log logger.Logger

	// rpc is the JSONRPC 2.0 server
	rpc *jsonrpc.JSONRPC

	// started indicates the start state
	// of the server
	started bool
}

// NewServer creates a new RPC server
func NewServer(db elldb.DB, addr string, cfg *config.EngineConfig, log logger.Logger) *Server {
	s := new(Server)
	s.db = db
	s.addr = addr
	s.log = log
	s.cfg = cfg
	s.rpc = jsonrpc.New(addr, cfg.RPC.SessionSecretKey, cfg.RPC.DisableAuth)
	return s
}

// GetAddr gets the address
func (s *Server) GetAddr() string {
	return s.addr
}

// Serve starts the server
func (s *Server) Serve() {
	s.AddAPI(s.APIs())
	s.started = true
	s.log.Info("RPC service started", "Address", s.addr)
	s.rpc.Serve()
}

// IsStarted returns the start state
func (s *Server) IsStarted() bool {
	return s.started
}

// Stop stops the server and frees resources
func (s *Server) Stop() {
	s.rpc.Stop()
	s.started = false
}

// AddAPI adds one or more API sets
func (s *Server) AddAPI(apis ...jsonrpc.APISet) {
	s.rpc.MergeAPISet(apis...)
}
