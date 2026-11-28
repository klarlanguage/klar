package klarlint

type File struct {
	*Config        // Global configuration
	Configurations []*Config
}

type Config struct {
	// Files to apply these settings to. If this configuration is at the top-level,
	// this controls what files will be checked. Glob patterns are supported. Use
	// `!` to exclude a file, directory, or module from being checked.
	Files []string
	// If you want to use a linter outside of Klar's builtin rules, such as one
	// from a Glas package or your own, pass the import path of the module.
	ExternalLinters []string
	// Rules and linters to enable, and their severity. Each key can be a rule,
	// linter, or [category](https://klarlanguage.github.io/docs/klarlint#categories).
	// To pass settings to a rule, set the value to a tuple with the severity, then
	// the options:
	//
	// ```klon
	// rules:
	//   - fileNaming: [true, { basedOnTypes: false }]
	// ```
	Rules map[string]RuleSetting

	// TODO: In the CLI, categories are counted as rules/linters. Should
	// the 'rules' option control categories, or they should be separated
	// into their own property?

	// Whether to automatically apply fixes and suggestions wherever possible. Not all
	// diagnostics will have automatic fixes, and only fixes that don't change your
	// code's behaviour will be applied.
	AutoFix bool
}

type RuleSetting struct {
	Level   DiagnosticLevel `options:"DiagnosticLevel"`
	Options map[string]any  // TODO: Use some equivalent to jsontext.Value
}

type DiagnosticLevel int

const (
	DefaultLevel  DiagnosticLevel = iota // Enable the rule with its intended severity
	SuppressLevel                        // Disable the rule and don't check for it
	ErrorLevel                           // Report an error and fail lint
	WarningLevel                         // Report a warning without failing lint
)
