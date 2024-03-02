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
	"math"
)

// == Code conjured by yyang, Feb, 2024 ==

// ------------------------------------------------------------------------------
// Lowering Pass
//

func (lir *LIR) resolvePhi(val *ssa.Value) {
	utils.Assert(val.Op == ssa.OpPhi, "sanity check")
	if len(val.Args) == 1 {
		// Replace phi with copy, this happens when optimization is disabled
		r := lir.NewVReg(val.Args[0])
		lir.NewInstr(val.Block.Id, LIR_Mov, r, r, r).comment(fmt.Sprintf("resolve %v", val.String()))
		lir.SetResult(val, r)
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
		lir.NewInstr(val.Block.Preds[i].Id, LIR_Mov, res, r, res).comment(fmt.Sprintf("resolve %v", val.String()))
	}
	lir.SetResult(val, res)
}

func getCondLirOp(ssaOp ssa.Op) LIROp {
	switch ssaOp {
	case ssa.OpCmpLE:
		return LIR_CmpLE
	case ssa.OpCmpLT:
		return LIR_CmpLT
	case ssa.OpCmpGE:
		return LIR_CmpGE
	case ssa.OpCmpGT:
		return LIR_CmpGT
	case ssa.OpCmpEQ:
		return LIR_CmpEQ
	case ssa.OpCmpNE:
		return LIR_CmpNE
	}
	utils.ShouldNotReachHere()
	return 0
}

func getCondJumpLirOp(ssaOp ssa.Op) LIROp {
	switch ssaOp {
	case ssa.OpCmpLE:
		return LIR_Jle
	case ssa.OpCmpLT:
		return LIR_Jlt
	case ssa.OpCmpGE:
		return LIR_Jge
	case ssa.OpCmpGT:
		return LIR_Jgt
	case ssa.OpCmpEQ:
		return LIR_Jeq
	case ssa.OpCmpNE:
		return LIR_Jne
	}
	utils.ShouldNotReachHere()
	return 0
}

func (lir *LIR) lowerCompare(val *ssa.Value) {
	left := lir.NewVReg(val.Args[0])
	right := lir.NewVReg(val.Args[1])
	lirOp := getCondLirOp(val.Op)
	// compare is used by value or block explicitly?
	if len(val.Uses) != 0 || val.UseBlock != nil {
		res := lir.NewVReg(val)
		lir.NewInstr(val.Block.Id, lirOp, res, right, left).comment(val)
		lir.SetResult(val, res)
	} else {
		// compare is used by control flow only
		lir.NewInstr(val.Block.Id, lirOp, right, right, left).comment(val)
		lir.SetResult(val, right)
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
		lir.NewInstr(val.Block.Id, LIR_Mov, result, left, result).comment(val)
		lir.NewInstr(val.Block.Id, lirOp, result, right, result).comment(val)
		lir.SetResult(val, result)
	case ssa.OpLShift, ssa.OpRShift:
		// Multiply dst by 2, CL times.
		// Signed divide dst by 2, CL times.
		left := lir.NewVReg(val.Args[0])
		right := lir.NewVReg(val.Args[1])
		result := lir.NewVReg(val)
		var reg Register
		for _, r := range []Register{RCX, ECX, CX, CL} {
			if r.GetType() == GetLIRType(val.Type) {
				reg = r
				break
			}
		}
		// move left to result
		lir.NewInstr(val.Block.Id, LIR_Mov, result, left, result).comment(val)
		// move right(shift count) to reg
		lir.NewInstr(val.Block.Id, LIR_Mov, reg, right, reg).comment(val)
		// CL is mandatory for rshift even if reg is ECX or RCX, etc
		lirOp := LIR_LShift
		if val.Op == ssa.OpRShift {
			lirOp = LIR_RShift
		}
		// shift left/right by CL, stored in result
		lir.NewInstr(val.Block.Id, lirOp, result, CL, result).comment(val)
		lir.SetResult(val, result)
	case ssa.OpNot:
		left := lir.NewVReg(val.Args[0])
		result := lir.NewVReg(val)
		lir.NewInstr(val.Block.Id, LIR_Mov, result, left, result).comment(val)
		lir.NewInstr(val.Block.Id, LIR_Not, result, result).comment(val)
	case ssa.OpMul:
		left := lir.NewVReg(val.Args[0])
		right := lir.NewVReg(val.Args[1])
		result := lir.NewVReg(val)
		// The dst of mul must be register, only the source can optionally be
		// memory, so we should load dst into a physical regsiter, this is different
		// from other instructions
		freeRegs := CallerSaveRegs(GetLIRType(val.Type))
		tempReg := freeRegs[0]
		lir.NewInstr(val.Block.Id, LIR_Mov, tempReg, left, tempReg).comment(val)
		lir.NewInstr(val.Block.Id, LIR_Mul, tempReg, right, tempReg).comment(val)
		lir.NewInstr(val.Block.Id, LIR_Mov, result, tempReg, result).comment(val)
		lir.SetResult(val, result)
	case ssa.OpDiv, ssa.OpMod:
		// Divides the signed value in the RAX (dividend) by the source operand
		// (divisor) and stores the result in the RAX
		left := lir.NewVReg(val.Args[0])
		right := lir.NewVReg(val.Args[1])
		result := lir.NewVReg(val)

		// Find the register that can hold the dividend
		var dividendReg Register
		for _, reg := range []Register{RAX, EAX, AX, AL} {
			if reg.GetType() == GetLIRType(val.Type) {
				dividendReg = reg
				break
			}
		}
		lir.NewInstr(val.Block.Id, LIR_Mov, dividendReg, left, dividendReg).comment(val)
		lir.NewInstr(val.Block.Id, LIR_Div, right, right).comment(val)
		if val.Op == ssa.OpDiv {
			// Quotient is stored in RAX
			lir.NewInstr(val.Block.Id, LIR_Mov, result, dividendReg, result).comment(val)
		} else {
			// Rem is stored in RDX
			var remReg Register
			for _, reg := range []Register{RDX, EDX, DX, DL} {
				if reg.GetType() == GetLIRType(val.Type) {
					remReg = reg
					break
				}
			}
			lir.NewInstr(val.Block.Id, LIR_Mov, result, remReg, result).comment(val)
		}
		lir.SetResult(val, result)
	case ssa.OpNegate:
		left := lir.NewVReg(val.Args[0])
		result := lir.NewVReg(val)
		lir.NewInstr(val.Block.Id, LIR_Mov, result, left, result).comment(val)
		lir.NewInstr(val.Block.Id, LIR_Xor, result, lir.NewImm(1), result).comment(val)
	default:
		utils.Unimplement()
	}
}

