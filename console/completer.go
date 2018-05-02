package console

import prompt "github.com/c-bata/go-prompt"

func completer(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "exit", Description: "Exit the console"},
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}
