package node

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/olebedev/emitter"

	"github.com/ellcrys/elld/blockchain/common"
	d_crypto "github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/node/histcache"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/wire"

	"github.com/ellcrys/elld/txpool"

	"github.com/ellcrys/elld/util/logger"

	"github.com/ellcrys/elld/config"

	"github.com/thoas/go-funk"

	crypto "github.com/libp2p/go-libp2p-crypto"
	peer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"

	"github.com/ellcrys/elld/util"
	libp2p "github.com/libp2p/go-libp2p"
	host "github.com/libp2p/go-libp2p-host"
	inet "github.com/libp2p/go-libp2p-net"
	protocol "github.com/libp2p/go-libp2p-protocol"
	ma "github.com/multiformats/go-multiaddr"
)

// Node represents a network node
type Node struct {
	mtx                     *sync.RWMutex
	cfg                     *config.EngineConfig    // node config
	address                 ma.Multiaddr            // node multiaddr
	IP                      net.IP                  // node ip
	host                    host.Host               // node libp2p host
	wg                      sync.WaitGroup          // wait group for preventing the main thread from exiting
	localNode               *Node                   // local node
	peerManager             *Manager                // node manager for managing connections to other remote peers
	gProtoc                 types.Gossip            // gossip protocol instance
	remote                  bool                    // remote indicates the node represents a remote peer
	Timestamp               time.Time               // the last time this node was seen/active
	isHardcodedSeed         bool                    // whether the node was hardcoded as a seed
	stopped                 bool                    // flag to tell if node has stopped
	log                     logger.Logger           // node logger
	rSeed                   []byte                  // random 256 bit seed to be used for seed random operations
	db                      elldb.DB                // used to access and modify local database
	signatory               *d_crypto.Key           // signatory address used to get node ID and for signing
	historyCache            *histcache.HistoryCache // Used to track objects and behaviours
	event                   *emitter.Emitter        // Provides access event emitting service
	openTransactionsSession map[string]struct{}     // Holds the id of transactions awaiting endorsement. Protected by mtx.
	transactionsPool        *txpool.TxPool          // the transaction pool for transactions
	txsRelayQueue           *txpool.TxQueue         // stores transactions waiting to be relayed
	bchain                  common.Blockchain       // The blockchain manager
}

// NewNode creates a node instance at the specified port
func newNode(db elldb.DB, config *config.EngineConfig, address string, signatory *d_crypto.Key, log logger.Logger) (*Node, error) {

	if signatory == nil {
		return nil, fmt.Errorf("signatory address required")
	}

	sk, _ := signatory.PrivKey().Marshal()
	priv, err := crypto.UnmarshalPrivateKey(sk)
	if err != nil {
		return nil, err
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
		cfg:       config,
		address:   util.FullAddressFromHost(host),
		host:      host,
		wg:        sync.WaitGroup{},
		log:       log,
		rSeed:     util.RandBytes(64),
		signatory: signatory,
		db:        db,
		event:     &emitter.Emitter{},
		mtx:       &sync.RWMutex{},
		openTransactionsSession: make(map[string]struct{}),
		transactionsPool:        txpool.NewTxPool(config.TxPool.Capacity),
		txsRelayQueue:           txpool.NewQueueNoSort(config.TxPool.Capacity),
	}

	node.localNode = node
	node.peerManager = NewManager(config, node, node.log)
	node.IP = node.ip()

	hc, err := histcache.NewHistoryCache(5000)
	if err != nil {
		return nil, fmt.Errorf("failed to create history cache. %s", err)
	}

	node.historyCache = hc

	log.Info("Opened local database", "Backend", "LevelDB")

	return node, nil
}

// NewNode creates a Node instance
func NewNode(config *config.EngineConfig, address string, signatory *d_crypto.Key, log logger.Logger) (*Node, error) {
	return newNode(nil, config, address, signatory, log)
}

