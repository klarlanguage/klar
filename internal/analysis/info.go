package analysis

import "github.com/ProCode-Software/klar/internal/ast"

type Info struct {
	Expressions map[ast.Expression]*Expr
	Tracked     map[*Object]TrackedInfo
	// Keys may be [*ast.Program], [*ast.Block] or [ast.Statement]
	Blocks    map[ast.Node]*Block // Node where each context begins
	InitOrder []*Object           // For top-level objects
}

type Block struct {
	Context *Context
	Returns *[]returnStmt
	// Whether the block contains a 'next', 'stop' or 'return' statement
	HasControlFlow bool
	Flags          stmtFlags
}

func (c *Checker) recordBlock(node ast.Node, sctx *stmtContext) {
	c.Module.Info.Blocks[node] = &Block{
		Context:        sctx.ctx,
		Returns:        sctx.returns,
		Flags:          sctx.flags,
		HasControlFlow: false, // TODO
	}
}
