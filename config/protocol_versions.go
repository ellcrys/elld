package config

import "os"

// DefaultNetVersion is the default network
// version used when no network version is provided.
const DefaultNetVersion = "0001"

var (
	// Versions contains protocol handlers versions information
	Versions *ProtocolVersions
)

// SetVersions sets the protocol handler version
func SetVersions() {

	var netVersion = DefaultNetVersion
	if ver := os.Getenv("ELLD_NET_VERSION"); len(ver) > 0 {
		netVersion = ver
	}

	Versions = &ProtocolVersions{
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

func init() {
	SetVersions()
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

// var (

// 	// ProtocolVersion is the protocol version
// 	// spoken by this client
// 	ProtocolVersion = "/" + GetNetworkVersion()

// 	// HandshakeVersion is the message version for
// 	// handling wire.Handshake
// 	HandshakeVersion = ProtocolVersion + "/handshake/1"

// 	// PingVersion is the version for handling
// 	// wire.Ping messages
// 	PingVersion = ProtocolVersion + "/ping/1"

// 	// GetAddrVersion is the message version
// 	// for handling wire.GetAddr messages
// 	GetAddrVersion = ProtocolVersion + "/getaddr/1"

// 	// AddrVersion is the message version for handling
// 	// wire.Addr messages
// 	AddrVersion = ProtocolVersion + "/addr/1"

// 	// IntroVersion is the message version for handling
// 	// wire.Intro messages
// 	IntroVersion = ProtocolVersion + "/intro/1"

// 	// TxVersion is the message version for handing
// 	// wire.Transaction messages
// 	TxVersion = ProtocolVersion + "/tx/1"

// 	// BlockBodyVersion is the message version for
// 	// handling wire.BlockBody messages
// 	BlockBodyVersion = ProtocolVersion + "/blockbody/1"

// 	// GetBlockHashesVersion is the message version for
// 	// handling wire.BlockHashes messages
// 	GetBlockHashesVersion = ProtocolVersion + "/getblockhashes/1"

// 	// RequestBlockVersion is the message version for
// 	// handling wire.RequestBlock messages
// 	RequestBlockVersion = ProtocolVersion + "/requestblock/1"

// 	// GetBlockBodiesVersion is the message version for
// 	// handling wire.GetBlockBodies messages
// 	GetBlockBodiesVersion = ProtocolVersion + "/getblockbodies/1"
// )
