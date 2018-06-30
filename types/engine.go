package types

import (
	evbus "github.com/asaskevich/EventBus"
	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/database"
	"github.com/ellcrys/elld/txpool"
)

// Engine represents node functionalities not provided by the
// protocol. This can include peer discovery, configuration,
// APIs etc.
type Engine interface {
	SetLogicBus(bus evbus.Bus)                // Set the event bus used to perform logical operations against the blockchain
	GetUnsignedTxRelayQueue() *txpool.TxQueue // Returns the unsigned transaction relay queue
	GetUnSignedTxPool() *txpool.TxPool        // Returns the unsigned transaction pool
	AddTxSession(txID string)                 // Adds a transaction to the transaction session collection
	Cfg() *config.EngineConfig                // Returns the engine configuration
	DB() database.DB                          // The engine's database instance
}
