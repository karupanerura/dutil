package command

import (
	"github.com/manifoldco/promptui"
)

func confirm(message string) bool {
	prompt := promptui.Select{
		Label: message + " [Yes/No]",
		Items: []string{"Yes", "No"},
	}
	_, result, err := prompt.Run()
	if err != nil {
		panic(err)
	}
	return result == "Yes"
}
