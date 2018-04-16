package peer

import (
	"bufio"
	"context"
	"fmt"

	"github.com/ellcrys/druid/util"
	"github.com/ellcrys/druid/wire"
	net "github.com/libp2p/go-libp2p-net"
	pc "github.com/multiformats/go-multicodec/protobuf"
)

// sendGetAddr sends a wire.GetAddr message to a remote peer
func (protoc *Inception) sendGetAddr(remotePeer *Peer) error {

	remotePeerID := remotePeer.ShortID()
	s, err := protoc.LocalPeer().addToPeerStore(remotePeer).newStream(context.Background(), remotePeer.ID(), util.GetAddrVersion)
	if err != nil {
		protocLog.Debugw("GetAddr message failed. failed to connect to peer", "Err", err, "PeerID", remotePeerID)
		return fmt.Errorf("getaddr failed. failed to connect to peer. %s", err)
	}
	defer s.Close()

	w := bufio.NewWriter(s)
	msg := &wire.GetAddr{}
	msg.Sig = protoc.sign(msg)
	if err := pc.Multicodec(nil).Encoder(w).Encode(msg); err != nil {
		protocLog.Debugw("GetAddr failed. failed to write to stream", "Err", err, "PeerID", remotePeerID)
		return fmt.Errorf("getaddr failed. failed to write to stream")
	}
	w.Flush()

	protoc.log.Infow("GetAddr message sent to peer", "PeerID", remotePeerID)

	// receive response
	resp := &wire.Addr{}
	decoder := pc.Multicodec(nil).Decoder(bufio.NewReader(s))
	if err := decoder.Decode(resp); err != nil {
		protocLog.Debugw("Failed to read GetAddr response response", "Err", err, "PeerID", remotePeerID)
		return fmt.Errorf("failed to read GetAddr response")
	}

	// verify message signature
	sig := resp.Sig
	resp.Sig = nil
	if err := protoc.verify(resp, sig, s.Conn().RemotePublicKey()); err != nil {
		protoc.log.Debugw("failed to verify message signature", "Err", err, "PeerID", remotePeerID)
		return fmt.Errorf("failed to verify message signature")
	}

	protoc.PM().AddOrUpdatePeer(remotePeer)

	// validate address before adding
	invalidAddrs := 0
	for _, addr := range resp.Addresses {
		if !util.IsValidAddress(addr.Address) {
			invalidAddrs++
			protoc.log.Debugw("Found invalid address in GetAddr response", "Addr", addr, "RemotePeerID", remotePeerID)
			continue
		}
		p, _ := protoc.LocalPeer().PeerFromAddr(addr.Address, true)
		protoc.PM().AddOrUpdatePeer(p)
	}

	protoc.log.Infow("Received GetAddr response from peer", "PeerID", remotePeerID, "NumAddrs", len(resp.Addresses), "InvalidAddrs", invalidAddrs)

	return nil
}

// SendGetAddr sends GetAddr message to peers.
func (protoc *Inception) SendGetAddr(remotePeers []*Peer) error {

	if !protoc.PM().NeedMorePeers() {
		return nil
	}

	for _, remotePeer := range remotePeers {
		go protoc.sendGetAddr(remotePeer)
	}

	return nil
}

// OnGetAddr processes a wire.GetAddr request.
// Sends a list of active addresses to the sender
func (protoc *Inception) OnGetAddr(s net.Stream) {

	remotePeerIDShort := util.ShortID(s.Conn().RemotePeer())
	remoteAddr := util.FullRemoteAddressFromStream(s)
	remotePeer := NewRemotePeer(remoteAddr, protoc.LocalPeer())
	defer s.Close()

	protoc.log.Infow("Received GetAddr message", "PeerID", remotePeerIDShort)

	msg := &wire.GetAddr{}
	if err := pc.Multicodec(nil).Decoder(bufio.NewReader(s)).Decode(msg); err != nil {
		protoc.log.Errorw("failed to read getaddr message", "Err", err, "PeerID", remotePeerIDShort)
		return
	}

	// verify signature
	sig := msg.Sig
	msg.Sig = nil
	if err := protoc.verify(msg, sig, s.Conn().RemotePublicKey()); err != nil {
		protoc.log.Debugw("failed to verify getaddr message signature", "Err", err, "PeerID", remotePeerIDShort)
		return
	}

	protoc.PM().AddOrUpdatePeer(remotePeer)

	activePeers := protoc.PM().GetActivePeers(0)
	if len(activePeers) > 2500 {
		activePeers = protoc.PM().GetRandomActivePeers(2500)
	}

	// send getaddr response message
	getAddrResp := &wire.Addr{}
	for _, peer := range activePeers {
		if !protoc.PM().IsLocalPeer(peer) && !peer.IsSame(remotePeer) {
			getAddrResp.Addresses = append(getAddrResp.Addresses, &wire.Address{
				Address:   peer.GetMultiAddr(),
				Timestamp: peer.Timestamp.Unix(),
			})
		}
	}

	getAddrResp.Sig = protoc.sign(getAddrResp)
	w := bufio.NewWriter(s)
	enc := pc.Multicodec(nil).Encoder(w)
	if err := enc.Encode(getAddrResp); err != nil {
		protoc.log.Errorw("failed to send GetAddr response", "Err", err)
		return
	}

	protoc.log.Infow("Sent GetAddr response to peer", "PeerID", remotePeerIDShort)

	w.Flush()
}
