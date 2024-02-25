// Copyright (c) 2024 The Falcon Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.
package ssa

import (
	"falcon/ast"
	"falcon/utils"
	"fmt"
)

// == Code conjured by yyang, Feb, 2024 ==

//------------------------------------------------------------------------------
// SSA based HIR construction
//
// See "Simple and Efficient Construction of Static Single Assignment Form" for
// more details. It transforms AST to SSA form in a simple manner.

type GraphBuilder struct {
	fn *Func
	// Block and Name identify unique variable
	names map[*Block]map[string]*Value
	// Sealed block means all its predecessors have been processed
	sealed map[*Block]bool
	// current block for SSA instruction generation
	current *Block
	// operand-less phis, i.e. orphan phis are those phis that are not yet complete
	orphanPhi map[*Block]map[string]*Value
	// skip the next seal operation, this is used to avoid sealing the loop header
	// automatically when the condition is generated
	skipNextSeal bool
	// support to build loop form
	scopes []*BlockScope
}

// LoopScope is used to construct loop form and related control flow alterations
type BlockScope struct {
	exit *Block
	post *Block
}

func NewGraphBuilder(fn *Func) *GraphBuilder {
	return &GraphBuilder{
		fn:           fn,
		names:        make(map[*Block]map[string]*Value),
		sealed:       make(map[*Block]bool),
		orphanPhi:    make(map[*Block]map[string]*Value),
		skipNextSeal: false,
		scopes:       make([]*BlockScope, 0),
	}
}

func (g *GraphBuilder) eliminateTrivialPhi(phi *Value) *Value {
	utils.Assert(phi.Op == OpPhi, "sanity check")
	// v1 = phi v2? This is a copy propagation, we can replace all uses of v1
	if len(phi.Args) == 1 {
		phi.ReplaceUses(phi.Args[0])
		return phi.Args[0]
	}
	// Consider the following IR snippet
	//
	// loop entry:
	// v1 = CInt #3
	//
	// loop header:
	// ....
	//
	// loop exit:
	// use(v1)
	//
	// use of v1 lives in loop exit, when sealing loop exit, we would look up
	// its single predecessor loop header, since there is no definition for
	// v1 in loop header, we would create an operand-less phi
	//
	// loop entry:
	// v1 = CInt #3
	//
	// loop header:
	// v3 = phi
	// ....
	//
	// loop exit:
	// use(v3)
	//
	// Since loop header has multiple predecessors, we would add operand by
	// looking up all its predecessors, finally it looks like
	//
	// loop entry:
	// v1 = CInt #3
	//
	// loop header:
	// v3 = phi v1, v3
	// ....
	//
	// loop exit:
	// use(v3)
	//
	// v3 references itself and v1, for such case, we consider it as a trivial
	// phi and replace all uses of v3 with v1
	var trivial *Value
	for _, arg := range phi.Args {
		if arg == phi {
			continue
		}
		if arg != phi {
			if trivial == nil {
				trivial = arg
			} else if trivial != arg {
				// Non-trivial phi, do nothing then
				return nil
			}
		}
	}
	if trivial != nil {
		phi.ReplaceUses(trivial)
		return trivial
	}
	return nil
}

func (g *GraphBuilder) lookupVar(name string, block *Block) *Value {
	// Good, we can find such variable in current block, just return it
	if _, exist := g.names[block][name]; exist {
		return g.names[block][name]
	}

	// Hard case, we can not find such variable in current block, we need to
	// search it recursively and carefully take ssa phi into consideration
	if _, sealed := g.sealed[block]; !sealed {
		// This is not-yet-complete CFG, add orphan phi
		val := block.NewValue(OpPhi, nil /*Should be typed later*/)
		g.orphanPhi[block][name] = val
		g.names[block][name] = val
		return val
	} else if len(block.Preds) == 1 {
		val := g.lookupVar(name, block.Preds[0])
		g.names[block][name] = val
		return val
	} else {
		// To avoid infinite recursion, create dummy phi node before looking up
		val := block.NewValue(OpPhi, nil /*Should be typed later*/)
		//g.orphanPhi[block][Name] = val
		g.names[block][name] = val
		g.addPhiOperand(name, val)
		return val
	}
}

func propagatePhiType(phi *Value, t *ast.Type) {
	// t is valid type?
	if t != nil {
		// The first time to set type for phi?
		if phi.Type == nil {
			// Set type for phi
			phi.Type = t
			// Also set type for all its uses
			for _, use := range phi.Uses {
				propagatePhiType(use, t)
			}
		}
	}
}

