package main

import (
	. "github.com/ellcrys/go.stub"
)

// BlockcodeEx defines a blockcode
type BlockcodeEx struct {
}

// OnInit initializes the blockcode
func (b *BlockcodeEx) OnInit() {
	On("add", b.add)
}

func (b *BlockcodeEx) add() (interface{}, error) {
	return 10, nil
}

func main() {
	Run(new(BlockcodeEx))
}
