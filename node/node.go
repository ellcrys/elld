// Package node provides an engine that combines
// other components to make up the client.
package node

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/ellcrys/elld/node/peermanager"

	"github.com/shopspring/decimal"

	"github.com/olebedev/emitter"

	"github.com/ellcrys/elld/blockchain"
	"github.com/ellcrys/elld/blockchain/txpool"
	d_crypto "github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/node/gossip"
	"github.com/ellcrys/elld/params"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"

	"github.com/ellcrys/elld/util/cache"
	"github.com/ellcrys/elld/util/logger"
	"github.com/ellcrys/elld/util/queue"

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
	syncMtx             sync.RWMutex
	cfg                 *config.EngineConfig // node config
	address             util.NodeAddr        // node address
	IP                  net.IP               // node ip
	host                host.Host            // node libp2p host
	wg                  sync.WaitGroup       // wait group for preventing the main thread from exiting
	localNode           *Node                // local node
	peerManager         *peermanager.Manager // node manager for managing connections to other remote peers
	gProtoc             core.Gossip          // gossip protocol instance
	remote              bool                 // remote indicates the node represents a remote peer
	lastSeen            time.Time            // the last time this node was seen
	createdAt           time.Time            // the first time this node was seen/added
	stopped             bool                 // flag to tell if node has stopped
	log                 logger.Logger        // node logger
	txsPool             *txpool.TxPool
	rSeed               []byte              // random 256 bit seed to be used for seed random operations
	db                  elldb.DB            // used to access and modify local database
	signatory           *d_crypto.Key       // signatory address used to get node ID and for signing
	history             *cache.Cache        // Used to track things we want to remember
	event               *emitter.Emitter    // Provides access event emitting service
	txsRelayQueue       *txpool.TxContainer // stores transactions waiting to be relayed
	bChain              types.Blockchain    // The blockchain manager
	blockHashQueue      *queue.Queue        // Contains headers collected during block syncing
	bestRemoteBlockInfo *core.BestBlockInfo // Holds information about the best known block heard from peers
	syncing             bool                // Indicates the process of syncing the blockchain with peers
	inbound             bool                // Indicates this that this node initiated the connection with the local node
	intros              *cache.Cache        // Stores peer ids received in wire.Intro messages
	tickerDone          chan bool
	hardcodedPeers      map[string]struct{}
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
		syncMtx:        sync.RWMutex{},
		cfg:            cfg,
		address:        util.AddressFromHost(host),
		host:           host,
		wg:             sync.WaitGroup{},
		log:            log,
		rSeed:          util.RandBytes(64),
		signatory:      coinbase,
		db:             db,
		event:          &emitter.Emitter{},
		txsRelayQueue:  txpool.NewQueueNoSort(cfg.TxPool.Capacity),
		blockHashQueue: queue.New(),
		history:        cache.NewActiveCache(5000),
		intros:         cache.NewActiveCache(50000),
		tickerDone:     make(chan bool),
		createdAt:      time.Now(),
		hardcodedPeers: make(map[string]struct{}),
	}

	node.localNode = node
	node.peerManager = peermanager.NewManager(cfg, node, node.log)
	node.IP = node.ip()

	g := gossip.NewGossip(node, log)
	g.SetPeerManager(node.peerManager)
	node.SetGossipManager(g)
	node.SetProtocolHandler(config.Versions.Handshake, g.Handle(g.OnHandshake))
	node.SetProtocolHandler(config.Versions.Ping, g.Handle(g.OnPing))
	node.SetProtocolHandler(config.Versions.GetAddr, g.Handle(g.OnGetAddr))
	node.SetProtocolHandler(config.Versions.Addr, g.Handle(g.OnAddr))
	node.SetProtocolHandler(config.Versions.Intro, g.Handle(g.OnIntro))
	node.SetProtocolHandler(config.Versions.Tx, g.Handle(g.OnTx))
	node.SetProtocolHandler(config.Versions.BlockBody, g.Handle(g.OnBlockBody))
	node.SetProtocolHandler(config.Versions.RequestBlock, g.Handle(g.OnRequestBlock))
	node.SetProtocolHandler(config.Versions.GetBlockHashes, g.Handle(g.OnGetBlockHashes))
	node.SetProtocolHandler(config.Versions.GetBlockBodies, g.Handle(g.OnGetBlockBodies))

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
		syncMtx:        sync.RWMutex{},
		localNode:      n,
		gProtoc:        n.gProtoc,
		createdAt:      time.Now(),
		lastSeen:       time.Now(),
		hardcodedPeers: make(map[string]struct{}),
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
		syncMtx:        sync.RWMutex{},
		localNode:      localNode,
		createdAt:      time.Now(),
		lastSeen:       time.Now(),
		hardcodedPeers: make(map[string]struct{}),
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
		syncMtx:        sync.RWMutex{},
		hardcodedPeers: make(map[string]struct{}),
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

	n.db = elldb.NewDB(n.cfg.DataDir())
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

