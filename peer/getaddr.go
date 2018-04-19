package peer

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

// sendGetAddr sends a wire.GetAddr message to a remote peer
func (pt *Inception) sendGetAddr(remotePeer *Peer) error {

	remotePeerIDShort := remotePeer.ShortID()
	s, err := pt.LocalPeer().addToPeerStore(remotePeer).newStream(context.Background(), remotePeer.ID(), util.GetAddrVersion)
	if err != nil {
		pt.log.Debugw("GetAddr message failed. failed to connect to peer", "Err", err, "PeerID", remotePeerIDShort)
		return fmt.Errorf("getaddr failed. failed to connect to peer. %s", err)
	}
	defer s.Close()

	w := bufio.NewWriter(s)
	msg := &wire.GetAddr{}
	msg.Sig = pt.sign(msg)
	if err := pc.Multicodec(nil).Encoder(w).Encode(msg); err != nil {
		pt.log.Debugw("GetAddr failed. failed to write to stream", "Err", err, "PeerID", remotePeerIDShort)
		return fmt.Errorf("getaddr failed. failed to write to stream")
	}
	w.Flush()

	pt.log.Infow("GetAddr message sent to peer", "PeerID", remotePeerIDShort)

	// receive response
	resp := &wire.Addr{}
	decoder := pc.Multicodec(nil).Decoder(bufio.NewReader(s))
	if err := decoder.Decode(resp); err != nil {
		pt.log.Debugw("Failed to read GetAddr response response", "Err", err, "PeerID", remotePeerIDShort)
		return fmt.Errorf("failed to read GetAddr response")
	}

	// verify message signature
	sig := resp.Sig
	resp.Sig = nil
	if err := pt.verify(resp, sig, s.Conn().RemotePublicKey()); err != nil {
		pt.log.Debugw("Failed to verify message signature", "Err", err, "PeerID", remotePeerIDShort)
		return fmt.Errorf("failed to verify message signature")
	}

	if len(resp.Addresses) > pt.LocalPeer().cfg.Peer.MaxAddrsExpected {
		pt.log.Debugw("Too many addresses received. Ignoring addresses", "Err", err, "PeerID", remotePeerIDShort, "NumAddrReceived", len(resp.Addresses))
		return fmt.Errorf("too many addresses received. Ignoring addresses")
	}

	invalidAddrs := 0
	for _, addr := range resp.Addresses {

		p, _ := pt.LocalPeer().PeerFromAddr(addr.Address, true)
		p.Timestamp = time.Unix(addr.Timestamp, 0)

		if p.IsBadTimestamp() {
			p.Timestamp = time.Now().Add(-1 * time.Hour * 24 * 5)
		}

		if pt.PM().AddOrUpdatePeer(p) != nil {
			invalidAddrs++
		}
	}

	pt.log.Infow("Received GetAddr response from peer", "PeerID", remotePeerIDShort, "NumAddrs", len(resp.Addresses), "InvalidAddrs", invalidAddrs)

	return nil
}

// SendGetAddr sends GetAddr message to peers.
func (pt *Inception) SendGetAddr(remotePeers []*Peer) error {

	if !pt.PM().NeedMorePeers() {
		return nil
	}

	for _, remotePeer := range remotePeers {
		go pt.sendGetAddr(remotePeer)
	}

	return nil
}

// OnGetAddr processes a wire.GetAddr request.
// Sends a list of active addresses to the sender
func (pt *Inception) OnGetAddr(s net.Stream) {

	remotePeerIDShort := util.ShortID(s.Conn().RemotePeer())
	remoteAddr := util.FullRemoteAddressFromStream(s)
	remotePeer := NewRemotePeer(remoteAddr, pt.LocalPeer())
	defer s.Close()

	if !remotePeer.IsKnown() && !pt.LocalPeer().isDevMode() {
		s.Conn().Close()
		return
	}

	pt.log.Infow("Received GetAddr message", "PeerID", remotePeerIDShort)

	msg := &wire.GetAddr{}
	if err := pc.Multicodec(nil).Decoder(bufio.NewReader(s)).Decode(msg); err != nil {
		pt.log.Errorw("failed to read getaddr message", "Err", err, "PeerID", remotePeerIDShort)
		return
	}

	// verify signature
	sig := msg.Sig
	msg.Sig = nil
	if err := pt.verify(msg, sig, s.Conn().RemotePublicKey()); err != nil {
		pt.log.Debugw("failed to verify getaddr message signature", "Err", err, "PeerID", remotePeerIDShort)
		return
	}

	activePeers := pt.PM().GetActivePeers(0)
	if len(activePeers) > 2500 {
		activePeers = pt.PM().GetRandomActivePeers(2500)
	}

	// send getaddr response message
	getAddrResp := &wire.Addr{}
	for _, peer := range activePeers {
		if !pt.PM().IsLocalPeer(peer) && !peer.IsSame(remotePeer) && !peer.isHardcodedSeed {
			getAddrResp.Addresses = append(getAddrResp.Addresses, &wire.Address{
				Address:   peer.GetMultiAddr(),
				Timestamp: peer.Timestamp.Unix(),
			})
		}
	}

	getAddrResp.Sig = pt.sign(getAddrResp)
	w := bufio.NewWriter(s)
	enc := pc.Multicodec(nil).Encoder(w)
	if err := enc.Encode(getAddrResp); err != nil {
		pt.log.Errorw("failed to send GetAddr response", "Err", err)
		return
	}

	pt.log.Infow("Sent GetAddr response to peer", "PeerID", remotePeerIDShort)

	w.Flush()
}