func (lir *LIR) lowerCall(val *ssa.Value) {
	utils.Assert(val.Op == ssa.OpCall, "sanity check")

	for i, arg := range val.Args {
		r := lir.NewVReg(arg)
		lir.NewInstr(val.Block.Id,
			LIR_Mov,
			ArgReg(i, GetLIRType(arg.Type)),
			r,
			ArgReg(i, GetLIRType(arg.Type)),
		).comment(val)
	}
	retReg := ReturnReg(GetLIRType(val.Type))
	lir.NewInstr(val.Block.Id, LIR_Call, retReg, Symbol{val.Sym.(string)}).comment(val)
	res := lir.NewVReg(val)
	if retReg != NoReg {
		// mov ret_val, res
		lir.NewInstr(val.Block.Id, LIR_Mov, res, retReg, res).comment(val)
	}
	lir.SetResult(val, res)
}

func (lir *LIR) lowerConst(val *ssa.Value) {
	utils.Assert(val.Op == ssa.OpConst, "sanity check")
	t := val.Type
	switch {
	case t.IsInt():
		r := Imm{LIRTypeDWord, val.Sym.(int)}
		res := lir.NewVReg(val)
		lir.NewInstr(val.Block.Id, LIR_Mov, res, r, res).comment(val)
		lir.SetResult(val, res)
	case t.IsShort():
		r := Imm{LIRTypeWord, val.Sym.(int16)}
		res := lir.NewVReg(val)
		lir.NewInstr(val.Block.Id, LIR_Mov, res, r, res).comment(val)
		lir.SetResult(val, res)
	case t.IsLong():
		r := Imm{LIRTypeQWord, val.Sym.(int64)}
		res := lir.NewVReg(val)
		lir.NewInstr(val.Block.Id, LIR_Mov, res, r, res).comment(val)
		lir.SetResult(val, res)
	case t.IsBool():
		b := 0
		if val.Sym.(bool) {
			b = 1
		}
		r := Imm{LIRTypeDWord, b}
		res := lir.NewVReg(val)
		lir.NewInstr(val.Block.Id, LIR_Mov, res, r, res).comment(val)
		lir.SetResult(val, res)
	case t.IsChar():
		r := Imm{LIRTypeByte, val.Sym.(int8)}
		res := lir.NewVReg(val)
		lir.NewInstr(val.Block.Id, LIR_Mov, res, r, res).comment(val)
		lir.SetResult(val, res)
	case t.IsFloat():
		utils.Unimplement()
	case t.IsDouble():
		hex := fmt.Sprintf("%x", math.Float64bits(val.Sym.(float64)))
		text := lir.NewText(fmt.Sprintf("0x%s", hex), TextFloat)
		// Load double literal from rodata
		addr := lir.NewAddr(LIRTypeVector16D, RIP, NoReg, text)
		res := lir.NewVReg(val)
		lir.NewInstr(val.Block.Id, LIR_Mov, res, addr, res).comment(val)
		lir.SetResult(val, res)
	case t.IsString():
		// arg0: ptr of string
		str := val.Sym.(string)
		ptrArg := ArgReg(0, LIRTypeDWord)
		lir.NewInstr(val.Block.Id, LIR_Mov, ptrArg, lir.NewText(str, TextString), ptrArg).comment(val)
		// arg1: len of string
		lenArg := ArgReg(1, LIRTypeDWord)
		lir.NewInstr(val.Block.Id, LIR_Mov, lenArg, lir.NewImm(len(str)), lenArg).comment(val)
		// call runtime stub
		retReg := ReturnReg(LIRTypeQWord)
		lir.NewInstr(val.Block.Id, LIR_Call, retReg, Symbol{"runtime_new_string"}).comment(val)
		// save result string
		res := lir.NewVReg(val)
		lir.NewInstr(val.Block.Id, LIR_Mov, res, retReg, res).comment(val)
		lir.SetResult(val, res)
	case t.IsArray():
		// arg0: len of array
		lenArg := ArgReg(0, LIRTypeDWord)
		lir.NewInstr(val.Block.Id, LIR_Mov, lenArg, lir.NewImm(val.Sym.(int)), lenArg).comment(val)
		// call runtime stub
		// FIXME: Return type should be arch dependent
		retReg := ReturnReg(LIRTypeQWord)
		lir.NewInstr(val.Block.Id, LIR_Call, retReg, Symbol{"runtime_new_array"}).comment(val)
		res := lir.NewVReg(val)
		lir.NewInstr(val.Block.Id, LIR_Mov, res, retReg, res).comment(val)
		lir.SetResult(val, res)
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
			lir.NewInstr(val.Block.Id, LIR_Mov, addr, elem, addr).comment(val)
		}
	case ssa.OpLoadIndex:
		if argVar.Type.IsString() {
			// load char from string
			// deference string to get data field
			base := lir.NewVReg(argVar)
			dataAddr := lir.NewAddr(LIRTypeQWord, base, NoReg, lir.NewOffset(0))
			freeRegs := CallerSaveRegs(LIRTypeQWord)
			dataRes := freeRegs[0]
			lir.NewInstr(val.Block.Id, LIR_Mov, dataRes, dataAddr, dataRes).comment("load string.data")
			// load element from dataAddr
			result := lir.NewVReg(val)
			index := lir.NewVReg(argIndex)
			charAddr := lir.NewAddr(LIRTypeByte, dataRes, index, lir.NewOffset(0))
			lir.NewInstr(val.Block.Id, LIR_Mov, result, charAddr, result).comment("load str.data[index]")
		} else {
			// load element from array
			base := lir.NewVReg(argVar)
			index := lir.NewVReg(argIndex)
			addr := lir.NewAddr(GetLIRType(val.Type), base, index, lir.NewOffset(0))
			result := lir.NewVReg(val)
			lir.NewInstr(val.Block.Id, LIR_Mov, result, addr, result).comment(val)
		}
	default:
		utils.ShouldNotReachHere()
	}
}