func (g *GraphBuilder) addPhiOperand(name string, phi *Value) {
	for _, pred := range phi.Block.Preds {
		input := g.lookupVar(name, pred)
		phi.AddArg(input)

		// Propagate type information to uses if possible
		if t := input.Type; t != nil {
			propagatePhiType(phi, t)
		}
	}

	g.eliminateTrivialPhi(phi)
}

func (g *GraphBuilder) setControl(b *Block) {
	utils.Assert(g.current != b, "control remains the same")
	// Seal previous control block if necessary
	if !g.skipNextSeal {
		oldControl := g.current
		if _, sealed := g.sealed[oldControl]; !sealed {
			g.sealBlock(oldControl)
		}
	} else {
		g.skipNextSeal = false
	}
	// Alter the control flow then
	g.current = b
}

func (g *GraphBuilder) getControl() *Block {
	return g.current
}

func (g *GraphBuilder) stopControl() {
	g.setControl(nil)
}

func (g *GraphBuilder) isStopControl() bool {
	return g.current == nil
}

func (g *GraphBuilder) enterBlockScope() *BlockScope {
	scope := &BlockScope{}

	g.scopes = append(g.scopes, scope)
	return scope
}

func (g *GraphBuilder) exitBlockScope() {
	g.scopes = g.scopes[:len(g.scopes)-1]
}

func (g *GraphBuilder) getBlockScope() *BlockScope {
	if len(g.scopes) == 0 {
		return nil
	}
	return g.scopes[len(g.scopes)-1]
}

func (g *GraphBuilder) sealBlock(block *Block) {
	// Seal this block, add operand for orphan phis
	for name, phi := range g.orphanPhi[block] {
		g.addPhiOperand(name, phi)
	}
	g.sealed[block] = true
}

func (g *GraphBuilder) recordBlock(blocks ...*Block) {
	for _, block := range blocks {
		g.names[block] = make(map[string]*Value)
		g.orphanPhi[block] = make(map[string]*Value)
	}
}

func addEdge(from, to *Block) {
	if from == nil || to == nil {
		return
	}
	if from.Kind == BlockReturn || from.Kind == BlockDead {
		return
	}
	from.WireTo(to)
}

func (g *GraphBuilder) verify() {
	// Final block should be BlockReturn
	if g.current.Kind != BlockReturn {
		utils.Fatal("final block is not BlockReturn")
	}
	// All blocks should be sealed
	for _, block := range g.fn.Blocks {
		if _, sealed := g.sealed[block]; !sealed {
			utils.Fatal("block not sealed %v", block)
		}
	}
}

func (g *GraphBuilder) printVars() {
	println("[[Names]]")
	for block, m := range g.names {
		for name, val := range m {
			fmt.Printf("b%d: %s : %s\n", block.Id, name, val.String())
		}
	}
}

func (g *GraphBuilder) buildConst(n ast.AstExpr) *Value {
	switch n.(type) {
	case *ast.IntExpr:
		val := g.getControl().NewValue(OpCInt, n.GetType())
		val.Sym = n.(*ast.IntExpr).Value
		return val
	case *ast.LongExpr:
		utils.Unimplement()
	case *ast.FloatExpr:
		utils.Unimplement()
	case *ast.DoubleExpr:
		val := g.getControl().NewValue(OpCDouble, n.GetType())
		val.Sym = n.(*ast.DoubleExpr).Value
		return val
	case *ast.BoolExpr:
		val := g.getControl().NewValue(OpCBool, n.GetType())
		val.Sym = n.(*ast.BoolExpr).Value
		val.Type = n.GetType()
		return val
	case *ast.StrExpr:
		val := g.getControl().NewValue(OpCString, n.GetType())
		val.Sym = n.(*ast.StrExpr).Value
		return val
	case *ast.ArrayExpr:
		val := g.getControl().NewValue(OpCArray, n.GetType())
		val.Sym = len(n.(*ast.ArrayExpr).Elems)
		for idx, elem := range n.(*ast.ArrayExpr).Elems {
			elem := g.build(elem)
			index := g.getControl().NewValue(OpCInt, ast.BasicTypes[ast.TypeInt])
			index.Sym = idx
			st := g.getControl().NewValue(OpStoreIndex, elem.Type)
			st.AddArg(val, index, elem)
		}
		return val
	default:
		utils.Unimplement()
	}
	return nil
}

