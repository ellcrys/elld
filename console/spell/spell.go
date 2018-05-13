package spell

import "net/rpc"

// Spell provides implementation for various node and blockchain service
type Spell struct {
	Ell     *ELLService
	Account *AccountService
	client  *rpc.Client
}

// NewSpell creates a new Spell instance
// Returns error if unable to dial RPC server.
func NewSpell() *Spell {
	spell := new(Spell)
	spell.Ell = new(ELLService)
	spell.Account = new(AccountService)
	return spell
}

// SetClient sets the rpc client on all services that
// need to interact with the rpc server
func (spell *Spell) SetClient(c *rpc.Client) {
	spell.client = c
	spell.Ell.client = c
	spell.Account.client = c
}
