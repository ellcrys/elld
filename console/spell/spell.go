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
func NewSpell() *Spell {
	spell := new(Spell)
	spell.EllService = new(ELLService)
	return spell
}

// SetClient sets the rpc client
func (spell *Spell) SetClient(c *rpc.Client) {
	spell.client = c
	spell.EllService.client = c
}
