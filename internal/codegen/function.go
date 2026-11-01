package codegen

import (
	"github.com/klarlanguage/klar/internal/analysis"
	"github.com/klarlanguage/klar/internal/ast"
	"github.com/klarlanguage/klar/internal/ir/jsir"
)

func (g *Generator) appendOverloads(dst []jsir.Statement, ovs []*analysis.Overload) []jsir.Statement {
	if len(ovs) > 1 {
		// TODO: Merge overloads
		return dst
	}
	ov := ovs[0]
	stmt := ov.Node().(*ast.FunctionDeclaration)
	// The self type will be disregarded as the generated function is put
	// into the JS class body
	fn := &jsir.FunctionDeclaration{
		Name:   ov.Name,
		Params: make([]*jsir.FunctionParam, 0, len(ov.Params)+len(ov.LabelledParams)),
	}
	_ = stmt
	return append(dst, fn)
}
