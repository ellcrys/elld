package types

import (
	"time"

	"github.com/olebedev/emitter"

	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"
	peer "github.com/libp2p/go-libp2p-peer"
	ma "github.com/multiformats/go-multiaddr"
)

// TxPool represents a transactions pool
type TxPool interface {
	SetEventEmitter(ee *emitter.Emitter)
	Put(tx core.Transaction) error
	Has(tx core.Transaction) bool
	HasByHash(hash string) bool
	SenderHasTxWithSameNonce(address util.String, nonce uint64) bool
	Select(maxSize int64) (txs []core.Transaction)
	ByteSize() int64
	Size() int64
}

// Engine represents node functionalities not provided by the
// protocol. This can include peer discovery, configuration,
// APIs etc.
type Engine interface {
	SetEventEmitter(*emitter.Emitter)     // Set the event emitter used to broadcast/receive events
	DB() elldb.DB                         // The engine's database instance
	GetTxPool() TxPool                    // Returns the transaction pool
	StringID() string                     // Returns the engine ID
	ShortID() string                      // Return the short version of the engine ID
	ID() peer.ID                          // Get the ID as issued by libp2p
	GetIP4TCPAddr() ma.Multiaddr          // Returns the ipv4 address of the engine
	GetTimestamp() time.Time              // Returns the timestamp of the engine
	SetTimestamp(time.Time)               // Set the engine's timestamp
	IsSame(Engine) bool                   // Checks whether an remote node has same ID as the engine
	IsHardcodedSeed() bool                // Checks whether the engine is an hardcoded seed peer
	GetMultiAddr() string                 // Returns the multiaddr of the node
	Connected() bool                      // Returns true if engine is connected to its local node
	GetBlockchain() core.Blockchain       // Returns the blockchain manager
	SetBlockchain(bchain core.Blockchain) // Set the blockchain manager
	ProdMode() bool                       // Checks whether the current mode is ModeProd
}
