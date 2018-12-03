package core

const (
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

	// EventPeerChainInfo indicates an event about a peer's chain info
	EventPeerChainInfo = "event.newPeerChainInfo"
)
