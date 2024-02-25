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
	"os"
	"strings"
)

// -----------------------------------------------------------------------------
// SSA Value in HIR
// This is the basic unit of SSA form, it represents a value in SSA form. It
// could be a constant, a variable, or an instruction.

type Value struct {
	Id       int
	Op       Op
	Args     []*Value
	Sym      interface{}
	Block    *Block
	Uses     []*Value // values that use this value as arguments
	UseBlock []*Block // block that uses this value as branching condition
	Type     *ast.Type
}

type Op int

const (
	OpAdd Op = iota
	OpSub
	OpMul
	OpDiv
	OpMod

	OpAnd
	OpOr
	OpXor
	OpNot
	OpLShift
	OpRShift

	OpCmpLE
	OpCmpLT
	OpCmpGE
	OpCmpGT
	OpCmpEQ
	OpCmpNE

	OpCInt
	OpCLong
	OpCShort
	OpCFloat
	OpCDouble
	OpChar
	OpCBool
	OpCByte
	OpCString
	OpCArray

	OpPhi
	OpCopy
	OpCall
	OpParam
	OpLoad
	OpStore
	OpLoadIndex
	OpStoreIndex
)

func (x Op) String() string {
	switch x {
	case OpAdd:
		return "Add"
	case OpSub:
		return "Sub"
	case OpMul:
		return "Mul"
	case OpDiv:
		return "Div"
	case OpMod:
		return "Mod"
	case OpAnd:
		return "And"
	case OpOr:
		return "Or"
	case OpXor:
		return "Xor"
	case OpNot:
		return "Not"
	case OpLShift:
		return "LShift"
	case OpRShift:
		return "RShift"
	case OpCmpLE:
		return "CmpLE"
	case OpCmpLT:
		return "CmpLT"
	case OpCmpGE:
		return "CmpGE"
	case OpCmpGT:
		return "CmpGT"
	case OpCmpEQ:
		return "CmpEQ"
	case OpCmpNE:
		return "CmpNE"
	case OpCInt:
		return "CInt"
	case OpCLong:
		return "CLong"
	case OpCShort:
		return "CShort"
	case OpCFloat:
		return "CFloat"
	case OpCDouble:
		return "CDouble"
	case OpChar:
		return "Char"
	case OpCBool:
		return "CBool"
	case OpCByte:
		return "CByte"
	case OpCString:
		return "CString"
	case OpCArray:
		return "CArray"
	case OpPhi:
		return "Phi"
	case OpCopy:
		return "Copy"
	case OpCall:
		return "Call"
	case OpParam:
		return "Param"
	case OpLoad:
		return "Load"
	case OpStore:
		return "Store"
	case OpLoadIndex:
		return "LoadIndex"
	case OpStoreIndex:
		return "StoreIndex"
	}
	return "<Unknown>"
}

func (v *Value) String() string {
	str := fmt.Sprintf("v%v = %v", v.Id, v.Op)
	if v.Type != nil {
		str += fmt.Sprintf("<%v>", v.Type)

	}
	for _, arg := range v.Args {
		str += " "
		str += fmt.Sprintf("v%d", arg.Id)
	}
	if v.Sym != nil {
		str += fmt.Sprintf(" @%v", v.Sym)
	}
	return str
}

func (v *Value) AddArg(args ...*Value) {
	for _, arg := range args {
		v.Args = append(v.Args, arg)
		arg.Uses = append(arg.Uses, v)
	}
}

func (v *Value) AddUseBlock(block *Block) {
	v.UseBlock = append(v.UseBlock, block)
	block.Ctrl = v
}

func (v *Value) RemoveUseBlock(block *Block) {
	for idx, b := range v.UseBlock {
		if b == block {
			v.UseBlock = append(v.UseBlock[:idx], v.UseBlock[idx+1:]...)
			break
		}
	}
	block.Ctrl = nil
}

func (v *Value) RemoveUse(value *Value) {
	for i := len(v.Uses) - 1; i >= 0; i-- {
		if v.Uses[i] == value {
			v.Uses = append(v.Uses[:i], v.Uses[i+1:]...)
		}
	}
}

func (v *Value) ReplaceUses(value *Value) {
	for idx, use := range v.Uses {
		for i, arg := range use.Args {
			if arg == v {
				use.Args[i] = value
				v.Uses[idx] = nil
				value.Uses = append(value.Uses, use)
				break // "use" may uses "v" multiple times
			}
		}
	}
	for i := len(v.Uses) - 1; i >= 0; i-- {
		if v.Uses[i] == nil {
			v.Uses = append(v.Uses[:i], v.Uses[i+1:]...)
		}
	}
	if v.UseBlock != nil {
		value.UseBlock = append(value.UseBlock, v.UseBlock...)
		for _, ub := range value.UseBlock {
			ub.Ctrl = value
		}
		v.UseBlock = nil
	}
	utils.Assert(len(v.Uses) == 0, "v has no uses")
	utils.Assert(len(v.UseBlock) == 0, "v has no use block")
}

