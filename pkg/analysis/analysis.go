package analysis

import (
	"github.com/klarlanguage/klar/internal/analysis"
	"github.com/klarlanguage/klar/internal/ast"
	"github.com/klarlanguage/klar/internal/klarerrs"
	"github.com/klarlanguage/klar/internal/target"
	"github.com/klarlanguage/klar/internal/version"
)

type CheckOptions struct {
	File        string
	Target      target.Target
	KlarVersion version.Version
	Path        string
	*analysis.Options
}

func CheckProgram(prog *ast.Program, opts CheckOptions) []*klarerrs.Error {
	mod := analysis.NewModule(
		opts.File, opts.Path, nil,
		opts.KlarVersion, []target.Target{opts.Target},
	)
	files := map[string]*ast.Program{opts.File: prog}
	c := analysis.NewChecker(mod, files, opts.Options)
	c.Check()
	return c.Errors
}
