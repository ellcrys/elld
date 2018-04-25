package peer

import (
	"bufio"
	"context"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"golang.org/x/crypto/sha3"

	"github.com/ellcrys/druid/util"
	"github.com/ellcrys/druid/wire"
	net "github.com/libp2p/go-libp2p-net"
	ma "github.com/multiformats/go-multiaddr"
	pc "github.com/multiformats/go-multicodec/protobuf"
)

// onAddr processes "addr" message
func (pt *Inception) onAddr(s net.Stream) ([]*wire.Address, error) {

	remoteAddr := util.FullRemoteAddressFromStream(s)
	remotePeer := NewRemotePeer(remoteAddr, pt.LocalPeer())
	remotePeerIDShort := remotePeer.ShortID()
	resp := &wire.Addr{}
	decoder := pc.Multicodec(nil).Decoder(bufio.NewReader(s))
	if err := decoder.Decode(resp); err != nil {
		pt.log.Debug("Failed to read Addr response response", "Err", err, "PeerID", remotePeerIDShort)
		return nil, fmt.Errorf("failed to read Addr response")
	}

	sig := resp.Sig
	resp.Sig = nil
	if err := pt.verify(resp, sig, s.Conn().RemotePublicKey()); err != nil {
		pt.log.Debug("Failed to verify message signature", "Err", err, "PeerID", remotePeerIDShort)
		return nil, fmt.Errorf("failed to verify message signature")
	}

	// we need to ensure the amount of addresses does not exceed the max. address expected
	if len(resp.Addresses) > pt.LocalPeer().cfg.Peer.MaxAddrsExpected {
		pt.log.Debug("Too many addresses received. Ignoring addresses", "PeerID", remotePeerIDShort, "NumAddrReceived", len(resp.Addresses))
		return nil, fmt.Errorf("too many addresses received. Ignoring addresses")
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
			continue
		}
	}

	pt.log.Info("Received Addr message from peer", "PeerID", remotePeerIDShort, "NumAddrs", len(resp.Addresses), "InvalidAddrs", invalidAddrs)
	return resp.Addresses, nil
}

// OnAddr handles incoming addr messages.
// Received addresses are relayed.
func (pt *Inception) OnAddr(s net.Stream) {

	remoteAddr := util.FullRemoteAddressFromStream(s)
	remotePeer := NewRemotePeer(remoteAddr, pt.LocalPeer())
	if pt.LocalPeer().isDevMode() && !util.IsDevAddr(remotePeer.IP) {
		pt.log.Debug("Can't accept message from non local or private IP in development mode", "Addr", remotePeer.GetMultiAddr(), "Msg", "Addr")
		return
	}

	addresses, err := pt.onAddr(s)
	if err != nil {
		return
	}

	if len(addresses) > 0 {
		go pt.RelayAddr(addresses)
	}
}

