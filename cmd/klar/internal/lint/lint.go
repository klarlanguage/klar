package lint

import "github.com/klarlanguage/klar/pkg/argparse"

func Run(r *argparse.Parser) {
}

var Flags = argparse.NewParser("[inputs...]").
	ListFlag("rules", "Rules to check for (prefix a rule with '!' to disable)", "rule", nil, "r").
	ListFlag("external-linters", "Import paths of linters to include rules from", "import-path", nil, "l").
	BoolFlag("watch", "Recheck when files in the project are changed", false, "w").
	ListFlag("settings", "Settings to pass to linters", "rule.setting=value", nil, "s").
	BoolFlag("fix", "Automatically apply fixes wherever possible", false, "f").
	BoolFlag("json", "Output results as JSON", false).
	StringFlag("config", "Path to 'klar.lint' config ('-' for standard input)", "path", "", "c").
	BoolFlag("verbose", "Enable verbose logging", false, "v").
	ListFlag("error", "Display some diagnostics as errors", "rule", nil, "E").
	ListFlag("warn", "Display some diagnostics as warnings", "rule", nil, "W").
	BoolFlag("list", "List the enabled linters and available rules (instead of running)", false).
	// Note: The config is created/updated based on the other CLI flags. Lint still runs.
	BoolFlag("init-config", "Also create or update a 'klar.lint' config with current settings", false)

const LongDescription = ""
