package run

import "github.com/klarlanguage/klar/pkg/argparse"

var Flags = argparse.NewParser("[input]").
	StringFlag("runtime", "The JavaScript runtime to use", "command", "", "R")
