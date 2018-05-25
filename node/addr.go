package node

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
	remotePeer := NewRemoteNode(remoteAddr, pt.LocalPeer())
	remotePeerIDShort := remotePeer.ShortID()

	resp := &wire.Addr{}
	decoder := pc.Multicodec(nil).Decoder(bufio.NewReader(s))
	if err := decoder.Decode(resp); err != nil {
		s.Reset()
		pt.log.Debug("Failed to read Addr response response", "Err", err, "PeerID", remotePeerIDShort)
		return nil, fmt.Errorf("failed to read Addr response")
	}

	// we need to ensure the amount of addresses does not exceed the max. address expected
	if int64(len(resp.Addresses)) > pt.LocalPeer().cfg.Node.MaxAddrsExpected {
		s.Reset()
		pt.log.Debug("Too many addresses received. Ignoring addresses", "PeerID", remotePeerIDShort, "NumAddrReceived", len(resp.Addresses))
		return nil, fmt.Errorf("too many addresses received. Ignoring addresses")
	}

	invalidAddrs := 0
	for _, addr := range resp.Addresses {

		p, _ := pt.LocalPeer().NodeFromAddr(addr.Address, true)
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

	defer s.Close()

	remoteAddr := util.FullRemoteAddressFromStream(s)
	remotePeer := NewRemoteNode(remoteAddr, pt.LocalPeer())
	if pt.LocalPeer().isDevMode() && !util.IsDevAddr(remotePeer.IP) {
		s.Reset()
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
func (pt *Inception) getAddrRelayPeers(candidateAddrs []*wire.Address) [2]*Node {

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
			p, _ := pt.LocalPeer().NodeFromAddr(sortCandidateAddrs[0].Meta.(*wire.Address).Address, true)
			p.Timestamp = time.Unix(sortCandidateAddrs[0].Meta.(*wire.Address).Timestamp, 0)
			pt.addrRelayPeers[0] = p

			if len(sortCandidateAddrs) >= 2 {
				p2, _ := pt.LocalPeer().NodeFromAddr(sortCandidateAddrs[1].Meta.(*wire.Address).Address, true)
				p2.Timestamp = time.Unix(sortCandidateAddrs[1].Meta.(*wire.Address).Timestamp, 0)
				pt.addrRelayPeers[1] = p2
			}
		}

		pt.lastRelayPeersSelectionTime = time.Now()

		return pt.addrRelayPeers
	}

	return pt.addrRelayPeers
}

func makeAddrRelayHistoryKey(addr *wire.Addr, peer *Node) MultiKey {
	return []interface{}{util.SerializeMsg(addr), peer.StringID()}
}

// RelayAddr relays addrs under the following rules:
// * "addr" message must contain not more than 10 addrs.
// * all addresses must be valid and different from the local peer address
// * Only addresses within 60 minutes from the current time.
// * Only routable addresses are allowed.
func (pt *Inception) RelayAddr(addrs []*wire.Address) []error {

	var errs []error
	var relayable []*wire.Address
	now := time.Now()

	if len(addrs) > 10 {
		errs = append(errs, fmt.Errorf("too many addresses in the message"))
		return errs
	}

	for _, addr := range addrs {

		if !util.IsValidAddr(addr.Address) {
			errs = append(errs, fmt.Errorf("address {%s} is not valid", addr.Address))
			continue
		}

		mAddr, _ := ma.NewMultiaddr(addr.Address)
		if pt.LocalPeer().IsSameID(util.IDFromAddr(mAddr).Pretty()) {
			errs = append(errs, fmt.Errorf("address {%s} is the same as local peer's", addr.Address))
			continue
		}

		addrTime := time.Unix(addr.Timestamp, 0)
		if now.Add(60 * time.Minute).Before(addrTime) {
			errs = append(errs, fmt.Errorf("address {%s} is over 60 minutes old", addr.Address))
			continue
		}

		if !pt.LocalPeer().DevMode() && !util.IsRoutableAddr(addr.Address) {
			errs = append(errs, fmt.Errorf("address {%s} is not routable", addr.Address))
			continue
		}

		relayable = append(relayable, addr)
	}

	if len(relayable) == 0 {
		errs = append(errs, fmt.Errorf("no addr to relay"))
		return errs
	}

	// get the peers to relay address to
	relayPeers := pt.getAddrRelayPeers(relayable)
	numRelayPeers := len(relayPeers)
	for _, p := range relayPeers {
		if p == nil {
			numRelayPeers--
		}
	}

	pt.log.Debug("Relaying addresses", "NumAddrsToRelay", len(relayable), "RelayPeers", numRelayPeers)

	successfullyRelayed := 0
	addrMsg := &wire.Addr{Addresses: relayable}
	for _, remotePeer := range relayPeers {

		if remotePeer == nil {
			continue
		}

		historyKey := makeAddrRelayHistoryKey(addrMsg, remotePeer)

		// ensure we have not relayed same message to this peer before
		if pt.LocalPeer().History().Has(historyKey) {
			errs = append(errs, fmt.Errorf("already sent same Addr to node"))
			pt.log.Debug("Already sent same Addr to node. Skipping.", "PeerID", remotePeer.ShortID())
			continue
		}

		s, err := pt.LocalPeer().addToPeerStore(remotePeer).newStream(context.Background(), remotePeer.ID(), util.AddrVersion)
		if err != nil {
			errs = append(errs, fmt.Errorf("Addr message failed. failed to connect to peer {%s}", remotePeer.ShortID()))
			pt.log.Debug("Addr message failed. failed to connect to peer", "Err", err, "PeerID", remotePeer.ShortID())
			continue
		}

		w := bufio.NewWriter(s)
		if err := pc.Multicodec(nil).Encoder(w).Encode(addrMsg); err != nil {
			s.Reset()
			errs = append(errs, fmt.Errorf("AAddr failed. failed to write to stream to peer {%s}", remotePeer.ShortID()))
			pt.log.Debug("Addr failed. failed to write to stream", "Err", err, "PeerID", remotePeer.ShortID())
			continue
		}

		// add new history
		pt.LocalPeer().History().Add(historyKey)

		w.Flush()
		s.Close()
		successfullyRelayed++
	}

	pt.log.Info("Relay completed", "NumAddrsToRelay", len(relayable), "NumRelayed", successfullyRelayed)

	return errs
}
