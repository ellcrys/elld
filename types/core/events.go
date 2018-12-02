package core

const (
	// EventMinerProposedBlockAborted defines an event about an aborted PoW operation
	EventMinerProposedBlockAborted = "event.aborted"

	// EventNewBlock represents an event about a new block
	// that was successfully added to the main chain
	EventNewBlock = "event.newBlock"

	// EventNewTransaction represents an event about a new
	// transaction that was received
	EventNewTransaction = "event.newTx"

	// EventOrphanBlock represents an event about an orphan block.
	EventOrphanBlock = "event.orphanBlock"

	// EventFoundBlock represents an event about a
	// block that has just been mined by the client
	EventFoundBlock = "event.foundBlock"
)
