package configdir

// PeerConfig represents peer configuration
type PeerConfig struct {
	BootstrapNodes   []string `json:"boostrapNodes"`
	Dev              bool     `json:"dev"`
	Test             bool     `json:"-"`
	GetAddrInterval  int64    `json:"getAddrInt"`
	PingInterval     int64    `json:"pingInt"`
	SelfAdvInterval  int64    `json:"selfAdvInt"`
	CleanUpInterval  int64    `json:"cleanUpInt"`
	ConnEstInterval  int64    `json:"conEstInt"`
	MaxAddrsExpected int64    `json:"maxAddrsExpected"`
	MaxConnections   int64    `json:"maxConnections"`
}

// TxPoolConfig defines configuration for the transaction pool
type TxPoolConfig struct {
	Capacity int64 `json:"cap"`
}

// Config represents the client's configuration
type Config struct {
	Node      *PeerConfig   `json:"peer"`
	TxPool    *TxPoolConfig `json:"txPool"`
	configDir string
}

// SetConfigDir sets the config directory
func (c *Config) SetConfigDir(d string) {
	c.configDir = d
}

// ConfigDir returns the config directory
func (c *Config) ConfigDir() string {
	return c.configDir
}

var defaultConfig = Config{}

func init() {
	defaultConfig.Node = &PeerConfig{}
}
