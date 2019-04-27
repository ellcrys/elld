// Package node provides an engine that combines
// other components to make up the client.
package node

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/dustinkirkland/golang-petname"

	"github.com/ellcrys/elld/node/peermanager"

	"github.com/olebedev/emitter"

	"github.com/ellcrys/elld/blockchain/txpool"
	d_crypto "github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/node/gossip"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"

	"github.com/ellcrys/elld/util/cache"
	"github.com/ellcrys/elld/util/logger"

	"github.com/ellcrys/elld/config"

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
	mtx                 sync.RWMutex
	cfg                 *config.EngineConfig // node config
	address             util.NodeAddr        // node address
	IP                  net.IP               // node ip
	host                host.Host            // node libp2p host
	wg                  sync.WaitGroup       // wait group for preventing the main thread from exiting
	localNode           *Node                // local node
	peerManager         *peermanager.Manager // node manager for managing connections to other remote peers
	gossipMgr           core.Gossip          // gossip protocol instance
	remote              bool                 // remote indicates the node represents a remote peer
	lastSeen            time.Time            // the last time this node was seen
	createdAt           time.Time            // the first time this node was seen/added
	stopped             bool                 // flag to tell if node has stopped
	log                 logger.Logger        // node logger
	txsPool             *txpool.TxPool
	rSeed               []byte              // random 256 bit seed to be used for seed random operations
	db                  elldb.DB            // used to access and modify local database
	coinbase            *d_crypto.Key       // signatory address used to get node ID and for signing
	history             *cache.Cache        // Used to track things we want to remember
	event               *emitter.Emitter    // Provides access event emitting service
	bChain              types.Blockchain    // The blockchain manager
	bestRemoteBlockInfo *core.BestBlockInfo // Holds information about the best known block heard from peers
	inbound             bool                // Indicates this that this node initiated the connection with the local node
	blockManager        *BlockManager       // Block manager for handling block events
	txManager           *TxManager          // Transaction manager for handling transaction events
	noNet               bool                // Indicates whether the host is listening for connections
	Name                string              // Random name for this node
	hardcodedPeers      map[string]struct{} // A collection of seed peers that were manually provided
	syncMode            core.SyncMode
}

// NewNode creates a node instance at the specified port
func newNode(db elldb.DB, cfg *config.EngineConfig, address string,
	coinbase *d_crypto.Key, log logger.Logger) (*Node, error) {

	if coinbase == nil {
		return nil, fmt.Errorf("signatory address required")
	}

	sk, _ := coinbase.PrivKey().Marshal()
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

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/%s/tcp/%s", h, port)),
		libp2p.Identity(priv),
	}

	host, err := libp2p.New(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create host > %s", err)
	}

	node := &Node{
		mtx:            sync.RWMutex{},
		cfg:            cfg,
		address:        util.AddressFromHost(host),
		host:           host,
		wg:             sync.WaitGroup{},
		log:            log,
		rSeed:          util.RandBytes(64),
		coinbase:       coinbase,
		db:             db,
		event:          &emitter.Emitter{},
		history:        cache.NewActiveCache(5000),
		createdAt:      time.Now(),
		hardcodedPeers: make(map[string]struct{}),
		Name:           petname.Generate(3, "-"),
		syncMode:       NewDefaultSyncMode(false),
	}

	node.localNode = node
	node.peerManager = peermanager.NewManager(cfg, node, node.log)
	node.IP = node.ip()

	g := gossip.NewGossip(node, log)
	g.SetPeerManager(node.peerManager)
	node.SetGossipManager(g)
	node.SetProtocolHandler(config.GetVersions().Handshake, g.Handle(g.OnHandshake))
	node.SetProtocolHandler(config.GetVersions().Ping, g.Handle(g.OnPing))
	node.SetProtocolHandler(config.GetVersions().GetAddr, g.Handle(g.OnGetAddr))
	node.SetProtocolHandler(config.GetVersions().Addr, g.Handle(g.OnAddr))
	node.SetProtocolHandler(config.GetVersions().Tx, g.Handle(g.OnTx))
	node.SetProtocolHandler(config.GetVersions().BlockInfo, g.Handle(g.OnBlockInfo))
	node.SetProtocolHandler(config.GetVersions().BlockBody, g.Handle(g.OnBlockBody))
	node.SetProtocolHandler(config.GetVersions().RequestBlock, g.Handle(g.OnRequestBlock))
	node.SetProtocolHandler(config.GetVersions().GetBlockHashes, g.Handle(g.OnGetBlockHashes))
	node.SetProtocolHandler(config.GetVersions().GetBlockBodies, g.Handle(g.OnGetBlockBodies))

	log.Info("Opened local database", "Backend", "LevelDB")

	return node, nil
}

