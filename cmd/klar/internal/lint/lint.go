package lint

import (
	buildcmd "github.com/klarlanguage/klar/cmd/klar/internal/build"
	"github.com/klarlanguage/klar/internal/build"
	"github.com/klarlanguage/klar/internal/cli"
	"github.com/klarlanguage/klar/pkg/argparse"
)

func Run(p *argparse.Parser) {
	if p.Flag("json").Set {
		if p.Flag("diff").Set {
			cli.Failure("'--json' and '--diff' can't be used together")
		}
		if p.Flag("compact").Set {
			cli.Failure("'--json' and '--compact' can't be used together")
		}
	}

	cwd, err := build.Cwd()
	if err != nil {
		cli.FailureError(err)
	}
	c := build.NewCompiler(build.ModeAnalyze, cwd)
	pc := build.NewProjectCompiler(c)
	res, err := buildcmd.RunBuild(pc, p.VarArgByName("inputs"), &buildcmd.CLIFlags{
		Parser:         p,
		ConfigFlagName: "", // TODO: No --klarbuild flag yet
	})
	switch {
	case false && p.Flag("json").Bool():
		// TODO: Output build failure as JSON
	case err != nil:
		buildcmd.HandleCriticalBuildError(err, c)
	case len(res.Errors) > 0:
		buildcmd.ShowResult(res, c)
		cli.Exit(1)
	}
}
