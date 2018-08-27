package console

import (
	prompt "github.com/c-bata/go-prompt"
)

var initialSuggestions = []prompt.Suggest{
	{Text: ".exit", Description: "Exit the console"},
	{Text: ".help", Description: "Print the help message"},
	{Text: "ell.auth()", Description: "Get a session token"},
	{Text: "ell.minerStart()", Description: "Start mining"},
	{Text: "ell.minerStop()", Description: "Stop mining"},
	{Text: "ell.mining()", Description: "Check if miner is active"},
	{Text: "ell.minerHashrate()", Description: "Get current miner hashrate"},
	{Text: "ell.listAccounts()", Description: "List all accounts"},
	{Text: "ell.sendTx()", Description: "Send a transaction"},
	{Text: "ell.join()", Description: "Connect to a peer"},
	{Text: "ell.numConnections()", Description: "Get number of active connections"},
	{Text: "ell.getActivePeers()", Description: "Get a list of active peers"},
	{Text: "ell.getChains()", Description: "Get all chains"},
	{Text: "ell.getBlock()", Description: "Get a block by number"},
	{Text: "ell.getBlockByHash()", Description: "Get a block by hash"},
	{Text: "ell.getOrphans()", Description: "Get a list of orphans"},
	{Text: "ell.getBestChain()", Description: "Get the best chain"},
	{Text: "ell.getReOrgs()", Description: "Get a list of re-organization events"},
	{Text: "methods", Description: "List all RPC methods"},
	{Text: "login()", Description: "Login to begin an authenticated session"},
}

var commonFunc = [][]string{
	{".help", "Print the help message"},
	{".exit", "Exit the console"},
}

// SuggestionManager manages suggestions
type SuggestionManager struct {
	suggestions []prompt.Suggest
}

// NewSuggestionManager creates a suggestion manager. Initialize suggestions
// by providing initial suggestion as argument
func newSuggestionManager(initialSuggestions []prompt.Suggest) *SuggestionManager {
	sm := new(SuggestionManager)
	sm.suggestions = initialSuggestions
	return sm
}

func (sm *SuggestionManager) completer(d prompt.Document) []prompt.Suggest {
	if words := d.GetWordBeforeCursor(); len(words) > 1 {
		return prompt.FilterHasPrefix(sm.suggestions, words, true)
	}
	return nil
}
