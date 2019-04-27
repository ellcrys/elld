package blockchain

// setSkipDecideBestChain disables or enables best chain
// choice selection. This method is only useful in
// preventing unwanted re-organisation when creating
// mock chain hierarchy in integration tests.
func (b *Blockchain) setSkipDecideBestChain(v bool) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.skipDecideBestChain = v
}

// canDecideBestChain checks whether best chain choice
// can be carried out.
func (b *Blockchain) canDecideBestChain() bool {
	b.lock.RLock()
	defer b.lock.RUnlock()
	return !b.skipDecideBestChain
}
