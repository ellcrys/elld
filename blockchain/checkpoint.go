package blockchain

import "github.com/ellcrys/elld/config"

// findPreviousCheckpoint finds the checkpoint from the list of
// checkpoints that is the most recent in the best chain
func (b *Blockchain) findPreviousCheckpoint() *config.CheckPoint {
	return nil
	// // starting from the last checkpoint, look for it in the
	// // main chain. If we find any of the checkpoint, then that's
	// // the previous checkpoint
	// numCheckpoints := len(checkpoints)
	// for i := numCheckpoints; i > 0; i-- {
	// 	checkpoint := checkpoints[i-1]
	// 	exists, err := b.bestChain.hasBlock(checkpoint.Hash)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	if exists {
	// 		return checkpoint
	// 	}
	// }
}
