package codegen

import (
	"maps"
	"slices"

	"github.com/klarlanguage/klar/internal/analysis"
	"github.com/klarlanguage/klar/internal/ast"
	"github.com/klarlanguage/klar/internal/ir/jsir"
)

func (g *Generator) convertFuncDecl(o *analysis.Object, b *block) {
	// Some overloads may have different names via the @name attribute.
	// Group the overloads by name and convert those grouped ones to JavaScript.
	var nameGroups map[string][]*analysis.Overload
	fnType := o.Type.(*analysis.Function)
	for _, ov := range fnType.Overloads {
		if ov.Name != o.Name {
			if nameGroups == nil {
				nameGroups = make(map[string][]*analysis.Overload)
			}
			nameGroups[ov.Name] = append(nameGroups[ov.Name], ov)
		}
	}
	if nameGroups == nil {
		// No name changes
		b.growStmts(len(fnType.Overloads) + 1)
		g.appendOverloads(fnType.Overloads, o.Public, b)
	} else {
		// Some may still have the original name, so add that to nameGroups
		for _, fnType := range fnType.Overloads {
			if fnType.Name == o.Name {
				nameGroups[fnType.Name] = append(nameGroups[fnType.Name], fnType)
			}
		}
		b.growStmts(len(nameGroups) + len(fnType.Overloads))
		// Sort the overloads by declaration order
		for _, name := range slices.SortedFunc(maps.Keys(nameGroups), func(a, b string) int {
			return int(nameGroups[a][0].Order) - int(nameGroups[b][0].Order)
		}) {
			g.appendOverloads(nameGroups[name], o.Public, b)
		}
	}
	// The root overload is already appended
}

func (g *Generator) appendOverloads(ovs []*analysis.Overload, public bool, b *block) {
	if len(ovs) > 1 {
		// TODO: Merge overloads
		return
	}
	ov := ovs[0]
	node := ov.Node().(*ast.FunctionDeclaration)
	// The self type will be disregarded as the generated function is put
	// into the JS class body
	fn := &jsir.FunctionDeclaration{
		Name:   ov.Name,
		Params: make([]*jsir.FunctionParam, 0, len(ov.Params)+len(ov.LabelledParams)),
	}
	if public {
		// TODO: Don't if this is a method
		b.insertStmt(&jsir.ExportModifierStatement{false, fn})
	} else {
		b.insertStmt(fn)
	}
	g.convertOverloadParams(ov, fn)
	switch {
	case node.Body != nil:
		fn.Body = g.convertBlock(node.Body, ov.InnerContext)
	case node.Expression != nil:
		// A block is needed because the expression may need to
		// generate other statements
		b := &block{ctx: ov.InnerContext}
		expr := g.convertExpression(node.Expression, b)
		// Don't return expressions that return Nothing
		if g.Info.Expressions[node.Expression].Kind() == analysis.NothingType {
			b.stmts = append(b.stmts, &jsir.ExpressionStatement{expr})
		} else {
			b.stmts = append(b.stmts, &jsir.ReturnStatement{expr})
		}
		fn.Body = &jsir.Block{b.stmts}
	}
}

// TODO: Types and default values
func (g *Generator) convertOverloadParams(ov *analysis.Overload, fn *jsir.FunctionDeclaration) {
	for _, vr := range ov.Params {
		p := &jsir.FunctionParam{Name: &jsir.Symbol{vr.Object.Name}}
		fn.Params = append(fn.Params, p)
	}
	var variadicParam *analysis.Variable
	if ov.Variadic() {
		variadicParam = ov.Params[len(ov.Params)-1]
	}
	for _, lp := range ov.LabelledParams {
		p := &jsir.FunctionParam{Name: &jsir.Symbol{lp.Object.Name}}
		// A labelled parameter can become variadic if there are no other variadic
		// positional or labelled params.
		if lp.Object.Flags.Has(analysis.VariadicParam) && variadicParam == nil {
			var hasVariadicLabelled bool
			for _, lp2 := range ov.LabelledParams {
				if lp2.Object.Flags.Has(analysis.VariadicParam) && lp2 != lp {
					hasVariadicLabelled = true
					break
				}
			}
			if !hasVariadicLabelled {
				variadicParam = lp.Variable
				continue
			}
		}
		fn.Params = append(fn.Params, p)
	}

	// Add the variadic param
	if variadicParam != nil {
		fn.VariadicParam = &jsir.FunctionParam{
			Name: &jsir.Symbol{variadicParam.Object.Name},
		}
	}
}

func (g *Generator) addMethods(cls *jsir.ClassDeclaration, o *analysis.Object) {
	meths := o.TypeName().Type.(analysis.SupportsMethods).GetMethods()
	cls.Methods = slices.Grow(cls.Methods, len(meths))
	b := &block{ctx: o.Context}
	for _, m := range meths {
		if _, ok := m.Type.(*analysis.FunctionAlias); ok {
			continue // TODO
		}
		g.convertFuncDecl(m, b)
	}
	for _, stmt := range b.stmts {
		cls.Methods = append(cls.Methods, stmt.(*jsir.FunctionDeclaration))
	}
}
