package core

const (
	// EventNewBlock represents an event about a new block
	// that was successfully added to the main chain
	EventNewBlock = "event.newBlock"

	// EventProcessBlock represents an event about
	// a block that was relayed by a peer.
	EventProcessBlock = "event.relayedBlock"

	// EventNewTransaction represents an event about a new
	// transaction that was received
	EventNewTransaction = "event.newTx"

	// EventTransactionReceived indicates that a transaction
	// has been received.
	EventTransactionReceived = "event.txReceived"

	// EventTransactionPooled indicates that a transaction
	// has been added to the transaction pool
	EventTransactionPooled = "event.txPooled"

	// EventTransactionInvalid indicates that a transaction
	// has been declared invalid
	EventTransactionInvalid = "event.txInvalid"

	// EventOrphanBlock represents an event about an orphan block.
	EventOrphanBlock = "event.orphanBlock"

	// EventPeerChainInfo indicates an event about a peer's chain info
	EventPeerChainInfo = "event.newPeerChainInfo"

	// EventBlockProcessed describes an event about
	// a processed block
	EventBlockProcessed = "event.blockProcessed"
)
