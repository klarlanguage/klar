package build

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/klarlanguage/klar/internal/build"
	"github.com/klarlanguage/klar/internal/cli"
	"github.com/klarlanguage/klar/internal/cli/ansi"
	"github.com/klarlanguage/klar/internal/cli/icons"
	"github.com/klarlanguage/klar/internal/config/klarbuild"
	"github.com/klarlanguage/klar/internal/klarerrs/jsonerrors"
	"github.com/klarlanguage/klar/internal/module"
	"github.com/klarlanguage/klar/internal/target"
	"github.com/klarlanguage/klar/internal/util"
	"github.com/klarlanguage/klar/pkg/argparse"
	"github.com/klarlanguage/klar/pkg/klon"
)

// Run executes the "klar build" command.
func Run(p *argparse.Parser) {
	inputArgs := p.VarArgByName("inputs")
	cwd, err := build.Cwd()
	if err != nil {
		cli.FailureError(err)
	}
	c := build.NewCompiler(build.ModeBuild, cwd)
	pc := build.NewProjectCompiler(c)

	// 1. Logging
	jsonOutput := p.Flag("json-output").Bool()
	if err := build.SetLogger(c, p.Flag("verbose").Bool(), jsonOutput); err != nil {
		cli.FailureError(err)
	}
	defer func() {
		if err := c.CloseLogger(); err != nil {
			cli.Failure("Failed to write log file:", err)
		}
	}()
	if !jsonOutput {
		c.Progress = NewBuildStatus(cwd)
	}

	// Steps 2-4 are shared by other commands like run, test, lint, and format
	res, err := RunBuild(pc, inputArgs, &CLIFlags{
		Parser:         p,
		ConfigFlagName: "config",
		JSFlags:        jsFlags,
	})

	// 5. Print error/success messages
	// =========

	// If we're showing the compiler's progress, clear the line before showing errors
	c.Progress.Clear()
	switch {
	case jsonOutput:
		if err := jsonerrors.WriteTo(os.Stdout, res, err); err != nil {
			cli.Error("Failed to write JSON errors:", err)
		}
		fmt.Println() // jsonerrors doesn't append a newline
		if err != nil || len(res.Errors) > 0 {
			cli.Exit(1)
		}
	case err != nil:
		// Critical error
		HandleCriticalBuildError(err, c)
	default:
		// Successes, errors, and/or warnings
		if ok := ShowResult(res, c); !ok {
			cli.Exit(1)
		}
	}
}

type CLIFlags struct {
	*argparse.Parser
	ConfigFlagName string // Flag that specifies a klar.build path, or empty if unsupported
	JSFlags        []string
}

func RunBuild(
	pc *build.ProjectCompiler, inputArgs []string, flags *CLIFlags,
) (res *build.Result, err error) {
	c := pc.Compiler
	c.StartTime = time.Now() // Start timer at resolution process

	// 2. --config flag
	// Providing it will use the klar.build for all inputs, otherwise, klar.build
	// files will be searched inside each input and module.
	if flags != nil && flags.ConfigFlagName != "" {
		configFlag := flags.Flag(flags.ConfigFlagName)
		if configFlag.String() != "" {
			configPath := configFlag.String()
			var warn []*klon.Error
			if c.ForcedKlarBuild, warn, err = klarbuild.Parse(configPath); err != nil {
				c.FailWithError(err)
			}
			c.PrintKlonWarnings(warn, configPath)
		} else if configFlag.Set {
			// Empty --config flag: use default klar.build
			c.ForcedKlarBuild = klarbuild.Default()
		}
	}
	if flags != nil {
		// Apply CLI overrides, whether the --config flag was used or not
		c.ForcedKlarBuild, c.KlarBuildOverrides = ParseJSFlags(flags, c.ForcedKlarBuild, true)

		// Apply CLI override for the target, since it isn't considered a JS flag
		if _, ok := flags.FlagDefs["target"]; ok && flags.Flag("target").Set {
			if c.ForcedKlarBuild == nil {
				c.ForcedKlarBuild = &klarbuild.File{}
			}
			c.ForcedKlarBuild.Target = flags.Flag("target").EnumValue().(target.Target)
			c.KlarBuildOverrides |= klarbuild.TargetOverriden
		}
	}

	// 3. Resolve command-line inputs
	pc.Inputs = make([]*build.Input, 0, max(len(inputArgs), 1))
	addInput := func(path string) {
		input, err := pc.ResolveInput(path)
		if err != nil {
			if err, ok := err.(*build.FilesystemError); ok && err.IsNotExist() {
				cli.ErrNotFound(path, "input")
			}
			c.FailWithError(err)
		}
		pc.Inputs = append(pc.Inputs, input)

		// Hide progress if stdin is an input
		if input.Kind == build.KindStdin {
			// Clear the line created by ResolvingInput
			c.Progress.Clear()
			c.Progress = build.HiddenProgress{}
		}
	}
	for i, path := range inputArgs {
		if path == "" {
			continue
		}
		c.Progress.ResolvingInput(path, i+1, len(inputArgs))
		addInput(path)
	}
	if len(pc.Inputs) == 0 {
		// If no inputs were provided, compile the current *package*
		pkgPath, _ := module.PackageRoot(c.WorkDir)
		c.Info("Building current package", slog.String("package", pkgPath))
		addInput(pkgPath)
	}

	// 4. Compile!
	// TODO: error if --output is file and there are multiple inputs
	res, err = pc.Compile()
	return
}

