package protocol

import (
	"github.com/ellcrys/garagecoin/modules/types"
	"github.com/kr/pretty"
	net "github.com/libp2p/go-libp2p-net"
	protocol "github.com/libp2p/go-libp2p-protocol"
)

// HandleHandshake processes handshake message from a remote peer
func (protoc *Inception) HandleHandshake(m *types.Message, protocol protocol.ID, conn net.Conn) {
	var opMsg types.HandshakeMsg
	m.Scan(&opMsg)
	pretty.Println(opMsg)
}
