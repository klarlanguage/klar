package format

import "github.com/klarlanguage/klar/pkg/argparse"

func Run(c *argparse.Parser) {}

var Flags = argparse.NewParser("[...inputs]").
	StringFlag("config",
		"Path to a 'klar.format' configuration (use '-' to read from standard input)",
		"file", "", "c",
	)

const LongDescription = ``
