package klarnew

import (
	"cmp"
	"os"
	"path/filepath"

	"github.com/klarlanguage/klar/internal/cli"
	"github.com/klarlanguage/klar/internal/cli/ansi"
	"github.com/klarlanguage/klar/internal/cli/prompt"
	"github.com/klarlanguage/klar/internal/klarerrs"
	"github.com/klarlanguage/klar/pkg/argparse"
	"golang.org/x/term"
)

type createInfo struct {
	dir string
	// If provided, a subpackage will be created for the specified project path
	rootProject string
	pkgName     string
	exec        bool // False if a library is being created
	flags       flags
}

func Run(c *argparse.Parser) {
	var info createInfo
	// If --name is provided, don't run interactively
	if name := c.Flag("name").String(); name != "" {
		if !validPackageName(name) {
			cli.Failure(makeInvalidPkgNameMsg(name))
		}
		info = createInfo{
			dir:     cmp.Or(c.ArgByName("dir"), "."),
			pkgName: name,
			exec:    c.Flag("type").EnumValue() == 2,
			flags:   makeFlags(c),
		}
	} else if !term.IsTerminal(int(os.Stdout.Fd())) {
		cli.Failure("Can't run 'klar new' interactively. Try providing the '--name' flag. Use '--help' for more information.")
	} else {
		info = runWizard(c)
	}
	info.create()
}

func runWizard(c *argparse.Parser) (info createInfo) {
	info.flags = makeFlags(c)

	form := prompt.NewForm(ansi.CodeBlue, prompt.ConnectedDiamonds)
	form.ShowFormTitle("New project", "")

	// 1. Prompt for directory?
	if info.dir = c.ArgByName("dir"); info.dir == "" {
		info.dir = form.Text("Where to create the project?", "", ".")
	}
	var err error
	if info.dir, err = filepath.Abs(info.dir); err != nil {
		cli.Failure("Failed to get absolute path:", err)
	}

	// 2. Package name
	nameFlag := c.Flag("name").String()
	if nameFlag != "" && !validPackageName(nameFlag) {
		form.Log(ansi.Red(makeInvalidPkgNameMsg(nameFlag)))
		nameFlag = "" // Ask the user to pick a new name
	}
	if nameFlag == "" {
		def := generateValidPkgName(filepath.Base(info.dir))
		for {
			info.pkgName = form.Text("What will your package be named?", "", def)
			if !validPackageName(info.pkgName) {
				form.Log(ansi.Red(makeInvalidPkgNameMsg(info.pkgName)))
				continue
			}
			break
		}
	}

	// 3. Project type
	// TODO: Use a choose from menu input instead
	info.exec = form.ChooseLetter("Project type", map[byte]string{
		'e': "Executable/command", 'l': "Library",
	}, 'e') == 'e'

	// 4. Addons
	addons := []struct {
		label, cliFlag string
		option         flags
	}{
		{"Tests", "with-tests", addTests},
		{"'klar.build' configuration", "with-klarbuild", addKlarBuild},
		{"GitHub CI", "with-github-ci", addGitHubAction},
	}
	opts := make(map[string]bool, len(addons))
	for _, addon := range addons {
		opts[addon.label] = c.Flag(addon.cliFlag).Bool()
	}
	form.Checkboxes("What else would you'd like in your project?", "", opts, false)
	for _, addon := range addons {
		if on := opts[addon.label]; on {
			info.flags |= addon.option
		} else {
			// Because we applied the CLI-provided flags already, we have to unapply the flag
			info.flags &^= addon.option
		}
	}

	// 5. Ask to initialize a Git repo
	// TODO: Don't if there's already a repo
	if form.Confirm("Initialize a Git repo?", true) {
		info.flags |= initGitRepo
	}

	return info
}

// TODO
func validPackageName(name string) bool {
	return true
}

// TODO
func generateValidPkgName(from string) string {
	return from
}

func makeInvalidPkgNameMsg(name string) string {
	if name == "" {
		return "The package name can't be empty"
	}
	return klarerrs.Quote(name) + ` isn't a valid Klar package name. A package name can contain letters, numbers, underscores, and optionally 1 slash in between.`
}
