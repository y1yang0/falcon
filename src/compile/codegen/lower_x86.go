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
package codegen

import (
	"falcon/compile/ssa"
	"falcon/utils"
	"fmt"
)

// == Code conjured by yyang, Feb, 2024 ==

type LIR struct {
	bid          int                    // current block id
	vid          int                    // global virtual register id
	v2r          map[int]Register       // Value id to virtual register
	roid         int                    // global read-only section id
	Name         string                 // function Name
	Instructions map[int][]*Instruction // blocks of instructions, order of blocks is important
	Labels       map[int]Label          // labels for each block for continuation point
	Texts        []Text                 // read-only section literals(string/quads/longs)
}

func (lir *LIR) String() string {
	str := ""
	for idx, instrs := range lir.Instructions {
		str += fmt.Sprintf("b%d:\n", idx)
		for _, instr := range instrs {
			str += fmt.Sprintf("  %v %v", instr.Op, instr.Result)
			for _, arg := range instr.Args {
				str += fmt.Sprintf(", %v", arg)
			}
			str += fmt.Sprintf("  # %v", instr.Comment)
			str += "\n"
		}
	}
	return str
}

func (lir *LIR) Bind(val *ssa.Value, result IOperand) {
	lir.v2r[val.Id] = result.(Register)
}

func (lir *LIR) NewInstr(op LIROp, args ...IOperand) *Instruction {
	return lir.NewInstrTo(lir.bid, op, args...)
}

func (lir *LIR) NewInstrTo(idx int, op LIROp, args ...IOperand) *Instruction {
	utils.Assert(len(args) > 0, "at least one argument")
	result := args[0]
	instr := &Instruction{Op: op, Result: result, Args: args[1:]}
	lir.Instructions[idx] = append(lir.Instructions[idx], instr)
	return instr
}

func (lir *LIR) NewLabel(idx int) Label {
	target := Label{Name: fmt.Sprintf("L%d", idx)}

	lir.Labels[idx] = target
	return target
}

func (lir *LIR) NewText(value string, kind TextKind) Text {
	target := Text{Value: value, Id: lir.roid, Kind: kind}
	lir.Texts = append(lir.Texts, target)
	lir.roid++
	return target
}

func (lir *LIR) NewJmp(op LIROp, block *ssa.Block) *Instruction {
	return lir.NewJmpTo(lir.bid, op, block)
}

func (lir *LIR) NewJmpTo(idx int, op LIROp, block *ssa.Block) *Instruction {
	var target Label
	if label, ok := lir.Labels[block.Id]; ok {
		target = label
	} else {
		target = lir.NewLabel(block.Id)
	}
	instr := &Instruction{Op: op, Result: target}
	lir.Instructions[idx] = append(lir.Instructions[idx], instr)
	lir.Labels[block.Id] = target
	return instr
}

func (lir *LIR) NewOffset(offset int) Offset {
	return Offset{offset}
}

func (lir *LIR) NewImm(val interface{}) IOperand {
	// we don't really need lir receiver, but leave it for consistency
	switch val.(type) {
	case int:
		return Imm{LIRTypeDWord, val.(int)}
	case bool:
		if val.(bool) {
			return Imm{LIRTypeDWord, 1}
		}
		return Imm{LIRTypeDWord, 0}
	default:
		utils.Unimplement()
	}
	return Imm{}
}

func (lir *LIR) NewAddr(t *LIRType, base Register, index Register, disp IOperand) Addr {
	return Addr{t, base, index, t.Width, disp}
}

func (lir *LIR) NewVReg(v *ssa.Value) Register {
	if r, ok := lir.v2r[v.Id]; ok {
		return r
	}
	r := Register{Index: lir.vid, Virtual: true, Type: GetLIRType(v.Type)}
	lir.vid++
	lir.v2r[v.Id] = r
	return r
}

func (x *Instruction) comment(v interface{}) {
	switch v := v.(type) {
	case *ssa.Block:
		x.Comment = fmt.Sprintf("b%d", v.Id)
	case *ssa.Value:
		x.Comment = fmt.Sprintf("%v", v)
	case string:
		x.Comment = v
	default:
		utils.Unimplement()
	}
}

func VerifyLIR(lir *LIR) {
	// verify that all instructions have a result
	for _, instrs := range lir.Instructions {
		for _, instr := range instrs {
			utils.Assert(instr.Result != nil, "miss result")
			utils.Assert(len(instr.Args) >= 0 && len(instr.Args) <= 2, "miss args")
		}
	}
}

