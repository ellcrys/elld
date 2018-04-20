package peer

import (
	"bufio"
	"fmt"
	"time"

	"github.com/ellcrys/druid/util"
	"github.com/ellcrys/druid/wire"
	net "github.com/libp2p/go-libp2p-net"
	pc "github.com/multiformats/go-multicodec/protobuf"
)

// onAddr handles "addr" message
func (pt *Inception) onAddr(s net.Stream) error {

	remoteAddr := util.FullRemoteAddressFromStream(s)
	remotePeer := NewRemotePeer(remoteAddr, pt.LocalPeer())
	remotePeerIDShort := remotePeer.ShortID()

	resp := &wire.Addr{}
	decoder := pc.Multicodec(nil).Decoder(bufio.NewReader(s))
	if err := decoder.Decode(resp); err != nil {
		pt.log.Debugw("Failed to read Addr response response", "Err", err, "PeerID", remotePeerIDShort)
		return fmt.Errorf("failed to read Addr response")
	}

	sig := resp.Sig
	resp.Sig = nil
	if err := pt.verify(resp, sig, s.Conn().RemotePublicKey()); err != nil {
		pt.log.Debugw("Failed to verify message signature", "Err", err, "PeerID", remotePeerIDShort)
		return fmt.Errorf("failed to verify message signature")
	}

	// we need to ensure the amount of addresses does not exceed the max. address expected
	if len(resp.Addresses) > pt.LocalPeer().cfg.Peer.MaxAddrsExpected {
		pt.log.Debugw("Too many addresses received. Ignoring addresses", "PeerID", remotePeerIDShort, "NumAddrReceived", len(resp.Addresses))
		return fmt.Errorf("too many addresses received. Ignoring addresses")
	}

	invalidAddrs := 0
	validAddrs := []*wire.Address{}
	for _, addr := range resp.Addresses {

		p, _ := pt.LocalPeer().PeerFromAddr(addr.Address, true)
		p.Timestamp = time.Unix(addr.Timestamp, 0)

		if p.IsBadTimestamp() {
			p.Timestamp = time.Now().Add(-1 * time.Hour * 24 * 5)
		}

		if pt.PM().AddOrUpdatePeer(p) != nil {
			invalidAddrs++
			continue
		}

		validAddrs = append(validAddrs, addr)
	}

	if len(validAddrs) > 0 {
		go pt.RelayAddr(validAddrs)
	}

	pt.log.Infow("Received Addr from peer", "PeerID", remotePeerIDShort, "NumAddrs", len(resp.Addresses), "InvalidAddrs", invalidAddrs)

	return nil
}

// OnAddr handles incoming addr messages
func (pt *Inception) OnAddr(s net.Stream) {
	pt.onAddr(s)
}

// RelayAddr relays addrs
func (pt *Inception) RelayAddr(addresses []*wire.Address) {
	fmt.Println("Relay", addresses)
}
