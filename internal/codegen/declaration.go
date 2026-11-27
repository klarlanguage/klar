package codegen

import (
	"fmt"

	"github.com/klarlanguage/klar/internal/analysis"
	"github.com/klarlanguage/klar/internal/ast"
	"github.com/klarlanguage/klar/internal/ir/jsir"
)

func (g *Generator) convertObject(o *analysis.Object, b *block) (res jsir.Statement) {
	defer func() {
		if res != nil {
			if o.Public && o.Context.File.TopLevel() {
				res = &jsir.ExportModifierStatement{Declaration: res}
			}
			b.stmts = append(b.stmts, res)
		}
	}()
	switch stmt := o.Node().(type) {
	case *ast.FunctionDeclaration:
		g.convertFuncDecl(o, b)
		return nil
	case *ast.FuncAliasDeclaration:
		return g.convertFuncAlias(o, stmt)
	case *ast.VariableDeclaration:
		return g.convertVarDecl(o, stmt, b)
	case ast.TypeDeclaration:
		return g.convertTypeDecl(o, stmt, b)
	default:
		panic(fmt.Sprintf("invalid analysis.Object statement: %T", stmt))
	}
}

func (g *Generator) convertTypeDecl(
	o *analysis.Object, decl ast.TypeDeclaration, b *block,
) jsir.Statement {
	typ := o.TypeName().Type
	switch decl := decl.(type) {
	case *ast.StructDeclaration:
		str := typ.(*analysis.Struct)
		cls := g.convertStruct(o, str)
		if len(str.Methods) > 0 {
			g.addMethods(cls, o)
		}
		return cls
	case *ast.EnumDeclaration:
		enum := typ.(*analysis.Enum)
		class := g.convertEnum(o, enum, b)
		if len(enum.Methods) > 0 {
			g.addMethods(class, o)
		}
		return class
	case *ast.InterfaceDeclaration:
		return nil // TODO: Return one for TypeScript
	case *ast.TagDeclaration:
		return g.convertTagDecl(o)
	case *ast.TypeAliasDeclaration:
	default:
		panic(fmt.Sprintf("invalid ast.TypeDeclaration node: %T", decl))
	}
	return &jsir.FunctionDeclaration{
		Name: o.Name,
		Body: &jsir.Block{},
	}
}

func (g *Generator) convertVarDecl(
	o *analysis.Object, decl *ast.VariableDeclaration, b *block,
) jsir.Statement {
	// Use 'let' for all declarations unless they are public, top-level constants
	if len(decl.Variables) == 1 {
		// Convert `_ := expr` to `void expr`
		// TODO: [DCE] Don't generate anything at all if the value has no side effects
		if _, ok := decl.Variables[0].(*ast.Discard); ok {
			return &jsir.ExpressionStatement{&jsir.UnaryExpression{
				Operator: jsir.OpVoid,
				Operand:  g.convertExpression(decl.Values[0], b),
			}}
		}
		return &jsir.BindingDeclaration{
			Kind: jsir.LetBinding,
			// Destructurability validated at analysis time
			Variable: g.convertDestructure(
				decl.Variables[0].(ast.Destructurable), !o.Public,
			),
			Value: g.convertExpression(decl.Values[0], b),
		}
	}

	jsDecl := &jsir.MultiBindingDeclaration{
		Kind:     jsir.LetBinding,
		Bindings: make([]*jsir.Binding, 0, len(decl.Variables)),
	}
	var singleRHS jsir.Expression
	if decl.IsSingleRHS() {
		singleRHS = g.convertExpression(decl.Values[0], b)
	}
	for i, name := range decl.Variables {
		// Don't declare discards. Remember that '_' is a valid identifier in JS.
		if _, ok := name.(*ast.Discard); ok {
			continue
		}
		var val jsir.Expression
		if singleRHS != nil {
			val = singleRHS
		} else {
			val = g.convertExpression(decl.Values[i], b)
		}
		jsDecl.Bindings = append(jsDecl.Bindings, &jsir.Binding{
			Variable: g.convertDestructure(name.(ast.Destructurable), !o.Public),
			Value:    val,
		})
	}
	if len(jsDecl.Bindings) == 0 {
		// All variables declared are discards, this just becomes a 'void' expression
		return &jsir.ExpressionStatement{&jsir.UnaryExpression{
			Operator: jsir.OpVoid,
			Operand:  g.convertExpression(decl.Values[0], b),
		}}
	}
	return jsDecl
}

// TODO
func (g *Generator) convertFuncAlias(
	o *analysis.Object, decl *ast.FuncAliasDeclaration,
) *jsir.BindingDeclaration {
	return &jsir.BindingDeclaration{
		Kind:     jsir.ConstBinding,
		Variable: escapedSymbol(o.Name, !o.Public),
		Value:    &jsir.NullLiteral{},
	}
}

func (g *Generator) convertTagDecl(o *analysis.Object) *jsir.BindingDeclaration {
	keyword := jsir.LetBinding
	if o.Public {
		keyword = jsir.ConstBinding
	}
	// Tags are converted to unique Symbols that are set as properties
	// on implementing classes
	// 	const Node = Symbol('Node')
	return &jsir.BindingDeclaration{
		Kind:     keyword,
		Variable: escapedSymbol(o.Name, !o.Public),
		Value: &jsir.CallExpression{
			g.getGlobal("Symbol", o.Context), []jsir.Expression{
				// Use single quotes because the name is a valid identifier
				&jsir.StringLiteral{jsir.SingleQuote, o.Name},
			},
		},
	}
}

func (g *Generator) getGlobal(name string, ctx *analysis.Context) jsir.Expression {
	if ctx.LookupRecursive(name) != nil {
		return &jsir.IndexExpression{
			Object:   &jsir.Symbol{"globalThis"},
			Property: &jsir.Symbol{name},
		}
	}
	return &jsir.Symbol{name}
}
