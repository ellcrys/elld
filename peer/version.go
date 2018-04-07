package peer

var (
	// ProtocolVersion is the protocol version spoken by this client
	ProtocolVersion = "/inception/1"

	// HandshakeVersion is the current handshake algorithm
	HandshakeVersion = ProtocolVersion + "/handshake/1"
)