func (g *GraphBuilder) buildAssignExpr(expr *ast.AssignExpr) *Value {

	if _, yes := expr.Left.(*ast.IndexExpr); yes {
		left := g.build(expr.Left)
		right := g.build(expr.Right)
		// arr[idx] = val
		block := g.getControl()
		st := block.NewValue(OpStoreIndex, right.Type)
		st.AddArg(left.Args[0] /*array*/, left.Args[1] /*index*/, right /*elem*/)
		// Remove fake LoadIndex
		block.RemoveValue(left)
		return st
	} else {
		// var = val
		right := g.build(expr.Right)
		block := g.getControl()
		switch expr.Opt {
		case ast.TK_ASSIGN:
			g.names[block][expr.Left.(*ast.VarExpr).Name] = right
			return right
		case ast.TK_PLUS_AGN:
			left := g.lookupVar(expr.Left.(*ast.VarExpr).Name, block)
			val := block.NewValue(OpAdd, right.Type, left, right)
			g.names[block][expr.Left.(*ast.VarExpr).Name] = val
			return val
		case ast.TK_MINUS_AGN:
			left := g.lookupVar(expr.Left.(*ast.VarExpr).Name, block)
			val := block.NewValue(OpSub, right.Type, left, right)
			g.names[block][expr.Left.(*ast.VarExpr).Name] = val
			return val
		default:
			utils.Fatal("unimplement %v", expr.Opt)
		}
	}

	return nil
}

func (g *GraphBuilder) buildFunCallExpr(expr *ast.FuncCallExpr) *Value {
	args := make([]*Value, len(expr.Args))
	for i, arg := range expr.Args {
		args[i] = g.build(arg)
	}
	block := g.getControl()
	val := block.NewValue(OpCall, expr.Type)
	val.Sym = expr.Name
	val.AddArg(args...)
	return val
}

func (g *GraphBuilder) buildIndexExpr(expr *ast.IndexExpr) *Value {
	block := g.getControl()
	array := g.lookupVar(expr.Name, block)
	index := g.build(expr.Index)
	val := block.NewValue(OpLoadIndex, expr.GetType())
	val.AddArg(array, index)
	return val
}

func (g *GraphBuilder) buildUnaryExpr(node *ast.UnaryExpr) *Value {
	switch node.Opt {
	case ast.TK_MINUS:
		token2ssaOp := map[ast.TokenKind]Op{
			ast.TK_MINUS: OpSub,
		}
		ssaOp, exist := token2ssaOp[node.Opt]
		utils.Assert(exist, "unimplement %v", node.Opt.String())
		zero := &ast.IntExpr{Value: 0}
		zero.SetType(ast.BasicTypes[ast.TypeInt])
		left := g.buildConst(zero)
		right := g.build(node.Left)
		block := g.getControl()
		return block.NewValue(ssaOp, right.Type, left, right)
	case ast.TK_BITNOT:
		arg := g.build(node.Left)
		block := g.getControl()
		return block.NewValue(OpNot, arg.Type, arg)
	default:
		utils.Unimplement()
	}
	return nil
}

func (g *GraphBuilder) buildLogicalExpr(node *ast.BinaryExpr) *Value {
	// Logical and/or are short-circuit operators
	cond1 := g.build(node.Left)
	cond1Block := g.getControl()
	cond1Block.Kind = BlockIf
	cond1.AddUseBlock(cond1Block)

	cond2Block := g.fn.NewBlock(BlockGoto)
	phi1Block := g.fn.NewBlock(BlockGoto)
	g.recordBlock(cond2Block, phi1Block)

	// The order of CFG edges are important, it indicates where is the true path and vice versa
	if node.Opt == ast.TK_LOGOR {
		//	       cond1
		//	       /  \
		//	      /    ▼
		//	      \  cond2
		//	       \  /
		//	         ▼
		//	  phi1(cond1,cond2)
		addEdge(cond1Block, phi1Block)
		addEdge(cond1Block, cond2Block)

		g.setControl(cond2Block)
		cond2 := g.build(node.Right)
		cond2Block = g.getControl()
		addEdge(cond2Block, phi1Block)

		// Create phi at cfg merge point
		g.setControl(phi1Block)
		phi1 := phi1Block.NewValue(OpPhi, cond1.Type)
		phi1.AddArg(cond1, cond2) // cond1 first
		phi1.AddUseBlock(phi1Block)
		utils.Assert(cond1.Type == cond2.Type, "type mismatch")
		return phi1
	} else {
		//	       cond1
		//	       /  \
		//	      ▼    \
		//	    cond2  /
		//	       \  /
		//	         ▼
		//	  phi1(cond2,cond1)
		addEdge(cond1Block, cond2Block)

		g.setControl(cond2Block)
		cond2 := g.build(node.Right)
		cond2Block = g.getControl()
		addEdge(cond2Block, phi1Block) // now first pred of phi1block is cond2block
		addEdge(cond1Block, phi1Block) // then safe to add edge from cond1block to phi1block

		// Create phi at cfg merge point
		g.setControl(phi1Block)
		phi1 := phi1Block.NewValue(OpPhi, cond2.Type)
		phi1.AddArg(cond2, cond1) // cond2 first
		phi1.AddUseBlock(phi1Block)
		utils.Assert(cond1.Type == cond2.Type, "type mismatch")
		return phi1
	}
}

