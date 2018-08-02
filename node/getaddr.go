package node

import (
	"context"
	"fmt"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/wire"
	net "github.com/libp2p/go-libp2p-net"
)

// StreamOptions defines options for handling a stream
type StreamOptions struct {

	// Ctx represents the streams context
	Ctx context.Context

	// RemotePeer is the peer we intend to send a
	// message to
	RemotePeer types.Engine

	// MsgVersion represents the message version of the
	// message to be sent.
	MsgVersion string

	// Writable is the message to be set. If set it
	// will be sent to the stream
	Writable interface{}

	// OnWriteFailed is called when write attempt fails.
	// The error is passed to the provided callback
	OnWriteFailed func(error)

	// ReadTo defines the container to hold incoming messages
	ReadTo interface{}

	// OnReadFailed is called when read attempt fails.
	// The error is passed to the provided callback
	OnReadFailed func(error)
}

// Stream creates a new stream or processes an existing stream. If existingStream
// is not set, a new stream is created using parameters provided in streamOpts.
func (g *Gossip) Stream(existingStream net.Stream, streamOpts StreamOptions) error {
	var stream = existingStream
	var err error

	// if stream is nil, we assume the caller intends for us
	// to create a new stream.
	if stream == nil {
		stream, err = g.newStream(streamOpts.Ctx, streamOpts.RemotePeer, streamOpts.MsgVersion)
		if err != nil {
			return fmt.Errorf("stream creation error: %s", err)
		}
	}

	// Attempt to write a message to the stream if provided.
	// If write fails, call the OnWriteFailed callback if set.
	if streamOpts.Writable != nil {
		if err := writeStream(stream, streamOpts.Writable); err != nil {
			if streamOpts.OnWriteFailed != nil {
				streamOpts.OnWriteFailed(err)
			}
			stream.Reset()
			return fmt.Errorf("getaddr failed. failed to write to stream")
		}
	}

	// If ReadTo is set, copy the stream data to the provided
	// struct/map in ReadTo.
	if streamOpts.ReadTo != nil {
		if err = readStream(stream, streamOpts.ReadTo); err != nil {
			if streamOpts.OnReadFailed != nil {
				streamOpts.OnReadFailed(err)
			}
			stream.Reset()
			return fmt.Errorf("stream read failed: %s", err)
		}
	}

	return nil
}

// sendGetAddr sends a wire.GetAddr message to a remote peer.
// The remote peer will respond with a wire.Addr message which the function
// must process using the OnAddr handler and return the response.
func (g *Gossip) sendGetAddr(remotePeer types.Engine) ([]*wire.Address, error) {

	remotePeerIDShort := remotePeer.ShortID()

	s, err := g.newStream(context.Background(), remotePeer, config.GetAddrVersion)
	if err != nil {
		g.log.Debug("GetAddr message failed. failed to connect to peer", "Err", err, "PeerID", remotePeerIDShort)
		return nil, fmt.Errorf("getaddr failed. failed to connect to peer. %s", err)
	}
	defer s.Close()

	msg := &wire.GetAddr{}
	if err := writeStream(s, msg); err != nil {
		s.Reset()
		g.log.Debug("GetAddr failed. failed to write to stream", "Err", err, "PeerID", remotePeerIDShort)
		return nil, fmt.Errorf("getaddr failed. failed to write to stream")
	}

	g.log.Debug("GetAddr message sent to peer", "PeerID", remotePeerIDShort)

	return g.onAddr(s)
}

// SendGetAddr sends GetAddr message to peers in separate goroutines.
// GetAddr returns with a list of addr that should be relayed to other peers.
func (g *Gossip) SendGetAddr(remotePeers []types.Engine) error {

	// we need to know if wee need more peers before we requests
	// more addresses from other peers.
	if !g.PM().NeedMorePeers() {
		return nil
	}

	// for each remore peers, send the GetAddr message in different goroutines
	for _, remotePeer := range remotePeers {
		rp := remotePeer
		go func() {

			// send GetAddr and receive a list of address
			addressToRelay, err := g.sendGetAddr(rp)
			if err != nil {
				return
			}

			// As per discovery protocol, relay the addresses received
			if len(addressToRelay) > 0 {
				g.RelayAddr(addressToRelay)
			}
		}()
	}

	return nil
}

// OnGetAddr processes a wire.GetAddr request.
// Sends a list of active addresses to the sender
func (g *Gossip) OnGetAddr(s net.Stream) {

	defer s.Close()

	remotePeerIDShort := util.ShortID(s.Conn().RemotePeer())
	remoteAddr := util.FullRemoteAddressFromStream(s)
	remotePeer := NewRemoteNode(remoteAddr, g.engine)

	// check whether we can interact with this remote peer
	if yes, reason := g.engine.canAcceptPeer(remotePeer); !yes {
		s.Reset()
		g.log.Debug(fmt.Sprintf("Can't accept message from peer: %s", reason), "Addr", remotePeer.GetMultiAddr(), "Msg", "GetAddr")
		return
	}

	g.log.Debug("Received GetAddr message", "PeerID", remotePeerIDShort)

	// read the message
	msg := &wire.GetAddr{}
	if err := readStream(s, msg); err != nil {
		s.Reset()
		g.log.Error("failed to read getaddr message", "Err", err, "PeerID", remotePeerIDShort)
		return
	}

	// get active addresses we know about. If we have more 2500
	// addresses, then we select 2500 randomly
	activePeers := g.PM().GetActivePeers(0)
	if len(activePeers) > 2500 {
		activePeers = g.PM().GetRandomActivePeers(2500)
	}

	// Construct an Addr message and add addresses to it
	addr := &wire.Addr{}
	for _, peer := range activePeers {
		// Ignore an address if it is the same with the local node
		// and if it is a hardcoded seed address
		if !g.PM().IsLocalNode(peer) && !peer.IsSame(remotePeer) && !peer.IsHardcodedSeed() {
			addr.Addresses = append(addr.Addresses, &wire.Address{
				Address:   peer.GetMultiAddr(),
				Timestamp: peer.GetTimestamp().Unix(),
			})
		}
	}

	if err := writeStream(s, addr); err != nil {
		s.Reset()
		g.log.Error("failed to send GetAddr response", "Err", err)
		return
	}

	g.log.Debug("Sent GetAddr response to peer", "PeerID", remotePeerIDShort)
}
