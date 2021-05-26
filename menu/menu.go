package menu

import (
	"github.com/manifoldco/promptui"
)

type Option struct {
	Name  string
	Value interface{}
}

func Run(menuTitle, pointerChar string, options []Option) Option {
	list := promptui.Select{
		Label: menuTitle,
		Items: options,
		Templates: &promptui.SelectTemplates{
			Active:   `🎣  {{ .Name | cyan | bold }}`,
			Inactive: `    {{ .Name | cyan }}`,
			Selected: `{{ "✔ Selected" | green | bold }}: {{ .Name | cyan }}`,
		},
	}

	idx, _, _ := list.Run()
	return options[idx]
}