func NewLIR(fn *ssa.Func) *LIR {
	return &LIR{
		bid:          0,
		vid:          0,
		roid:         0,
		v2r:          make(map[int]Register),
		Name:         fn.Name,
		Instructions: make(map[int][]*Instruction, len(fn.Blocks)), //order is important
		Labels:       make(map[int]Label),
		Texts:        make([]Text, 0),
	}
}

// -----------------------------------------------------------------------------
// HIR Destruction
//
// splitCritical eliminates critical edges in the CFG by inserting new safe lands
// between the critical edges. This is a necessary step before LIR generation to
// resolve phi nodes, as phi nodes may inserts move instructions in the incoming
// blocks, which breaks program semantics in the presence of critical edges.
func splitCritical(fn *ssa.Func) {
	//  b1  b2
	//  / \ /
	//     b3
	// To resolve phi in b3, we need to split critical edge from b1 to b3, i.e.
	// create safe land to insert move instruction for phi
	//  b1   b2
	//  / \   |
	//     b4 |
	//       \|
	//       b3
	oldBlocks := make([]*ssa.Block, len(fn.Blocks))
	copy(oldBlocks, fn.Blocks)
	for _, block := range oldBlocks {
		if len(block.Preds) > 1 {
			for _, pred := range block.Preds {
				if len(pred.Succs) > 1 {
					// Create new safe land
					safeLand := fn.NewBlock(ssa.BlockGoto)
					// Insert safe land between pred and block
					safeLand.InsertBetween(pred, block)
				}
			}
		}
	}
}

func (lir *LIR) thawPhi(val *ssa.Value) {
	utils.Assert(val.Op == ssa.OpPhi, "sanity check")
	if len(val.Args) == 1 {
		// Replace phi with copy, this happens when optimization is disabled
		r := lir.NewVReg(val.Args[0])
		lir.NewInstr(LIR_Mov, r, r, r).comment(fmt.Sprintf("resolve %v", val.String()))
		lir.Bind(val, r)
		return
	}
	// Before
	//  v1 = ... v2 = ...
	//    \       /
	//     \     /
	// v3 = phi v1, v2
	//
	// After
	//  r1 = ... r1 = ...
	//    \       /
	//     \     /
	//   mov r2, r1
	res := lir.NewVReg(val)
	for i := 0; i < len(val.Block.Preds); i++ {
		// Find the incoming value
		r := lir.NewVReg(val.Args[i])
		// Insert a move instruction at pred block
		lir.NewInstrTo(val.Block.Preds[i].Id, LIR_Mov, res, r, res).comment(fmt.Sprintf("resolve %v", val.String()))
	}
	lir.Bind(val, res)
}

func (lir *LIR) lowerCompare(val *ssa.Value) {
	left := lir.NewVReg(val.Args[0])
	right := lir.NewVReg(val.Args[1])
	lirOp := getCondLirOp(val.Op)
	// compare is used by value or block explicitly?
	if len(val.Uses) != 0 || val.UseBlock != nil {
		res := lir.NewVReg(val)
		lir.NewInstr(lirOp, res, right, left).comment(val)
		lir.Bind(val, res)
	} else {
		// compare is used by control flow only
		lir.NewInstr(lirOp, right, right, left).comment(val)
		lir.Bind(val, right)
	}
}