// -----------------------------------------------------------------------------
// Basic Block
// A basic block is a straight-line code sequence with no branches in except to
// the entry and no branches out except at the exit. It has a single entry point
// and a single exit point.

type BlockKind int

const (
	BlockIf     BlockKind = iota // block has two successors
	BlockGoto                    // block has only one successor
	BlockReturn                  // block has no successors
	BlockDead                    // block is dead and should be removed later
)

func (kind BlockKind) String() string {
	switch kind {
	case BlockIf:
		return "If"
	case BlockGoto:
		return "Goto"
	case BlockReturn:
		return "Return"
	}
	return "<Unknown>"
}

type BlockHint int

const (
	None BlockHint = iota
	HintEntry
	HintLoopHeader
)

type Block struct {
	Func   *Func
	Id     int
	Kind   BlockKind
	Values []*Value
	Succs  []*Block
	Preds  []*Block
	Ctrl   *Value
	Hint   BlockHint
}

func (block *Block) String() string {
	var str string
	if len(block.Preds) > 0 {
		str = fmt.Sprintf("b%v: [", block.Id)
		for i, pred := range block.Preds {
			if i == len(block.Preds)-1 {
				str += fmt.Sprintf("b%d", pred.Id)
			} else {
				str += fmt.Sprintf("b%d ", pred.Id)
			}
		}
		str += "]\n"
	} else {
		str = fmt.Sprintf("b%v: \n", block.Id)

	}
	var ctrl *Value
	for _, val := range block.Values {
		str += fmt.Sprintf(" %v\n", val)
		for _, buse := range val.UseBlock {
			if buse == block {
				ctrl = val
				break
			}
		}
	}
	if ctrl != nil {
		str += fmt.Sprintf(" %s v%d ", block.Kind.String(), ctrl.Id)
	} else {
		str += fmt.Sprintf(" %s ", block.Kind.String())
	}
	if len(block.Succs) > 0 {
		str += "["
		for i, succ := range block.Succs {
			if i == len(block.Succs)-1 {
				str += fmt.Sprintf("b%d", succ.Id)
			} else {
				str += fmt.Sprintf("b%d ", succ.Id)
			}
		}
		str += "]"
	}
	return str
}

func (block *Block) WireTo(to *Block) {
	block.Succs = append(block.Succs, to)
	to.Preds = append(to.Preds, block)
}

func (block *Block) NewValue(op Op, t *ast.Type, args ...*Value) *Value {
	val := &Value{Id: block.Func.globalValueId, Block: block}
	block.Func.globalValueId++
	val.Op = op
	val.Type = t
	val.Args = make([]*Value, 0)
	for _, arg := range args {
		val.AddArg(arg)
	}
	if op == OpPhi {
		block.Values = append([]*Value{val}, block.Values...)
	} else {
		block.Values = append(block.Values, val)
	}
	return val
}

func (block *Block) RemoveValue(val *Value) {
	for idx, v := range block.Values {
		if v == val {
			// Find all defs and remove it from uses list
			for _, def := range val.Args {
				def.RemoveUse(val)
			}
			// Remove this value from block
			block.Values = append(block.Values[:idx], block.Values[idx+1:]...)
			break
		}
	}
}

func (block *Block) RemoveSucc(succ *Block) bool {
	for idx, s := range block.Succs {
		if s == succ {
			block.Succs = append(block.Succs[:idx], block.Succs[idx+1:]...)
			return true
		}
	}
	return false
}

func (block *Block) RemovePred(pred *Block) bool {
	for idx, p := range block.Preds {
		if p == pred {
			block.Preds = append(block.Preds[:idx], block.Preds[idx+1:]...)
			return true
		}
	}
	return false
}

// -----------------------------------------------------------------------------
// HIR Function
// Abstraction of a function in SSA form.

type Func struct {
	globalValueId int
	globalBlockId int
	Name          string
	Entry         *Block
	Blocks        []*Block
}

func NewFunc(name string) *Func {
	fn := &Func{
		Name:   name,
		Blocks: make([]*Block, 0),
	}
	return fn
}

func (fn *Func) NewBlock(kind BlockKind) *Block {
	block := &Block{
		fn,
		fn.globalBlockId,
		kind,
		make([]*Value, 0),
		make([]*Block, 0),
		make([]*Block, 0),
		nil,
		None,
	}
	fn.globalBlockId++
	fn.Blocks = append(fn.Blocks, block)
	return block
}

func (fn *Func) RemoveBlock(block *Block) {
	for i := len(fn.Blocks) - 1; i >= 0; i-- {
		if fn.Blocks[i] == block {
			fn.Blocks = append(fn.Blocks[:i], fn.Blocks[i+1:]...)
			break
		}
	}
	for i := len(block.Values) - 1; i >= 0; i-- {
		block.RemoveValue(block.Values[i])
	}
}

func (f *Func) String() string {
	var s string
	s += fmt.Sprintf("func %s:\n", f.Name)
	for _, block := range f.Blocks {
		s += fmt.Sprintf("%s\n", block.String())
	}
	return s
}

//------------------------------------------------------------------------------
// Debugging And Verification

