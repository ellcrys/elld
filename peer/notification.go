package peer

import (
	net "github.com/libp2p/go-libp2p-net"
	ma "github.com/multiformats/go-multiaddr"
)

// Notification implements net.Notifiee
type Notification struct {
}

// Listen is called when network starts listening on an address
func (n *Notification) Listen(net.Network, ma.Multiaddr) {

}

// ListenClose is called when network stops listening on an address
func (n *Notification) ListenClose(net.Network, ma.Multiaddr) {

}

// Connected is called when a connection is opened
func (n *Notification) Connected(net.Network, net.Conn) {

}

// Disconnected is called when a connection is closed
func (n *Notification) Disconnected(nt net.Network, c net.Conn) {
}

// OpenedStream is called when a stream is openned
func (n *Notification) OpenedStream(net.Network, net.Stream) {

}

// ClosedStream is called when a stream is openned
func (n *Notification) ClosedStream(nt net.Network, s net.Stream) {
}
