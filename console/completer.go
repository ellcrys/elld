package console

import prompt "github.com/c-bata/go-prompt"

func completer(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: ".exit", Description: "Exit the console"},
	}
	words := d.GetWordBeforeCursor()
	if len(words) < 1 {
		return nil
	}
	return prompt.FilterHasPrefix(s, words, true)
}
