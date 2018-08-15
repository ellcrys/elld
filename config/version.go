package config

var (

	// ClientVersion is the version of this current code base
	ClientVersion = "0.0.1"

	// ProtocolVersion is the protocol version spoken by this client
	ProtocolVersion = "/inception/1"

	// HandshakeVersion is the current handshake message version
	HandshakeVersion = ProtocolVersion + "/handshake/1"

	// PingVersion is the current ping message version
	PingVersion = ProtocolVersion + "/ping/1"

	// GetAddrVersion is the current getaddr message version
	GetAddrVersion = ProtocolVersion + "/getaddr/1"

	// AddrVersion is the current addr message version
	AddrVersion = ProtocolVersion + "/addr/1"

	// TxVersion is the current tx message version
	TxVersion = ProtocolVersion + "/tx/1"

	// BlockVersion is the current block message version
	BlockVersion = ProtocolVersion + "/block/1"
)
