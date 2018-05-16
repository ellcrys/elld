package util

var (

	// ClientVersion is the version of this current code base
	ClientVersion = "0.0.1"

	// ProtocolVersion is the protocol version spoken by this client
	ProtocolVersion = "/inception/1"

	// HandshakeVersion is the current handshake message handler
	HandshakeVersion = ProtocolVersion + "/handshake/1"

	// PingVersion is the current ping message handler
	PingVersion = ProtocolVersion + "/ping/1"

	// GetAddrVersion is the current getaddr message handler
	GetAddrVersion = ProtocolVersion + "/getaddr/1"

	// AddrVersion is the current addr message handler
	AddrVersion = ProtocolVersion + "/addr/1"

	// TxVersion is the current tx message handler
	TxVersion = ProtocolVersion + "/tx/1"
)
