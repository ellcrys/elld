package gossip

import (
	"fmt"
	"math/big"
	"sort"
	"time"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/util/cache"
	net "github.com/libp2p/go-libp2p-net"
)

// onAddr processes core.Addr message
func (g *Manager) onAddr(s net.Stream, rp core.Engine) ([]*core.Address, error) {

	resp := &core.Addr{}
	if err := ReadStream(s, resp); err != nil {
		return nil, g.logErr(err, rp, "[OnAddr] Failed to read stream")
	}

	// we need to ensure the amount of addresses does
	// not exceed the maximum expected
	if int64(len(resp.Addresses)) > g.engine.GetCfg().Node.MaxAddrsExpected {
		g.log.Debug("Too many addresses received. Ignoring addresses",
			"PeerID", rp.ShortID(),
			"NumAddrReceived", len(resp.Addresses))
		return nil, fmt.Errorf("too many addresses received. Ignoring addresses")
	}

	invalidAddrs := 0

	// Validate each address before we addthem to the peer list
	for _, addr := range resp.Addresses {

		p := g.engine.NewRemoteNode(addr.Address)

		if !addr.Address.IsValid() || (!g.engine.TestMode() &&
			!addr.Address.IsRoutable()) {
			invalidAddrs++
			continue
		}

		// Check whether we know this node as a banned peer
		if g.PM().IsBanned(p) {
			invalidAddrs++
			continue
		}

		g.PM().AddOrUpdateNode(rp)
	}

	g.log.Debug("Received addresses", "PeerID", rp.ShortID(),
		"NumAddrs", len(resp.Addresses),
		"InvalidAddrs", invalidAddrs)

	return resp.Addresses, nil
}

// OnAddr handles incoming core.Addr message.
// Received addresses are relayed.
func (g *Manager) OnAddr(s net.Stream, rp core.Engine) error {

	defer s.Close()

	// process the stream and return the addresses set
	addresses, err := g.onAddr(s, rp)
	if err != nil {
		g.engine.GetEventEmitter().Emit(EventAddrProcessed, err)
		return err
	}

	// As long as we have more that one address,
	// we should attempt to relay them to other peers
	if len(addresses) > 0 {
		g.RelayAddresses(addresses)
	}

	g.engine.GetEventEmitter().Emit(EventAddrProcessed)
	return nil
}

// PickBroadcasters selects N random addresses from
// the given slice of addresses and caches them to
// be used as broadcasters.
// They are returned on subsequent calls and only
// renewed when there are less than N addresses or the
// cache is over 24 hours since it was last updated.
func (g *Manager) PickBroadcasters(addresses []*core.Address, n int) *core.BroadcastPeers {
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
		node := g.engine.NewRemoteNode(info.address)
		node.SetLastSeen(time.Unix(info.timestamp, 0))
		g.broadcasters.Add(node)
		if g.broadcasters.Len() == n {
			break
		}
	}

	g.broadcastersUpdatedAt = time.Now()

	return g.broadcasters

}

// PickBroadcastersFromPeers is like PickBroadcasters except it
// accepts a slice of peer engine objects.
func (g *Manager) PickBroadcastersFromPeers(peers []core.Engine, n int) *core.BroadcastPeers {
	peerAddrs := []*core.Address{}
	for _, peer := range peers {
		peerAddrs = append(peerAddrs, &core.Address{
			Address:   peer.GetAddress(),
			Timestamp: peer.GetLastSeen().Unix(),
		})
	}
	return g.PickBroadcasters(peerAddrs, n)
}

func makeAddrRelayHistoryKey(addr *core.Addr, peer core.Engine) []interface{} {
	return []interface{}{util.SerializeMsg(addr), peer.StringID()}
}

// RelayAddresses relays core.Address under
// the following rules:
// * core.Address message must contain not more
//   than 10 addrs.
// * all addresses must be valid and
//   different from the local peer address
// * Only addresses within 60 minutes from
//   the current time.
// * Only routable addresses are allowed.
func (g *Manager) RelayAddresses(addrs []*core.Address) []error {

	var errs []error
	var relayable []*core.Address
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

		// Check whether the node is associated with a banned peer
		rn := g.engine.NewRemoteNode(addr.Address)
		if g.PM().IsBanned(rn) {
			errs = append(errs, fmt.Errorf("address {%s} associated with a banned peer",
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

		msg := &core.Addr{}

		for _, p := range relayable {

			// We must have no history sending/receiving this
			// address to/from this peer in recent time
			hk := []interface{}{p.Address.String(), rp.StringID()}
			if g.engine.GetHistory().HasMulti(hk) {
				continue
			}

			// Remote peer address must not be same as rp
			if p.Address.Equal(rp.GetAddress()) {
				continue
			}

			msg.Addresses = append(msg.Addresses, p)
		}

		s, c, err := g.NewStream(rp, config.Versions.Addr)
		if err != nil {
			err := g.logConnectErr(err, rp, "[RelayAddresses] Failed to connect to peer")
			errs = append(errs, err)
			c()
			continue
		}
		defer c()
		defer s.Close()

		if err := WriteStream(s, msg); err != nil {
			err := g.logErr(err, rp, "[RelayAddresses] Failed to write to peer")
			errs = append(errs, err)
			continue
		}

		for _, p := range relayable {
			hk := []interface{}{p.Address.String(), rp.StringID()}
			g.engine.GetHistory().AddMulti(cache.Sec(3600), hk)
		}

		relayed++

	}

	g.log.Debug("Address relayed", "NumAddrs", len(relayable), "NumRelayed", relayed)
	g.engine.GetEventEmitter().Emit(EventAddressesRelayed)

	return errs
}
