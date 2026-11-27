package codegen

import (
	"fmt"
	"slices"

	"github.com/klarlanguage/klar/internal/analysis"
	"github.com/klarlanguage/klar/internal/ast"
	"github.com/klarlanguage/klar/internal/ir/jsir"
)

func (g *Generator) convertExpression(expr ast.Expression, b *block) jsir.Expression {
	switch expr := expr.(type) {
	case *ast.AssertExpression:
	case *ast.AwaitExpression:
	case *ast.BinaryExpression:
		return &jsir.BinaryExpression{
			Left:     g.convertExpression(expr.Left, b),
			Operator: ConvertOperator(expr.Operator.Kind),
			Right:    g.convertExpression(expr.Right, b),
		}
	case *ast.BooleanLiteral:
		return &jsir.BooleanLiteral{expr.Value}
	case *ast.CallExpression:
		args := make([]jsir.Expression, len(expr.Args))
		for i, fp := range expr.Args {
			args[i] = g.convertExpression(fp.Value, b)
		}
		return &jsir.CallExpression{
			Callee:    g.convertExpression(expr.Callee, b),
			Arguments: args,
		}
	case *ast.EnumLiteral:
	case *ast.ForExpression:
	case *ast.FloatLiteral:
		return &jsir.NumericLiteral{expr.Value}
	case *ast.IntegerLiteral:
		return &jsir.NumericLiteral{float64(expr.Value)}
	case *ast.GoExpression:
	case *ast.IndexExpression:
		computed := expr.Computed
		lhs := g.convertExpression(expr.Object, b)
		var rhs jsir.Expression
		if sym, ok := expr.Property.(*ast.Symbol); ok && !expr.Computed {
			if _, ok := jsir.JSKeywords[sym.Identifier]; ok {
				rhs = &jsir.StringLiteral{QuoteStyle: jsir.AutoQuote, Content: sym.Identifier}
				computed = true
			}
		}
		if rhs == nil {
			rhs = g.convertExpression(expr.Property, b)
		}
		return &jsir.IndexExpression{Object: lhs, Property: rhs, Computed: computed}
	case *ast.LambdaExpression:
	case *ast.ListCastExpression:
	case *ast.ListLiteral:
		items := make([]jsir.Expression, len(expr.Items))
		for i, node := range expr.Items {
			items[i] = g.convertExpression(node, b)
		}
		return &jsir.ArrayLiteral{items}
	case *ast.MapCastExpression:
	case *ast.MapLiteral:
		return &jsir.UnaryExpression{
			Operator: jsir.OpNew,
			Operand: &jsir.CallExpression{
				Callee: g.getGlobal("Map", b.ctx),
			},
		}
	case *ast.NilLiteral:
		// TODO: Should be 'undefined' when assigning to a map. We could also check for this
		// when converting assignments
		return &jsir.NullLiteral{}
	case *ast.ObjectPipeline:
	case *ast.PipelineExpression:
	case *ast.ParenExpression:
		return g.convertExpression(ast.Unparen(expr), b)
	case *ast.RangeExpression:
	case *ast.RegexLiteral:
		// TODO: This must be a `new RegExp()` call if the expression has interpolations
		return &jsir.RegExpLiteral{Flags: expr.Flags, Pattern: "(?:)"}
	case *ast.RelationalExpression:
		return g.convertRelationalExpr(expr, b)
	case *ast.RestExpression:
		return &jsir.SpreadExpression{g.convertExpression(expr.Expression, b)}
	case *ast.SliceExpression:
	case *ast.StringLiteral:
	case *ast.StructDotInit:
	case *ast.Symbol:
		name := expr.Identifier
		// Convert references to 'self' to 'this'
		if typ := g.Info.Expressions[expr]; typ != nil {
			if o, ok := typ.Type.(*analysis.Object); ok {
				if vr, ok := o.Type.(*analysis.Variable); ok && vr.VarKind == analysis.SelfVar {
					name = "this"
				}
			}
		}
		return &jsir.Symbol{name}
	case *ast.TryExpression:
	case *ast.TupleLiteral:
		// Tuples are converted into JavaScript arrays
		items := make([]jsir.Expression, len(expr.Values))
		for i, node := range expr.Values {
			items[i] = g.convertExpression(node, b)
		}
		return &jsir.ArrayLiteral{items}
	case *ast.UnaryExpression:
		return &jsir.UnaryExpression{
			Operator: ConvertOperator(expr.Operator.Kind),
			Operand:  g.convertExpression(expr.Right, b),
		}
	case *ast.WhenExpression:
	default:
		panic(fmt.Sprintf("unhandled ast.Expression type: %T", expr))
	}
	return &jsir.NullLiteral{}
}

func (g *Generator) convertRelationalExpr(expr *ast.RelationalExpression, b *block) jsir.Expression {
	// TODO: Should a fast path be created for simple expressions with 2 operands? Benchmark

	// 	a < b < c < d
	// is converted to:
	// 	a < b && (b < c && (c < d))
	var final *jsir.BinaryExpression
	prevLHS := g.convertExpression(expr.Expressions[len(expr.Expressions)-1], b)
	for i, op := range slices.Backward(expr.Operators) {
		lhs := g.convertExpression(expr.Expressions[i], b)
		relExpr := &jsir.BinaryExpression{
			Left:     lhs,
			Operator: ConvertOperator(op.Kind),
			Right:    prevLHS,
		}
		prevLHS = lhs

		if final == nil {
			final = relExpr
			continue
		} else {
			final = &jsir.BinaryExpression{
				Left:     relExpr,
				Operator: jsir.OpLogicalAnd,
				Right:    final,
			}
		}
	}
	return final
}
