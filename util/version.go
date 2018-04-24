package util

var (

	// ClientVersion is the version of this current code base
	ClientVersion = "0.0.1"

	// ProtocolVersion is the protocol version spoken by this client
	ProtocolVersion = "/inception/1"

	// HandshakeVersion is the current handshake algorithm
	HandshakeVersion = ProtocolVersion + "/handshake/1"

	// PingVersion is the current ping algorithm
	PingVersion = ProtocolVersion + "/ping/1"

	// GetAddrVersion is the current getaddr algorithm
	GetAddrVersion = ProtocolVersion + "/getaddr/1"

	// AddrVersion is the current addr algorithm
	AddrVersion = ProtocolVersion + "/addr/1"
)
