package prompt

import (
	"bufio"
	"bytes"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"
	"unicode"

	"github.com/klarlanguage/klar/internal/cli/ansi"
	"golang.org/x/term"
)

var defaultForm = NewForm("", Single)

func Text(msg, desc, def string) string {
	return defaultForm.Text(msg, desc, def)
}

func (f *Form) Text(msg, desc, def string) string {
	f.inputTitle(msg, "")
	if desc != "" {
		ansi.ColorPrint(ansi.CodeDim, " ", desc)
	}
	if def != "" {
		fmt.Print(" ", ansi.Dim("(default: "+def+")"))
	}
	fmt.Print(": ")
	res, err := bufio.NewReader(os.Stdin).ReadString('\n')
	res = strings.TrimSpace(res)
	if err != nil || res == "" {
		return def
	}
	return res
}

func Confirm(msg string, defaultRes bool) bool {
	return defaultForm.Confirm(msg, defaultRes)
}

func (f *Form) Confirm(msg string, defaultRes bool) bool {
	// (y/n) display
	var defaultStr string
	if defaultRes {
		defaultStr = ansi.ColorSprintf(ansi.CodeDim, "(%s/n)", ansi.BoldBrightGreen("Y"))
	} else {
		defaultStr = ansi.ColorSprintf(ansi.CodeDim, "(y/%s)", ansi.BoldBrightRed("N"))
	}

	// Prompt
	f.inputTitle(msg, " ")
	fmt.Print(defaultStr, ": ")
	defer fmt.Println() // Final newline

	// The terminal has to be made raw so we can read a single character without
	// the user pressing Enter
	if oldState, err := term.MakeRaw(int(os.Stdin.Fd())); err == nil {
		defer term.Restore(int(os.Stdin.Fd()), oldState)
	}
	for {
		res := make([]byte, 1)
		os.Stdin.Read(res)
		switch bytes.ToLower(res)[0] {
		case ' ', '\n', '\t', 'r':
			continue
		case 'y', 't', '1':
			return true
		case 'n', 'f', '0':
			return false
		default:
			return defaultRes
		}
	}
}

func ChooseLetter(msg string, opts map[byte]string, def byte) byte {
	return defaultForm.ChooseLetter(msg, opts, def)
}

func (f *Form) ChooseLetter(msg string, opts map[byte]string, def byte) byte {
	f.inputTitle(msg, " ")
	ansi.ColorPrint(ansi.CodeDim, "[")
	for i, char := range slices.Sorted(maps.Keys(opts)) {
		if i > 0 {
			ansi.ColorPrint(ansi.CodeDim, " | ")
		}
		label := opts[char]
		var bold string
		if char == def {
			bold = "<** y!>"
			char = byte(unicode.ToUpper(rune(char)))
		}
		ansi.TagPrintf("%s<y>(%c)</y> %s</>", bold, char, label)
	}
	ansi.ColorPrint(ansi.CodeDim, "] ")
	for {
		// No raw mode
		buf, err := bufio.NewReader(os.Stdin).ReadBytes('\n')
		if err != nil {
			return def
		}
		got := bytes.ToLower(buf)[0]
		if got == def || got == ' ' || got == '\n' ||
			got == '\t' || got == '\x1b' || got == '\r' {
			return def
		}
		if _, ok := opts[got]; ok {
			return got
		}
		f.divider()
		ansi.ColorPrint(ansi.CodeRed, "Please choose a valid option: ")
	}
}