func (lir *LIR) lowerArithmetic(val *ssa.Value) {
	switch val.Op {
	case ssa.OpAdd, ssa.OpSub,
		ssa.OpAnd, ssa.OpOr, ssa.OpXor:
		// dest := dest Op src
		ssaOp2LIROp := map[ssa.Op]LIROp{
			ssa.OpAdd: LIR_Add,
			ssa.OpSub: LIR_Sub,
			ssa.OpAnd: LIR_And,
			ssa.OpOr:  LIR_Or,
			ssa.OpXor: LIR_Xor,
		}
		lirOp, exist := ssaOp2LIROp[val.Op]
		utils.Assert(exist, "unimplemented arithmetic op %v", val.Op)
		left := lir.NewVReg(val.Args[0])
		right := lir.NewVReg(val.Args[1])
		result := lir.NewVReg(val)
		lir.NewInstr(LIR_Mov, result, left, result).comment(val)
		lir.NewInstr(lirOp, result, right, result).comment(val)
		lir.Bind(val, result)
	case ssa.OpLShift, ssa.OpRShift:
		// Multiply dst by 2, CL times.
		// Signed divide dst by 2, CL times.
		left := lir.NewVReg(val.Args[0])
		right := lir.NewVReg(val.Args[1])
		result := lir.NewVReg(val)
		reg := RCX.Cast(GetLIRType(val.Type))
		// move left to result
		lir.NewInstr(LIR_Mov, result, left, result).comment(val)
		// move right(shift count) to reg
		lir.NewInstr(LIR_Mov, reg, right, reg).comment(val)
		// CL is mandatory for rshift even if reg is ECX or RCX, etc
		lirOp := LIR_LShift
		if val.Op == ssa.OpRShift {
			lirOp = LIR_RShift
		}
		// shift left/right by CL, stored in result
		lir.NewInstr(lirOp, result, CL, result).comment(val)
		lir.Bind(val, result)
	case ssa.OpNot:
		left := lir.NewVReg(val.Args[0])
		result := lir.NewVReg(val)
		lir.NewInstr(LIR_Mov, result, left, result).comment(val)
		lir.NewInstr(LIR_Not, result, result).comment(val)
	case ssa.OpMul:
		left := lir.NewVReg(val.Args[0])
		right := lir.NewVReg(val.Args[1])
		result := lir.NewVReg(val)
		// The dst of mul must be register, only the source can optionally be
		// memory, so we should load dst into a physical regsiter, this is different
		// from other instructions
		freeRegs := CallerSaveRegs(GetLIRType(val.Type))
		tempReg := freeRegs[0]
		lir.NewInstr(LIR_Mov, tempReg, left, tempReg).comment(val)
		lir.NewInstr(LIR_Mul, tempReg, right, tempReg).comment(val)
		lir.NewInstr(LIR_Mov, result, tempReg, result).comment(val)
		lir.Bind(val, result)
	case ssa.OpDiv, ssa.OpMod:
		// Divides the signed value in the RAX (dividend) by the source operand
		// (divisor) and stores the result in the RAX
		left := lir.NewVReg(val.Args[0])
		right := lir.NewVReg(val.Args[1])
		result := lir.NewVReg(val)

		// Find the register that can hold the dividend or remainder
		dividendReg := RAX.Cast(GetLIRType(val.Type))
		lir.NewInstr(LIR_Mov, dividendReg, left, dividendReg).comment(val)
		lir.NewInstr(LIR_Div, right, right).comment(val)
		// Store quotient/remainder in result
		outputReg := dividendReg
		if val.Op == ssa.OpMod {
			// Remainder is stored in RDX, while quotient is stored in RAX
			outputReg = RDX.Cast(GetLIRType(val.Type))
		}
		lir.NewInstr(LIR_Mov, result, outputReg, result).comment(val)
		lir.Bind(val, result)
	case ssa.OpNegate:
		left := lir.NewVReg(val.Args[0])
		result := lir.NewVReg(val)
		lir.NewInstr(LIR_Mov, result, left, result).comment(val)
		lir.NewInstr(LIR_Xor, result, lir.NewImm(1), result).comment(val)
	default:
		utils.Unimplement()
	}
}

func (lir *LIR) lowerCall(val *ssa.Value) {
	utils.Assert(val.Op == ssa.OpCall, "sanity check")

	for i, arg := range val.Args {
		r := lir.NewVReg(arg)
		argReg := ArgReg(i, GetLIRType(arg.Type))
		lir.NewInstr(LIR_Mov, argReg, r, argReg).comment(val)
	}
	retReg := ReturnReg(GetLIRType(val.Type))
	lir.NewInstr(LIR_Call, retReg, Symbol{val.Sym.(string)}).comment(val)
	res := lir.NewVReg(val)
	if retReg != NoReg {
		// mov ret_val, res
		lir.NewInstr(LIR_Mov, res, retReg, res).comment(val)
	}
	lir.Bind(val, res)
}

