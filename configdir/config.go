package configdir

// PeerConfig represents peer configuration
type PeerConfig struct {
	BootstrapNodes []string `json:"boostrapNodes"`
}

// Config represents the client's configuration
type Config struct {
	Peer *PeerConfig `json:"peer"`
}

var defaultConfig = Config{}

func init() {
	defaultConfig.Peer = &PeerConfig{}
}