// getAddrRelayPeers returns two addresses that we will relay any incoming Addr message
// for the next 24 hours. If we haven't determined these addresses or it has been 24 hours since
// we last selected an addr, we randomly select new addresses from the candidate addresses otherwise,
// we return the current relay addresses.
func (pt *Inception) getAddrRelayPeers(candidateAddrs []*wire.Address) [2]*Peer {

	now := time.Now()
	pt.arm.Lock()
	defer pt.arm.Unlock()

	if pt.lastRelayPeersSelectionTime.Add(24 * time.Hour).Before(time.Now()) { // select new addresses

		var sortCandidateAddrs []*util.BigIntWithMeta
		for _, c := range candidateAddrs {

			// ensure local peer is not a candidate
			mAddr, _ := ma.NewMultiaddr(c.Address)
			if pt.LocalPeer().IsSameID(util.IDFromAddr(mAddr).Pretty()) {
				continue
			}

			h := sha3.New256()
			h.Write([]byte(c.Address))               // add the address
			h.Write([]byte(strconv.Itoa(now.Day()))) // the current day
			h.Write(pt.LocalPeer().rSeed)            // random seed

			sortCandidateAddrs = append(sortCandidateAddrs, &util.BigIntWithMeta{
				Int:  big.NewInt(0).SetBytes(h.Sum(nil)),
				Meta: c,
			})
		}

		// order sortCandidateAddrs in ascending order
		util.AscOrderBigIntMeta(sortCandidateAddrs)

		if len(sortCandidateAddrs) >= 1 {
			p, _ := pt.LocalPeer().PeerFromAddr(sortCandidateAddrs[0].Meta.(*wire.Address).Address, true)
			p.Timestamp = time.Unix(sortCandidateAddrs[0].Meta.(*wire.Address).Timestamp, 0)
			pt.addrRelayPeers[0] = p

			if len(sortCandidateAddrs) >= 2 {
				p2, _ := pt.LocalPeer().PeerFromAddr(sortCandidateAddrs[1].Meta.(*wire.Address).Address, true)
				p2.Timestamp = time.Unix(sortCandidateAddrs[1].Meta.(*wire.Address).Timestamp, 0)
				pt.addrRelayPeers[1] = p2
			}
		}

		pt.lastRelayPeersSelectionTime = time.Now()

		return pt.addrRelayPeers
	}

	return pt.addrRelayPeers
}

// RelayAddr relays addrs under the following rules:
// * "addr" message must contain not more than 10 addrs.
// * all addresses must be valid and different from the local peer address
// * Only addresses within 60 minutes from the current time.
// * Only routable addresses are allowed.
func (pt *Inception) RelayAddr(addrs []*wire.Address) error {

	var relayable []*wire.Address
	now := time.Now()

	if len(addrs) > 10 {
		return fmt.Errorf("too many items in addr message")
	}

	for _, addr := range addrs {

		if !util.IsValidAddr(addr.Address) {
			continue
		}

		mAddr, _ := ma.NewMultiaddr(addr.Address)
		if pt.LocalPeer().IsSameID(util.IDFromAddr(mAddr).Pretty()) {
			continue
		}

		addrTime := time.Unix(addr.Timestamp, 0)
		if now.Add(60 * time.Minute).Before(addrTime) {
			continue
		}

		if !pt.LocalPeer().DevMode() && !util.IsRoutableAddr(addr.Address) {
			continue
		}

		relayable = append(relayable, addr)
	}

	if len(relayable) == 0 {
		return fmt.Errorf("no addr to relay")
	}

	// get the peers to relay address to
	relayPeers := pt.getAddrRelayPeers(relayable)
	numRelayPeers := len(relayPeers)
	if relayPeers[0] == nil {
		numRelayPeers--
	}
	if relayPeers[1] == nil {
		numRelayPeers--
	}

	pt.log.Debug("Relaying addresses", "NumAddrsToRelay", len(relayable), "RelayPeers", numRelayPeers)

	successfullyRelayed := 0
	addrMsg := &wire.Addr{Addresses: relayable}
	addrMsg.Sig = pt.sign(addrMsg)
	for _, remotePeer := range relayPeers {

		if remotePeer == nil {
			continue
		}

		s, err := pt.LocalPeer().addToPeerStore(remotePeer).newStream(context.Background(), remotePeer.ID(), util.AddrVersion)
		if err != nil {
			pt.log.Debug("Addr message failed. failed to connect to peer", "Err", err, "PeerID", remotePeer.ShortID())
			continue
		}

		w := bufio.NewWriter(s)
		if err := pc.Multicodec(nil).Encoder(w).Encode(addrMsg); err != nil {
			pt.log.Debug("Addr failed. failed to write to stream", "Err", err, "PeerID", remotePeer.ShortID())
			continue
		}

		w.Flush()
		s.Close()
		successfullyRelayed++
	}

	pt.log.Info("Relay completed", "NumAddrsToRelay", len(relayable), "NumRelayed", successfullyRelayed)

	return nil
}
