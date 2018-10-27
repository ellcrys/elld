package config

import (
	"github.com/ellcrys/elld/util"
)

const (
	// ModeProd refers to production mode
	ModeProd = iota
	// ModeDev refers to development mode
	ModeDev
	// ModeTest refers to test mode
	ModeTest
)

// PeerConfig represents peer configuration
type PeerConfig struct {

	// BootstrapAddresses sets addresses to connect to
	BootstrapAddresses []string `json:"addresses"`

	// Mode determines the current environment type
	Mode int `json:"dev"`

	// GetAddrInterval is the interval between GetAddr messages
	GetAddrInterval int64 `json:"getAddrInt"`

	// PingInterval is the interval between Ping messages
	PingInterval int64 `json:"pingInt"`

	// SelfAdvInterval is the interval self advertisement messages
	SelfAdvInterval int64 `json:"selfAdvInt"`

	// CleanUpInterval is the interval between address clean ups
	CleanUpInterval int64 `json:"cleanUpInt"`

	// MaxAddrsExpected is the maximum number addresses
	// expected from a remote peer
	MaxAddrsExpected int64 `json:"maxAddrsExpected"`

	// MaxOutboundConnections is the maximum number of outbound
	// connections the node is allowed
	MaxOutboundConnections int64 `json:"maxOutConnections"`

	// MaxOutboundConnections is the maximum number of inbound
	// connections the node is allowed
	MaxInboundConnections int64 `json:"maxInConnections"`

	// ConnEstInterval is the time interval when the node
	ConnEstInterval int64 `json:"conEstInt"`

	// MessageTimeout is the number of seconds to attempt to
	// connect to or read from a peer.
	MessageTimeout int64
}

// RPCConfig defines configuration for the RPC component
type RPCConfig struct {

	// DisableAuth determines whether to
	// perform authentication for requests
	// attempting to invoke private methods.
	DisableAuth bool `json:"disableAuth"`

	// Username is the RPC username
	Username string `json:"username"`

	// Password is the RPC password
	Password string `json:"password"`

	// SessionSecretKey is the key used to sign the
	// session tokens. Must be kept secret.
	SessionSecretKey string `json:"sessionSecretKey"`
}

// TxPoolConfig defines configuration for the transaction pool
type TxPoolConfig struct {

	// Capacity is the maximum amount of item the transaction pool can hold
	Capacity int64 `json:"cap"`
}

// ConsensusConfig defines configuration for consensus processes
type ConsensusConfig struct {

	// MaxEndorsementPeriodInBlocks is the amount of blocks after ticket maturity an endorser can
	// continue to perform endorsement functions.
	MaxEndorsementPeriodInBlocks uint `json:"maxEndorsementPeriodInBlocks"`

	// NumBlocksForTicketMaturity is the number of blocks before an endorser ticket
	// is considered mature.
	NumBlocksForTicketMaturity uint `json:"numBlocksForTicketMaturity"`
}

// MinerConfig defines configuration for mining
type MinerConfig struct {

	// Mode describes the mining mode
	Mode uint `json:"-"`
}

// VersionInfo describes the clients
// components and runtime version information
type VersionInfo struct {
	BuildVersion string
	BuildCommit  string
	BuildDate    string
	GoVersion    string
}

// EngineConfig represents the client's configuration
type EngineConfig struct {

	// Node holds the node configurations
	Node *PeerConfig `json:"peer"`

	// TxPool holds transaction pool configurations
	TxPool *TxPoolConfig `json:"txPool"`

	// Consensus holds consensus related configurations
	Consensus *ConsensusConfig `json:"consensus"`

	// Chain holds blockchain related configurations
	Chain *ChainConfig `json:"chain"`

	// Miner holds mining configurations
	Miner *MinerConfig `json:"mining"`

	// RPC holds rpc configurations
	RPC *RPCConfig `json:"rpc"`

	// configDir is where the node's config and data is stored
	configDir string

	// VersionInfo holds version information
	VersionInfo *VersionInfo `json:"-"`
}

// SetConfigDir sets the config directory
func (c *EngineConfig) SetConfigDir(d string) {
	c.configDir = d
}

// DataDir returns the config directory
func (c *EngineConfig) DataDir() string {
	return c.configDir
}

// CheckPoint describes a point on the chain. We use it to prevent
// blocks dated far back in the history of the chain from causing a
// chain reorganization.
type CheckPoint struct {
	Number uint64 `json:"number"`
	Hash   string `json:"hash"`
}

// ChainConfig includes parameters for the chain
type ChainConfig struct {

	// Checkpoints includes a collection of points on the chain of
	// which blocks are supposed to exists after or before.
	Checkpoints []*CheckPoint `json:"checkpoints"`
}

var defaultConfig = EngineConfig{}

func init() {

	defaultConfig.Node = &PeerConfig{
		GetAddrInterval:        60,
		PingInterval:           60,
		SelfAdvInterval:        120,
		CleanUpInterval:        600,
		MaxAddrsExpected:       1000,
		MaxOutboundConnections: 10,
		MaxInboundConnections:  115,
		ConnEstInterval:        10,
		Mode:                   ModeProd,
		MessageTimeout:         60,
	}

	defaultConfig.Consensus = &ConsensusConfig{
		MaxEndorsementPeriodInBlocks: 21,
		NumBlocksForTicketMaturity:   21,
	}

	defaultConfig.TxPool = &TxPoolConfig{
		Capacity: 10000,
	}

	defaultConfig.Chain = &ChainConfig{
		Checkpoints: nil,
	}

	defaultConfig.Miner = &MinerConfig{
		Mode: 0,
	}

	defaultConfig.RPC = &RPCConfig{
		Username:         "admin",
		Password:         "admin",
		SessionSecretKey: util.RandString(32),
	}
}
