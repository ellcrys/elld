package node

import (
	"context"
	"fmt"
	"math/big"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/shopspring/decimal"

	"gopkg.in/oleiade/lane.v1"

	"github.com/olebedev/emitter"

	"github.com/ellcrys/elld/blockchain/txpool"
	d_crypto "github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/params"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/types/core/objects"

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

// BestBlockInfo represent best block
// heard by the engine from other peers
type BestBlockInfo struct {
	BestBlockHash            util.Hash
	BestBlockTotalDifficulty *big.Int
	BestBlockNumber          uint64
}

// SyncStateInfo describes the current state
// and progress of ongoing blockchain synchronization
type SyncStateInfo struct {
	TargetTD           *big.Int `json:"targetTotalDifficulty"`
	TargetChainHeight  uint64   `json:"targetChainHeight" msgpack:"targetChainHeight"`
	CurrentTD          *big.Int `json:"currentTotalDifficulty" msgpack:"currentTotalDifficulty"`
	CurrentChainHeight uint64   `json:"currentChainHeight" msgpack:"currentChainHeight"`
	ProgressPercent    float64  `json:"progressPercent" msgpack:"progressPercent"`
}

// Node represents a network node
type Node struct {
	mtx                 *sync.RWMutex
	cfg                 *config.EngineConfig // node config
	address             ma.Multiaddr         // node multiaddr
	IP                  net.IP               // node ip
	host                host.Host            // node libp2p host
	wg                  sync.WaitGroup       // wait group for preventing the main thread from exiting
	localNode           *Node                // local node
	peerManager         *Manager             // node manager for managing connections to other remote peers
	gProtoc             *Gossip              // gossip protocol instance
	remote              bool                 // remote indicates the node represents a remote peer
	Timestamp           time.Time            // the last time this node was seen/active
	isHardcodedSeed     bool                 // whether the node was hardcoded as a seed
	stopped             bool                 // flag to tell if node has stopped
	log                 logger.Logger        // node logger
	rSeed               []byte               // random 256 bit seed to be used for seed random operations
	db                  elldb.DB             // used to access and modify local database
	signatory           *d_crypto.Key        // signatory address used to get node ID and for signing
	history             *cache.Cache         // Used to track things we want to remember
	event               *emitter.Emitter     // Provides access event emitting service
	transactionsPool    *txpool.TxPool       // the transaction pool for transactions
	txsRelayQueue       *txpool.TxContainer  // stores transactions waiting to be relayed
	bchain              core.Blockchain      // The blockchain manager
	blockHashQueue      *lane.Deque          // Contains headers collected during block syncing
	bestRemoteBlockInfo *BestBlockInfo       // Holds information about the best known block heard from peers
	syncing             bool                 // Indicates the process of syncing the blockchain with peers
	acquainted          bool                 // Indicates that the node has "handshooked" the local node
	tickerDone          chan bool
}

// NewNode creates a node instance at the specified port
func newNode(db elldb.DB, config *config.EngineConfig, address string, coinbase *d_crypto.Key, log logger.Logger) (*Node, error) {

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
		mtx:              &sync.RWMutex{},
		cfg:              config,
		address:          util.FullAddressFromHost(host),
		host:             host,
		wg:               sync.WaitGroup{},
		log:              log,
		rSeed:            util.RandBytes(64),
		signatory:        coinbase,
		db:               db,
		event:            &emitter.Emitter{},
		transactionsPool: txpool.New(config.TxPool.Capacity),
		txsRelayQueue:    txpool.NewQueueNoSort(config.TxPool.Capacity),
		blockHashQueue:   lane.NewDeque(),
		history:          cache.NewActiveCache(5000),
		tickerDone:       make(chan bool),
	}

	node.localNode = node
	node.peerManager = NewManager(config, node, node.log)
	node.IP = node.ip()

	log.Info("Opened local database", "Backend", "LevelDB")

	return node, nil
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
func NewRemoteNode(address ma.Multiaddr, localNode *Node) *Node {
	node := &Node{
		address:   address,
		localNode: localNode,
		remote:    true,
		mtx:       &sync.RWMutex{},
	}
	node.IP = node.ip()
	return node
}

// NewAlmostEmptyNode returns a node with
// almost all its field uninitialized.
func NewAlmostEmptyNode() *Node {
	return &Node{
		mtx: &sync.RWMutex{},
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

// setSyncing sets the sync status
func (n *Node) setSyncing(syncing bool) {
	n.mtx.Lock()
	defer n.mtx.Unlock()
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

// updateSyncInfo sets a given remote best
// block info as the best known remote block
// only when it is better than the local best block.
// Using this information, it can tell when syncing
// has stopped and as such, updates the syncing status.
func (n *Node) updateSyncInfo(bi *BestBlockInfo) {
	n.mtx.Lock()
	defer n.mtx.Unlock()

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

// getSyncStateInfo generates status and progress
// information about the current blockchain sync operation
func (n *Node) getSyncStateInfo() *SyncStateInfo {

	// No need to compute when we are
	// not currently syncing
	if !n.isSyncing() {
		return nil
	}

	if n.bestRemoteBlockInfo == nil {
		return nil
	}

	var syncState = &SyncStateInfo{}

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
	n.mtx.RLock()
	defer n.mtx.RUnlock()
	return n.syncing
}

// Gossip returns the set protocol
func (n *Node) Gossip() *Gossip {
	return n.gProtoc
}

// PM returns the peer manager
func (n *Node) PM() *Manager {
	return n.peerManager
}

// GetHistory return a cache for
// holding arbitrary objects we
// want to keep track of
func (n *Node) GetHistory() *cache.Cache {
	return n.history
}

// IsSame checks if p is the same as node
func (n *Node) IsSame(node types.Engine) bool {
	return n.StringID() == node.StringID()
}

// GetBlockchain returns the
// blockchain manager
func (n *Node) GetBlockchain() core.Blockchain {
	return n.bchain
}

// SetBlockchain sets the blockchain
func (n *Node) SetBlockchain(bchain core.Blockchain) {
	n.bchain = bchain
}

// IsHardcodedSeed checks whether
// the node is an hardcoded seed node
func (n *Node) IsHardcodedSeed() bool {
	return n.isHardcodedSeed
}

// MakeHardcoded sets the node has
// an hardcoded seed node
func (n *Node) MakeHardcoded() {
	n.isHardcodedSeed = true
}

// SetTimestamp sets the timestamp value
func (n *Node) SetTimestamp(newTime time.Time) {
	n.mtx.Lock()
	defer n.mtx.Unlock()
	n.Timestamp = newTime
}

// GetEventEmitter gets the
// event emitter
func (n *Node) GetEventEmitter() *emitter.Emitter {
	return n.event
}

// GetTimestamp gets the nodes timestamp
func (n *Node) GetTimestamp() time.Time {
	n.mtx.RLock()
	defer n.mtx.RUnlock()
	return n.Timestamp
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
func (n *Node) SetEventEmitter(ee *emitter.Emitter) {
	n.event = ee
	n.transactionsPool.SetEventEmitter(ee)
}

// SetLocalNode sets the local peer
func (n *Node) SetLocalNode(node *Node) {
	n.localNode = node
}

// canAcceptPeer determines whether we
// can continue to interact with a
// remote node. This is a good place
// to check if a remote node has been
// blacklisted etc
func (n *Node) canAcceptPeer(remotePeer *Node) (bool, error) {
	// In non-production mode, we cannot interact with a remote peer with a public IP
	if !n.ProdMode() && !util.IsDevAddr(remotePeer.IP) {
		return false, fmt.Errorf("in development mode, we cannot interact with peers with public IP")
	}

	// If the local peer does not know the remotePeer, it cannot interact with it.
	// This only applies in production mode.
	if n.ProdMode() && !remotePeer.IsKnown() {
		return false, fmt.Errorf("remote peer is unknown")
	}

	return true, nil
}

// addToPeerStore adds a remote node
// to the host's peerstore
func (n *Node) addToPeerStore(remote types.Engine) *Node {
	n.localNode.Peerstore().AddAddr(remote.ID(), remote.GetIP4TCPAddr(), pstore.PermanentAddrTTL)
	return n
}

// newStream creates a stream to a peer
func (n *Node) newStream(ctx context.Context, peerID peer.ID, protocolID string) (inet.Stream, error) {
	return n.Host().NewStream(ctx, peerID, protocol.ID(protocolID))
}

// SetTransactionPool sets the
// transaction pool
func (n *Node) SetTransactionPool(txp *txpool.TxPool) {
	n.transactionsPool = txp
}

// SetGossipProtocol sets the
// gossip protocol implementation
func (n *Node) SetGossipProtocol(protoc *Gossip) {
	n.gProtoc = protoc
}

// GetHost returns the node's host
func (n *Node) GetHost() host.Host {
	return n.host
}

// GetBlockHashQueue returns
// the block hash queue
func (n *Node) GetBlockHashQueue() *lane.Deque {
	return n.blockHashQueue
}

// Peerstore returns the Peerstore
// of the node
func (n *Node) Peerstore() pstore.Peerstore {
	if h := n.Host(); h != nil {
		return h.Peerstore()
	}
	return nil
}

// Host returns the internal
// host instance
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

// StringID is like ID() but
// it returns string
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

// Connected checks whether the node
// is connected to the local node.
// Returns false if node is the local node.
func (n *Node) Connected() bool {
	if n.localNode == nil {
		return false
	}
	return len(n.localNode.host.Network().ConnsToPeer(n.ID())) > 0
}

// IsKnown checks whether a peer is
// known to the local node
func (n *Node) IsKnown() bool {
	if n.localNode == nil {
		return false
	}
	return n.localNode.PM().GetPeer(n.StringID()) != nil
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

// GetMultiAddr returns the full multi address of the node
func (n *Node) GetMultiAddr() string {
	if n.host == nil && !n.remote {
		return ""
	} else if n.remote {
		return n.address.String()
	}
	hostAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s",
		n.host.ID().Pretty()))
	return n.host.Addrs()[0].Encapsulate(hostAddr).String()
}

// GetAddr returns the host and port
// of the node as "host:port"
func (n *Node) GetAddr() string {
	hostAddr := n.host.Addrs()[0].String()
	parts := strings.Split(strings.Trim(hostAddr, "/"), "/")
	return fmt.Sprintf("%s:%s", parts[1], parts[3])
}

// GetIP4TCPAddr returns ip4 and tcp
// parts of the host's multi address
func (n *Node) GetIP4TCPAddr() ma.Multiaddr {
	ipfsPart := fmt.Sprintf("/ipfs/%s", n.ID().Pretty())
	ipfsAddr, _ := ma.NewMultiaddr(ipfsPart)
	return n.address.Decapsulate(ipfsAddr)
}

// GetBindAddress returns the bind address
func (n *Node) GetBindAddress() string {
	return n.address.String()
}

// validateAddress checks whether
// an address is valid for the
// current engine mode.
func validateAddress(engine types.Engine, address string) error {

	// Check whether the address is valid
	if !util.IsValidAddr(address) {
		return fmt.Errorf("not a valid multiaddr")
	}

	// In non-production mode, only
	// local/private addresses are allowed
	if !engine.ProdMode() && !util.IsDevAddr(util.GetIPFromAddr(address)) {
		return fmt.Errorf("public addresses are " +
			"not allowed in development mode")
	}

	// In production mode, only routable
	// addresses are allowed
	if engine.ProdMode() && !util.IsRoutableAddr(address) {
		return fmt.Errorf("local or private addresses " +
			"are not allowed in production mode")
	}

	return nil
}

// AddAddresses adds addresses that can be
// connected to when new connections need to
// be established.
func (n *Node) AddAddresses(peerAddresses []string, hardcoded bool) error {

	for _, addr := range peerAddresses {

		if err := validateAddress(n, addr); err != nil {
			n.log.Info("Invalid Bootstrap Address",
				"Err", err.Error(), "Address", addr)
			continue
		}

		pAddr, _ := ma.NewMultiaddr(addr)
		rp := NewRemoteNode(pAddr, n)
		rp.isHardcodedSeed = hardcoded
		rp.gProtoc = n.gProtoc
		n.peerManager.AddPeer(rp)
	}
	return nil
}

// connectToNode handshake to each bootstrap
// peer. Then send GetAddr message if
// handshake is successful
func (n *Node) connectToNode(remote types.Engine) error {
	if n.gProtoc.SendHandshake(remote) == nil {
		n.gProtoc.SendGetAddr([]types.Engine{remote})
	}
	return nil
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
			go n.gProtoc.RelayTx(tx, n.peerManager.GetActivePeers(0))
		case <-n.tickerDone:
			ticker.Stop()
			return
		}
	}
}

// Acquainted indicates that a node
// has undergone the handshake ritual
func (n *Node) Acquainted() {
	n.mtx.Lock()
	defer n.mtx.Unlock()
	n.acquainted = true
}

// IsAcquainted checks whether the node
// is acquainted with its local node.
func (n *Node) IsAcquainted() bool {
	return n.acquainted
}

// Start starts the node.
func (n *Node) Start() {

	// Start the peer manager
	n.PM().Manage()

	// Attempt to connect to peers
	for _, node := range n.PM().Peers() {
		go n.connectToNode(node)
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
func (n *Node) relayBlock(block core.Block) {
	if block.GetNumber() > 1 {
		n.gProtoc.RelayBlock(block, n.peerManager.GetActivePeers(0))
	}
}

func (n *Node) handleNewBlockEvent() {
	for {
		select {
		case evt := <-n.event.Once(core.EventNewBlock):
			n.relayBlock(evt.Args[0].(core.Block))
		}
	}
}

func (n *Node) handleNewTransactionEvent() {
	for {
		select {
		case evt := <-n.event.Once(core.EventNewTransaction):
			if !n.GetTxRelayQueue().Add(evt.Args[0].(core.Transaction)) {
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
			orphanBlock := evt.Args[0].(*objects.Block)
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
			abortedBlock := evt.Args[0].(*objects.Block)
			n.log.Debug("Attempting to re-add transactions "+
				"in aborted miner block",
				"NumTx", len(abortedBlock.Transactions))
			for _, tx := range abortedBlock.Transactions {
				if err := n.addTransaction(tx); err != nil {
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
			var broadcaster types.Engine
			otherBlockHashes := []interface{}{}

			// Collect hash of headers sent by a
			// particular broadcaster. Temporarily
			// keep the others in a cache to be added back
			// in the queue when we have collected some hashes
			for !n.blockHashQueue.Empty() && int64(len(hashes)) <
				params.MaxGetBlockBodiesHashes {
				bh := n.blockHashQueue.Shift().(*BlockHash)

				if broadcaster != nil && bh.Broadcaster.StringID() !=
					broadcaster.StringID() {
					otherBlockHashes = append(otherBlockHashes, bh)
					continue
				}

				hashes = append(hashes, bh.Hash)
				broadcaster = bh.Broadcaster
			}

			// append the others that were not selected
			// back to the block hash queue
			for _, bh := range otherBlockHashes {
				n.blockHashQueue.Append(bh)
			}

			// send block body request
			go n.gProtoc.SendGetBlockBodies(broadcaster, hashes)

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

	// Close the database.
	if n.db != nil {

		// Wait a few seconds for active
		// operations to complete before
		// closing the database
		time.Sleep(2 * time.Second)

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
		return nil, fmt.Errorf("invalid address provided")
	}
	nAddr, _ := ma.NewMultiaddr(addr)
	return &Node{
		address:   nAddr,
		localNode: n,
		gProtoc:   n.gProtoc,
		remote:    remote,
		mtx:       &sync.RWMutex{},
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
	n.mtx.RLock()
	defer n.mtx.RUnlock()

	if n.Timestamp.IsZero() {
		return true
	}

	now := time.Now().UTC()
	if n.Timestamp.After(now.Add(time.Minute*10)) ||
		n.Timestamp.Before(now.Add(-3*time.Hour)) {
		return true
	}

	return false
}

// GetTxRelayQueue returns the transaction relay queue
func (n *Node) GetTxRelayQueue() *txpool.TxContainer {
	return n.txsRelayQueue
}

// GetTxPool returns the unsigned transaction pool
func (n *Node) GetTxPool() core.TxPool {
	return n.transactionsPool
}
