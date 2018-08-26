package config

var (

	// ClientVersion is the version of this code base
	ClientVersion = "0.0.1"

	// ProtocolVersion is the protocol version spoken by this client
	ProtocolVersion = "/inception/1"

	// HandshakeVersion is the message version for handling
	// incoming handshake request
	HandshakeVersion = ProtocolVersion + "/handshake/1"

	// PingVersion is the version for handling
	// incoming ping messages
	PingVersion = ProtocolVersion + "/ping/1"

	// GetAddrVersion is the message version
	// for handling incoming request to send addresses
	GetAddrVersion = ProtocolVersion + "/getaddr/1"

	// AddrVersion is the message version for handling
	// incoming addresses
	AddrVersion = ProtocolVersion + "/addr/1"

	// TxVersion is the message version for handing
	// incoming transactions
	TxVersion = ProtocolVersion + "/tx/1"

	// BlockBodyVersion is the message version for
	// handling incoming block
	BlockBodyVersion = ProtocolVersion + "/blockbody/1"

	// GetBlockHeaders is the message version for
	// handling request for block headers
	GetBlockHeaders = ProtocolVersion + "/getblockheaders/1"

	// RequestBlockVersion is the message version for handling
	// block requests
	RequestBlockVersion = ProtocolVersion + "/requestblock/1"
)