func (g *GraphBuilder) buildBinaryExpr(node *ast.BinaryExpr) *Value {
	switch node.Opt {
	case ast.TK_LOGOR, ast.TK_LOGAND:
		return g.buildLogicalExpr(node)
	default:
		token2ssaOp := map[ast.TokenKind]Op{
			ast.TK_PLUS:  OpAdd,
			ast.TK_MINUS: OpSub,
			ast.TK_TIMES: OpMul,
			ast.TK_DIV:   OpDiv,
			ast.TK_MOD:   OpMod,

			ast.TK_BITAND: OpAnd,
			ast.TK_BITOR:  OpOr,
			ast.TK_BITXOR: OpXor,
			ast.TK_LSHIFT: OpLShift,
			ast.TK_RSHIFT: OpRShift,

			ast.TK_LE: OpCmpLE,
			ast.TK_LT: OpCmpLT,
			ast.TK_GE: OpCmpGE,
			ast.TK_GT: OpCmpGT,
			ast.TK_EQ: OpCmpEQ,
			ast.TK_NE: OpCmpNE,
		}
		ssaOp, exist := token2ssaOp[node.Opt]
		utils.Assert(exist, "unimplement %v", node.Opt.String())
		left := g.build(node.Left)
		right := g.build(node.Right)
		block := g.getControl()
		val := block.NewValue(ssaOp, right.Type, left, right)
		return val
	}
}

// ----------------------------------------------------------------------------
// The Loop Form
//
// The natural loop usually looks like in below IR form:
//
//	 loop entry
//	     │
//	     │  ┌───loop latch
//	     ▼  ▼       ▲
//	loop header     │
//	     │  │       │
//	     │  └──►loop body
//	     ▼
//	 loop exit
//
// In the terminology, loop entry dominates the entire loop, loop header contains
// the loop conditional test, loop body refers to the code that is repeated, loop
// latch contains the backedge to loop header, for simple loops, the loop body is
// equal to loop latch, and loop exit refers to the block that dominated by the
// entire loop.
func (g *GraphBuilder) buildLoop(init, cond, body, post ast.AstNode) {
	loopHeader := g.fn.NewBlock(BlockIf)
	loopHeader.Hint = HintLoopHeader
	loopBody := g.fn.NewBlock(BlockGoto)
	loopExit := g.fn.NewBlock(BlockGoto)
	g.recordBlock(loopHeader, loopBody, loopExit)

	// Construct the control flow
	loopEntry := g.getControl()
	loopEntry.Kind = BlockGoto //repair loop entry
	addEdge(loopEntry, loopHeader)

	// Build the loop initialization expr
	if init != nil {
		g.build(init)
	}
	g.setControl(loopHeader)
	// @@ Note, we don't want to seal up the loop header here, because there are
	// still unvisited predecessor block from backedge, so we set a flag to skip
	// the next seal operation during setControl
	g.skipNextSeal = true
	val := g.build(cond)

	// Build the loop condition test
	// @@ Don't use loop header here, because the generation of condition alters
	// the control flow, adding intermediate edge between loop header and loop body
	// For example, if the condition is a short-circuit operator, the CFGs looks
	// like below form
	//
	//	 loop entry
	//	     │
	//	     │   ┌──loop latch
	//	     ▼   ▼     ▲
	//	 loop header   │
	//	     │   │     │
	//	   cond2 │     │
	//	     │   │     │
	//	     ▼   ▼     │
	//	 merge block   │
	//	     │   │     │
	//	     │   │     │
	//	     │   └──►loop body
	//	     ▼
	//
	// The current control block is merge block, and we should wire edges from it
	// to loop body and loop exit rather than starting from loop header
	loopHeaderTail := g.getControl()
	loopHeaderTail.Kind = BlockIf
	val.AddUseBlock(loopHeaderTail)
	addEdge(loopHeaderTail, loopBody) // order is important , true path first
	addEdge(loopHeaderTail, loopExit)

	// Build loop body
	g.setControl(loopBody)

	scope := g.enterBlockScope()
	scope.exit = loopExit
	scope.post = loopHeader
	g.build(body)
	g.exitBlockScope()

	if !g.isStopControl() {
		// Build loop post only when present and loop is not breaked
		if post != nil {
			g.build(post)
		}
	}
	// Add backedge from tail of loop body to loop header
	loopBodyTail := g.getControl()
	addEdge(loopBodyTail, loopHeader)

	g.setControl(loopExit)

	// Backedge from loop latch to loop header has been processed, we can seal
	// the loop headerwhich in turns complete orphan phis
	g.sealBlock(loopHeader)
}

