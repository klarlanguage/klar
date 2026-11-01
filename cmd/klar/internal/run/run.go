package run

import (
	"os"

	"github.com/klarlanguage/klar/internal/cli"
	"github.com/klarlanguage/klar/internal/cli/ansi"
	"github.com/klarlanguage/klar/pkg/argparse"
)

func Run(r *argparse.Parser) {
	input := r.ArgByName("input")
	if input != "" {
		if _, err := os.Stat(input); err != nil && os.Args[1] != "run" && input[0] != '.' {
			// Run via `klar ./file` rather than `klar run ./file`. Show
			// a more helpful error
			cli.ColorErrorfln("<**>Can't find a Klar command named <c!>%s</c!></**>", input)
			ansi.TagPrintln("\nRun <c!>klar help</c!> to see available commands.")
		}
	}
}

const LongDescription = ``
