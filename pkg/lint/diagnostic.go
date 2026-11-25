package lint

import (
	"fmt"

	"github.com/klarlanguage/klar/internal/klarerrs"
	"github.com/klarlanguage/klar/internal/ranges"
	"github.com/klarlanguage/klar/pkg/klarerrors/reporter"
)

// A Diagnostic is a warning or error reported from a linter. The diagnostic may be
// mutated outside the linter.
//
// Diagnostic implements [reporter.Error].
type Diagnostic struct {
	Rule  Rule
	Msg   string
	File  string
	Range ranges.Range

	Label      string               // Text to show beside the range
	Desc       string               // Extended message to show below the error
	Highlights []klarerrs.Highlight // Additional highlights to display
	Details    []klarerrs.Detail    // Additional files context to display
	Hints      []klarerrs.Hint      // Hints to display

	Params map[string]any
}

var _ reporter.Error = (*Diagnostic)(nil)

func (d *Diagnostic) Title() string                         { return "Error" }
func (d *Diagnostic) Message() string                       { return d.Msg }
func (d *Diagnostic) Description() string                   { return d.Desc }
func (d *Diagnostic) ErrorCode() string                     { return d.Rule.String() }
func (d *Diagnostic) IsWarning() bool                       { return false }
func (d *Diagnostic) FilePath() string                      { return d.File }
func (d *Diagnostic) Location() ranges.Range                { return d.Range }
func (d *Diagnostic) MainHighlight() string                 { return d.Label }
func (d *Diagnostic) ErrorDetails() []klarerrs.Detail       { return d.Details }
func (d *Diagnostic) ErrorHighlights() []klarerrs.Highlight { return d.Highlights }
func (d *Diagnostic) ErrorHints() []klarerrs.Hint           { return d.Hints }

// MarkAutoFix marks the i'th hint in [Diagnostic.Hints] as the autofix for
// the diagnostic. he hint must have a diff.
func (d *Diagnostic) MarkAutoFix(i int) {
	// Bounds check + ensure the hint has a diff
	if d.Hints[i].Diff == nil {
		panic(fmt.Sprintf("Diagnostic.MarkAutoFix(%d): hint #%[1]d doesn't have a diff", i))
	}
	if d.Params == nil {
		d.Params = make(map[string]any)
	}
	d.Params["autofixIndex"] = i
}
