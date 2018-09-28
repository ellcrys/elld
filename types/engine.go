package types

import (
	"time"

	"github.com/olebedev/emitter"

	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/types/core"
	peer "github.com/libp2p/go-libp2p-peer"
	ma "github.com/multiformats/go-multiaddr"
)

// Engine represents node functionalities not provided by the
// protocol. This can include peer discovery, configuration,
// APIs etc.
type Engine interface {
	SetEventEmitter(*emitter.Emitter)     // Set the event emitter used to broadcast/receive events
	DB() elldb.DB                         // The engine's database instance
	GetTxPool() core.TxPool               // Returns the transaction pool
	StringID() string                     // Returns the engine ID
	ShortID() string                      // Return the short version of the engine ID
	ID() peer.ID                          // Get the ID as issued by libp2p
	GetIP4TCPAddr() ma.Multiaddr          // Returns the ipv4 address of the engine
	GetLastSeen() time.Time               // Returns the timestamp of the engine
	SetLastSeen(time.Time)                // Set the engine's timestamp
	CreatedAt() time.Time                 // Returns the time the node was created
	SetCreatedAt(t time.Time)             // Set the time the node was created locally
	IsSame(Engine) bool                   // Checks whether an remote node has same ID as the engine
	IsHardcodedSeed() bool                // Checks whether the engine is an hardcoded seed node
	GetMultiAddr() string                 // Returns the multiaddr of the node
	Connected() bool                      // Returns true if engine is connected to its local node
	GetBlockchain() core.Blockchain       // Returns the blockchain manager
	SetBlockchain(bchain core.Blockchain) // Set the blockchain manager
	ProdMode() bool                       // Checks whether the current mode is ModeProd
	Acquainted()                          // Sets the acquainted status to true
	IsAcquainted() bool                   // Returns the acquainted status
}
