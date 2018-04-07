package peer

import (
	"bufio"
	"context"

	"github.com/ellcrys/gcoin/util"
	pb "github.com/ellcrys/gcoin/wire"
	"github.com/kr/pretty"
	net "github.com/libp2p/go-libp2p-net"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	protocol "github.com/libp2p/go-libp2p-protocol"
	pc "github.com/multiformats/go-multicodec/protobuf"
)

// DoSendHandshake sends an introduction message to a peer
func (protoc *Inception) DoSendHandshake(remotePeer *Peer) {

	remotePeer.localPeer.Peerstore().AddAddr(remotePeer.ID(), remotePeer.GetIP4TCPAddr(), pstore.PermanentAddrTTL)
	s, err := remotePeer.localPeer.host.NewStream(context.Background(), remotePeer.ID(), protocol.ID(HandshakeVersion))
	if err != nil {
		protocLog.Debugw("handshake failed. failed to connect to peer", "Err", err, "PeerID", remotePeer.IDPretty())
		return
	}
	defer s.Close()

	// send handshake
	w := bufio.NewWriter(s)
	msg := &pb.Handshake{Address: remotePeer.localPeer.GetMultiAddr()}
	if err := pc.Multicodec(nil).Encoder(w).Encode(msg); err != nil {
		protocLog.Debugw("handshake failed. failed to write to stream", "Err", err, "PeerID", remotePeer.IDPretty())
		return
	}
	w.Flush()

	// receive response
	resp := &pb.HandshakeResponse{}
	decoder := pc.Multicodec(nil).Decoder(bufio.NewReader(s))
	if err := decoder.Decode(resp); err != nil {
		protocLog.Debugw("failed to read handshake response", "Err", err, "PeerID", remotePeer.IDPretty())
		return
	}

	pretty.Println("Resp:", resp)
}

// OnHandshake handles incoming handshake request
func (protoc *Inception) OnHandshake(s net.Stream) {
	defer s.Close()

	msg := &pb.Handshake{}
	if err := pc.Multicodec(nil).Decoder(bufio.NewReader(s)).Decode(msg); err != nil {
		protoc.log.Errorw("failed to read message", "Err", err, "PeerID", s.Conn().RemotePeer().Pretty())
		return
	}

	var addresses []string
	peers := protoc.PM().ActivePeers()
	for _, p := range peers {
		addresses = append(addresses, p.GetMultiAddr())
	}

	protoc.PM().AddPeer(&Peer{
		address:   util.FullRemoteAddressFromStream(s),
		localPeer: protoc.LocalPeer(),
	})

	// create response message, sign it and add the signature to the message
	addrMsg := &pb.HandshakeResponse{Addresses: addresses}
	addrMsg.Sig = protoc.sign(addrMsg)

	w := bufio.NewWriter(s)
	enc := pc.Multicodec(nil).Encoder(w)
	if err := enc.Encode(addrMsg); err != nil {
		protoc.log.Errorw("failed to send handshake response", "Err", err)
		return
	}

	w.Flush()
}
