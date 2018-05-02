package console

import prompt "github.com/c-bata/go-prompt"

var suggestions = []prompt.Suggest{
	{Text: ".exit", Description: "Exit the console"},
	{Text: ".help", Description: "Print the help message"},
}

var commonFunc = [][]string{
	{".help", "Print the help message"},
	{".exit", "Exit the console"},
}

func completer(d prompt.Document) []prompt.Suggest {
	words := d.GetWordBeforeCursor()
	if len(words) < 1 {
		return nil
	}
	return prompt.FilterHasPrefix(suggestions, words, true)
}
