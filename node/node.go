package node

import (
	"context"
	"fmt"
	mrand "math/rand"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/ellcrys/druid/txpool"

	"github.com/ellcrys/druid/database"
	"github.com/ellcrys/druid/util/logger"

	"github.com/ellcrys/druid/configdir"

	"github.com/thoas/go-funk"

	crypto "github.com/libp2p/go-libp2p-crypto"
	peer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"

	"github.com/ellcrys/druid/util"
	libp2p "github.com/libp2p/go-libp2p"
	host "github.com/libp2p/go-libp2p-host"
	inet "github.com/libp2p/go-libp2p-net"
	protocol "github.com/libp2p/go-libp2p-protocol"
	ma "github.com/multiformats/go-multiaddr"
)

// Node represents a network node
type Node struct {
	cfg             *configdir.Config // node config
	address         ma.Multiaddr      // node multiaddr
	IP              net.IP            // node ip
	host            host.Host         // node libp2p host
	wg              sync.WaitGroup    // wait group for preventing the main thread from exiting
	localNode       *Node             // local node
	peerManager     *Manager          // node manager for managing connections to other remote peers
	protoc          Protocol          // protocol instance
	remote          bool              // remote indicates the node represents a remote peer
	Timestamp       time.Time         // the last time this node was seen/active
	isHardcodedSeed bool              // whether the node was hardcoded as a seed
	log             logger.Logger     // node logger
	rSeed           []byte            // random 256 bit seed to be used for seed random operations
	db              database.DB
	txPool          *txpool.TxPool
}

// NewNode creates a node instance at the specified port
func NewNode(config *configdir.Config, address string, idSeed int64, log logger.Logger) (*Node, error) {

	// generate node identity
	priv, _, err := util.GenerateKeyPair(mrand.New(mrand.NewSource(idSeed)))
	if err != nil {
		return nil, fmt.Errorf("failed to create keypair")
	}

	h, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, fmt.Errorf("failed to parse address. Expects 'ip:port' format")
	}

	if h == "" {
		h = "127.0.0.1"
	}

	// construct host options
	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/%s/tcp/%s", h, port)),
		libp2p.Identity(priv),
	}

	// create host
	host, err := libp2p.New(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create host > %s", err)
	}

	node := &Node{
		cfg:     config,
		address: util.FullAddressFromHost(host),
		host:    host,
		wg:      sync.WaitGroup{},
		log:     log,
		rSeed:   util.RandBytes(64),
		txPool:  txpool.NewTxPool(config.TxPool.Capacity),
	}

	node.localNode = node
	node.peerManager = NewManager(config, node, node.log)
	node.IP = node.ip()

	log.Info("Opened local database", "Backend", "LevelDB")

	return node, nil
}

// NewRemoteNode creates a Node that represents a remote node
func NewRemoteNode(address ma.Multiaddr, localNode *Node) *Node {
	node := &Node{
		address:   address,
		localNode: localNode,
		remote:    true,
	}
	node.IP = node.ip()
	return node
}

// OpenDB opens the database
func (n *Node) OpenDB() error {
	if n.db != nil {
		return fmt.Errorf("db already open")
	}
	n.db = database.NewGeneralDB(n.cfg.ConfigDir())
	return n.db.Open()
}

// PM returns the peer manager
func (n *Node) PM() *Manager {
	return n.peerManager
}

// IsSame checks if p is the same as node
func (n *Node) IsSame(node *Node) bool {
	return n.StringID() == node.StringID()
}

// DevMode checks whether the node is in dev mode
func (n *Node) DevMode() bool {
	return n.cfg.Node.Dev
}

// IsSameID is like IsSame except it accepts string
func (n *Node) IsSameID(id string) bool {
	return n.StringID() == id
}

// SetLocalNode sets the local peer
func (n *Node) SetLocalNode(node *Node) {
	n.localNode = node
}

// addToPeerStore adds a remote node to the host's peerstore
func (n *Node) addToPeerStore(remote *Node) *Node {
	n.localNode.Peerstore().AddAddr(remote.ID(), remote.GetIP4TCPAddr(), pstore.PermanentAddrTTL)
	return n
}

