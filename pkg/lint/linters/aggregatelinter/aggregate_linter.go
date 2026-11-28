// Package aggregatelinter provides a [lint.Linter] implementation that
// runs multiple linters.
package aggregatelinter

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"github.com/klarlanguage/klar/pkg/lint"
)

type AggregateLinter struct {
	Linters []*lint.Linter

	Name             string
	ShortDescription string
	LongDescription  string // The list of included linters will be appended to this
	DocsURL          string

	// TODO: Settings
}

func (al *AggregateLinter) Linter() *lint.Linter {
	rules := []lint.Rule{}
	// Include the names and descriptions of each linter in the long description
	var b strings.Builder
	b.WriteString(al.LongDescription)
	b.WriteString("\n\nThe following linters are included:\n")
	sortByName := func(a, b *lint.Linter) int { return cmp.Compare(a.Name, b.Name) }
	for _, l := range slices.SortedFunc(slices.Values(al.Linters), sortByName) {
		fmt.Fprintf(&b, "\n  - %s - %s", l.Name, l.ShortDescription)
		// TODO: Should the returned Linter's Rules point to the returned linter,
		// or the linter where the rule actually came from?
		rules = append(rules, l.Rules...)
	}

	return &lint.Linter{
		Name:             al.Name,
		ShortDescription: al.ShortDescription,
		LongDescription:  b.String(),
		DocsURL:          al.DocsURL,
		Run:              al.run,
		Rules:            rules,
		Settings:         nil, // TODO
	}
}

func (al *AggregateLinter) run(lc *lint.LinterContext, opts *lint.LinterSettings) error {
	return nil
}
