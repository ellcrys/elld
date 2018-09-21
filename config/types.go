package config

import (
	"github.com/ellcrys/elld/miner/blakimoto"
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

	// BootstrapNodes are the list of nodes to join in other to gain access to the network
	BootstrapNodes []string `json:"boostrapNodes"`

	// Mode determines the current environment type
	Mode int `json:"dev"`

	// GetAddrInterval is the time interval when the node sends a GetAddr message to peers
	GetAddrInterval int64 `json:"getAddrInt"`

	// PingInterval is the time interval when the node sends Ping messages to peers
	PingInterval int64 `json:"pingInt"`

	// SelfAdvInterval is the time interval when the node sends a self advertisement Addr message to peers
	SelfAdvInterval int64 `json:"selfAdvInt"`

	// CleanUpInterval is the time interval when the node cleans up disconnected, old peers and updates its address list
	CleanUpInterval int64 `json:"cleanUpInt"`

	// MaxAddrsExpected is the maximum address the node expects to receive from a remote node
	MaxAddrsExpected int64 `json:"maxAddrsExpected"`

	// MaxConnections is the maximum number of connections the node is allowed to maintain
	MaxConnections int64 `json:"maxConnections"`

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
	SessionSecretKey string `json:"sessionSecretKet"`
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

	// Mode describes the blakimoto mining mode
	Mode blakimoto.Mode `json:"-"`
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
}

// SetConfigDir sets the config directory
func (c *EngineConfig) SetConfigDir(d string) {
	c.configDir = d
}

// ConfigDir returns the config directory
func (c *EngineConfig) ConfigDir() string {
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
		GetAddrInterval:  60,
		PingInterval:     60,
		SelfAdvInterval:  120,
		CleanUpInterval:  600,
		MaxAddrsExpected: 1000,
		MaxConnections:   100,
		ConnEstInterval:  10,
		Mode:             ModeProd,
		MessageTimeout:   15,
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
		Mode: blakimoto.ModeNormal,
	}

	defaultConfig.RPC = &RPCConfig{
		Username:         "admin",
		Password:         "admin",
		SessionSecretKey: util.RandString(32),
	}
}
