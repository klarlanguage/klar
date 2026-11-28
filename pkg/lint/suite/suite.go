// Package suite defines the default linters used by the `klar lint` command,
// and the categories of linters it supports.
package suite

import (
	"github.com/klarlanguage/klar/pkg/lint"
	"github.com/klarlanguage/klar/pkg/lint/linters/filenames"
)

// The linters used by `klar lint` that are enabled by default.
var DefaultLinters = []*lint.Linter{
	filenames.Linter,
}
