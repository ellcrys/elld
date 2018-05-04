package console

import (
	"fmt"
	"regexp"

	prompt "github.com/c-bata/go-prompt"
)

var initialSuggestions = []prompt.Suggest{
	{Text: ".exit", Description: "Exit the console"},
	{Text: ".help", Description: "Print the help message"},
	{Text: "spell", Description: "Ellcrys console services"},
	{Text: "spell.ell", Description: "Blockchain interaction module"},
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

// extend the current suggestions and return it
func (sm *SuggestionManager) extend(suggestions []prompt.Suggest) {

	// new suggestions might include symbols that existing in the initial
	// suggestions but have been overridden. We need to remove such symbols
	// from the initial suggestions.
	// This implementation can potentially be slow if the suggestions grow too large.
	var updatedInitialSuggestions []prompt.Suggest
	for _, i := range sm.initialSuggestions {
		found := false
		for _, j := range suggestions {
			m, _ := regexp.Match(fmt.Sprintf("%s.*", j.Text), []byte(i.Text))
			if m {
				found = true
				break
			}
		}
		if !found {
			updatedInitialSuggestions = append(updatedInitialSuggestions, i)
		}
		found = false
	}

	sm.suggestions = append(updatedInitialSuggestions, suggestions...)
	return
}
