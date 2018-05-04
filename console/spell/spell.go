package spell

import "net/rpc"

// Spell provides implementation for various actions
// performed in the console.
type Spell struct {
	EllService *ELLService
	client     *rpc.Client
}

// NewSpell creates a new Spell instance
// Returns error if unable to dial RPC server.
func NewSpell(rpcAddr string) (*Spell, error) {

	client, err := rpc.DialHTTP("tcp", rpcAddr)
	if err != nil {
		return nil, err
	}

	spell := new(Spell)
	spell.EllService = new(ELLService)
	spell.EllService.client = client

	return spell, nil
}