func (g *GraphBuilder) buildIf(cond, thenB, elseB ast.AstNode, hasResult bool) *Value {
	// If then and else blocks are all presented, the CFG looks like a diamond
	//
	//       entry
	//       /  \
	//      ▼    ▼
	//     then  else
	//       \  /
	//         ▼
	//       merge
	//
	// If else block is not presented, the CFG looks like a triangle
	//
	//       entry
	//       /  \
	//      ▼    \
	//    then   /
	//       \  /
	//         ▼
	//       merge
	//

	// Construct the control flow
	val := g.build(cond)
	entry := g.getControl()
	entry.Kind = BlockIf
	val.AddUseBlock(entry)

	// Build the then block nevertheless
	var ifThen, ifElse *Block
	ifThen = g.fn.NewBlock(BlockGoto)
	// We need to wire the control flow from entry before setting control
	// in order to avoid sealing the entry block while its predecessor
	// block is not linked
	addEdge(entry, ifThen)
	g.recordBlock(ifThen)
	g.setControl(ifThen)
	var thenVal, elseVal *Value
	thenVal = g.build(thenB)
	mergeThen, mergeElse := g.getControl(), (*Block)(nil)

	// BUild the else block if it is presented
	if elseB != nil {
		ifElse = g.fn.NewBlock(BlockGoto)
		addEdge(entry, ifElse)
		g.recordBlock(ifElse)
		g.setControl(ifElse)
		elseVal = g.build(elseB)
		mergeElse = g.getControl()
	} else {
		// No else block, mergeElse aliases to entry block
		mergeElse = entry
	}

	// Merge point
	merge := g.fn.NewBlock(BlockGoto)
	g.recordBlock(merge)
	addEdge(mergeThen, merge)
	addEdge(mergeElse, merge)
	g.setControl(merge)

	// Good! We've done. See if we need to create a phi value for merge block
	if hasResult {
		utils.Assert(thenVal != nil && elseVal != nil, "sanity check")
		utils.Assert(thenVal.Type == elseVal.Type, "type mismatch")
		phi := merge.NewValue(OpPhi, thenVal.Type)
		phi.AddArg(thenVal, elseVal)
		return phi
	}
	return nil
}

func (g *GraphBuilder) buildBreakStmt(node *ast.BreakStmt) {
	utils.Assert(g.getBlockScope() != nil, "break statement not in loop")
	// Find break target from inner most scope to outer most scope
	var breakTarget *Block
	scope := g.getBlockScope()
	breakTarget = scope.exit
	addEdge(g.getControl(), breakTarget)
	// Stop control flow from now.
	g.stopControl()
}

func (g *GraphBuilder) buildContinueStmt(node *ast.ContinueStmt) {
	utils.Assert(g.getBlockScope() != nil, "continue statement not in loop")
	// Find continue target from inner most scope to outer most scope
	var continueTarget *Block
	scope := g.getBlockScope()
	continueTarget = scope.post
	addEdge(g.getControl(), continueTarget)
	// Stop control flow from now.
	g.stopControl()
}

func (g *GraphBuilder) buildLetStmt(node *ast.LetStmt) *Value {
	block := g.getControl()
	varName := node.Var.Name
	val := g.build(node.Init)
	g.names[block][varName] = val
	return val
}

func (g *GraphBuilder) buildReturnStmt(node *ast.ReturnStmt) {
	if node.Expr == nil {
		// No return value, just stop control flow then.
		block := g.getControl()
		block.Kind = BlockReturn
		g.stopControl()
		return
	}
	// Evaluate return value and stop control flow then.
	val := g.build(node.Expr)
	block := g.getControl()
	block.Kind = BlockReturn
	val.AddUseBlock(block)
}

