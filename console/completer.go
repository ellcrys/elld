package console

import (
	prompt "github.com/c-bata/go-prompt"
)

var initialSuggestions = []prompt.Suggest{
	{Text: ".exit", Description: "Exit the console"},
	{Text: ".help", Description: "Print the help message"},
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
	words := d.GetWordBeforeCursor()
	if len(words) < 1 {
		return nil
	}
	return prompt.FilterHasPrefix(sm.suggestions, words, true)
}

// Extend the current suggestions and return it
func (sm *SuggestionManager) extend(suggestions []prompt.Suggest) {
	sm.suggestions = append(initialSuggestions, suggestions...)
	return
}
