package klarnew

import "github.com/klarlanguage/klar/pkg/argparse"

var Flags = argparse.NewParser("[dir]").
	EnumFlag("type", "The type of package to create", "", map[string]any{
		"library": 1, "lib": 1,
		"executable": 2, "exec": 2, "cmd": 2, "command": 2,
	}, "executable", "t").
	StringFlag("name", "The name of the package (defaults to the directory's name)", "", "", "n").
	BoolFlag("with-tests", "Add test templates to the project", true).
	BoolFlag("with-klarbuild", "Add a 'klar.build' configuration for the project", false).
	BoolFlag("git-init", "Initialize a new Git repository for the project", false).
	BoolFlag("with-github-ci", "Create a GitHub Actions workflow to test and lint the project", false)

// TODO: Possibly add a --with-npm-package flag to create a package.json file for the
// purpose of creating a JS library

type flags uint8

const (
	initGitRepo flags = 1 << iota
	addTests
	localCache
	addKlarBuild
	addGitHubAction
)

func makeFlags(c *argparse.Parser) (f flags) {
	for _, pair := range []struct {
		name   string
		option flags
	}{
		{"with-tests", addTests},
		{"with-klarbuild", addKlarBuild},
		{"local-cache", localCache},
		{"with-github-ci", addGitHubAction},
		{"git-init", initGitRepo},
	} {
		if c.Flag(pair.name).Bool() {
			f |= pair.option
		}
	}
	return f
}

const LongDescription = ""
