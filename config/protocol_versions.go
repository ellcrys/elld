package config

import "sync"

// DefaultNetVersion is the default network
// version used when no network version is provided.
const DefaultNetVersion = "0001"

// MainNetVersion is the main net version number
const MainNetVersion = "0001"

var (
	// versions contains protocol handlers versions information
	versions      *ProtocolVersions
	cfgLck        = &sync.RWMutex{}
	curNetVersion = DefaultNetVersion
	burnerMainnet = true
)

// IsMainNetVersion checks whether a given version represents the mainnet version
func IsMainNetVersion(version string) bool {
	return version == MainNetVersion
}

// IsMainNet returns true if the current network is the mainnet
func IsMainNet() bool {
	return GetNetVersion() == MainNetVersion
}

// SetBurnerMainnet sets whether the burner chain is a mainnet
func SetBurnerMainnet(mainnet bool) {
	burnerMainnet = mainnet
}

// IsBurnerMainnet returns true if the burner node is on the mainnet
func IsBurnerMainnet() bool {
	return burnerMainnet
}

// SetVersions sets the protocol version.
// All protocol handlers will be prefixed
// with the version to create a
func SetVersions(netVersion string) {
	cfgLck.Lock()
	defer cfgLck.Unlock()

	if netVersion == "" {
		netVersion = DefaultNetVersion
	} else {
		curNetVersion = netVersion
	}

	versions = &ProtocolVersions{
		Protocol:       netVersion,
		Handshake:      netVersion + "/handshake/1",
		Ping:           netVersion + "/ping/1",
		GetAddr:        netVersion + "/getaddr/1",
		Addr:           netVersion + "/addr/1",
		Tx:             netVersion + "/tx/1",
		BlockBody:      netVersion + "/blockbody/1",
		GetBlockHashes: netVersion + "/getblockhashes/1",
		RequestBlock:   netVersion + "/requestblock/1",
		GetBlockBodies: netVersion + "/getblockbodies/1",
	}
}

// GetVersions returns the protocol version object
func GetVersions() *ProtocolVersions {
	cfgLck.RLock()
	defer cfgLck.RUnlock()
	return versions
}

// GetNetVersion returns the current network version
func GetNetVersion() string {
	return curNetVersion
}

func init() {
	SetVersions("")
}

// ProtocolVersions hold protocol message handler versions
type ProtocolVersions struct {

	// Protocol is the network version supported by this client
	Protocol string

	// Handshake is the message version for handling wire.Handshake
	Handshake string

	// Ping is the version for handling wire.Ping messages
	Ping string

	// GetAddr is the message version for handling wire.GetAddr messages
	GetAddr string

	// Addr is the message version for handling wire.Addr messages
	Addr string

	// Tx is the message version for handing wire.Transaction messages
	Tx string

	// BlockInfo is the message version for handling wire.BlockInfo messages
	BlockInfo string

	// BlockBody is the message version for handling wire.BlockBody messages
	BlockBody string

	// GetBlockHashes is the message version for handling wire.BlockHashes messages
	GetBlockHashes string

	// RequestBlock is the message version for handling wire.RequestBlock messages
	RequestBlock string

	// GetBlockBodies is the message version for handling wire.GetBlockBodies messages
	GetBlockBodies string
}
