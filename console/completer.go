package console

import (
	prompt "github.com/c-bata/go-prompt"
)

var initialSuggestions = []prompt.Suggest{
	{Text: ".exit", Description: "Exit the console"},
	{Text: ".help", Description: "Print the help message"},
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

// hasSuggestion checks whether a suggest with
// the given text exists.
func (sm *SuggestionManager) hasSuggestion(text string) bool {
	for _, s := range sm.suggestions {
		if s.Text == text {
			return true
		}
	}
	return false
}

// add adds a suggestion if it does not
// already exists.
func (sm *SuggestionManager) add(sugs ...prompt.Suggest) {
	for _, s := range sugs {
		if sm.hasSuggestion(s.Text) {
			continue
		}
		sm.suggestions = append(sm.suggestions, s)
	}
}

// completer is used to filter the what suggestion
// to show in the completer
func (sm *SuggestionManager) completer(d prompt.Document) []prompt.Suggest {
	if words := d.GetWordBeforeCursor(); len(words) > 1 {
		return prompt.FilterHasPrefix(sm.suggestions, words, true)
	}
	return nil
}
