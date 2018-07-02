package node

// AddTxSession adds a transaction id to the session map
func (n *Node) AddTxSession(txID string) {
	if !n.HasTxSession(txID) {
		n.mtx.Lock()
		n.openTransactionsSession[txID] = struct{}{}
		n.mtx.Unlock()
		n.log.Info("New transaction session has been opened", "TxID", txID, "NumOpenedSessions", n.CountTxSession())
		return
	}
}

// HasTxSession checks whether a transaction has an open session
func (n *Node) HasTxSession(txID string) bool {
	n.mtx.Lock()
	defer n.mtx.Unlock()
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
	n.mtx.Lock()
	defer n.mtx.Unlock()
	return len(n.openTransactionsSession)
}
