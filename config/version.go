package config

var (

	// ProtocolVersion is the protocol version
	// spoken by this client
	ProtocolVersion = "/inception/1"

	// HandshakeVersion is the message version for
	// handling wire.Handshake
	HandshakeVersion = ProtocolVersion + "/handshake/1"

	// PingVersion is the version for handling
	// wire.Ping messages
	PingVersion = ProtocolVersion + "/ping/1"

	// GetAddrVersion is the message version
	// for handling wire.GetAddr messages
	GetAddrVersion = ProtocolVersion + "/getaddr/1"

	// AddrVersion is the message version for handling
	// wire.Addr messages
	AddrVersion = ProtocolVersion + "/addr/1"

	// IntroVersion is the message version for handling
	// wire.Intro messages
	IntroVersion = ProtocolVersion + "/intro/1"

	// TxVersion is the message version for handing
	// wire.Transaction messages
	TxVersion = ProtocolVersion + "/tx/1"

	// BlockBodyVersion is the message version for
	// handling wire.BlockBody messages
	BlockBodyVersion = ProtocolVersion + "/blockbody/1"

	// GetBlockHashesVersion is the message version for
	// handling wire.BlockHashes messages
	GetBlockHashesVersion = ProtocolVersion + "/getblockhashes/1"

	// RequestBlockVersion is the message version for
	// handling wire.RequestBlock messages
	RequestBlockVersion = ProtocolVersion + "/requestblock/1"

	// GetBlockBodiesVersion is the message version for
	// handling wire.GetBlockBodies messages
	GetBlockBodiesVersion = ProtocolVersion + "/getblockbodies/1"
)
