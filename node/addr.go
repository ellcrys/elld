package node

import (
	"fmt"
	"math/big"
	"sort"
	"time"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/util/cache"
	"github.com/ellcrys/elld/wire"
	net "github.com/libp2p/go-libp2p-net"
)

// onAddr processes wire.Addr message
func (g *Gossip) onAddr(s net.Stream) ([]*wire.Address, error) {

	remoteAddr := util.RemoteAddrFromStream(s)
	remotePeer := NewRemoteNode(remoteAddr, g.engine)
	remotePeerIDShort := remotePeer.ShortID()

	resp := &wire.Addr{}
	if err := ReadStream(s, resp); err != nil {
		g.log.Debug("Failed to read Addr response response", "Err", err, "PeerID", remotePeerIDShort)
		return nil, fmt.Errorf("failed to read Addr response: %s", err)
	}

	g.PM().UpdateLastSeen(remotePeer)

	// we need to ensure the amount of
	// addresses does not exceed the
	// maximum address expected
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
		p.lastSeen = time.Unix(addr.Timestamp, 0)

		// Check if the timestamp us acceptable according to
		// the discovery protocol rules
		if p.IsBadTimestamp() {
			p.lastSeen = time.Now().UTC().Add(-1 * time.Hour * 24 * 5)
		}

		// Add the remote peer to the peer manager's list
		if g.PM().UpdateLastSeen(p) != nil {
			invalidAddrs++
			continue
		}
	}

	g.log.Info("Received Addr message from peer",
		"PeerID", remotePeerIDShort,
		"NumAddrs", len(resp.Addresses),
		"InvalidAddrs", invalidAddrs)

	return resp.Addresses, nil
}

// OnAddr handles incoming wire.Addr message.
// Received addresses are relayed.
func (g *Gossip) OnAddr(s net.Stream) {

	defer s.Close()

	remoteAddr := util.RemoteAddrFromStream(s)
	remotePeer := NewRemoteNode(remoteAddr, g.engine)

	// check whether we are are allowed to
	//  interact with the remote peer
	if ok, err := g.engine.canAcceptPeer(remotePeer); !ok {
		g.log.Debug(fmt.Sprintf("Can't accept message from peer: %s", err.Error()),
			"Addr", remotePeer.GetAddress(), "Msg", "GetAddr")
		return
	}

	// process the stream and return
	// the addresses set
	addresses, err := g.onAddr(s)
	if err != nil {
		g.engine.event.Emit(EventAddrProcessed, err)
		return
	}

	// As long as we have more that one
	// address, we should attempt to relay
	// it/them to other peers
	if len(addresses) > 0 {
		go g.RelayAddresses(addresses)
	}

	g.engine.event.Emit(EventAddrProcessed)
}

// SelectRelayPeers returns two random remote
// nodes to broadcast wire.Addr messages to.
// These peers are selected from the given
// candidate addresses.
// The selected peers are cached for up to
// 24 hours after which the peers are
// reselected.
func (g *Gossip) SelectRelayPeers(candidates []*wire.Address) []*Node {
	g.mtx.Lock()
	defer g.mtx.Unlock()

	now := time.Now().UTC()

	// If the last time we selected the peers
	// have not surpassed 24 hours, return the
	// last selected peers
	if !g.relayPeerSelectedAt.Add(24 * time.Hour).Before(now) {
		return g.RelayPeers
	}

	type addrInfo struct {
		hash      *big.Int
		address   util.NodeAddr
		timestamp int64
	}

	var candidatesInfo []addrInfo
	for _, c := range candidates {

		// Make sure the address isn't the same
		// as the address of the local node
		if g.engine.IsSameID(c.Address.ID().Pretty()) {
			continue
		}

		// We need to get a numeric representation
		// of the address. Preferable as big.Int
		addrHash := util.Blake2b256([]byte(c.Address))
		addrBigInt := new(big.Int).SetBytes(addrHash)

		// Add the address along with other
		// info into to slice of valid addresses
		candidatesInfo = append(candidatesInfo, addrInfo{
			hash:      addrBigInt,
			address:   c.Address,
			timestamp: c.Timestamp,
		})
	}

	// sort the filtered candidates
	// in ascending order.
	sort.Slice(candidatesInfo, func(i, j int) bool {
		return candidatesInfo[i].hash.Cmp(candidatesInfo[j].hash) == -1
	})

	// Clear the relay peer cache if we
	// have at least 2 candidates
	if len(candidatesInfo) >= 2 {
		g.RelayPeers = []*Node{}
	}

	// Create an remote engine object
	// from the first 2 addresses.
	for _, info := range candidatesInfo {
		n, _ := g.engine.NodeFromAddr(info.address, true)
		n.lastSeen = time.Unix(info.timestamp, 0)
		g.RelayPeers = append(g.RelayPeers, n)
		if len(g.RelayPeers) == 2 {
			break
		}
	}

	// Update the relay peer selection time
	g.relayPeerSelectedAt = time.Now().UTC()

	return g.RelayPeers

}