// GetListenAddresses gets the address at which the node listens
func (n *Node) GetListenAddresses() (addrs []util.NodeAddr) {
	lAddrs, _ := n.host.Network().InterfaceListenAddresses()
	for _, addr := range lAddrs {
		ipfsPart := fmt.Sprintf("/ipfs/%s", n.host.ID().Pretty())
		hostAddr, _ := ma.NewMultiaddr(ipfsPart)
		fullAddr := addr.Encapsulate(hostAddr).String()
		addrs = append(addrs, util.NodeAddr(fullAddr))
	}
	return
}

// NewNode creates a Node instance
func NewNode(config *config.EngineConfig, address string,
	signatory *d_crypto.Key, log logger.Logger) (*Node, error) {
	return newNode(nil, config, address, signatory, log)
}

// NewNodeWithDB is like NewNode but it accepts a db instance
func NewNodeWithDB(db elldb.DB, config *config.EngineConfig, address string,
	signatory *d_crypto.Key, log logger.Logger) (*Node, error) {
	return newNode(db, config, address, signatory, log)
}

// NewRemoteNode creates a Node that represents a remote node
func (n *Node) NewRemoteNode(address util.NodeAddr) core.Engine {
	node := &Node{
		address:        address,
		remote:         true,
		mtx:            sync.RWMutex{},
		localNode:      n,
		gossipMgr:      n.gossipMgr,
		createdAt:      time.Now(),
		lastSeen:       time.Now(),
		hardcodedPeers: make(map[string]struct{}),
		syncMode:       NewDefaultSyncMode(false),
	}
	node.IP = node.ip()
	return node
}

// NewRemoteNodeFromMultiAddr is like NewRemoteNode
// excepts it accepts a Multiaddr
func NewRemoteNodeFromMultiAddr(address ma.Multiaddr, localNode *Node) *Node {
	node := &Node{
		address:        util.NodeAddr(address.String()),
		remote:         true,
		mtx:            sync.RWMutex{},
		localNode:      localNode,
		createdAt:      time.Now(),
		lastSeen:       time.Now(),
		hardcodedPeers: make(map[string]struct{}),
		syncMode:       NewDefaultSyncMode(false),
	}
	node.IP = node.ip()
	return node
}

// NewAlmostEmptyNode returns a node with
// almost all its field uninitialized.
func NewAlmostEmptyNode() *Node {
	return &Node{
		createdAt:      time.Now(),
		mtx:            sync.RWMutex{},
		hardcodedPeers: make(map[string]struct{}),
		syncMode:       NewDefaultSyncMode(false),
	}
}

// NewTestNodeWithAddress returns a node with
// with the given address set
func NewTestNodeWithAddress(address ma.Multiaddr) *Node {
	return &Node{
		createdAt:      time.Now(),
		mtx:            sync.RWMutex{},
		address:        util.NodeAddr(address.String()),
		hardcodedPeers: make(map[string]struct{}),
		syncMode:       NewDefaultSyncMode(false),
	}
}

