package config

import (
	"fmt"
	"path/filepath"
)

const (
	// ModeProd refers to production mode
	ModeProd = iota
	// ModeDev refers to development mode
	ModeDev
	// ModeTest refers to test mode
	ModeTest
)

// NodeConfig represents node's configuration
type NodeConfig struct {

	// BootstrapAddresses sets addresses to connect to
	BootstrapAddresses []string `json:"bootstrapAddrs" mapstructure:"bootstrapAddrs"`

	// ListeningAddr is the address the node binds on to listen to incoming messages
	ListeningAddr string `json:"address" mapstructure:"address"`

	// Mode determines the current environment type
	Mode int `json:"mode" mapstructure:"mode"`

	// GetAddrInterval is the interval between GetAddr messages
	GetAddrInterval int64 `json:"getAddrInt" mapstructure:"getAddrInt"`

	// PingInterval is the interval between Ping messages
	PingInterval int64 `json:"pingInt" mapstructure:"pingInt"`

	// SelfAdvInterval is the interval self advertisement messages
	SelfAdvInterval int64 `json:"selfAdvInt" mapstructure:"selfAdvInt"`

	// CleanUpInterval is the interval between address clean ups
	CleanUpInterval int64 `json:"cleanUpInt" mapstructure:"cleanUpInt"`

	// MaxAddrsExpected is the maximum number addresses
	// expected from a remote peer
	MaxAddrsExpected int64 `json:"maxAddrsExpected" mapstructure:"maxAddrsExpected"`

	// MaxOutboundConnections is the maximum number of outbound
	// connections the node is allowed
	MaxOutboundConnections int64 `json:"maxOutConnections" mapstructure:"maxOutConnections"`

	// MaxOutboundConnections is the maximum number of inbound
	// connections the node is allowed
	MaxInboundConnections int64 `json:"maxInConnections" mapstructure:"maxInConnections"`

	// ConnEstInterval is the time interval when the node
	ConnEstInterval int64 `json:"conEstInt" mapstructure:"conEstInt"`

	// MessageTimeout is the number of seconds to attempt to
	// connect to or read from a peer.
	MessageTimeout int64 `json:"messageTimeout" mapstructure:"messageTimeout"`

	// Key is the address of the node key to use for start up
	Key string `json:"key" mapstructure:"key"`

	// UTXOKeeperIndexInterval is the number of seconds between
	// every burner accounts utxo indexation execution
	UTXOKeeperIndexInterval int64 `json:"utxoKeeperIndexInt" mapstructure:"utxoKeeperIndexInt"`

	// BurnerBlockIndexInterval is the number of seconds between
	// every burner block indexation execution
	BurnerBlockIndexInterval int64 `json:"burnerBlockIndexInt" mapstructure:"burnerBlockIndexInt"`

	// BurnerTicketIndexInterval is the number of seconds between
	// every ticket indexation execution
	BurnerTicketIndexInterval int64 `json:"burnerTicketIndexInt" mapstructure:"burnerTicketIndexInt"`
}

// RPCConfig defines configuration for the RPC component
type RPCConfig struct {

	// DisableAuth determines whether to
	// perform authentication for requests
	// attempting to invoke private methods.
	DisableAuth bool `json:"disableAuth" mapstructure:"disableAuth"`

	// Username is the RPC username
	Username string `json:"username" mapstructure:"username"`

	// Password is the RPC password
	Password string `json:"password" mapstructure:"password"`

	// SessionSecretKey is the key used to sign the
	// session tokens. Must be kept secret.
	SessionSecretKey string `json:"sessionSecretKey" mapstructure:"sessionSecretKey"`

	// SessionTTL is the duration a session can
	// remain active before it is considered expired.
	SessionTTL int64 `json:"sessionTTL" mapstructure:"sessionTTL"`
}

// TxPoolConfig defines configuration for the transaction pool
type TxPoolConfig struct {

	// Capacity is the maximum amount of item the transaction pool can hold
	Capacity int64 `json:"capacity" mapstructure:"capacity"`
}

// MinerConfig defines configuration for mining
type MinerConfig struct {

	// Mode describes the mining mode
	Mode uint `json:"-" mapstructure:"-"`
}

// VersionInfo describes the clients
// components and runtime version information
type VersionInfo struct {
	BuildVersion string `json:"buildVersion" mapstructure:"buildVersion"`
	BuildCommit  string `json:"buildCommit" mapstructure:"buildCommit"`
	BuildDate    string `json:"buildDate" mapstructure:"buildDate"`
	GoVersion    string `json:"goVersion" mapstructure:"goVersion"`
}

// EngineConfig represents the client's configuration
type EngineConfig struct {

	// Node holds the node configurations
	Node *NodeConfig `json:"node" mapstructure:"node"`

	// TxPool holds transaction pool configurations
	TxPool *TxPoolConfig `json:"txPool" mapstructure:"txPool"`

	// Miner holds mining configurations
	Miner *MinerConfig `json:"mining" mapstructure:"mining"`

	// RPC holds rpc configurations
	RPC *RPCConfig `json:"rpc" mapstructure:"rpc"`

	// dataDir is where the node's config and network data is stored
	dataDir string

	// dataDir is where the network's data is stored
	netDataDir string

	// VersionInfo holds version information
	VersionInfo *VersionInfo `json:"-" mapstructure:"-"`

	// g stores references to global objects that can be
	// used anywhere a config is required. Can help to reduce
	// the complexity method definition
	g *Globals
}

// SetNetDataDir sets the network's data directory
func (c *EngineConfig) SetNetDataDir(d string) {
	c.netDataDir = d
}

// NetDataDir returns the network's data directory
func (c *EngineConfig) NetDataDir() string {
	return c.netDataDir
}

// DataDir returns the application's data directory
func (c *EngineConfig) DataDir() string {
	return c.dataDir
}

// SetDataDir sets the application's data directory
func (c *EngineConfig) SetDataDir(d string) {
	c.dataDir = d
}

// GetDBDir returns the path where database files are stored
func (c *EngineConfig) GetDBDir() string {
	var ns string
	var dbFile = "data%s.db"
	if c.Node.Mode == ModeDev {
		ns = "_" + c.g.NodeKey.Addr().String()
	}
	return filepath.Join(c.NetDataDir(), fmt.Sprintf(dbFile, ns))
}

// GetTicketDBPath returns the path to the ticket database
func (c *EngineConfig) GetTicketDBPath() string {
	var ns string
	var dbFile = "ticket%s.db"
	if c.Node.Mode == ModeDev {
		ns = "_" + c.g.NodeKey.Addr().String()
	}
	return filepath.Join(c.NetDataDir(), fmt.Sprintf(dbFile, ns))
}

// IsDev checks whether the current environment is 'development'
func (c *EngineConfig) IsDev() bool {
	return c.Node.Mode == ModeDev
}
