package analysis

import (
	"strconv"

	"github.com/klarlanguage/klar/internal/ast"
	"github.com/klarlanguage/klar/internal/ranges"
)

// TODO: This is a fast and quick implementation. It is nowhere near complete.
// TODO: This should check for duplicate and unreachable patterns
func (c *Checker) checkWhenExhaustiveness(
	whenExpr *ast.WhenExpression, subjects []*whenSubject, required bool, fid FileID,
) bool {
	subj := subjects[0]
	switch {
	case len(subjects) > 1:
		return false // TODO: Not implemented yet
	case subj.Default != nil:
		// A default case makes the 'when' exhaustive, unless it has a fallthrough
		return true
	case subj.IsImplicitTrue():
		return false // No subject requires a default case
	}

	var ec exhaustiveChecker
	switch subj.Kind() {
	case BoolType:
		ec = &boolExhaustiveChecker{}
	case KindEnum:
		enum := As[EnumParenter](subj.Type)
		ec = newEnumExhaustiveChecker(enum.EnumParent())
	default:
		return false // Not implemented for non-enums or bools
	}
	for _, cs := range whenExpr.Cases {
		for _, subjExpr := range cs.Options {
			for _, opt := range subjExpr {
				optType := c.Module.Info.Expressions[opt].Type.(*WhenPattern).Expr
				if !isConstant(optType.Type) {
					continue
				}
				var item any
				switch subj.Kind() {
				case BoolType:
					item = optType.ConstExpr().Value.ConstValue()
				case KindEnum:
				}
				if dupe, orig := ec.addPattern(item, opt.GetRange()); dupe {
					_ = orig
				}
			}
		}
	}
	missing := ec.missingCases()
	// If `required` is true, report an error if inexhaustive
	return len(missing) == 0
}

type exhaustiveChecker interface {
	addPattern(any, ranges.Range) (duplicate bool, orig ranges.Range)
	missingCases() []string
}

type boolExhaustiveChecker struct {
	covered map[bool]ranges.Range
}

func (bec *boolExhaustiveChecker) addPattern(
	val any, r ranges.Range,
) (duplicate bool, orig ranges.Range) {
	if bec.covered == nil {
		bec.covered = make(map[bool]ranges.Range, 2)
	}
	orig, duplicate = bec.covered[val.(bool)]
	if !duplicate {
		bec.covered[val.(bool)] = r
	}
	return
}

func (bec *boolExhaustiveChecker) missingCases() (missing []string) {
	if len(bec.covered) == 2 {
		return nil
	}
	for _, bool := range [...]bool{false, true} {
		if _, ok := bec.covered[bool]; !ok {
			missing = append(missing, strconv.FormatBool(bool))
		}
	}
	return
}

type enumExhaustiveChecker struct {
	enum    *Enum
	covered map[*EnumItem]ranges.Range
}

func newEnumExhaustiveChecker(enum *Enum) *enumExhaustiveChecker {
	return &enumExhaustiveChecker{enum: enum}
}

func (eec *enumExhaustiveChecker) addPattern(
	val any, r ranges.Range,
) (duplicate bool, orig ranges.Range) {
	if eec.covered == nil {
		eec.covered = make(map[*EnumItem]ranges.Range, len(eec.enum.Items))
	}
	ei := val.(*EnumItem)
	orig, duplicate = eec.covered[ei]
	if !duplicate {
		eec.covered[ei] = r
	}
	return
}

func (eec *enumExhaustiveChecker) missingCases() (missing []string) {
	for _, ei := range eec.enum.Items {
		if _, ok := eec.covered[ei]; !ok {
			missing = append(missing, "."+ei.Name)
		}
	}
	return
}
