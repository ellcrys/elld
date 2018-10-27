package core

import (
	"math/big"
	"time"

	"github.com/olebedev/emitter"
	lane "gopkg.in/oleiade/lane.v1"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/util/cache"
	host "github.com/libp2p/go-libp2p-host"
	peer "github.com/libp2p/go-libp2p-peer"
)

// Engine represents node functionalities not provided by the
// protocol. This can include peer discovery, configuration,
// APIs etc.
type Engine interface {

	// SetEventEmitter set the event emitter
	// used to broadcast/receive events
	SetEventEmitter(*emitter.Emitter)

	// DB is the laoded client database
	DB() elldb.DB

	// GetTxPool returns the transaction pool
	GetTxPool() types.TxPool

	// StringID is the peer ID of engine's
	// network host as a string
	StringID() string

	// ShortID is the short version the value
	// return by StringID. It meant to be
	// used for logging.
	ShortID() string

	// ID returns the peer ID of the engine's
	// network knows
	ID() peer.ID

	// GetLastSeen returns the time the engine
	// (or peer) was last seen
	GetLastSeen() time.Time

	// SetLastSeen sets the time the engine
	// (or peer) was last seen
	SetLastSeen(time.Time)

	// CreatedAt returns the time this engine
	// was first created.
	CreatedAt() time.Time

	// SetCreatedAt sets the time this engine
	// was first created
	SetCreatedAt(t time.Time)

	// IsSame checks whether the engine has
	// the same ID as another engine
	IsSame(Engine) bool

	// IsSameID is like IsSame but accepts a string
	IsSameID(id string) bool

	// IsHardcodedSeed indicates that the engine
	// was not discovered through the gossip protocol
	// but injected into the codebase
	IsHardcodedSeed() bool

	// GetAddress returns the listening address
	// of the host network
	GetAddress() util.NodeAddr

	// Connected checks whether the engine is connected
	// to the local node
	Connected() bool

	// GetBlockchain returns the blockchain instance
	GetBlockchain() types.Blockchain

	// SetBlockchain stores a reference to the
	// blockchain instance
	SetBlockchain(bchain types.Blockchain)

	// ProdMode checks whether the engine
	// is in production mode
	ProdMode() bool

	// TestMode checks whether the engine
	// is in test mode
	TestMode() bool

	// IsInbound checks whether the engine (or peer) is
	// considered an inbound connection to the local node
	IsInbound() bool

	// IsInbound checks whether the engine (or peer) is
	// considered an outbound connection to the local node
	SetInbound(v bool)

	// HasStopped checks whether the peer has stopped
	HasStopped() bool

	// GetHost returns the engine's network host
	GetHost() host.Host

	// Gossip returns the gossip manager
	Gossip() Gossip

	// NewRemoteNode creates a node that will be used to
	// represent a remote peer.
	NewRemoteNode(address util.NodeAddr) Engine

	// GetCfg returns the engine configuration
	GetCfg() *config.EngineConfig

	// GetEventEmitter gets the event emitter
	GetEventEmitter() *emitter.Emitter

	// GetHistory returns the general items cache
	GetHistory() *cache.Cache

	// SetSyncing sets the sync status
	SetSyncing(syncing bool)

	// UpdateSyncInfo updates the sync state
	UpdateSyncInfo(bi *BestBlockInfo)

	// GetBlockHashQueue returns the block hash queue
	GetBlockHashQueue() *lane.Deque

	// GetSyncStateInfo generates status and progress
	// information about the current blockchain sync operation
	GetSyncStateInfo() *SyncStateInfo

	// AddToPeerStore adds the ID of the engine
	// to the peerstore
	AddToPeerStore(node Engine) Engine

	// GetIntros returns the cache containing received intros
	GetIntros() *cache.Cache

	// AddTransaction validates and adds a
	// transaction to the transaction pool.
	AddTransaction(tx types.Transaction) error

	// IsHardCodedSeed sets the hardcoded seed state
	// of the engine.
	IsHardCodedSeed(v bool)

	// SetGossipManager sets the gossip manager
	SetGossipManager(m Gossip)
}

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
