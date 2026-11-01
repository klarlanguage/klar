package parser

import (
	"github.com/klarlanguage/klar/internal/ast"
	"github.com/klarlanguage/klar/internal/klarerrs"
	"github.com/klarlanguage/klar/internal/lexer"
	"github.com/klarlanguage/klar/internal/parser"
)

type (
	Error   = klarerrs.Error
	Options = parser.Options
)

func Parse(tokens []lexer.Token, options *parser.Options) (
	program *ast.Program, errors []*Error,
) {
	p := parser.New(tokens, options)
	program = p.Parse()
	errors = p.Errors
	return
}

func ParseString(src string) (program *ast.Program, errs []*Error) {
	return Parse(TokenizeString(src), nil)
}
