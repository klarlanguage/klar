package lint

import (
	"github.com/klarlanguage/klar/internal/build"
	"github.com/klarlanguage/klar/internal/config/klarlint"
)

// TODO: Create and use a more common Module type outside of the build package
type Module = build.Module

type Runner struct {
	Diagnostics chan *Diagnostic
	Modules     []*Module
	Configs     []*klarlint.Config
	Linters     []*Linter // All linters needed
}

func NewRunner(mods []*Module, conf *klarlint.File, linters ...*Linter) *Runner {
	r := &Runner{
		Diagnostics: make(chan *Diagnostic),
		Modules:     mods,
		Linters:     linters,
	}
	switch {
	case conf == nil:
	case conf.Config != nil:
		r.Configs = []*klarlint.Config{conf.Config}
	default:
		r.Configs = conf.Configurations
	}
	return r
}

type LinterContext struct {
	Diagnostics chan *Diagnostic
	rules       map[Rule]klarlint.DiagnosticLevel
	*Module
}

func (lc *LinterContext) RuleEnabled(r Rule) bool {
	return lc.rules[r] != klarlint.SuppressLevel
}

func (lc *LinterContext) Report(d *Diagnostic) {
	lc.Diagnostics <- d
}

func (r *Runner) AddLinters(linters ...*Linter) {
	r.Linters = append(r.Linters, linters...)
}

func (r *Runner) Run() {
}
