package spell

import (
	"fmt"
	"net/rpc"

	"github.com/robertkrimen/otto"

	"github.com/ellcrys/elld/crypto"
	"github.com/fatih/color"
)

// Spell provides implementation for various node and blockchain service
type Spell struct {
	key     *crypto.Key
	Balance *BalanceService
	Account *Account
	client  *rpc.Client
}

// NewSpell creates a new Spell instance
// Returns error if unable to dial RPC server.
func NewSpell(key *crypto.Key) *Spell {
	spell := new(Spell)
	spell.key = key
	spell.Balance = NewBalanceService(nil, key)
	spell.Account = NewAccount(nil, key)
	return spell
}

// SetClient sets the rpc client on all services that
// need to interact with the rpc server
func (spell *Spell) SetClient(c *rpc.Client) {
	spell.client = c
	spell.Balance.client = c
	spell.Account.client = c
}

// SetSignatory sets the key to be used to sign transactions
func (spell *Spell) SetSignatory(key *crypto.Key) {
	spell.key = key
	spell.Balance.key = key
	spell.Account.key = key
}

// RecoverFunc recovers from panics. Used only when
// we expect a spell function to panic but we want to only print an error
func RecoverFunc() {
	if r := recover(); r != nil {
		color.Red("Panic: %s", r)
	}
}

// Panic is a convenient method of panicking with a method path and message
func Panic(methodPath, msg string) {
	panic(fmt.Errorf(fmt.Sprintf("(%s) %s", methodPath, msg)))
}

// ConsoleErr creates and return a representation of a javascript error
func ConsoleErr(msg string, fields map[string]interface{}) otto.Value {

	var fieldConstruct = ""
	for k, v := range fields {
		switch v.(type) {
		case string:
			fieldConstruct += fmt.Sprintf("instance.%s = '%s';", k, v)
		case int, int64, float64:
			fieldConstruct += fmt.Sprintf("instance.%s = %d;", k, v)
		}
	}

	_, v, err := otto.Run(`
	function RPCError(msg){
		var instance = new Error(msg);
		instance.error = true;
		` + fieldConstruct + `
		return instance;
	} 
	new RPCError('` + msg + `');`)

	if err != nil {
		panic(err)
	}

	return v
}
