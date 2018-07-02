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
	SetLogicBus(bus evbus.Bus)     // Set the event bus used to perform logical operations against the blockchain
	Cfg() *config.EngineConfig     // Returns the engine configuration
	DB() database.DB               // The engine's database instance
	AddTxSession(txID string)      // Add new transaction session
	HasTxSession(txID string) bool // Check if a transaction has an existing session
	RemoveTxSession(txID string)   // Remove a transaction session
	CountTxSession() int           // Count the number of open transaction session
	GetTxPool() *txpool.TxPool     // Returns the transaction pool
}
