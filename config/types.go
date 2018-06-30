package config

// PeerConfig represents peer configuration
type PeerConfig struct {

	// BootstrapNodes are the list of nodes to join in other to gain access to the network
	BootstrapNodes []string `json:"boostrapNodes"`

	// Dev, when set to true starts the node on a development. In
	// dev mode, the node cannot communicate with nodes on the public, routable
	// internet. It will also not use the production config directory.
	Dev bool `json:"dev"`

	// Test enables or disables features when running the node in a test environment
	Test bool `json:"-"`

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
	MaxEndorsementPeriodInBlocks uint

	// NumBlocksForTicketMaturity is the number of blocks before an endorser ticket
	// is considered mature.
	NumBlocksForTicketMaturity uint
}

// EngineConfig represents the client's configuration
type EngineConfig struct {

	// Node holds the node configurations
	Node *PeerConfig `json:"peer"`

	// TxPool holds transaction pool configurations
	TxPool *TxPoolConfig `json:"txPool"`

	// Consensus holds consensus related configurations
	Consensus *ConsensusConfig `json:"consensus"`

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

var defaultConfig = EngineConfig{}

func init() {

	defaultConfig.Node = &PeerConfig{
		GetAddrInterval:  1800,
		PingInterval:     1800,
		SelfAdvInterval:  1800,
		CleanUpInterval:  600,
		MaxAddrsExpected: 1000,
		MaxConnections:   100,
		ConnEstInterval:  600,
	}

	defaultConfig.Consensus = &ConsensusConfig{
		MaxEndorsementPeriodInBlocks: 21,
		NumBlocksForTicketMaturity:   21,
	}

	defaultConfig.TxPool = &TxPoolConfig{
		Capacity: 1000,
	}
}