// NewNodeWithDB is like NewNode but it accepts a db instance
func NewNodeWithDB(db elldb.DB, config *config.EngineConfig, address string, signatory *d_crypto.Key, log logger.Logger) (*Node, error) {
	return newNode(db, config, address, signatory, log)
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

// OpenDB opens the database.
// In dev mode, create a namespace and open database file prefixed with the namespace.
func (n *Node) OpenDB() error {

	if n.db != nil {
		return fmt.Errorf("db already open")
	}

	n.db = elldb.NewDB(n.cfg.ConfigDir())
	var namespace string
	if n.DevMode() {
		namespace = n.StringID()
	}

	return n.db.Open(namespace)
}

// DB returns the database instance
func (n *Node) DB() elldb.DB {
	return n.db
}

// GossipProto returns the set protocol
func (n *Node) GossipProto() types.Gossip {
	return n.gProtoc
}

// PM returns the peer manager
func (n *Node) PM() *Manager {
	return n.peerManager
}

// Cfg returns the config object
func (n *Node) Cfg() *config.EngineConfig {
	return n.cfg
}

// History returns the cache holding items (messages etc) we have seen
func (n *Node) History() *histcache.HistoryCache {
	return n.historyCache
}

// IsSame checks if p is the same as node
func (n *Node) IsSame(node types.Engine) bool {
	return n.StringID() == node.StringID()
}

// GetBlockchain returns the blockchain manager
func (n *Node) GetBlockchain() common.Blockchain {
	return n.bchain
}

// SetBlockchain sets the blockchain
func (n *Node) SetBlockchain(bchain common.Blockchain) {
	n.bchain = bchain
}

// IsHardcodedSeed checks whether the node is an hardcoded seed node
func (n *Node) IsHardcodedSeed() bool {
	return n.isHardcodedSeed
}

// SetTimestamp sets the timestamp value
func (n *Node) SetTimestamp(newTime time.Time) {
	n.Timestamp = newTime
}

// DevMode checks whether the node is in dev mode
func (n *Node) DevMode() bool {
	return n.cfg.Node.Dev
}

// IsSameID is like IsSame except it accepts string
func (n *Node) IsSameID(id string) bool {
	return n.StringID() == id
}

// SetEventBus set the event bus used to broadcast events across the engine
func (n *Node) SetEventBus(ee *emitter.Emitter) {
	n.event = ee
}

// SetLocalNode sets the local peer
func (n *Node) SetLocalNode(node *Node) {
	n.localNode = node
}

// canAcceptPeer determines whether we can continue to interact with
// a remote node. This is a good place to check if a remote node
// has been blacklisted etc
func (n *Node) canAcceptPeer(remotePeer *Node) (bool, string) {

	// In dev mode, we cannot interact with a remote peer with a public IP
	if n.isDevMode() && !util.IsDevAddr(remotePeer.IP) {
		return false, "in development mode, we cannot interact with peers with public IP"
	}

	// If the local peer does not know the remotePeer, it cannot interact with it.
	// This does not apply in dev mode.
	if !remotePeer.IsKnown() && !n.isDevMode() {
		return false, "remote peer is unknown"
	}

	return true, ""
}

// addToPeerStore adds a remote node to the host's peerstore
func (n *Node) addToPeerStore(remote types.Engine) *Node {
	n.localNode.Peerstore().AddAddr(remote.ID(), remote.GetIP4TCPAddr(), pstore.PermanentAddrTTL)
	return n
}

// newStream creates a stream to a peer
func (n *Node) newStream(ctx context.Context, peerID peer.ID, protocolID string) (inet.Stream, error) {
	return n.Host().NewStream(ctx, peerID, protocol.ID(protocolID))
}

// SetGossipProtocol sets the gossip protocol implementation
func (n *Node) SetGossipProtocol(protoc types.Gossip) {
	n.gProtoc = protoc
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
		rp.gProtoc = n.gProtoc
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
func (n *Node) connectToNode(remote types.Engine) error {
	if n.gProtoc.SendHandshake(remote) == nil {
		return n.gProtoc.SendGetAddr([]types.Engine{remote})
	}
	return nil
}

// relayTx continuously relays transactions in the tx relay queue
func (n *Node) relayTx() {
	for !n.stopped {
		q := n.GetTxRelayQueue()
		if q.Size() == 0 {
			time.Sleep(5 * time.Second)
			continue
		}

		tx := q.First()
		n.gProtoc.RelayTx(tx, n.peerManager.GetActivePeers(0))
	}
}

// GetTimestamp returns the timestamp
func (n *Node) GetTimestamp() time.Time {
	return n.Timestamp
}

// Start starts the node.
// - Start the peer manager
// - Send handshake to each bootstrap node.
// - Set callback to queue transactions for relaying
func (n *Node) Start() {

	n.PM().Manage()

	for _, node := range n.PM().bootstrapNodes {
		go n.connectToNode(node)
	}

	// before a transaction is added to the tx pool, it must be successfully
	// added to the tx relay queue.
	n.GetTxPool().BeforeAppend(func(tx *wire.Transaction) error {
		if !n.GetTxRelayQueue().Append(tx) {
			return txpool.ErrQueueFull
		}
		return nil
	})

	go n.relayTx()
}

// Wait forces the current thread to wait for the node
func (n *Node) Wait() {
	n.wg.Add(1)
	n.wg.Wait()
}

// Stop stops the node and releases any held resources.
func (n *Node) Stop() {

	n.stopped = true

	if pm := n.PM(); pm != nil {
		pm.Stop()
	}

	if n.host != nil {
		n.host.Close()
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
		gProtoc:   n.gProtoc,
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

// GetTxRelayQueue returns the transaction relay queue
func (n *Node) GetTxRelayQueue() *txpool.TxQueue {
	return n.txsRelayQueue
}

// GetTxPool returns the unsigned transaction pool
func (n *Node) GetTxPool() *txpool.TxPool {
	return n.transactionsPool
}
