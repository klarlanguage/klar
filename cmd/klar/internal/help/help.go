package help

import (
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/klarlanguage/klar/internal/cli"
	"github.com/klarlanguage/klar/internal/cli/ansi"
	"github.com/klarlanguage/klar/internal/util"
	"github.com/klarlanguage/klar/pkg/argparse"
	"golang.org/x/term"
)

// Existing CLI commands already handled in main.go
func Run(c *argparse.Parser) {
	name := c.ArgByName("command-or-topic")
	switch {
	case c.Flag("topics").Bool():
		ListTopics()
	case c.Flag("error").Bool():
		ExplainError(name)
	case c.Flag("docs").Bool():
		if name == "" {
			cli.Failure("Please provide a command or topic name")
		}
		// Show documentation online
		cli.Eprintf(
			"Searching documentation for %s on %s\n",
			ansi.Cyan(name), ansi.Magenta("site.url"),
		)
	default:
		if name == "" {
			cli.Failure("Please provide a command or topic name")
		}
		if t := findTopic(name); t != nil {
			DisplayTopic(Topics[name])
			break
		}
		cli.Error("Can't find a CLI command or topic named " + ansi.Cyan(name))
		cli.HintIndent(
			"To search online documentation, use the " + ansi.Cyan("-d") + " flag.",
		)
		cli.Exit(1)
	}
}

func findTopic(query string) *Topic {
	if t, ok := Topics[strings.ToLower(query)]; ok {
		return t
	}
	for _, t := range Topics {
		if strings.EqualFold(t.Title, query) {
			return t
		}
	}
	return nil
}

func DisplayTopic(t *Topic) {
	termWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		termWidth = 0
	}
	ansi.ColorPrintln(ansi.CodeBold, t.Title)
	fmt.Println()
	util.Wrap(t.Description, util.WrapAllWriter(os.Stdout), termWidth, termWidth, 0)
	fmt.Println()
	if len(t.SeeAlso) > 0 {
		ansi.ColorPrintln(ansi.CodeBold, "\nSee also:")
		for _, name := range t.SeeAlso {
			fmt.Println("    ", ansi.Cyan(name))
		}
		ansi.TagPrintln("Learn more by running 'klar help <topic>'")
	}
}

func ListTopics() {
	ansi.ColorPrintln(ansi.CodeBold, "Help topics:")
	for _, id := range slices.Sorted(maps.Keys(Topics)) {
		t := Topics[id]
		ansi.TagPrintfln(
			"\n%s <d>(%s)</d>\n    %s",
			ansi.BoldBrightBlue(ansi.Hyperlink(t.Title, t.URL)), id,
			t.Description,
		)
	}
}

var Flags = argparse.NewParser("[command-or-topic]").
	BoolFlag("docs", "Show online documentation for the command or topic", false, "d", "web").
	BoolFlag("topics", "List the available topics", false).
	BoolFlag("error", "Explain a compiler error", false, "e")

const LongDescription = `Shows details about a Klar CLI command or topic, including its usage and available options. This includes the 'klar help' command itself, which displays this message.

Using 'klar help command' is equivalent to 'klar command --help'. If no command is passed, it shows the list of CLI commands and options.

When the '--docs' or '-d' flag is passed, the full documentation for the topic is opened in the browser.`
