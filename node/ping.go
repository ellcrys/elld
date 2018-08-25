package node

import (
	"context"
	"fmt"
	"time"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/wire"

	"github.com/ellcrys/elld/util"
	net "github.com/libp2p/go-libp2p-net"
)

func (g *Gossip) sendPing(remotePeer types.Engine) error {

	remotePeerIDShort := remotePeer.ShortID()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// create stream to the remote peer
	s, err := g.newStream(ctx, remotePeer, config.PingVersion)
	if err != nil {
		g.log.Debug("Ping failed. failed to connect to peer", "Err", err, "PeerID", remotePeerIDShort)
		return fmt.Errorf("ping failed. failed to connect to peer. %s", err)
	}
	defer s.Close()

	// construct the message and write it to the stream
	msg := &wire.Ping{}
	if err := writeStream(s, msg); err != nil {
		s.Reset()
		g.log.Debug("ping failed. failed to write to stream", "Err", err, "PeerID", remotePeerIDShort)
		return fmt.Errorf("ping failed. failed to write to stream")
	}

	g.log.Info("Sent ping to peer", "PeerID", remotePeerIDShort)

	// receive pong response from the remote peer
	pongMsg := &wire.Pong{}
	if err := readStream(s, pongMsg); err != nil {
		s.Reset()
		g.log.Debug("Failed to read pong response", "Err", err, "PeerID", remotePeerIDShort)
		return fmt.Errorf("failed to read pong response")
	}

	// update the remote peer's timestamp
	remotePeer.SetTimestamp(time.Now())

	// add the remote peer to the peer manager's list
	g.PM().AddOrUpdatePeer(remotePeer)

	g.log.Info("Received pong response from peer", "PeerID", remotePeerIDShort)

	return nil
}

// SendPing sends a ping message
func (g *Gossip) SendPing(remotePeers []types.Engine) {
	g.log.Info("Sending ping to peer(s)", "NumPeers", len(remotePeers))
	for _, remotePeer := range remotePeers {
		_remotePeer := remotePeer
		go func() {
			if err := g.sendPing(_remotePeer); err != nil {
				g.PM().onFailedConnection(_remotePeer)
			}
		}()
	}
}

// OnPing handles incoming ping message
func (g *Gossip) OnPing(s net.Stream) {

	defer s.Close()

	remotePeer := NewRemoteNode(util.FullRemoteAddressFromStream(s), g.engine)
	remotePeerIDShort := remotePeer.ShortID()

	g.log.Info("Received ping message", "PeerID", remotePeerIDShort)

	// read the message from the stream
	msg := &wire.Ping{}
	if err := readStream(s, msg); err != nil {
		s.Reset()
		g.log.Error("failed to read ping message", "Err", err, "PeerID", remotePeerIDShort)
		return
	}

	// send pong message
	pongMsg := &wire.Pong{}
	if err := writeStream(s, pongMsg); err != nil {
		s.Reset()
		g.log.Error("failed to send pong response", "Err", err)
		return
	}

	// update the remote peer's timestamp in the peer manager's list
	remotePeer.Timestamp = time.Now()
	g.PM().AddOrUpdatePeer(remotePeer)

	g.log.Info("Sent pong response to peer", "PeerID", remotePeerIDShort)
}