// SetSyncing sets the sync status
func (n *Node) SetSyncing(syncing bool) {
	n.syncMtx.Lock()
	defer n.syncMtx.Unlock()
	n.syncing = syncing
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

// UpdateSyncInfo sets a given remote best
// block info as the best known remote block
// only when it is better than the local best block.
// Using this information, it can tell when syncing
// has stopped and as such, updates the syncing status.
func (n *Node) UpdateSyncInfo(bi *core.BestBlockInfo) {
	n.syncMtx.Lock()
	defer n.syncMtx.Unlock()

	if bi == nil {
		goto compare
	}

	// If we have not seen block info of any
	// remote peer, we can set to the bi.
	// But if the current best remote block
	// has a lower total difficulty than the
	// latest, we update to the latests
	if n.bestRemoteBlockInfo == nil {
		n.bestRemoteBlockInfo = bi
	} else if n.bestRemoteBlockInfo.BestBlockTotalDifficulty.
		Cmp(bi.BestBlockTotalDifficulty) == -1 {
		n.bestRemoteBlockInfo = bi
	}

compare:

	// Do nothing if we still don't know the best remote
	// block info. This means the local blockchain is still
	// considered the better chain

	if n.bestRemoteBlockInfo == nil {
		return
	}

	// We need to compare the local best block
	// with the best remote block. If the local
	// block is equal or better, we set syncing status
	// to false
	localBestBlock, _ := n.GetBlockchain().ChainReader().Current()
	if localBestBlock.GetHeader().GetTotalDifficulty().
		Cmp(n.bestRemoteBlockInfo.BestBlockTotalDifficulty) > -1 {
		n.syncing = false
	}
}

// GetSyncStateInfo generates status and progress
// information about the current blockchain sync operation
func (n *Node) GetSyncStateInfo() *core.SyncStateInfo {

	// No need to compute when we are
	// not currently syncing
	if !n.isSyncing() {
		return nil
	}

	if n.bestRemoteBlockInfo == nil {
		return nil
	}

	var syncState = &core.SyncStateInfo{}

	// Get the current local best chain
	localBestBlock, _ := n.GetBlockchain().ChainReader().Current()
	syncState.TargetTD = n.bestRemoteBlockInfo.BestBlockTotalDifficulty
	syncState.TargetChainHeight = n.bestRemoteBlockInfo.BestBlockNumber
	syncState.CurrentTD = localBestBlock.GetHeader().GetTotalDifficulty()
	syncState.CurrentChainHeight = localBestBlock.GetNumber()

	// compute progress percentage based
	// on block height differences
	pct := float64(100) * (float64(syncState.CurrentChainHeight) /
		float64(syncState.TargetChainHeight))
	syncState.ProgressPercent, _ = decimal.NewFromFloat(pct).
		Round(1).Float64()

	return syncState
}

// isSyncing checks whether block
// synchronization is ongoing
func (n *Node) isSyncing() bool {
	n.syncMtx.RLock()
	defer n.syncMtx.RUnlock()
	return n.syncing
}

// Gossip returns the set protocol
func (n *Node) Gossip() core.Gossip {
	return n.gProtoc
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
func (n *Node) SetBlockchain(bchain types.Blockchain) {
	n.bChain = bchain
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

// SetLocalNode sets the node as the
// local node to n which makes n the "remote" node
func (n *Node) SetLocalNode(node *Node) {
	n.localNode = node
}

// CountIntros counts the number of
// intros received
func (n *Node) CountIntros() int {
	return n.intros.Len()
}

// GetIntros returns the cache containing received intros
func (n *Node) GetIntros() *cache.Cache {
	return n.intros
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
func (n *Node) SetGossipProtocol(protoc *gossip.Gossip) {
	n.gProtoc = protoc
}

// GetHost returns the node's host
func (n *Node) GetHost() host.Host {
	return n.host
}

// SetHost sets the host
func (n *Node) SetHost(h host.Host) {
	n.host = h
}

// GetBlockHashQueue returns the block hash queue
func (n *Node) GetBlockHashQueue() *queue.Queue {
	return n.blockHashQueue
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
		rp.SetHardcodedState(hardcoded)
		rp.SetGossipManager(n.gProtoc)
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

// SetHardcodedState sets the hardcoded seed state
// of the engine.
func (n *Node) SetHardcodedState(v bool) {
	if v {
		n.localNode.addHardcodedPeer(n)
		return
	}
	n.localNode.removeHardcodedPeer(n)
}

// SetGossipManager sets the gossip manager
func (n *Node) SetGossipManager(m core.Gossip) {
	n.gProtoc = m
}

// relayTx continuously relays transactions
// in the tx relay queue
func (n *Node) relayTx() {
	ticker := time.NewTicker(3 * time.Second)
	for {
		select {
		case <-ticker.C:
			q := n.GetTxRelayQueue()
			if q.Size() == 0 {
				continue
			}
			tx := q.First()
			go n.Gossip().RelayTx(tx, n.peerManager.GetActivePeers(0))
		case <-n.tickerDone:
			ticker.Stop()
			return
		}
	}
}

// Start starts the node.
func (n *Node) Start() {

	// Start the peer manager
	n.PM().Manage()

	// Attempt to connect to peers
	for _, node := range n.PM().GetActivePeers(0) {
		go n.peerManager.ConnectToPeer(node.StringID())
	}

	// Start the sub-routine that
	// relays transactions
	go n.relayTx()

	// Handle incoming events
	go n.handleEvents()

	// start a block body requester
	// workers
	for i := 0; i < params.NumBlockBodiesRequesters; i++ {
		go n.ProcessBlockHashes()
	}
}

// relayBlock attempts to relay non-genesis
//  a block to active peers.
func (n *Node) relayBlock(block types.Block) {
	if block.GetNumber() > 1 {
		n.Gossip().RelayBlock(block, n.peerManager.GetConnectedPeers())
	}
}

func (n *Node) handleNewBlockEvent() {
	for {
		select {
		case evt := <-n.event.Once(core.EventNewBlock):
			n.relayBlock(evt.Args[0].(types.Block))
		}
	}
}

func (n *Node) handleNewTransactionEvent() {
	for {
		select {
		case evt := <-n.event.Once(core.EventNewTransaction):
			if !n.GetTxRelayQueue().Add(evt.Args[0].(types.Transaction)) {
				n.log.Debug("Failed to add transaction to relay queue.",
					"Err", "Capacity reached")
			}
		}
	}
}

func (n *Node) handleOrphanBlockEvent() {
	for {
		select {
		case evt := <-n.event.Once(core.EventOrphanBlock):
			// We need to request the parent block from the
			// peer who sent it to us (a.k.a broadcaster)
			orphanBlock := evt.Args[0].(*core.Block)
			parentHash := orphanBlock.GetHeader().GetParentHash()
			n.log.Debug("Requesting orphan parent block from broadcaster",
				"BlockNo", orphanBlock.GetNumber(),
				"ParentBlockHash", parentHash.SS())
			n.gProtoc.RequestBlock(orphanBlock.Broadcaster, parentHash)
		}
	}
}

func (n *Node) handleAbortedMinerBlockEvent() {
	for {
		select {
		case evt := <-n.event.Once(core.EventOrphanBlock):
			// handle core.EventMinerProposedBlockAborted
			// listens for aborted miner blocks and attempt
			// to re-add the transactions to the pool.
			abortedBlock := evt.Args[0].(*core.Block)
			n.log.Debug("Attempting to re-add transactions "+
				"in aborted miner block",
				"NumTx", len(abortedBlock.Transactions), "BlockNo", abortedBlock.GetNumber())
			for _, tx := range abortedBlock.Transactions {
				if err := n.AddTransaction(tx); err != nil {
					n.log.Debug("failed to re-add transaction",
						"Err", err.Error())
				}
			}
		}
	}
}

func (n *Node) handleEvents() {
	go n.handleNewBlockEvent()
	go n.handleNewTransactionEvent()
	go n.handleOrphanBlockEvent()
	go n.handleAbortedMinerBlockEvent()
}

// ProcessBlockHashes collects hashes and request for their
// block bodies from the initial broadcaster if the headers.
func (n *Node) ProcessBlockHashes() {

	ticketDur := 5 * time.Second
	if n.TestMode() {
		ticketDur = 1 * time.Nanosecond
	}

	ticker := time.NewTicker(ticketDur)
	for {
		select {
		case <-ticker.C:

			if n.blockHashQueue.Empty() {
				continue
			}

			hashes := []util.Hash{}
			var broadcaster core.Engine
			otherBlockHashes := []queue.Item{}

			// Collect hash of headers sent by a
			// particular broadcaster. Temporarily
			// keep the others in a cache to be added back
			// in the queue when we have collected some hashes
			n.mtx.Lock()
			for !n.blockHashQueue.Empty() && int64(len(hashes)) <
				params.MaxGetBlockBodiesHashes {

				bh := n.blockHashQueue.Head()
				if bh == nil {
					continue
				}

				if broadcaster != nil && bh.(*core.BlockHash).Broadcaster.StringID() !=
					broadcaster.StringID() {
					otherBlockHashes = append(otherBlockHashes, bh)
					continue
				}

				hashes = append(hashes, bh.(*core.BlockHash).Hash)
				broadcaster = bh.(*core.BlockHash).Broadcaster
			}

			// append the others that were not selected
			// back to the block hash queue
			for _, bh := range otherBlockHashes {
				n.blockHashQueue.Append(bh)
			}

			n.mtx.Unlock()

			// send block body request
			if len(hashes) > 0 {
				n.gProtoc.SendGetBlockBodies(broadcaster, hashes)
			}

		case <-n.tickerDone:
			ticker.Stop()
			return
		}
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

// GetTxRelayQueue returns the transaction relay queue
func (n *Node) GetTxRelayQueue() *txpool.TxContainer {
	return n.txsRelayQueue
}

// GetTxPool returns the unsigned transaction pool
func (n *Node) GetTxPool() types.TxPool {
	return n.txsPool
}

// SetTxsPool sets the transaction pool
func (n *Node) SetTxsPool(txp *txpool.TxPool) {
	n.txsPool = txp
}

// AddTransaction validates and adds a
// transaction to the transaction pool.
func (n *Node) AddTransaction(tx types.Transaction) error {

	txValidator := blockchain.NewTxValidator(tx, n.GetTxPool(), n.GetBlockchain())
	if errs := txValidator.Validate(); len(errs) > 0 {
		return errs[0]
	}

	return n.GetTxPool().Put(tx)
}
