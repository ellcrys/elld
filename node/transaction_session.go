package node

// AddTxSession adds a transaction id to the session map
func (protoc *Gossip) AddTxSession(txID string) {
	if !protoc.HasTxSession(txID) {
		protoc.mtx.Lock()
		protoc.openTransactionsSession[txID] = struct{}{}
		protoc.mtx.Unlock()
		protoc.log.Info("New transaction session has been opened", "TxID", txID, "NumOpenedSessions", protoc.CountTxSession())
		return
	}
}

// HasTxSession checks whether a transaction has an open session
func (protoc *Gossip) HasTxSession(txID string) bool {
	protoc.mtx.Lock()
	defer protoc.mtx.Unlock()
	_, has := protoc.openTransactionsSession[txID]
	return has
}

// RemoveTxSession removes a transaction's session entry
func (protoc *Gossip) RemoveTxSession(txID string) {
	protoc.mtx.Lock()
	defer protoc.mtx.Unlock()
	delete(protoc.openTransactionsSession, txID)
}

// CountTxSession counts the number of opened transaction sessions
func (protoc *Gossip) CountTxSession() int {
	protoc.mtx.Lock()
	defer protoc.mtx.Unlock()
	return len(protoc.openTransactionsSession)
}