func (lir *LIR) lowerConst(val *ssa.Value) {
	utils.Assert(val.Op == ssa.OpConst, "sanity check")
	t := val.Type
	switch {
	case t.IsInt():
		r := Imm{LIRTypeDWord, val.Sym.(int)}
		res := lir.NewVReg(val)
		lir.NewInstr(LIR_Mov, res, r, res).comment(val)
		lir.Bind(val, res)
	case t.IsShort():
		r := Imm{LIRTypeWord, val.Sym.(int16)}
		res := lir.NewVReg(val)
		lir.NewInstr(LIR_Mov, res, r, res).comment(val)
		lir.Bind(val, res)
	case t.IsLong():
		r := Imm{LIRTypeQWord, val.Sym.(int64)}
		res := lir.NewVReg(val)
		lir.NewInstr(LIR_Mov, res, r, res).comment(val)
		lir.Bind(val, res)
	case t.IsBool():
		b := 0
		if val.Sym.(bool) {
			b = 1
		}
		r := Imm{LIRTypeDWord, b}
		res := lir.NewVReg(val)
		lir.NewInstr(LIR_Mov, res, r, res).comment(val)
		lir.Bind(val, res)
	case t.IsChar():
		r := Imm{LIRTypeByte, val.Sym.(int8)}
		res := lir.NewVReg(val)
		lir.NewInstr(LIR_Mov, res, r, res).comment(val)
		lir.Bind(val, res)
	case t.IsFloat():
		utils.Unimplement()
	case t.IsDouble():
		text := lir.NewText(utils.Float64ToHex(val.Sym.(float64)), TextFloat)
		// Load double literal from rodata
		addr := lir.NewAddr(LIRTypeVector16D, RIP, NoReg, text)
		res := lir.NewVReg(val)
		lir.NewInstr(LIR_Mov, res, addr, res).comment(val)
		lir.Bind(val, res)
	case t.IsString():
		// arg0: ptr of string
		str := val.Sym.(string)
		ptrArg := ArgReg(0, LIRTypeDWord)
		lir.NewInstr(LIR_Mov, ptrArg, lir.NewText(str, TextString), ptrArg).comment(val)
		// arg1: len of string
		lenArg := ArgReg(1, LIRTypeDWord)
		lir.NewInstr(LIR_Mov, lenArg, lir.NewImm(len(str)), lenArg).comment(val)
		// call runtime stub
		retReg := ReturnReg(LIRTypeQWord)
		lir.NewInstr(LIR_Call, retReg, Symbol{"runtime_new_string"}).comment(val)
		// save result string
		res := lir.NewVReg(val)
		lir.NewInstr(LIR_Mov, res, retReg, res).comment(val)
		lir.Bind(val, res)
	case t.IsArray():
		// arg0: len of array
		lenArg := ArgReg(0, LIRTypeDWord)
		lir.NewInstr(LIR_Mov, lenArg, lir.NewImm(val.Sym.(int)), lenArg).comment(val)
		// call runtime stub
		// FIXME: Return type should be arch dependent
		retReg := ReturnReg(LIRTypeQWord)
		lir.NewInstr(LIR_Call, retReg, Symbol{"runtime_new_array"}).comment(val)
		res := lir.NewVReg(val)
		lir.NewInstr(LIR_Mov, res, retReg, res).comment(val)
		lir.Bind(val, res)
	default:
		utils.Unimplement()
	}
}

func (lir *LIR) lowerIndexed(val *ssa.Value) {
	utils.Assert(val.Op == ssa.OpLoadIndex || val.Op == ssa.OpStoreIndex, "sanity check")
	argVar := val.Args[0]
	argIndex := val.Args[1]
	switch val.Op {
	case ssa.OpStoreIndex:
		if argVar.Type.IsString() {
			utils.Fatal("string is immutable")
		} else {
			argValue := val.Args[2]
			base := lir.NewVReg(argVar)
			index := lir.NewVReg(argIndex)
			elem := lir.NewVReg(argValue)
			addr := lir.NewAddr(elem.Type, base, index, lir.NewOffset(0))
			lir.NewInstr(LIR_Mov, addr, elem, addr).comment(val)
		}
	case ssa.OpLoadIndex:
		if argVar.Type.IsString() {
			// load char from string
			// deference string to get data field
			base := lir.NewVReg(argVar)
			dataAddr := lir.NewAddr(LIRTypeQWord, base, NoReg, lir.NewOffset(0))
			freeRegs := CallerSaveRegs(LIRTypeQWord)
			dataRes := freeRegs[0]
			lir.NewInstr(LIR_Mov, dataRes, dataAddr, dataRes).comment("load string.data")
			// load element from dataAddr
			result := lir.NewVReg(val)
			index := lir.NewVReg(argIndex)
			charAddr := lir.NewAddr(LIRTypeByte, dataRes, index, lir.NewOffset(0))
			lir.NewInstr(LIR_Mov, result, charAddr, result).comment("load str.data[index]")
		} else {
			// load element from array
			base := lir.NewVReg(argVar)
			index := lir.NewVReg(argIndex)
			addr := lir.NewAddr(GetLIRType(val.Type), base, index, lir.NewOffset(0))
			result := lir.NewVReg(val)
			lir.NewInstr(LIR_Mov, result, addr, result).comment(val)
		}
	default:
		utils.ShouldNotReachHere()
	}
}

