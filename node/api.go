package node

import (
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/wire"
)

// APIs returns all API handlers
func (n *Node) APIs() types.APISet {
	return map[string]types.APIFunc{
		"TransactionAdd": func(args ...interface{}) (interface{}, error) {
			return nil, n.addTransaction(args[0].(*wire.Transaction))
		},
	}
}
