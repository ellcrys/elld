package helpers

import (
	"math/big"
	"time"
)

// Block represents a block
type Block struct {
	Number     uint64
	Timestamp  time.Time
	Difficulty *big.Int
}
