package node

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"golang.org/x/crypto/sha3"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/node/histcache"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/wire"
	net "github.com/libp2p/go-libp2p-net"
	ma "github.com/multiformats/go-multiaddr"
)

// onAddr processes "addr" message
func (g *Gossip) onAddr(s net.Stream) ([]*wire.Address, error) {

	defer s.Reset()

	remoteAddr := util.FullRemoteAddressFromStream(s)
	remotePeer := NewRemoteNode(remoteAddr, g.engine)
	remotePeerIDShort := remotePeer.ShortID()

	// read message from the stream
	resp := &wire.Addr{}
	if err := readStream(s, resp); err != nil {
		g.log.Debug("Failed to read Addr response response", "Err", err, "PeerID", remotePeerIDShort)
		return nil, fmt.Errorf("failed to read Addr response: %s", err)
	}

	// we need to ensure the amount of addresses does not exceed the max. address expected
	if int64(len(resp.Addresses)) > g.engine.cfg.Node.MaxAddrsExpected {
		g.log.Debug("Too many addresses received. Ignoring addresses", "PeerID", remotePeerIDShort, "NumAddrReceived", len(resp.Addresses))
		return nil, fmt.Errorf("too many addresses received. Ignoring addresses")
	}

	invalidAddrs := 0

	// Validate each address before we add them to the peer
	// list maintained by the peer manager
	for _, addr := range resp.Addresses {

		// first construct a remote node and set the node's timestamp
		p, _ := g.engine.NodeFromAddr(addr.Address, true)
		p.Timestamp = time.Unix(addr.Timestamp, 0)

		// Check if the timestamp us acceptable according to
		// the discovery protocol rules
		if p.IsBadTimestamp() {
			p.Timestamp = time.Now().Add(-1 * time.Hour * 24 * 5)
		}

		// Add the remote peer to the peer manager's list
		if g.PM().AddOrUpdatePeer(p) != nil {
			invalidAddrs++
			continue
		}
	}

	g.log.Info("Received Addr message from peer", "PeerID", remotePeerIDShort, "NumAddrs", len(resp.Addresses), "InvalidAddrs", invalidAddrs)

	return resp.Addresses, nil
}

// OnAddr handles incoming addr messages.
// Received addresses are relayed.
func (g *Gossip) OnAddr(s net.Stream) {

	defer s.Close()

	remoteAddr := util.FullRemoteAddressFromStream(s)
	remotePeer := NewRemoteNode(remoteAddr, g.engine)

	// check whether we are are allowed to interact with the remote peer
	if yes, reason := g.engine.canAcceptPeer(remotePeer); !yes {
		g.log.Debug(fmt.Sprintf("Can't accept message from peer: %s", reason), "Addr", remotePeer.GetMultiAddr(), "Msg", "GetAddr")
		return
	}

	// process the stream and return the addresses set
	addresses, err := g.onAddr(s)
	if err != nil {
		return
	}

	// As long as we have more that one address, we should attempt
	// to relay it to other peers
	if len(addresses) > 0 {
		go g.RelayAddr(addresses)
	}
}

// selectPeersToRelayTo returns two nodes that will be relayed any incoming Addr message
// for the next 24 hours. If we haven't determined these addresses or it has been 24 hours since
// we last selected an addr, we randomly select new addresses from the candidate addresses otherwise,
// we return the current relay addresses.
func (g *Gossip) selectPeersToRelayTo(candidateAddrs []*wire.Address) [2]*Node {

	now := time.Now()
	g.mtx.Lock()
	defer g.mtx.Unlock()

	if g.lastRelayPeersSelectionTime.Add(24 * time.Hour).Before(time.Now()) { // select new addresses

		var sortCandidateAddrs []*util.BigIntWithMeta
		for _, c := range candidateAddrs {

			// ensure local peer is not a candidate
			mAddr, _ := ma.NewMultiaddr(c.Address)
			if g.engine.IsSameID(util.IDFromAddr(mAddr).Pretty()) {
				continue
			}

			h := sha3.New256()
			h.Write([]byte(c.Address))               // add the address
			h.Write([]byte(strconv.Itoa(now.Day()))) // the current day
			h.Write(g.engine.rSeed)                  // random seed

			sortCandidateAddrs = append(sortCandidateAddrs, &util.BigIntWithMeta{
				Int:  big.NewInt(0).SetBytes(h.Sum(nil)),
				Meta: c,
			})
		}

		// order sortCandidateAddrs in ascending order
		util.AscOrderBigIntMeta(sortCandidateAddrs)

		if len(sortCandidateAddrs) >= 1 {
			p, _ := g.engine.NodeFromAddr(sortCandidateAddrs[0].Meta.(*wire.Address).Address, true)
			p.Timestamp = time.Unix(sortCandidateAddrs[0].Meta.(*wire.Address).Timestamp, 0)
			g.addrRelayPeers[0] = p

			if len(sortCandidateAddrs) >= 2 {
				p2, _ := g.engine.NodeFromAddr(sortCandidateAddrs[1].Meta.(*wire.Address).Address, true)
				p2.Timestamp = time.Unix(sortCandidateAddrs[1].Meta.(*wire.Address).Timestamp, 0)
				g.addrRelayPeers[1] = p2
			}
		}

		g.lastRelayPeersSelectionTime = time.Now()

		return g.addrRelayPeers
	}

	return g.addrRelayPeers
}

