package lint

type Linter struct {
	Name             string // Conventionally in camelCase
	ShortDescription string
	LongDescription  string
	DocsURL          string // Optional URL for documentation

	Rules []Rule // The rules the linter provides

	// Settings for the linter and/or rules, or nil if neither accept settings
	Settings func() *LinterSettings

	Run func(lc *LintContext, opts *LinterSettings) error

	// TODO: Linter dependencies (need to run before this) and option to allow running
	// on a program with errors
}

type LinterSettings struct {
	// The object to decode linter-wide settings to, or nil if the
	// linter doesn't accept settings.
	Linter       any
	RuleSettings map[Rule]any // Settings for individual rules, or nil
}

type Rule struct {
	Linter *Linter
	Code   string
}

func (r Rule) String() string { return r.Linter.Name + "/" + r.Code }

// NewRule returns a new [Rule] associated with the linter.
func (l *Linter) NewRule(code string) Rule {
	r := Rule{l, code}
	l.Rules = append(l.Rules, r)
	return r
}

type LintContext struct {
	Diagnostics []*Diagnostic
}

func (lc *LintContext) RuleEnabled(r Rule) bool {
	return true
}

func (lc *LintContext) Report(d *Diagnostic) {
}
