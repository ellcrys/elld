package node

import (
	"fmt"
	"math/big"
	"sort"
	"sync"
	"time"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/util/cache"
	"github.com/ellcrys/elld/wire"
	net "github.com/libp2p/go-libp2p-net"
)

// onAddr processes wire.Addr message
func (g *Gossip) onAddr(s net.Stream) ([]*wire.Address, error) {

	remoteAddr := util.RemoteAddrFromStream(s)
	rp := NewRemoteNode(remoteAddr, g.engine)
	rpIDStr := rp.ShortID()

	resp := &wire.Addr{}
	if err := ReadStream(s, resp); err != nil {
		return nil, g.logErr(err, rp, "[OnAddr] Failed to read stream")
	}

	g.PM().AddOrUpdateNode(rp)

	// we need to ensure the amount of addresses does
	// not exceed the maximum expected
	if int64(len(resp.Addresses)) > g.engine.cfg.Node.MaxAddrsExpected {
		g.log.Debug("Too many addresses received. Ignoring addresses",
			"PeerID", rpIDStr,
			"NumAddrReceived", len(resp.Addresses))
		return nil, fmt.Errorf("too many addresses received. Ignoring addresses")
	}

	invalidAddrs := 0

	// Validate each address before we addthem to the peer list
	for _, addr := range resp.Addresses {

		p, _ := g.engine.NodeFromAddr(addr.Address, true)

		if !addr.Address.IsValid() || (!g.engine.TestMode() && !addr.Address.IsRoutable()) {
			invalidAddrs++
			continue
		}

		// Check whether we know this node as a banned peer
		if g.PM().IsBanned(p) {
			invalidAddrs++
			continue
		}

		// Add the remote peer to the peer manager's list
		g.PM().AddOrUpdateNode(p)
	}

	g.log.Info("Received addresses", "PeerID", rpIDStr,
		"NumAddrs", len(resp.Addresses),
		"InvalidAddrs", invalidAddrs)

	return resp.Addresses, nil
}

// OnAddr handles incoming wire.Addr message.
// Received addresses are relayed.
func (g *Gossip) OnAddr(s net.Stream) {

	defer s.Close()
	remoteAddr := util.RemoteAddrFromStream(s)
	rp := NewRemoteNode(remoteAddr, g.engine)

	// check whether we are allowed to receive this peer's message
	if ok, err := g.PM().CanAcceptNode(rp); !ok {
		g.logErr(err, rp, "message unaccepted")
		return
	}

	// process the stream and return
	// the addresses set
	addresses, err := g.onAddr(s)
	if err != nil {
		g.engine.event.Emit(EventAddrProcessed, err)
		return
	}

	// As long as we have more that one address,
	// we should attempt to relay them to other peers
	if len(addresses) > 0 {
		g.RelayAddresses(addresses)
	}

	g.engine.event.Emit(EventAddrProcessed)
}

// PickBroadcasters selects N random addresses from
// the given slice of addresses and caches them to
// be used as broadcasters.
// They are returned on subsequent calls and only
// renewed when there are less than N addresses or the
// cache is over 24 hours since it was last updated.
func (g *Gossip) PickBroadcasters(addresses []*wire.Address, n int) *BroadcastPeers {
	g.mtx.Lock()
	defer g.mtx.Unlock()

	now := time.Now()
	if g.broadcasters.Len() == n && !g.broadcastersUpdatedAt.
		Add(24*time.Hour).Before(now) {
		return g.broadcasters
	}

	type addrInfo struct {
		hash      *big.Int
		address   util.NodeAddr
		timestamp int64
	}

	var candidatesInfo []addrInfo
	for _, c := range addresses {

		// Make sure the address isn't the same
		// as the address of the local node
		if g.engine.IsSameID(c.Address.ID().Pretty()) {
			continue
		}

		// We need to get a numeric representation
		// of the address
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

	// sort the filtered candidates in ascending order.
	sort.Slice(candidatesInfo, func(i, j int) bool {
		return candidatesInfo[i].hash.Cmp(candidatesInfo[j].hash) == -1
	})

	// Clear the cache if we have at least N cached
	if len(candidatesInfo) >= n {
		g.broadcasters.Clear()
	}

	// Create a remote engine object from the first N addresses.
	for _, info := range candidatesInfo {
		node, _ := g.engine.NodeFromAddr(info.address, true)
		node.lastSeen = time.Unix(info.timestamp, 0)
		g.broadcasters.Add(node)
		if g.broadcasters.Len() == n {
			break
		}
	}

	g.broadcastersUpdatedAt = time.Now()

	return g.broadcasters

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

	var rMtx = sync.Mutex{}
	var errs []error
	var relayable []*wire.Address
	now := time.Now()

	// Do not proceed if there are more than 10 addresses
	if len(addrs) > 10 {
		errs = append(errs, fmt.Errorf("too many addresses in the message"))
		return errs
	}

	for _, addr := range addrs {

		// We must ensure we don't relay invalid addresses
		if !addr.Address.IsValid() {
			errs = append(errs, fmt.Errorf("address {%s} is not valid", addr.Address))
			continue
		}

		// Ignore an address that matches the local node's address
		if g.engine.IsSameID(addr.Address.ID().Pretty()) {
			errs = append(errs, fmt.Errorf("address {%s} is the same as local peer's",
				addr.Address))
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

	// Select up to 2 peers to act as broadcasters
	broadcasters := g.PickBroadcasters(relayable, 2)

	g.log.Debug("Relaying addresses", "NumAddrs", len(relayable),
		"NumBroadcasters", broadcasters.Len())

	relayed := 0
	for _, rp := range broadcasters.Peers() {

		msg := &wire.Addr{}

		for _, p := range relayable {

			// We must have no history sending/receiving this
			// address to/from this peer in recent time
			hk := []interface{}{p.Address.String(), rp.StringID()}
			if g.engine.history.HasMulti(hk) {
				continue
			}

			// Remote peer address must not be same as rp
			if p.Address.Equal(rp.GetAddress()) {
				continue
			}

			msg.Addresses = append(msg.Addresses, p)
		}

		go func(rp types.Engine) {

			s, c, err := g.NewStream(rp, config.AddrVersion)
			if err != nil {
				rMtx.Lock()
				err := g.logConnectErr(err, rp, "[RelayAddresses] Failed to connect to peer")
				errs = append(errs, err)
				rMtx.Unlock()
				return
			}
			defer c()
			defer s.Close()

			if err := WriteStream(s, msg); err != nil {
				rMtx.Lock()
				err := g.logErr(err, rp, "[RelayAddresses] Failed to write to peer")
				errs = append(errs, err)
				rMtx.Unlock()
				return
			}

			g.PM().AddOrUpdateNode(rp)

			for _, p := range relayable {
				hk := []interface{}{p.Address.String(), rp.StringID()}
				g.engine.history.AddMulti(cache.Sec(3600), hk)
			}

			rMtx.Lock()
			relayed++
			rMtx.Unlock()

		}(rp)
	}

	g.log.Debug("Address relayed", "NumAddrs", len(relayable), "NumRelayed", relayed)
	g.engine.event.Emit(EventAddressesRelayed)

	rMtx.Lock()
	defer rMtx.Unlock()

	return errs
}