func makeAddrRelayHistoryKey(addr *wire.Addr, peer *Node) histcache.MultiKey {
	return []interface{}{util.SerializeMsg(addr), peer.StringID()}
}

// RelayAddr relays addrs under the following rules:
// * "addr" message must contain not more than 10 addrs.
// * all addresses must be valid and different from the local peer address
// * Only addresses within 60 minutes from the current time.
// * Only routable addresses are allowed.
func (g *Gossip) RelayAddr(addrs []*wire.Address) []error {

	var errs []error
	var relayable []*wire.Address
	now := time.Now()

	// Do not proceed if addresses to be relayed are more than 10
	if len(addrs) > 10 {
		errs = append(errs, fmt.Errorf("too many addresses in the message"))
		return errs
	}

	for _, addr := range addrs {

		// We must ensure we don't relay invalid addresses
		if !util.IsValidAddr(addr.Address) {
			errs = append(errs, fmt.Errorf("address {%s} is not valid", addr.Address))
			continue
		}

		// Ignore an address that matches the local
		mAddr, _ := ma.NewMultiaddr(addr.Address)
		if g.engine.IsSameID(util.IDFromAddr(mAddr).Pretty()) {
			errs = append(errs, fmt.Errorf("address {%s} is the same as local peer's", addr.Address))
			continue
		}

		// Ignore an address whose timestamp is over 60 minutes old
		addrTime := time.Unix(addr.Timestamp, 0)
		if now.Add(60 * time.Minute).Before(addrTime) {
			errs = append(errs, fmt.Errorf("address {%s} is over 60 minutes old", addr.Address))
			continue
		}

		// In non-production mode, we are allowed to relay non-routable addresses.
		// But we can't allow them in production
		if g.engine.ProdMode() && !util.IsRoutableAddr(addr.Address) {
			errs = append(errs, fmt.Errorf("address {%s} is not routable", addr.Address))
			continue
		}

		relayable = append(relayable, addr)
	}

	if len(relayable) == 0 {
		errs = append(errs, fmt.Errorf("no addr to relay"))
		return errs
	}

	// select two the peers from the list of relayable peers that
	// will be relayed to.
	relayPeers := g.selectPeersToRelayTo(relayable)
	numRelayPeers := len(relayPeers)
	for _, p := range relayPeers {
		if p == nil {
			numRelayPeers--
		}
	}

	g.log.Debug("Relaying addresses", "NumAddrsToRelay", len(relayable), "RelayPeers", numRelayPeers)

	successfullyRelayed := 0
	addrMsg := &wire.Addr{Addresses: relayable}

	for _, remotePeer := range relayPeers {

		if remotePeer == nil {
			continue
		}

		historyKey := makeAddrRelayHistoryKey(addrMsg, remotePeer)

		// ensure we have not relayed same message to this peer before
		if g.engine.History().Has(historyKey) {
			errs = append(errs, fmt.Errorf("already sent same Addr to node"))
			g.log.Debug("Already sent same Addr to node. Skipping.", "PeerID", remotePeer.ShortID())
			continue
		}

		s, err := g.newStream(context.Background(), remotePeer, config.AddrVersion)
		if err != nil {
			errs = append(errs, fmt.Errorf("Addr message failed. failed to connect to peer {%s}", remotePeer.ShortID()))
			g.log.Debug("Addr message failed. failed to connect to peer", "Err", err, "PeerID", remotePeer.ShortID())
			continue
		}
		defer s.Reset()

		if err := writeStream(s, addrMsg); err != nil {
			errs = append(errs, fmt.Errorf("AAddr failed. failed to write to stream to peer {%s}", remotePeer.ShortID()))
			g.log.Debug("Addr failed. failed to write to stream", "Err", err, "PeerID", remotePeer.ShortID())
			continue
		}

		// add new history
		g.engine.History().Add(historyKey)

		successfullyRelayed++
	}

	g.log.Info("Relay completed", "NumAddrsToRelay", len(relayable), "NumRelayed", successfullyRelayed)

	return errs
}