func (lir *LIR) lowerValue(val *ssa.Value) {
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
		lir.NewInstr(val.Block.Id, LIR_Mov,
			result, ArgReg(iarg, GetLIRType(val.Type)), result).comment(val)
		lir.SetResult(val, result)
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
	for _, val := range block.Values {
		if val.Op == ssa.OpPhi {
			lir.resolvePhi(val)
		} else {
			lir.lowerValue(val)
		}
	}
	for _, succ := range block.Succs {
		lir.lowerBlock(visited, succ)
	}
}

func (lir *LIR) lowerBlockControl(block *ssa.Block) {
	switch block.Kind {
	case ssa.BlockGoto:
		lir.NewJmp(block.Id, LIR_Jmp, block.Succs[0]).comment(block.Succs[0])
	case ssa.BlockReturn:
		ctrl := block.Ctrl
		if ctrl != nil {
			// Return with value
			left := lir.NewVReg(ctrl)
			retReg := ReturnReg(GetLIRType(ctrl.Type))
			lir.NewInstr(block.Id, LIR_Mov, retReg, left, retReg).comment(ctrl)
			lir.SetResult(ctrl, left)
		}
		// Pure return
		lir.NewInstr(block.Id, LIR_Ret, NoReg).comment("ret")
	case ssa.BlockIf:
		ctrl := block.Ctrl
		switch ctrl.Op {
		case ssa.OpCmpLT, ssa.OpCmpLE, ssa.OpCmpGT, ssa.OpCmpGE, ssa.OpCmpEQ, ssa.OpCmpNE:
			lir.NewJmp(block.Id, getCondJumpLirOp(ctrl.Op), block.Succs[0]).comment(block.Succs[0])
			lir.NewJmp(block.Id, LIR_Jmp, block.Succs[1]).comment(block.Succs[1])
		default:
			// @@ Note this is not much obvious, it jumps when condition is false!!
			// Imm(1)&0 => 0, set zf = 1
			// Imm(1)&1 => 1, set zf = 0
			// jeq jumps if zf = 1
			r := lir.NewImm(1)
			res := lir.NewVReg(ctrl)
			lir.NewInstr(block.Id, LIR_CmpEQ, res, r, res).comment(block)
			lir.SetResult(ctrl, res)
			lir.NewJmp(block.Id, LIR_Jeq, block.Succs[0]).comment(block.Succs[0])
			lir.NewJmp(block.Id, LIR_Jmp, block.Succs[1]).comment(block.Succs[1])
		}
	}
}

func Lower(fn *ssa.Func) *LIR {
	lir := NewLIR(fn)

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