// newStream creates a stream to a peer
func (n *Node) newStream(ctx context.Context, peerID peer.ID, protocolID string) (inet.Stream, error) {
	return n.Host().NewStream(ctx, peerID, protocol.ID(protocolID))
}

// SetProtocol sets the protocol implementation
func (n *Node) SetProtocol(protoc Protocol) {
	n.protoc = protoc
}

// GetHost returns the node's host
func (n *Node) GetHost() host.Host {
	return n.host
}

// Peerstore returns the Peerstore of the node
func (n *Node) Peerstore() pstore.Peerstore {
	if h := n.Host(); h != nil {
		return h.Peerstore()
	}
	return nil
}

// Host returns the internal host instance
func (n *Node) Host() host.Host {
	return n.host
}

// ID returns the peer id of the host
func (n *Node) ID() peer.ID {
	if n.address == nil {
		return ""
	}

	pid, _ := n.address.ValueForProtocol(ma.P_IPFS)
	id, _ := peer.IDB58Decode(pid)
	return id
}

// StringID is like ID() but returns string
func (n *Node) StringID() string {
	if n.address == nil {
		return ""
	}

	pid, _ := n.address.ValueForProtocol(ma.P_IPFS)
	return pid
}

// ShortID is like IDPretty but shorter
func (n *Node) ShortID() string {
	id := n.StringID()
	if len(id) == 0 {
		return ""
	}
	return id[0:12] + ".." + id[40:52]
}

// Connected checks whether the node is connected to the local node.
// Returns false if node is the local node.
func (n *Node) Connected() bool {
	if n.localNode == nil {
		return false
	}
	return len(n.localNode.host.Network().ConnsToPeer(n.ID())) > 0
}

func (n *Node) isDevMode() bool {
	return n.cfg.Node.Dev
}

// IsKnown checks whether a peer is known to the local node
func (n *Node) IsKnown() bool {
	if n.localNode == nil {
		return false
	}
	return n.localNode.PM().GetKnownPeer(n.StringID()) != nil
}

// PrivKey returns the node's private key
func (n *Node) PrivKey() crypto.PrivKey {
	return n.host.Peerstore().PrivKey(n.host.ID())
}

// PubKey returns the node's public key
func (n *Node) PubKey() crypto.PubKey {
	return n.host.Peerstore().PubKey(n.host.ID())
}

// SetProtocolHandler sets the protocol handler for a specific protocol
func (n *Node) SetProtocolHandler(version string, handler inet.StreamHandler) {
	n.host.SetStreamHandler(protocol.ID(version), handler)
}

// GetMultiAddr returns the full multi address of the node
func (n *Node) GetMultiAddr() string {
	if n.host == nil && !n.remote {
		return ""
	} else if n.remote {
		return n.address.String()
	}
	hostAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", n.host.ID().Pretty()))
	return n.host.Addrs()[0].Encapsulate(hostAddr).String()
}

// GetAddr returns the host and port of the node as "host:port"
func (n *Node) GetAddr() string {
	parts := strings.Split(strings.Trim(n.host.Addrs()[0].String(), "/"), "/")
	return fmt.Sprintf("%s:%s", parts[1], parts[3])
}

// GetIP4TCPAddr returns ip4 and tcp parts of the host's multi address
func (n *Node) GetIP4TCPAddr() ma.Multiaddr {
	ipfsAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", n.ID().Pretty()))
	return n.address.Decapsulate(ipfsAddr)
}

// GetBindAddress returns the bind address
func (n *Node) GetBindAddress() string {
	return n.address.String()
}