// OpenDB opens the database.
// In dev mode, create a namespace
// and open database file prefixed
// with the namespace.
func (n *Node) OpenDB() error {

	if n.db != nil {
		return fmt.Errorf("db already open")
	}

	n.db = elldb.NewDB(n.cfg.NetDataDir())
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

// DisableNetwork prevents incoming or outbound connections
func (n *Node) DisableNetwork() {
	n.noNet = true
}

// IsNetworkDisabled checks whether networking is disabled
func (n *Node) IsNetworkDisabled() bool {
	return n.noNet
}

// SetSyncMode sets the node's sync mode
func (n *Node) SetSyncMode(mode core.SyncMode) {
	n.syncMode = mode
}

// GetSyncMode returns the sync mode
func (n *Node) GetSyncMode() core.SyncMode {
	return n.syncMode
}

// GetName returns the pet name of the node
func (n *Node) GetName() string {
	return n.Name
}

// SetName sets the name of the node
func (n *Node) SetName(name string) {
	n.Name = name
}

// SetCfg sets the node's config
func (n *Node) SetCfg(cfg *config.EngineConfig) {
	*n.cfg = *cfg
}

// GetCfg returns the config
func (n *Node) GetCfg() *config.EngineConfig {
	return n.cfg
}

// SetInbound set the connection as inbound or not
func (n *Node) SetInbound(v bool) {
	n.inbound = v
}

// IsInbound checks whether the connection is inbound
func (n *Node) IsInbound() bool {
	return n.inbound
}

// Gossip returns the set protocol
func (n *Node) Gossip() core.Gossip {
	return n.gossipMgr
}

// PM returns the peer manager
func (n *Node) PM() *peermanager.Manager {
	return n.peerManager
}

// GetHistory return a cache for
// holding arbitrary objects we
// want to keep track of
func (n *Node) GetHistory() *cache.Cache {
	return n.history
}

// IsSame checks if p is the same as node
func (n *Node) IsSame(node core.Engine) bool {
	return n.StringID() == node.StringID()
}

// GetBlockchain returns the
// blockchain manager
func (n *Node) GetBlockchain() types.Blockchain {
	return n.bChain
}

// SetBlockchain sets the blockchain
func (n *Node) SetBlockchain(bChain types.Blockchain) {
	n.bChain = bChain
}

// IsHardcodedSeed checks whether
// the node is an hardcoded seed node
func (n *Node) IsHardcodedSeed() bool {
	n.mtx.RLock()
	defer n.mtx.RUnlock()
	_, ok := n.localNode.hardcodedPeers[n.StringID()]
	return ok
}

// SetLastSeen sets the timestamp value
func (n *Node) SetLastSeen(newTime time.Time) {
	n.mtx.Lock()
	defer n.mtx.Unlock()
	n.lastSeen = newTime
}

// CreatedAt returns the node's time
// of creation
func (n *Node) CreatedAt() time.Time {
	n.mtx.RLock()
	defer n.mtx.RUnlock()
	return n.createdAt
}

// SetCreatedAt sets the time the
// node was created
func (n *Node) SetCreatedAt(t time.Time) {
	n.mtx.Lock()
	defer n.mtx.Unlock()
	n.createdAt = t
}

// GetEventEmitter gets the event emitter
func (n *Node) GetEventEmitter() *emitter.Emitter {
	n.mtx.RLock()
	defer n.mtx.RUnlock()
	return n.event
}

// GetLastSeen gets the nodes timestamp
func (n *Node) GetLastSeen() time.Time {
	n.mtx.RLock()
	defer n.mtx.RUnlock()
	return n.lastSeen
}

// DevMode checks whether the
// node is in dev mode
func (n *Node) DevMode() bool {
	return n.cfg.Node.Mode == config.ModeDev
}

// TestMode checks whether the
// node is in test mode
func (n *Node) TestMode() bool {
	return n.cfg.Node.Mode == config.ModeTest
}

// ProdMode checks whether the
// node is in production mode
func (n *Node) ProdMode() bool {
	return n.cfg.Node.Mode == config.ModeProd
}

// IsSameID is like IsSame except
// it accepts string
func (n *Node) IsSameID(id string) bool {
	return n.StringID() == id
}

// SetEventEmitter set the event bus
// used to broadcast events across
// the engine
func (n *Node) SetEventEmitter(e *emitter.Emitter) {
	n.event = e
}

// SetBlockManager sets the block manager
func (n *Node) SetBlockManager(bm *BlockManager) {
	n.blockManager = bm
}

// SetTxManager sets the transaction manager
func (n *Node) SetTxManager(tm *TxManager) {
	n.txManager = tm
}

// SetLocalNode sets the node as the
// local node to n which makes n the "remote" node
func (n *Node) SetLocalNode(node *Node) {
	n.localNode = node
}

// AddToPeerStore adds the ID of the engine
// to the peerstore
func (n *Node) AddToPeerStore(node core.Engine) core.Engine {
	addr := node.GetAddress()
	n.localNode.Peerstore().AddAddr(node.ID(),
		addr.DecapIPFS(),
		pstore.PermanentAddrTTL)
	return n
}

// SetGossipProtocol sets the
// gossip protocol implementation
func (n *Node) SetGossipProtocol(mgr *gossip.Manager) {
	n.gossipMgr = mgr
}

// GetHost returns the node's host
func (n *Node) GetHost() host.Host {
	return n.host
}

// SetHost sets the host
func (n *Node) SetHost(h host.Host) {
	n.host = h
}

// Peerstore returns the Peerstore
// of the node
func (n *Node) Peerstore() pstore.Peerstore {
	if h := n.GetHost(); h != nil {
		return h.Peerstore()
	}
	return nil
}

// ID returns the peer id of the host
func (n *Node) ID() peer.ID {
	if n.address == "" {
		return peer.ID("")
	}
	return n.address.ID()
}

// StringID is like ID() but
// it returns string
func (n *Node) StringID() string {
	if n.address == "" {
		return ""
	}
	return n.ID().Pretty()
}

// ShortID is like IDPretty but shorter
func (n *Node) ShortID() string {
	if n.address == "" {
		return ""
	}
	id := n.StringID()
	return id[0:12] + ".." + id[40:52]
}

// Connected checks whether the node
// is connected to the local node.
// Returns false if node is the local node.
func (n *Node) Connected() bool {
	return len(n.localNode.host.Network().ConnsToPeer(n.ID())) > 0
}

// Connect connects n to
func (n *Node) Connect(rn core.Engine) error {
	h := n.GetHost()
	rnH := rn.GetHost()
	h.Peerstore().AddAddr(rnH.ID(), rn.GetAddress().DecapIPFS(), pstore.TempAddrTTL)
	return h.Connect(context.TODO(), h.Peerstore().PeerInfo(rn.ID()))
}

// PrivKey returns the node's private key
func (n *Node) PrivKey() crypto.PrivKey {
	return n.host.Peerstore().PrivKey(n.host.ID())
}

// PubKey returns the node's public key
func (n *Node) PubKey() crypto.PubKey {
	return n.host.Peerstore().PubKey(n.host.ID())
}

// SetProtocolHandler sets the protocol
// handler for a specific protocol
func (n *Node) SetProtocolHandler(version string,
	handler inet.StreamHandler) {
	n.host.SetStreamHandler(protocol.ID(version), handler)
}

// GetAddress returns the node's address
func (n *Node) GetAddress() util.NodeAddr {
	return n.address
}

// AddAddresses adds addresses which the engine can
// establish connections to.
func (n *Node) AddAddresses(connStrings []string, hardcoded bool) error {

	for _, connStr := range connStrings {

		// Resolve the address in the connection string
		// to an IP if it is currently a domain name
		connStr, err := util.ValidateAndResolveConnString(connStr)
		if err != nil {
			n.log.Warn("Could not add or resolve connection string", "Err", err)
			continue
		}

		addr := util.AddressFromConnString(connStr)

		// In production mode, only routable
		// addresses are allowed
		if n.ProdMode() && !addr.IsRoutable() {
			n.log.Warn("address is not routable", "Address", connStr)
			continue
		}

		// Convert the connection string to a valid
		// IPFS Multiaddr format
		rp := n.NewRemoteNode(addr)
		rp.SetHardcoded(hardcoded)
		rp.SetGossipManager(n.gossipMgr)
		n.peerManager.AddPeer(rp)
	}

	return nil
}

// addHardcodedPeer adds a peer to the hardcoded peer
// index maintained by this engine
func (n *Node) addHardcodedPeer(peer core.Engine) {
	n.mtx.Lock()
	n.hardcodedPeers[peer.StringID()] = struct{}{}
	n.mtx.Unlock()
}

// RemoveHardcodedPeer removes a peer from the hardcoded
// peer index maintained by this engine
func (n *Node) removeHardcodedPeer(peer core.Engine) {
	n.mtx.Lock()
	delete(n.hardcodedPeers, peer.StringID())
	n.mtx.Unlock()
}

// SetHardcoded sets the hardcoded seed state
// of the engine.
func (n *Node) SetHardcoded(v bool) {
	if v {
		n.localNode.addHardcodedPeer(n)
		return
	}
	n.localNode.removeHardcodedPeer(n)
}

// SetGossipManager sets the gossip manager
func (n *Node) SetGossipManager(m core.Gossip) {
	n.gossipMgr = m
}

// Start starts the node.
func (n *Node) Start() {

	// Start the peer manager
	n.PM().Manage()

	if n.noNet {
		return
	}

	// Attempt to connect to peers
	for _, node := range n.PM().GetActivePeers(0) {
		go n.peerManager.ConnectToPeer(node.StringID())
	}
}

// Wait forces the current thread to wait for the node
func (n *Node) Wait() {
	n.wg.Add(1)
	n.wg.Wait()
}

// HasStopped checks whether the node has stopped
func (n *Node) HasStopped() bool {
	n.mtx.RLock()
	defer n.mtx.RUnlock()
	return n.stopped
}

// Stop stops the node and releases any held resources.
func (n *Node) Stop() {

	n.mtx.Lock()
	n.stopped = true
	n.mtx.Unlock()

	// stop the peer manager
	// and its managed routines.
	if pm := n.PM(); pm != nil {
		pm.Stop()
	}

	// Shut down the host
	if n.host != nil {
		n.host.Close()
	}

	if n.db != nil {

		// Wait a few seconds for active
		// operations to complete before
		// closing the database
		time.Sleep(2 * time.Second)

		err := n.db.Close()
		if err != nil {
			n.log.Error("Failed to close database", "Err", err)
		} else {
			n.log.Info("Database has been closed")
		}
	}

	n.log.Info("Elld has stopped")

	if n.wg != (sync.WaitGroup{}) {
		n.wg.Done()
	}
}

// ip returns the IP address
func (n *Node) ip() net.IP {
	return n.address.IP()
}

// GetTxPool returns the unsigned transaction pool
func (n *Node) GetTxPool() types.TxPool {
	return n.txsPool
}

// SetTxsPool sets the transaction pool
func (n *Node) SetTxsPool(txp *txpool.TxPool) {
	n.txsPool = txp
}
