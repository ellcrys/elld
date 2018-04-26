package configdir

// PeerConfig represents peer configuration
type PeerConfig struct {
	BootstrapNodes   []string `json:"boostrapNodes"`
	Dev              bool     `json:"dev"`
	GetAddrInterval  int64    `json:"getAddrInt"`
	PingInterval     int64    `json:"pingInt"`
	SelfAdvInterval  int64    `json:"selfAdvInt"`
	CleanUpInterval  int64    `json:"cleanUpInt"`
	MaxAddrsExpected int64    `json:"maxAddrsExpected"`
	MaxConnections   int64    `json:"maxConnections"`
}

// Config represents the client's configuration
type Config struct {
	Peer *PeerConfig `json:"peer"`
}

var defaultConfig = Config{}

func init() {
	defaultConfig.Peer = &PeerConfig{}
}
