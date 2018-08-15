package node

import (
	"context"
	"fmt"
	"time"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/wire"
	net "github.com/libp2p/go-libp2p-net"
)

// SendHandshake sends an introduction message to a peer
func (g *Gossip) SendHandshake(remotePeer types.Engine) error {

	remotePeerIDShort := remotePeer.ShortID()

	g.log.Info("Sending handshake to peer", "PeerID", remotePeerIDShort)

	// create stream
	s, err := g.newStream(context.Background(), remotePeer, config.HandshakeVersion)
	if err != nil {
		g.log.Debug("Handshake failed. failed to connect to peer", "Err", err, "PeerID", remotePeerIDShort)
		return fmt.Errorf("handshake failed. failed to connect to peer. %s", err)
	}
	defer s.Close()

	// write to the stream
	msg := &wire.Handshake{SubVersion: config.ClientVersion}
	if err := writeStream(s, msg); err != nil {
		s.Reset()
		g.log.Debug("Handshake failed. failed to write to stream", "Err", err, "PeerID", remotePeerIDShort)
		return fmt.Errorf("handshake failed. failed to write to stream")
	}

	g.log.Debug("Sent handshake to peer", "PeerID", remotePeerIDShort)

	// receive handshake message from the remote peer.
	resp := &wire.Handshake{}
	if err := readStream(s, resp); err != nil {
		s.Reset()
		g.log.Debug("Failed to read handshake response", "Err", err, "PeerID", remotePeerIDShort)
		return fmt.Errorf("failed to read handshake response")
	}

	// update the timestamp of the peer
	remotePeer.SetTimestamp(time.Now())
	g.PM().AddOrUpdatePeer(remotePeer)

	g.log.Info("Handshake was successful", "PeerID", remotePeerIDShort, "SubVersion", resp.SubVersion)

	return nil
}

// OnHandshake handles incoming handshake request
func (g *Gossip) OnHandshake(s net.Stream) {

	defer s.Close()

	remotePeer := NewRemoteNode(util.FullRemoteAddressFromStream(s), g.Engine())
	remotePeerIDShort := remotePeer.ShortID()

	// In non-production mode, ensure messages from public addresses are ignored
	if !g.Engine().ProdMode() && !util.IsDevAddr(remotePeer.IP) {
		s.Reset()
		g.log.Debug("In development mode, we cannot interact with peers with public IP", "Addr", remotePeer.GetMultiAddr(), "Msg", "Handshake")
		return
	}

	g.log.Info("Received handshake message", "PeerID", remotePeerIDShort)

	// read the message from the stream
	msg := &wire.Handshake{}
	if err := readStream(s, msg); err != nil {
		s.Reset()
		g.log.Error("failed to read handshake message", "Err", err, "PeerID", remotePeerIDShort)
		return
	}

	// TODO: perform version compatibility checks with msg

	// send back this peer's client version so the other peer can decide
	// whether to keep the connection
	msg = &wire.Handshake{SubVersion: config.ClientVersion}
	if err := writeStream(s, msg); err != nil {
		s.Reset()
		g.log.Error("failed to send handshake response", "Err", err)
		return
	}

	// update the remote peer's timestamp and add it to the peer manager's list
	remotePeer.Timestamp = time.Now()
	g.PM().AddOrUpdatePeer(remotePeer)

	g.log.Info("Handshake has been acknowledged", "PeerID", remotePeerIDShort, "SubVersion", msg.SubVersion)

}
