package codegen

import (
	"github.com/klarlanguage/klar/internal/analysis"
	"github.com/klarlanguage/klar/internal/ast"
	"github.com/klarlanguage/klar/internal/ir/jsir"
	"github.com/klarlanguage/klar/internal/lexer"
)

func (g *Generator) convertStatement(stmt ast.Statement, ctx *analysis.Context) jsir.Statement {
	switch stmt := stmt.(type) {
	case *ast.VariableDeclaration:
		_ = stmt
	case *ast.WhileStatement:
		while := &jsir.WhileStatement{}
		var cond jsir.Expression
		if stmt.Condition != nil {
			cond = g.convertExpression(stmt.Condition)
		}

		// If the statement has a binding:
		// 	while data := read(), data.length > 0 {}
		// is converted to:
		// 	let data
		//  while ((data = read()).length > 0) {}
		// or:
		// 	let data
		//  while (data = read(), data.length > 0)
		var bindingDecl *jsir.VariableDeclaration
		var varAssign *jsir.AssignmentExpression
		if stmt.Variable != nil {
			bindingDecl = &jsir.VariableDeclaration{}
			varAssign = &jsir.AssignmentExpression{
				Assignee: bindingDecl.Variable,
				Operator: jsir.OpEqual,
				Value:    g.convertExpression(stmt.Variable.Values[0]),
			}
		}

		switch {
		case stmt.Variable == nil:
			while.Condition = cond
		case stmt.Condition == nil:
			while.Condition = varAssign
		default:
			while.Condition = &jsir.CommaExpression{[]jsir.Expression{varAssign, cond}}
		}

		while.Body = g.convertBlock(stmt.Body, g.Module.Info.Blocks[stmt].Context)
		if bindingDecl != nil {
			return &jsir.StatementChain{[]jsir.Statement{bindingDecl, while}}
		}
		return while
	case *ast.ForStatement:
	// for 10 {}           -> for (let i = 0; i < 10; i++) {}
	// for i in 1...10     -> for (let i = 1; i <= 10; i++) {}
	// for i in 1..<10...2 -> for (let i = 1; i < 10; i += 2) {}
	//
	// for item in [1, 2]    -> for (const item of [1, 2])
	// for [1, 2]            -> for (let i = 0; i < [1, 2].length; i++) {}
	// for i, _ in [1, 2]    -> for (const k in [1, 2])
	// for i, item in [1, 2] -> for (const [i, item] of [1, 2].entries())
	//                       or for (const i in [1, 2]) const item = [1, 2][i]
	//
	// for k, v in #{'x': 1} -> for (const [k, v] of new Map(...))
	// for k in #{'x': 1}    -> for (const [k] of new Map(...))
	//                       or for (const k of new Map(...).keys())
	case *ast.AssignmentStatement:
		var commaExpr *jsir.CommaExpression
		if len(stmt.Assignee) > 1 {
			commaExpr = &jsir.CommaExpression{make([]jsir.Expression, len(stmt.Assignee))}
		}
		var singleRHS jsir.Expression
		if stmt.IsSingleRHS() {
			singleRHS = g.convertExpression(stmt.Values[0])
		}
		for i, lhs := range stmt.Assignee {
			// TODO: JS doesn't support object destructures in assignments
			jsLHS := g.convertExpression(lhs)
			assignExpr := &jsir.AssignmentExpression{
				Assignee: jsLHS,
				Value:    singleRHS,
				Operator: ConvertOperator(stmt.Operator.Kind),
			}
			if !stmt.IsSingleRHS() {
				assignExpr.Value = g.convertExpression(stmt.Values[i])
			}
			if len(stmt.Assignee) == 1 {
				return &jsir.ExpressionStatement{assignExpr}
			}
			commaExpr.Items[i] = assignExpr
		}
		return &jsir.ExpressionStatement{commaExpr}
	case *ast.ReturnStatement:
		var expr jsir.Expression
		if stmt.Value != nil {
			expr = g.convertExpression(stmt.Value)
		}
		return &jsir.ReturnStatement{expr}
	case *ast.NextStatement:
		if stmt.Label != nil {
			return &jsir.ContinueStatement{stmt.Label.Name}
		}
		return &jsir.ContinueStatement{}
	case *ast.StopStatement:
		if stmt.Label != nil {
			return &jsir.BreakStatement{stmt.Label.Name}
		}
		return &jsir.BreakStatement{}
	case *ast.ExpressionStatement:
		return &jsir.ExpressionStatement{g.convertExpression(stmt.Expression)}
	case *ast.ImportStatement:
		// Don't know what to do yet
		// Because dead code elimination and bundling have to occur, the import
		// statements could be added only during the bundling phase.
		// Anyways, a few things to note about converting Klar imports to JS:
		// - When generating .dts files, only import modules used for types
		// - Differenciate between project modules and NPM/Node modules
	}
	return &jsir.FunctionDeclaration{Name: randomName(6), Body: &jsir.Block{}}
}

func ConvertOperator(op lexer.TokenType) jsir.Operator {
	return map[lexer.TokenType]jsir.Operator{
		lexer.Equal:          jsir.OpEqual,
		lexer.NotEqual:       jsir.OpNotEqual,
		lexer.GreaterThan:    jsir.OpGreaterThan,
		lexer.LessThan:       jsir.OpLessThan,
		lexer.GreaterEqualTo: jsir.OpGreaterEqual,
		lexer.LessEqualTo:    jsir.OpLessEqual,
		lexer.Plus:           jsir.OpPlus,
		lexer.Minus:          jsir.OpMinus,
		lexer.Asterisk:       jsir.OpTimes,
		lexer.Slash:          jsir.OpDivide,
		lexer.Percent:        jsir.OpRemainder,
		lexer.Caret:          jsir.OpExponent,
		lexer.AndAnd:         jsir.OpLogicalAnd,
		lexer.OrOr:           jsir.OpLogicalOr,
		lexer.Not:            jsir.OpLogicalNot,
		lexer.PlusEqual:      jsir.OpPlusEqual,
		lexer.MinusEqual:     jsir.OpMinusEqual,
		lexer.AsteriskEqual:  jsir.OpTimesEqual,
		lexer.SlashEqual:     jsir.OpDivideEqual,
		lexer.PercentEqual:   jsir.OpRemainderEqual,
		lexer.CaretEqual:     jsir.OpExponentEqual,
	}[op]
}
