package core

const (
	// EventAborted defines an event about an aborted PoW operation
	EventAborted = "event.aborted"

	// EventNewBlock represents an event about a new block
	// that was successfully added to the main chain
	EventNewBlock = "event.newBlock"

	// EventNewTransaction  represents an event about a new
	// transaction that was received 
	EventNewTransaction = "event.newTx"
)