func (lir *LIR) lowerValue(val *ssa.Value) {
	utils.Assert(lir.bid == val.Block.Id, "sanity check")
	switch val.Op {
	case ssa.OpConst:
		lir.lowerConst(val)
	case ssa.OpAdd, ssa.OpSub, ssa.OpMul, ssa.OpDiv, ssa.OpMod,
		ssa.OpAnd, ssa.OpOr, ssa.OpXor, ssa.OpNot, ssa.OpLShift, ssa.OpRShift,
		ssa.OpNegate:
		lir.lowerArithmetic(val)
	case ssa.OpPhi:
		// Phi should be already resolved
		utils.ShouldNotReachHere()
	case ssa.OpCmpLT, ssa.OpCmpLE, ssa.OpCmpGT, ssa.OpCmpGE, ssa.OpCmpEQ, ssa.OpCmpNE:
		lir.lowerCompare(val)
	case ssa.OpParam:
		iarg := val.Sym.(int)
		result := lir.NewVReg(val)
		lir.NewInstr(LIR_Mov,
			result, ArgReg(iarg, GetLIRType(val.Type)), result).comment(val)
		lir.Bind(val, result)
	case ssa.OpCall:
		lir.lowerCall(val)
	case ssa.OpStoreIndex, ssa.OpLoadIndex:
		lir.lowerIndexed(val)
	default:
		utils.Unimplement()
	}
}

func (lir *LIR) lowerBlock(visited map[*ssa.Block]bool, block *ssa.Block) {
	if _, exist := visited[block]; exist {
		return
	}
	visited[block] = true

	for _, pred := range block.Preds {
		if _, exist := visited[pred]; !exist {
			lir.lowerBlock(visited, pred)
			visited[pred] = true
		}
	}
	lir.bid = block.Id
	for _, val := range block.Values {
		if val.Op == ssa.OpPhi {
			lir.thawPhi(val)
		} else {
			lir.lowerValue(val)
		}
	}
	for _, succ := range block.Succs {
		lir.lowerBlock(visited, succ)
	}
}

func (lir *LIR) lowerBlockControl(block *ssa.Block) {
	lir.bid = block.Id
	switch block.Kind {
	case ssa.BlockGoto:
		lir.NewJmp(LIR_Jmp, block.Succs[0]).comment(block.Succs[0])
	case ssa.BlockReturn:
		ctrl := block.Ctrl
		if ctrl != nil {
			// Return with value
			left := lir.NewVReg(ctrl)
			retReg := ReturnReg(GetLIRType(ctrl.Type))
			lir.NewInstr(LIR_Mov, retReg, left, retReg).comment(ctrl)
			lir.Bind(ctrl, left)
		}
		// Pure return
		lir.NewInstr(LIR_Ret, NoReg).comment("ret")
	case ssa.BlockIf:
		ctrl := block.Ctrl
		switch ctrl.Op {
		case ssa.OpCmpLT, ssa.OpCmpLE, ssa.OpCmpGT, ssa.OpCmpGE, ssa.OpCmpEQ, ssa.OpCmpNE:
			lir.NewJmp(getCondJumpLirOp(ctrl.Op), block.Succs[0]).comment(block.Succs[0])
			lir.NewJmp(LIR_Jmp, block.Succs[1]).comment(block.Succs[1])
		default:
			// @@ Note this is not much obvious, it jumps when condition is false!!
			// Imm(1)&0 => 0, set zf = 1
			// Imm(1)&1 => 1, set zf = 0
			// jeq jumps if zf = 1
			r := lir.NewImm(1)
			res := lir.NewVReg(ctrl)
			lir.NewInstr(LIR_CmpEQ, res, r, res).comment(block)
			lir.Bind(ctrl, res)
			lir.NewJmp(LIR_Jeq, block.Succs[0]).comment(block.Succs[0])
			lir.NewJmp(LIR_Jmp, block.Succs[1]).comment(block.Succs[1])
		}
	}
}

func Lower(fn *ssa.Func) *LIR {
	lir := NewLIR(fn)

	// Split critical edge
	splitCritical(fn)
	ssa.VerifyHIR(fn)

	// Prepare all label
	for _, block := range fn.Blocks {
		lir.NewLabel(block.Id)
	}

	// Do LIR generation in pre-order
	lir.lowerBlock(make(map[*ssa.Block]bool), fn.Entry)

	// Post-processing all blocks according to their kind
	for _, block := range fn.Blocks {
		lir.lowerBlockControl(block)
	}

	VerifyLIR(lir)
	return lir
}
