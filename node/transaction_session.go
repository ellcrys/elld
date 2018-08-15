package node

// AddTxSession adds a transaction id to the session map
func (n *Node) AddTxSession(txID string) {
	n.mtx.Lock()
	defer n.mtx.Unlock()
	if _, has := n.openTransactionsSession[txID]; !has {
		n.openTransactionsSession[txID] = struct{}{}
		n.log.Info("New transaction session has been opened", "TxID", txID, "NumOpenedSessions", len(n.openTransactionsSession))
		return
	}
}

// HasTxSession checks whether a transaction has an open session
func (n *Node) HasTxSession(txID string) bool {
	n.mtx.RLock()
	defer n.mtx.RUnlock()
	_, has := n.openTransactionsSession[txID]
	return has
}

// RemoveTxSession removes a transaction's session entry
func (n *Node) RemoveTxSession(txID string) {
	n.mtx.Lock()
	defer n.mtx.Unlock()
	delete(n.openTransactionsSession, txID)
}

// CountTxSession counts the number of opened transaction sessions
func (n *Node) CountTxSession() int {
	n.mtx.RLock()
	defer n.mtx.RUnlock()
	return len(n.openTransactionsSession)
}
