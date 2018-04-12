package peer

import (
	"github.com/ellcrys/druid/util"
	inet "github.com/libp2p/go-libp2p-net"
	ma "github.com/multiformats/go-multiaddr"
)

// Notification implements net.Notifiee
type Notification struct {
	pm *Manager
}

// Listen is called when network starts listening on an address
func (n *Notification) Listen(inet.Network, ma.Multiaddr) {

}

// ListenClose is called when network stops listening on an address
func (n *Notification) ListenClose(inet.Network, ma.Multiaddr) {

}

// Connected is called when a connection is opened
func (n *Notification) Connected(net inet.Network, conn inet.Conn) {
	fullAddr := util.FullRemoteAddressFromConn(conn)
	go n.pm.onPeerConnect(fullAddr)
}

// Disconnected is called when a connection is closed
func (n *Notification) Disconnected(net inet.Network, conn inet.Conn) {
	fullAddr := util.FullRemoteAddressFromConn(conn)
	go n.pm.onPeerDisconnect(fullAddr)
}

// OpenedStream is called when a stream is openned
func (n *Notification) OpenedStream(inet.Network, inet.Stream) {

}

// ClosedStream is called when a stream is openned
func (n *Notification) ClosedStream(nt inet.Network, s inet.Stream) {
}