// HandleCriticalBuildError handles an error returned from [build.ProjectCompiler.Compile]
func HandleCriticalBuildError(err error, c *build.Compiler) {
	switch err := err.(type) {
	case *build.InterfaceError:
		// For InterfaceErrors: print a prettier error
		c.PrintInterfaceOrKlonError(err)
		cli.Exit(1)
	case *build.FilesystemError:
		cli.FailureError(err)
	default:
		// Errors should be a struct such as InterfaceError or FilesystemError
		panic(fmt.Sprintf("error %T should be wrapped: %[1]v", err))
	}
}

// TODO: Change uses of the word 'build' when running lint, format, test, or run.
func ShowResult(res *build.Result, c *build.Compiler) (ok bool) {
	var (
		warnCount, errCount = len(res.Warnings), len(res.Errors)
		icon, format        string
	)
	formatIcon := func(icon rune, color byte) string {
		return fmt.Sprintf("<%c>%c</%[1]c>", color, icon)
	}
	formatCount := func(n int, kind string) string {
		switch {
		case n == 1:
			return fmt.Sprintf("%d %s", n, kind)
		case kind == "error" && res.IsMaxErrors:
			return fmt.Sprintf("%d+ %ss", build.MaxErrors, kind)
		}
		return fmt.Sprintf("%d %ss", n, kind)
	}
	switch {
	case errCount == 0 && warnCount == 0:
		// Succeeded
		icon, format = formatIcon(icons.Check, 'g'), "<g!>succeeded</g!>"
	case errCount == 0 && warnCount > 0:
		// Succeeded with warnings
		icon, format = formatIcon(icons.Warning, 'y'),
			"<y!>succeeded</y!> with <y!>"+formatCount(warnCount, "warning")+"</y!>"
	case errCount > 0 && warnCount == 0:
		// Failed
		icon = formatIcon(icons.ThinXLarge, 'r')
		format = "<r!>failed</r!> with <r!>" + formatCount(errCount, "error") + "</r!>"
	case errCount > 0 && warnCount > 0:
		// Failed with errors and warnings
		icon = formatIcon(icons.ThinXLarge, 'r')
		format = "<r!>failed</r!> with <r!>" + formatCount(errCount, "error") +
			"</r!> and <y!>" + formatCount(warnCount, "warning") + "</y!>"
	default:
		panic(fmt.Sprintf("unreachable: %d errors and %d warnings", errCount, warnCount))
	}
	dur := util.FormatDuration(res.Elapsed)
	fmt.Fprintln(os.Stderr, ansi.Colorize(
		"<**>"+icon+" Build "+format+"</**> in <c>"+dur+"</c>",
	))

	if errCount > 0 || warnCount > 0 {
		fmt.Fprintln(os.Stderr)
	}
	c.PrintAllErrors(res.Errors)
	c.PrintAllErrors(res.Warnings)
	if res.IsMaxErrors {
		fmt.Fprintln(os.Stderr, ansi.BoldBrightRed(
			ansi.Underline("\nBuild stopped early due to too many errors"),
		))
	}
	if errCount > 0 {
		return false
	}
	return true
}

// ParseJSFlags parses flags from r into o.
func ParseJSFlags(r *CLIFlags, kb *klarbuild.File, hasJSInput bool) (
	_ *klarbuild.File, overrides klarbuild.Overrides,
) {
	for _, setting := range r.JSFlags {
		v, ok := r.Flags[setting]
		if !ok || !v.Set {
			continue
		}
		if kb == nil {
			kb = &klarbuild.File{}
		}
		// Check if a JavaScript flag was used when no inputs target JS
		if !hasJSInput {
			// TODO: I would like to use the first JS flag the user provided
			cli.Failure(fmt.Sprintf(
				"Can't use JavaScript flag '--%s' on the %s target",
				setting, kb.Target.Name(),
			))
		}
		switch setting {
		case "declaration":
			kb.JS.Declaration = v.Value.(bool)
		case "minify":
			kb.JS.Minify = v.Bool()
		case "inline-sourcemap":
			if v.Bool() {
				kb.JS.Sourcemap = klarbuild.SourceMapInline
			}
		case "sourcemap":
			if v.Bool() {
				kb.JS.Sourcemap = klarbuild.SourceMapEnabled
			} else {
				kb.JS.Sourcemap = klarbuild.SourceMapDisabled
			}
		case "jsdoc":
			kb.JS.JSDoc = v.Bool()
		case "copy-node-modules":
			kb.JS.CopyNodeModules = v.Bool()
		case "banner":
			kb.JS.Banner = v.String()
		case "bundle":
			kb.JS.Bundle = v.EnumValue().(klarbuild.BundleMode)
		case "declaration-path":
			kb.JS.DeclarationPath = v.String()
		default:
			panic("unhandled flag: " + setting)
		}
	}
	return kb, overrides
}

const LongDescription = `Compiles Klar source files at the provided file or module paths. If none are provided, inputs defined in 'klar.build', the build configuration file, are used.

An input passed to 'klar build' can be a directory path, to compile a module or package; a file path, to compile an individual file; '-', to read from standard input and compile it as an individual file; or a name prefixed with '@' to resolve a module by its name and compile it.

A 'klar.build' is used to customize the build process and how files are compiled. For more information on build settings, run 'klar help klar.build'.
For each input, its closest 'klar.build' file is used to configure the build. The '--config' flag can be used to override the configuration for all inputs. If the '--config' flags is provided, but empty, the default settings are used without looking for 'klar.build' files. Common build options are provided as flags to override klar.build options.

Currently, Klar files can be compiled to JavaScript.`
