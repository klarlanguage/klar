package filenames

import "github.com/klarlanguage/klar/pkg/lint"

var Linter = &lint.Linter{
	Name:             "fileNames",
	ShortDescription: "Ensures .klar file names follow naming conventions",
	LongDescription:  "", // TODO (maybe also a docs URL)
	Settings: func() *lint.LinterSettings {
		return &lint.LinterSettings{
			Linter: &Settings{BasedOnTypes: true},
		}
	},
	Run: Run,
}

type Settings struct {
	// If a file mainly provides a type, ensure the names of the file and the type match.
	BasedOnTypes bool
}

var (
	RuleNameMustBeCamelCase = Linter.NewRule("nameMustBeCamelCase")
	RuleNameMustMatchType   = Linter.NewRule("nameMustMatchType")
)

func Run(lc *lint.LinterContext, opts *lint.LinterSettings) error {
	return nil
}