func makeAddrRelayHistoryKey(addr *wire.Addr, peer *Node) []interface{} {
	return []interface{}{util.SerializeMsg(addr), peer.StringID()}
}

// RelayAddresses relays wire.Address under
// the following rules:
// * wire.Address message must contain not more
//   than 10 addrs.
// * all addresses must be valid and
//   different from the local peer address
// * Only addresses within 60 minutes from
//   the current time.
// * Only routable addresses are allowed.
func (g *Gossip) RelayAddresses(addrs []*wire.Address) []error {

	var errs []error
	var relayable []*wire.Address
	now := time.Now().UTC()

	// Do not proceed if there are more
	// than 10 addresses
	if len(addrs) > 10 {
		errs = append(errs, fmt.Errorf("too many addresses"+
			" in the message"))
		return errs
	}

	for _, addr := range addrs {

		// We must ensure we don't relay invalid addresses
		if !addr.Address.IsValid() {
			errs = append(errs, fmt.Errorf("address {%s} is not valid",
				addr.Address))
			continue
		}

		// Ignore an address that matches the local
		if g.engine.IsSameID(addr.Address.ID().Pretty()) {
			errs = append(errs, fmt.Errorf("address {%s} is the same"+
				" as local peer's", addr.Address))
			continue
		}

		// Ignore an address whose timestamp is over 60 minutes old
		addrTime := time.Unix(addr.Timestamp, 0)
		if now.Add(60 * time.Minute).Before(addrTime) {
			errs = append(errs, fmt.Errorf("address {%s} is over 60 minutes old",
				addr.Address))
			continue
		}

		// In non-production mode, we are allowed
		// to relay non-routable addresses.
		// But we can't allow them in production
		if g.engine.ProdMode() && !addr.Address.IsRoutable() {
			errs = append(errs, fmt.Errorf("address {%s} is not routable",
				addr.Address))
			continue
		}

		relayable = append(relayable, addr)
	}

	// When no address is relayable, we
	// exist with an error
	if len(relayable) == 0 {
		errs = append(errs, fmt.Errorf("no addr to relay"))
		return errs
	}

	// select two peers from the list of
	// relayable peers that we will send the addresses to
	relayPeers := g.SelectRelayPeers(relayable)

	g.log.Debug("Relaying addresses", "NumAddrsToRelay", len(relayable),
		"RelayPeers", len(relayPeers))

	relayed := 0
	for _, remotePeer := range relayPeers {

		// Construct the address message.
		// Be sure to not include an address
		// matching the remote peer's
		addrMsg := &wire.Addr{}
		for _, p := range relayable {
			if !p.Address.Equal(remotePeer.GetAddress()) {
				addrMsg.Addresses = append(addrMsg.Addresses, p)
			}
		}

		historyKey := makeAddrRelayHistoryKey(addrMsg, remotePeer)

		// ensure we have not relayed same
		// message to this peer before
		if g.engine.history.HasMulti(historyKey...) {
			errs = append(errs, fmt.Errorf("already sent same Addr to node"))
			g.log.Debug("Already sent same Addr to node. Skipping.",
				"PeerID", remotePeer.ShortID())
			continue
		}

		s, c, err := g.NewStream(remotePeer, config.AddrVersion)
		if err != nil {
			errs = append(errs, fmt.Errorf("Addr message failed."+
				" failed to connect to peer {%s}", remotePeer.ShortID()))
			g.log.Debug("Addr message failed. failed to connect to peer",
				"Err", err, "PeerID", remotePeer.ShortID())
			continue
		}
		defer c()
		defer s.Close()

		if err := WriteStream(s, addrMsg); err != nil {
			errs = append(errs, fmt.Errorf("Addr failed. failed to "+
				"write to stream to peer {%s}", remotePeer.ShortID()))
			g.log.Debug("Addr failed. failed to write to stream",
				"Err", err, "PeerID", remotePeer.ShortID())
			continue
		}

		g.PM().UpdateLastSeen(remotePeer)

		// add new history
		g.engine.history.AddMulti(cache.Sec(600), historyKey...)

		relayed++
	}

	g.log.Debug("Relay completed", "NumAddrsToRelay",
		len(relayable), "NumRelayed", relayed)
	defer g.engine.event.Emit(EventAddressesRelayed)

	return errs
}
