package command

import "github.com/c-bata/go-prompt"

func confirm(message string) bool {
	result := prompt.Choose(message+" ", []string{"Yes", "No"}, prompt.OptionShowCompletionAtStart(), prompt.OptionCompletionOnDown())
	return result == "Yes"
}
