package blockchain

import "github.com/ellcrys/elld/config"

// checkpoints
var checkpoints = []*config.CheckPoint{
	&config.CheckPoint{
		Number: 3000,
		Hash:   "abc",
	},
}
