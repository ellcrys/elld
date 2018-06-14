package node

// AddTxSession adds a transaction id to the session map
func (protoc *Inception) AddTxSession(txID string) {
	if !protoc.HasTxSession(txID) {
		protoc.mtx.Lock()
		defer protoc.mtx.Unlock()
		protoc.openTxSessions[txID] = struct{}{}
		protoc.log.Info("New transaction session has been opened", "TxID", txID, "NumOpenedSessions", protoc.CountTxSession())
	}
}

// HasTxSession checks whether a transaction has an open session
func (protoc *Inception) HasTxSession(txID string) bool {
	protoc.mtx.Lock()
	defer protoc.mtx.Unlock()
	_, has := protoc.openTxSessions[txID]
	return has
}

// RemoveTxSession removes a transaction's session entry
func (protoc *Inception) RemoveTxSession(txID string) {
	protoc.mtx.Lock()
	defer protoc.mtx.Unlock()
	delete(protoc.openTxSessions, txID)
}

// CountTxSession counts the number of opened transaction sessions
func (protoc *Inception) CountTxSession() int {
	protoc.mtx.Lock()
	defer protoc.mtx.Unlock()
	return len(protoc.openTxSessions)
}
