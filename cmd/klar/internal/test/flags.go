package test

import "github.com/klarlanguage/klar/pkg/argparse"

var Flags = argparse.NewParser("[input]").
	StringFlag("runtime",
		"The JavaScript runtime to use (may be a path to an executable)", "command", "", "R",
	).
	StringFlag("config",
		"Path to a 'klar.build' config file. By default, ones inside each input are used",
		"file", "", "c",
	)

const LongDescription = ``
