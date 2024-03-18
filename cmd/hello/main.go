package hello

import (
	"fmt"

	"go-cli-app/internal/config"
	"go-cli-app/pkg/utils"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func Run(name string) {
	defer utils.PrintSeparator()
	if config.GlobalConfig.CapitalizeName {
		// Parse the language from the configuration
		lang := language.Make(config.GlobalConfig.Language)
		caser := cases.Title(lang)
		name = caser.String(name)
	}

	fmt.Println("Hello", name)
}