func (g *GraphBuilder) build(n ast.AstNode) *Value {
	if g.isStopControl() {
		return nil
	}
	switch n.(type) {
	case *ast.FuncDecl:
		g.build(n.(*ast.FuncDecl).Block)
	case *ast.BlockDecl:
		for _, stmt := range n.(*ast.BlockDecl).Stmts {
			g.build(stmt)
		}

	case *ast.LetStmt:
		g.buildLetStmt(n.(*ast.LetStmt))
	case *ast.ForStmt:
		s := n.(*ast.ForStmt)
		g.buildLoop(s.Init, s.Cond, s.Body, s.Post)
	case *ast.WhileStmt:
		s := n.(*ast.WhileStmt)
		g.buildLoop(nil, s.Cond, s.Body, nil)
	case *ast.SimpleStmt:
		g.build(n.(*ast.SimpleStmt).Expr)
	case *ast.ReturnStmt:
		g.buildReturnStmt(n.(*ast.ReturnStmt))
	case *ast.IfStmt:
		s := n.(*ast.IfStmt)
		g.buildIf(s.Cond, s.Then, s.Else, false /*NoResult*/)
	case *ast.BreakStmt:
		g.buildBreakStmt(n.(*ast.BreakStmt))
	case *ast.ContinueStmt:
		g.buildContinueStmt(n.(*ast.ContinueStmt))

	case *ast.UnaryExpr:
		return g.buildUnaryExpr(n.(*ast.UnaryExpr))
	case *ast.BinaryExpr:
		return g.buildBinaryExpr(n.(*ast.BinaryExpr))
	case *ast.VarExpr:
		return g.lookupVar(n.(*ast.VarExpr).Name, g.getControl())
	case *ast.IntExpr, *ast.DoubleExpr, *ast.BoolExpr, *ast.StrExpr, *ast.ArrayExpr:
		return g.buildConst(n.(ast.AstExpr))
	case *ast.AssignExpr:
		return g.buildAssignExpr(n.(*ast.AssignExpr))
	case *ast.FuncCallExpr:
		return g.buildFunCallExpr(n.(*ast.FuncCallExpr))
	case *ast.IndexExpr:
		return g.buildIndexExpr(n.(*ast.IndexExpr))
	case *ast.TernaryExpr:
		e := n.(*ast.TernaryExpr)
		return g.buildIf(e.Cond, e.Then, e.Else, true /*HasResult*/)
	default:
		utils.Fatal("unimplement %v", n)
	}
	return nil
}

func (g *GraphBuilder) buildParams(params []ast.AstExpr) {
	entry := g.getControl()
	utils.Assert(entry == g.fn.Entry, "sanity check")
	for idx, param := range params {
		val := entry.NewValue(OpParam, param.GetType())
		valName := param.(*ast.VarExpr).Name
		val.Sym = idx
		g.names[entry][valName] = val
	}
}

func CleanHIR(fn *Func) {
	opt := &Optimizer{Func: fn, Debug: false}
	opt.dce()
}

func BuildHIR(funcDecl *ast.FuncDecl) *Func {
	fn := NewFunc(funcDecl.Name)
	entry := fn.NewBlock(BlockReturn)
	entry.Hint = HintEntry
	fn.Entry = entry

	g := NewGraphBuilder(fn)
	g.recordBlock(entry)

	g.setControl(entry)
	g.buildParams(funcDecl.Params)
	g.build(funcDecl.Block)

	finalBlock := g.getControl()
	g.sealBlock(finalBlock)
	finalBlock.Kind = BlockReturn //terminate the program

	g.verify()
	// g.printVars()
	return fn
}

func Compile(funcDecl *ast.FuncDecl, debug bool) *Func {
	fn := BuildHIR(funcDecl)
	CleanHIR(fn)
	VerifyHIR(fn)
	if debug {
		fmt.Printf("== HIR(%s) ==\n", funcDecl.Name)
		fmt.Printf("[[BuildHIR]]\n%v", fn.String())
	}
	OptimizeHIR(fn, debug)
	VerifyHIR(fn)
	if debug {
		fmt.Printf("[[Ideal]]\n%v", fn.String())
		//fmt.Printf("==DU after ideal==\n")
		//fn.PrintDefUses()
	}
	return fn
}
