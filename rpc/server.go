package rpc

import (
	"net"

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

	// conn is the listener
	conn net.Listener

	// rpc is the JSONRPC 2.0 server
	rpc *jsonrpc.JSONRPC
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

// Serve starts the server
func (s *Server) Serve() error {
	s.log.Info("RPC service started", "Address", s.addr)
	s.AddAPI(s.APIs())
	s.rpc.Serve()
	return nil
}

// Stop stops the server and frees resources
func (s *Server) Stop() {
	if s != nil && s.conn != nil {
		s.conn.Close()
	}
}

// AddAPI adds one or more API sets
func (s *Server) AddAPI(apis ...jsonrpc.APISet) {
	s.rpc.MergeAPISet(apis...)
}
