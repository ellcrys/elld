package console

import (
	prompt "github.com/c-bata/go-prompt"
)

var initialSuggestions = []prompt.Suggest{
	{Text: ".exit", Description: "Exit the console"},
	{Text: ".help", Description: "Print the help message"},
	{Text: "spell", Description: "Ellcrys console services"},
	{Text: "spell.balance", Description: "Balance module - Send ELL, Check balance etc"},
	{Text: "spell.balance.send", Description: "Send ELL to a non-contract account"},
	{Text: "spell.accounts", Description: "List all accounts"},
	{Text: "spell.account", Description: "Object"},
	{Text: "spell.account.getAccounts", Description: "List all accounts"},
}

var commonFunc = [][]string{
	{".help", "Print the help message"},
	{".exit", "Exit the console"},
}

// SuggestionManager manages suggestions
type SuggestionManager struct {
	initialSuggestions []prompt.Suggest
	suggestions        []prompt.Suggest
}

// NewSuggestionManager creates a suggestion manager. Initialize suggestions
// by providing initial suggestion as argument
func NewSuggestionManager(initialSuggestions []prompt.Suggest) *SuggestionManager {
	sm := new(SuggestionManager)
	sm.initialSuggestions = initialSuggestions
	sm.suggestions = initialSuggestions
	return sm
}

func (sm *SuggestionManager) completer(d prompt.Document) []prompt.Suggest {
	if words := d.GetWordBeforeCursor(); len(words) > 1 {
		return prompt.FilterHasPrefix(sm.suggestions, words, true)
	}
	return nil
}