func (fn *Func) PrintDefUses() {
	for _, block := range fn.Blocks {
		fmt.Printf("b%d: ", block.Id)
		for _, val := range block.Values {
			if val.UseBlock != nil && len(val.Uses) > 0 {
				ub := ""
				for _, b := range val.UseBlock {
					ub += fmt.Sprintf("b%d ", b.Id)
				}
				fmt.Printf("%v: uses %v ub %v\n", val, val.Uses, ub)
			} else {
				fmt.Printf("%v: uses %v\n", val, val.Uses)
			}
		}
		fmt.Printf("\n")
	}
}

func DumpSSAToDotFile(ssa *Func) {
	f, err := os.Create(fmt.Sprintf("ssa_%s.dot", ssa.Name))
	if err != nil {
		panic(err)
	}
	defer func() {
		utils.ExecuteCmd(".",
			"dot",
			"-Tpng",
			fmt.Sprintf("ssa_%s.dot", ssa.Name),
			"-o",
			fmt.Sprintf("ssa_%s.png", ssa.Name),
		)
		f.Close()
		os.Remove(fmt.Sprintf("ssa_%s.dot", ssa.Name))
	}()
	f.WriteString("digraph G {\n")
	f.WriteString("  graph [ dpi = 500 ];\n")
	for _, block := range ssa.Blocks {
		for i, succ := range block.Succs {
			if i == 1 {
				f.WriteString(fmt.Sprintf("  b%d -> b%d [label=\"F\"]\n", block.Id, succ.Id))
			} else {
				f.WriteString(fmt.Sprintf("  b%d -> b%d\n", block.Id, succ.Id))
			}
		}
	}
	for _, block := range ssa.Blocks {
		blockStr := strings.ReplaceAll(block.String(), "\n", "\\l")
		blockStr = strings.ReplaceAll(blockStr, "<", "\\<")
		blockStr = strings.ReplaceAll(blockStr, ">", "\\>")
		color := "black"
		if block.Hint == HintLoopHeader {
			color = "red"
		}
		f.WriteString(fmt.Sprintf("b%d [shape=record,label=\"{ %s }\",color=\"%s\"]\n",
			block.Id, blockStr, color))
	}
	f.WriteString("}\n")
}

// Continuing on the wrong thing will only lead to more mistakes, we need to stop
// and check step by step, and then proceed further. This is my life philosophy.
func VerifyHIR(fn *Func) {
	// All blocks are reachable from entry now
	reachable := FindReachableBlocks(fn.Entry)
	for _, block := range fn.Blocks {
		if !reachable[block] {
			fmt.Printf("%v", fn)
			utils.Fatal("block b%d is unreachable during verification", block.Id)
		}

	}
	// Incoming arguments of phi match with predecessors
	for _, block := range fn.Blocks {
		for _, val := range block.Values {
			if val.Op != OpPhi {
				continue
			}
			if len(val.Args) != len(block.Preds) {
				fmt.Printf("%v", fn)
				utils.Fatal("phi args mismatch with predecessors")
			}
		}
	}

	// Verify saneness of CFG edges
	for _, block := range fn.Blocks {
		switch block.Kind {
		case BlockGoto:
			if len(block.Succs) != 1 {
				fmt.Printf("%v", fn)
				utils.Fatal("sanity check")
			}
		case BlockIf:
			if len(block.Succs) != 2 {
				fmt.Printf("%v", fn)
				utils.Fatal("sanity check")
			}
			utils.Assert(len(block.Succs) == 2, "sanity check")
		case BlockReturn:
			if len(block.Succs) != 0 {
				fmt.Printf("%v", fn)
				utils.Fatal("sanity check")
			}
		default:
			utils.Unimplement()
		}
	}

	// Ensure def-uses chains are correct
	utils.Assert(len(fn.Entry.Preds) == 0, "sanity check")
	defUses := make(map[*Value][]*Value)
	for _, block := range fn.Blocks {
		for _, val := range block.Values {
			for _, arg := range val.Args {
				defUses[arg] = append(defUses[arg], val)
			}
		}
	}
	// for val, uses := range defUses {
	// 	if len(val.Uses) != len(uses) {
	// 		fmt.Printf("== Broken defUses ==\n%v", fn.String())
	// 		fmt.Printf("=== v.Uses:\n")
	// 		for _, block := range fn.Blocks {
	// 			for _, val := range block.Values {
	// 				fmt.Printf("v%d:%v\n", val.Id, val.Uses)
	// 			}
	// 		}
	// 		fmt.Printf("=== defUses:\n")
	// 		for val, uses := range defUses {
	// 			fmt.Printf("v%d:%v\n", val.Id, uses)
	// 		}
	// 		fmt.Printf("== Bad %v\n", val)
	// 		utils.Fatal("def-uses chains are broken")
	// 	}
	// }

	// All SSA values are typed then..
	for _, block := range fn.Blocks {
		for _, val := range block.Values {
			if val.Type == nil {
				fmt.Printf("%v", fn)
				DumpSSAToDotFile(fn)
				utils.Fatal("SSA value %v is untyped", val)
			}
		}
	}

	// All def dominates its uses
	VerifyDom(fn)
}
