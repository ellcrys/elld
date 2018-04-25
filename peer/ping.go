package peer

import (
	"bufio"
	"context"
	"fmt"
	"time"

	"github.com/ellcrys/druid/wire"

	"github.com/ellcrys/druid/util"
	net "github.com/libp2p/go-libp2p-net"
	pc "github.com/multiformats/go-multicodec/protobuf"
)

func (pt *Inception) sendPing(remotePeer *Peer) error {

	remotePeerIDShort := remotePeer.ShortID()
	s, err := pt.LocalPeer().addToPeerStore(remotePeer).newStream(context.Background(), remotePeer.ID(), util.PingVersion)
	if err != nil {
		pt.log.Debug("Ping failed. failed to connect to peer", "Err", err, "PeerID", remotePeerIDShort)
		return fmt.Errorf("ping failed. failed to connect to peer. %s", err)
	}
	defer s.Close()

	w := bufio.NewWriter(s)
	msg := &wire.Ping{}
	msg.Sig = pt.sign(msg)
	if err := pc.Multicodec(nil).Encoder(w).Encode(msg); err != nil {
		pt.log.Debug("ping failed. failed to write to stream", "Err", err, "PeerID", remotePeerIDShort)
		return fmt.Errorf("ping failed. failed to write to stream")
	}
	w.Flush()

	pt.log.Info("Sent ping to peer", "PeerID", remotePeerIDShort)

	// receive pong response
	pongMsg := &wire.Pong{}
	decoder := pc.Multicodec(nil).Decoder(bufio.NewReader(s))
	if err := decoder.Decode(pongMsg); err != nil {
		pt.log.Debug("Failed to read pong response", "Err", err, "PeerID", remotePeerIDShort)
		return fmt.Errorf("failed to read pong response")
	}

	sig := pongMsg.Sig
	pongMsg.Sig = nil
	if err := pt.verify(pongMsg, sig, s.Conn().RemotePublicKey()); err != nil {
		pt.log.Debug("failed to verify message signature", "Err", err, "PeerID", remotePeerIDShort)
		return fmt.Errorf("failed to verify message signature")
	}

	remotePeer.Timestamp = time.Now()
	pt.PM().AddOrUpdatePeer(remotePeer)

	pt.log.Info("Received pong response from peer", "PeerID", remotePeerIDShort)

	return nil
}

// SendPing sends a ping message
func (pt *Inception) SendPing(remotePeers []*Peer) {
	pt.log.Info("Sending ping to peer(s)", "NumPeers", len(remotePeers))
	for _, remotePeer := range remotePeers {
		_remotePeer := remotePeer
		go func() {
			if err := pt.sendPing(_remotePeer); err != nil {
				pt.PM().onFailedConnection(_remotePeer)
			}
		}()
	}
}

// OnPing handles incoming ping message
func (pt *Inception) OnPing(s net.Stream) {

	remotePeer := NewRemotePeer(util.FullRemoteAddressFromStream(s), pt.LocalPeer())
	remotePeerIDShort := remotePeer.ShortID()
	defer s.Close()

	pt.log.Info("Received ping message", "PeerID", remotePeerIDShort)

	msg := &wire.Ping{}
	if err := pc.Multicodec(nil).Decoder(bufio.NewReader(s)).Decode(msg); err != nil {
		pt.log.Error("failed to read ping message", "Err", err, "PeerID", remotePeerIDShort)
		return
	}

	// verify signature
	sig := msg.Sig
	msg.Sig = nil
	if err := pt.verify(msg, sig, s.Conn().RemotePublicKey()); err != nil {
		pt.log.Debug("failed to verify ping message signature", "Err", err, "PeerID", remotePeerIDShort)
		return
	}

	// send pong message
	pongMsg := &wire.Pong{}
	pongMsg.Sig = pt.sign(pongMsg)
	w := bufio.NewWriter(s)
	enc := pc.Multicodec(nil).Encoder(w)
	if err := enc.Encode(pongMsg); err != nil {
		pt.log.Error("failed to send pong response", "Err", err)
		return
	}

	remotePeer.Timestamp = time.Now()
	pt.PM().AddOrUpdatePeer(remotePeer)
	pt.log.Info("Sent pong response to peer", "PeerID", remotePeerIDShort)

	w.Flush()
}
