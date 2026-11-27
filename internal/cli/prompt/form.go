package prompt

import (
	"bytes"
	"fmt"
	"unicode/utf8"

	"github.com/klarlanguage/klar/internal/cli/ansi"
	"github.com/klarlanguage/klar/internal/cli/icons"
)

type Form struct {
	Accent    string
	TextColor string
	Theme     FormTheme
}

func NewForm(accent string, theme FormTheme) *Form {
	return &Form{Accent: accent, Theme: theme, TextColor: ansi.CodeBold}
}

type FormTheme int

const (
	ConnectedDiamonds FormTheme = iota // Show '◆' in front of questions
	Classic                            // Shows '?' in front of questions
	Single                             // Don't show icons before questions
)

func (f *Form) ShowFormTitle(title, desc string) {
	if ansi.DisableColor {
		fmt.Printf(
			"%s\n%s\n%s\n", title, desc,
			bytes.Repeat([]byte{'='}, utf8.RuneCountInString(title)),
		)
		return
	}
	if f.Theme == ConnectedDiamonds {
		ansi.ColorPrint(ansi.CodeGray, string(icons.BoxTopLeft), " ")
	}
	fmt.Printf("\x1b[34;1;7m %s \x1b[m\n", title)
	if desc != "" {
		f.divider()
		fmt.Println(desc)
	}
	if f.Theme != ConnectedDiamonds {
		fmt.Println()
	}
}

func (f *Form) divider() {
	if f.Theme == ConnectedDiamonds {
		ansi.ColorPrint(ansi.CodeGray, string(icons.BoxVertical), " ")
	}
}

func (f *Form) dividerNewline() {
	if f.Theme == ConnectedDiamonds {
		f.divider()
		fmt.Println()
	}
}

func (f *Form) inputTitle(text, terminator string) {
	if text == "" {
		return
	}
	text = ansi.Color(f.TextColor, text)
	switch f.Theme {
	case ConnectedDiamonds:
		f.dividerNewline()
		fmt.Printf("%s %s%s", ansi.Color(f.Accent, string(icons.Rhombus)), text, terminator)
	case Classic:
		fmt.Printf("%s %s%s", ansi.Color(f.Accent, "?"), text, terminator)
	case Single:
		ansi.ColorPrint(ansi.CodeBold, text, terminator)
	}
}

func (f *Form) Log(text string) {
	switch f.Theme {
	case ConnectedDiamonds:
		f.dividerNewline()
		fmt.Printf("%s %s\n", ansi.Yellow(string(icons.Rhombus)), text)
	case Classic:
		fmt.Printf("%s %s\n", ansi.Yellow("!"), text)
	case Single:
		ansi.ColorPrintln(ansi.CodeBold, text)
	}
}

func (f *Form) invalidTheme() {
	panic(fmt.Sprintf("invalid theme: %d", f.Theme))
}
