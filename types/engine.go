package types

import (
	"time"

	evbus "github.com/asaskevich/EventBus"
	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/txpool"
	peer "github.com/libp2p/go-libp2p-peer"
	ma "github.com/multiformats/go-multiaddr"
)

// Engine represents node functionalities not provided by the
// protocol. This can include peer discovery, configuration,
// APIs etc.
type Engine interface {
	SetLogicBus(bus evbus.Bus)              // Set the event bus used to perform logical operations against the blockchain
	Cfg() *config.EngineConfig              // Returns the engine configuration
	DB() elldb.DB                           // The engine's database instance
	AddTxSession(txID string)               // Add new transaction session
	HasTxSession(txID string) bool          // Check if a transaction has an existing session
	RemoveTxSession(txID string)            // Remove a transaction session
	CountTxSession() int                    // Count the number of open transaction session
	GetTxPool() *txpool.TxPool              // Returns the transaction pool
	StringID() string                       // Returns the engine ID
	ShortID() string                        // Return the short version of the engine ID
	ID() peer.ID                            // Get the ID as issued by libp2p
	GetIP4TCPAddr() ma.Multiaddr            // Returns the ipv4 address of the engine
	GetTimestamp() time.Time                // Returns the timestamp of the engine
	SetTimestamp(time.Time)                 // Set the engine's timestamp
	IsSame(Engine) bool                     // Checks whether an remote node has same ID as the engine
	IsHardcodedSeed() bool                  // Checks whether the engine is an hardcoded seed peer
	GetMultiAddr() string                   // Returns the multiaddr of the node
	Connected() bool                        // Returns true if engine is connected to its local node
	GetBlockchain() common.Blockchain       // Returns the blockchain manager
	SetBlockchain(bchain common.Blockchain) // Set the blockchain manager
}