// AddBootstrapNodes sets the initial nodes to communicate to
func (n *Node) AddBootstrapNodes(peerAddresses []string, hardcoded bool) error {

	for _, addr := range peerAddresses {

		if !util.IsValidAddr(addr) {
			n.log.Debug("Invalid bootstrap peer address", "PeerAddr", addr)
			continue
		}

		if n.isDevMode() && !util.IsDevAddr(util.GetIPFromAddr(addr)) {
			n.log.Debug("Only local or private address are allowed in dev mode", "Addr", addr)
			continue
		}

		if !n.DevMode() && !util.IsRoutableAddr(addr) {
			n.log.Debug("Invalid bootstrap peer address", "PeerAddr", addr)
			continue
		}

		pAddr, _ := ma.NewMultiaddr(addr)
		rp := NewRemoteNode(pAddr, n)
		rp.isHardcodedSeed = hardcoded
		rp.protoc = n.protoc
		n.peerManager.AddBootstrapPeer(rp)
	}
	return nil
}

// PeersPublicAddr gets all the peers' public address.
// It will ignore any peer whose ID is specified in peerIDsToIgnore
func (n *Node) PeersPublicAddr(peerIDsToIgnore []string) (peerAddrs []ma.Multiaddr) {
	for _, _p := range n.host.Peerstore().Peers() {
		if !funk.Contains(peerIDsToIgnore, _p.Pretty()) {
			if _pAddrs := n.host.Peerstore().Addrs(_p); len(_pAddrs) > 0 {
				peerAddrs = append(peerAddrs, _pAddrs[0])
			}
		}
	}
	return
}

// connectToNode handshake to each bootstrap peer.
// Then send GetAddr message if handshake is successful
func (n *Node) connectToNode(remote *Node) error {
	if n.protoc.SendHandshake(remote) == nil {
		return n.protoc.SendGetAddr([]*Node{remote})
	}
	return nil
}

// Start starts the node.
// Set Tx Pool relay callback
// Send handshake to each bootstrap node.
func (n *Node) Start() {

	n.txPool.OnQueued(n.protoc.RelayTx)

	n.PM().Manage()
	for _, node := range n.PM().bootstrapNodes {
		go n.connectToNode(node)
	}
}

// Wait forces the current thread to wait for the node
func (n *Node) Wait() {
	n.wg.Add(1)
	n.wg.Wait()
}

// Stop stops the node and releases any held resources.
func (n *Node) Stop() {

	if pm := n.PM(); pm != nil {
		pm.Stop()
	}

	if n.db != nil {
		err := n.db.Close()
		if err == nil {
			n.log.Info("Database has been closed")
		} else {
			n.log.Error("failed to close database", "Err", err)
		}
	}

	n.log.Info("Local node has stopped")

	if n.wg != (sync.WaitGroup{}) {
		n.wg.Done()
	}
}

// NodeFromAddr creates a Node from a multiaddr
func (n *Node) NodeFromAddr(addr string, remote bool) (*Node, error) {
	if !util.IsValidAddr(addr) {
		return nil, fmt.Errorf("addr is not valid")
	}
	nAddr, _ := ma.NewMultiaddr(addr)
	return &Node{
		address:   nAddr,
		localNode: n,
		protoc:    n.protoc,
		remote:    remote,
	}, nil
}

// ip returns the IP address
func (n *Node) ip() net.IP {
	addr := n.GetIP4TCPAddr()
	if addr == nil {
		return nil
	}
	ip, _ := addr.ValueForProtocol(ma.P_IP6)
	if ip == "" {
		ip, _ = addr.ValueForProtocol(ma.P_IP4)
	}
	return net.ParseIP(ip)
}

// IsBadTimestamp checks whether the timestamp of the node is bad.
// It is bad when:
// - It has no timestamp
// - The timestamp is 10 minutes in the future or over 3 hours ago
// TODO: Also check of history of failed connection attempts
func (n *Node) IsBadTimestamp() bool {
	if n.Timestamp.IsZero() {
		return true
	}

	now := time.Now()
	if n.Timestamp.After(now.Add(time.Minute*10)) || n.Timestamp.Before(now.Add(-3*time.Hour)) {
		return true
	}

	return false
}

func (n *Node) createTx() error {
	// tx := &wire.NewTransaction(wire.TxTypeRepoCreate, 1, "somebody", n.)
	// return n.txPool.Put()
	return nil
}
