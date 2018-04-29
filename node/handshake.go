package node

import (
	"bufio"
	"context"
	"fmt"
	"time"

	"github.com/ellcrys/druid/util"
	"github.com/ellcrys/druid/wire"
	net "github.com/libp2p/go-libp2p-net"
	pc "github.com/multiformats/go-multicodec/protobuf"
)

// SendHandshake sends an introduction message to a peer
func (pt *Inception) SendHandshake(remotePeer *Node) error {

	remotePeerIDShort := remotePeer.ShortID()
	pt.log.Info("Sending handshake to peer", "PeerID", remotePeerIDShort)

	s, err := pt.LocalPeer().addToPeerStore(remotePeer).newStream(context.Background(), remotePeer.ID(), util.HandshakeVersion)
	if err != nil {
		pt.log.Debug("Handshake failed. failed to connect to peer", "Err", err, "PeerID", remotePeerIDShort)
		return fmt.Errorf("handshake failed. failed to connect to peer. %s", err)
	}
	defer s.Close()

	w := bufio.NewWriter(s)
	msg := &wire.Handshake{SubVersion: util.ClientVersion}
	if err := pc.Multicodec(nil).Encoder(w).Encode(msg); err != nil {
		pt.log.Debug("Handshake failed. failed to write to stream", "Err", err, "PeerID", remotePeerIDShort)
		return fmt.Errorf("handshake failed. failed to write to stream")
	}
	w.Flush()

	pt.log.Debug("Sent handshake to peer", "PeerID", remotePeerIDShort)

	resp := &wire.Handshake{}
	decoder := pc.Multicodec(nil).Decoder(bufio.NewReader(s))
	if err := decoder.Decode(resp); err != nil {
		pt.log.Debug("Failed to read handshake response", "Err", err, "PeerID", remotePeerIDShort)
		return fmt.Errorf("failed to read handshake response")
	}

	remotePeer.Timestamp = time.Now()
	pt.PM().AddOrUpdatePeer(remotePeer)

	pt.log.Info("Handshake was successful", "PeerID", remotePeerIDShort, "SubVersion", resp.SubVersion)

	return nil
}

// OnHandshake handles incoming handshake request
func (pt *Inception) OnHandshake(s net.Stream) {

	remotePeer := NewRemoteNode(util.FullRemoteAddressFromStream(s), pt.LocalPeer())
	remotePeerIDShort := remotePeer.ShortID()
	defer s.Close()

	if pt.LocalPeer().isDevMode() && !util.IsDevAddr(remotePeer.IP) {
		pt.log.Debug("Can't accept message from non local or private IP in development mode", "Addr", remotePeer.GetMultiAddr(), "Msg", "Handshake")
		return
	}

	pt.log.Info("Received handshake message", "PeerID", remotePeerIDShort)

	msg := &wire.Handshake{}
	if err := pc.Multicodec(nil).Decoder(bufio.NewReader(s)).Decode(msg); err != nil {
		pt.log.Error("failed to read handshake message", "Err", err, "PeerID", remotePeerIDShort)
		return
	}

	ack := &wire.Handshake{SubVersion: util.ClientVersion}
	w := bufio.NewWriter(s)
	enc := pc.Multicodec(nil).Encoder(w)
	if err := enc.Encode(ack); err != nil {
		pt.log.Error("failed to send handshake response", "Err", err)
		return
	}

	remotePeer.Timestamp = time.Now()
	pt.PM().AddOrUpdatePeer(remotePeer)
	pt.log.Info("Handshake has been acknowledged", "PeerID", remotePeerIDShort, "SubVersion", msg.SubVersion)

	w.Flush()
}
