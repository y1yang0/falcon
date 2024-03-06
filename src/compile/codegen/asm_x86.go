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
	"falcon/utils"
	"fmt"
	"sort"
	"strings"
)

type Assembler struct {
	buf         string
	stackOffset int
	v2offset    map[int]int
	funcIndex   int

	scalarScratch []Register
	floatScratch  Register
	doubleScratch Register
}

func NewAssembler() *Assembler {
	asm := &Assembler{
		buf:         "",
		stackOffset: -8,
		v2offset:    make(map[int]int),
		funcIndex:   0,
	}
	// Since we don't have register allocation, all virtual registers are actually
	// a stack slot, i.e. a memory location, so we need to move from memory to
	// temporary register and then to destination register. Here we use the caller
	// save register r10/x0015. You might to ask what's the rationale behind this. No
	// rationale:) I just picked one. It's not important, you can pick any register
	// you like, as long as it's not used by the code you're generating. The only
	// requirement is that it's a caller save register, because we don't know if
	// the value in the register is still needed after the function call.
	asm.scalarScratch = []Register{R10, R10D, R10W, R10B}
	asm.floatScratch = XMM15S
	asm.doubleScratch = XMM15D
	return asm
}

func (asm *Assembler) GetScratchReg(t *LIRType) Register {
	switch t {
	case LIRTypeByte, LIRTypeWord, LIRTypeDWord, LIRTypeQWord:
		for _, reg := range asm.scalarScratch {
			if reg.Type.Width == t.Width {
				return reg
			}
		}
		return BadReg
	case LIRTypeVector16S:
		return asm.floatScratch
	case LIRTypeVector16D:
		return asm.doubleScratch
	default:
		utils.Unimplement()
	}
	return NoReg
}

func (asm *Assembler) comment(comment string) {
	asm.buf += fmt.Sprintf("  # %s\n", comment)
}

// For integer instructions, suffix is used to specify the width of the operand
// b byte
// w word (2 bytes)
// l long /doubleword (4 bytes)
// q quadword (8 bytes)
//
// For SSE/AVX instructions, there involes more suffixes. For example, we have
// 128bits XMM registers, and we can use them to store 4 single precision floating
// or 2 double precision floating point numbers
//
// [ 32bits ][ 32bits ][ 32bits ][ 32bits ] xmm0
// [ 32bits ][ 32bits ][ 32bits ][ 32bits ] xmm1
// [ 64bits ][ 64bits ] xmm0
// [ 64bits ][ 64bits ] xmm1
//
// ss suffix is used for scalar single precision floating point, all related
// instructions operates on the lower 32 bits of the register, the upper 96 bits
// are ignored.
// sd suffix is used for scalar double precision floating point, all related
// instructions operates on the lower 64 bits of the register, the upper 64 bits
// are ignored.
// ps suffix is used for packed single precision floating point, all related
// instructions operates each pair of 32 bits in the register.
// pd suffix is used for packed double precision floating point, all related
// instructions operates each 64 bits in the register.
func (asm *Assembler) suffix(t *LIRType) string {
	switch t {
	case LIRTypeByte:
		return "b"
	case LIRTypeWord:
		return "w"
	case LIRTypeDWord:
		return "l"
	case LIRTypeQWord:
		return "q"
	case LIRTypeVector16S:
		return "ss"
	case LIRTypeVector16D:
		return "sd"
	default:
		utils.Unimplement()
	}
	return ""
}

// ------------------------------------------------------------------------------
// Register allocation
// Naive "register allocation" for virtual registers. We don't actully
// allocate any physical register, but instead we use stack slot as a
// virtual register. This is a very naive approach, but it's enough for
// our purpose.
//
// Stack layout is as follows:
// | param n        |
// | ...            |
// | param 1        |
// | frame pointer  | fp
// | return address | fp-8
// | slot1          | fp-16
// | slot2          |
// | ...            |

func (asm *Assembler) allocateStackSlot(v Register) string {
	if !v.Virtual {
		return fmt.Sprintf("%%%s", v.String())
	}
	if offset, ok := asm.v2offset[v.Index]; ok {
		return fmt.Sprintf("%d(%%rbp)", offset)
	}
	off := asm.stackOffset
	asm.v2offset[v.Index] = off
	asm.stackOffset -= 8
	return fmt.Sprintf("%d(%%rbp)", off)
}

func (asm *Assembler) operand(operand IOperand) string {
	switch v := operand.(type) {
	case Register:
		if !v.Virtual {
			// Physical register is perfect to become an operand
			return fmt.Sprintf("%%%s", v.String())
		}
		// Allocate stack slot for virtual register
		return asm.allocateStackSlot(v)
	case Imm:
		// Immediate is a constant and can be used as operand directly
		return fmt.Sprintf("$%d", v.Value)
	case Offset:
		// Offset is a constant and can be used as operand directly
		return fmt.Sprintf("%d", v.Value)
	case Addr:
		// Since base and index are both virtual register, we need to pre-allocate
		// stack slot before we can use them as operand
		freeRegs := CallerSaveRegs(LIRTypeQWord)
		baseReg := v.Base
		if baseReg.Virtual {
			// Load base register to a scratch register
			baseReg = asm.loadToReg(v.Base, freeRegs[0])
		}
		// No index register
		if v.Index == NoReg {
			return fmt.Sprintf("%s(%%%s)", asm.operand(v.Disp), baseReg)
		}
		// Index register
		// Type of index register must be the same as base register
		utils.Assert(v.Index.Type == LIRTypeDWord, "index should be int")
		// Load index register to a scratch register
		indexReg := asm.loadToReg(v.Index, CallerSaveRegs(LIRTypeDWord)[1])
		// @@ GCC assembler requires base and index register have same width
		// so we can not construct address like "(%rax, %ecx, 4)"
		// Since we already load index register to a 32-bit register, we need
		// to pick up a 64-bit register annd clear upper 32 bits of it, then
		// use it as index register to construct address like "(%rax, %rcx, 4)"
		// Here we use "movl %ecx, %ecx" to clear upper 32 bits of %rcx, the
		// lower 32 bits of %rcx remains unchanged.
		indexReg64bits := CallerSaveRegs(LIRTypeQWord)[1]
		asm.mov(indexReg, indexReg)
		return fmt.Sprintf("%s(%%%s, %%%s, %d)",
			asm.operand(v.Disp), baseReg, indexReg64bits, v.Scale)
	case Label:
		// Same as asm.label, add per-function label prefix
		return fmt.Sprintf(".F%d_%s", asm.funcIndex, v.Name)
	case Symbol:
		// Symbol is un-manglable
		return v.Name
	case Text:
		if v.Kind == TextString {
			// Text is immediate by referenced by label
			return fmt.Sprintf("$.T_%d", v.Id)
		} else if v.Kind == TextFloat {
			// .T_0:
			//   ..quad 0x12345
			// movsd .T_0(%rip), %xmm0
			return fmt.Sprintf(".T_%d", v.Id)
		} else {
			utils.ShouldNotReachHere()
		}
	default:
		utils.ShouldNotReachHere()
	}
	return "<unknown>"
}

// loadToScratchReg loads typed operand to scratch register
func (asm *Assembler) loadToScratchReg(src IOperand) IOperand {
	switch src.(type) {
	case Register, Addr:
		if s, ok := src.(Register); ok && !s.Virtual {
			return s
		}
		// Virtual register or memory address, candidate for load
		srcType := src.GetType()
		// Pick the scratch register
		scratch0 := asm.GetScratchReg(srcType)
		source := asm.operand(src)
		// Move src to scratch register
		asm.buf += fmt.Sprintf("  mov%s %s, %s\n",
			asm.suffix(srcType), source, asm.operand(scratch0))
		// Return scratch register
		return scratch0
	default:
		// Immediate or symbol etc, no need to load
		return src
	}
}

// loadToReg loads typed operand to physical register
func (asm *Assembler) loadToReg(src IOperand, reg Register) Register {
	utils.Assert(!reg.Virtual, "reg is not a physical register")
	asm.buf += fmt.Sprintf("  mov%s %s, %s\n",
		asm.suffix(reg.GetType()),
		asm.operand(src), asm.operand(reg))
	return reg
}

// emit0 emits an instruction without operands
func (asm *Assembler) emit0(mnemonic string) {
	asm.buf += fmt.Sprintf("  %s\n", mnemonic)
}

// emit1 emits an instruction with one operand
func (asm *Assembler) emit1(mnemonic string, dst IOperand) {
	dstType := dst.GetType()
	if dstType.IsValid() {
		// if dst is typed operand, deduce suffix from it
		asm.buf += fmt.Sprintf("  %s%s %s\n",
			mnemonic, asm.suffix(dstType), asm.operand(dst))
	} else {
		// otherwise, deduce suffix by GCC assembler automtically
		asm.buf += fmt.Sprintf("  %s %s\n", mnemonic, asm.operand(dst))
	}
}

// emit2 emits an instruction with two operands
func (asm *Assembler) emit2(mnemonic string, src IOperand, dst IOperand) {
	dstType := dst.GetType()
	// Try to load source to scratch register if possible
	srcReg := asm.loadToScratchReg(src)
	if srcReg, isReg := srcReg.(Register); isReg {
		asm.buf += fmt.Sprintf("  %s%s %s, %s\n",
			mnemonic,
			asm.suffix(dstType),
			asm.operand(srcReg),
			asm.operand(dst),
		)
		return
	}
	// Fair enough, generate asm directly
	asm.buf += fmt.Sprintf("  %s%s %s, %s\n",
		mnemonic,
		asm.suffix(dstType),
		asm.operand(src),
		asm.operand(dst),
	)
}

var FrameSize = Symbol{Name: "FRAME_SIZE"}

func (asm *Assembler) patchSymbol(sym Symbol, operand IOperand) {
	asm.buf = strings.ReplaceAll(asm.buf, sym.Name, asm.operand(operand))
}

func (asm *Assembler) emitRoData(lir *LIR) {
	// Emit read-only section
	// .rodata
	// .T_0:
	//   .string "Hello, world\n"
	// .T_1:
	//   .string "Hello, world\n"
	// ...
	if len(lir.Texts) > 0 {
		asm.buf += "  .section .rodata\n"
		for _, text := range lir.Texts {
			asm.text(text)
		}
	}
}

func (asm *Assembler) emitPrologue(name string) {
	asm.buf += "  .text\n"
	asm.buf += fmt.Sprintf("  .globl %s\n", name)
	asm.buf += fmt.Sprintf("%s:\n", name)
	asm.comment("prologue")
	asm.push(RBP)
	asm.mov(RSP, RBP)
	asm.sub(FrameSize, RSP)
}

func (asm *Assembler) emitEpilogue() {
	asm.comment("epilogue")
	asm.add(FrameSize, RSP)
	asm.pop(RBP)
	asm.ret()
}

func (asm *Assembler) push(src IOperand) {
	asm.emit1("push", src)
}

func (asm *Assembler) pop(src IOperand) {
	asm.emit1("pop", src)
}

func (asm *Assembler) mov(src IOperand, dst IOperand) {
	// Suffix of "mov" can not be deduced by destination operand, so we need to
	// take over the responsibility of deducing suffix here
	asm.emit2("mov", src, dst)
}

func (asm *Assembler) and(src IOperand, dst IOperand) {
	asm.emit2("and", src, dst)
}

func (asm *Assembler) or(src IOperand, dst IOperand) {
	asm.emit2("or", src, dst)
}

func (asm *Assembler) xor(src IOperand, dst IOperand) {
	asm.emit2("xor", src, dst)
}

// bitwise not on src and store result in dst
func (asm *Assembler) not(src IOperand) {
	asm.emit1("not", src)
}

// shift arithmetic left (SAL)
func (asm *Assembler) sal(src IOperand, dst IOperand) {
	asm.emit2("sal", src, dst)
}

// shift arithmetic right (SAR)
func (asm *Assembler) sar(src IOperand, dst IOperand) {
	asm.emit2("sar", src, dst)
}

func (asm *Assembler) cmp(res IOperand, src IOperand, dst IOperand, op LIROp) {
	asm.emit2("cmp", src, dst)
	if res != src && res != dst {
		// Result of cmp is set and might be used later, so we need to save it
		freeRegs := CallerSaveRegs(res.GetType())
		asm.setcc(AL, op)
		if freeRegs[0].GetType().Width != 1 {
			asm.movzx(AL, freeRegs[0])
		}
		asm.mov(freeRegs[0], res)
	}
}

// Move with zero-extend
func (asm *Assembler) movzx(src IOperand, dst IOperand) {
	// AT&T syntax splits the movzx Intel instruction mnemonic into different
	// mnemonics for different source sizes
	srcType := src.GetType()
	switch srcType {
	case LIRTypeByte:
		asm.emit2("movzb", src, dst)
	default:
		utils.Unimplement()
	}
}

func (asm *Assembler) setcc(dst IOperand, cc LIROp) {
	set := "set"
	switch cc {
	case LIR_CmpLT:
		set += "l"
	case LIR_CmpLE:
		set += "le"
	case LIR_CmpGT:
		set += "g"
	case LIR_CmpGE:
		set += "ge"
	case LIR_CmpEQ:
		set += "e"
	case LIR_CmpNE:
		set += "ne"
	default:
		utils.Unimplement()
	}
	asm.emit1(set, dst)
}

func (asm *Assembler) jmp(op LIROp, dst IOperand) {
	// if dst > src, zf = 0, cf = 0, sf = of
	// if dst < src, zf = 0, cf = 1, sf != of
	// if dst = src, zf = 1, cf = 0, sf = of
	//
	// jg jumps if zf = 0 and sf = of
	// jge jumps if sf = of
	// jl jumps if sf != of
	// jle jumps if zf = 1 or sf != of
	// je jumps if zf = 1
	// jne jumps if zf = 0
	jump := ""
	switch op {
	case LIR_Jmp:
		jump = "jmp"
	case LIR_Jle:
		jump = "jle"
	case LIR_Jlt:
		jump = "jl"
	case LIR_Jge:
		jump = "jge"
	case LIR_Jgt:
		jump = "jg"
	case LIR_Jeq:
		jump = "je"
	case LIR_Jne:
		jump = "jne"
	default:
		utils.Unimplement()
	}
	asm.emit1(jump, dst)
}

func (asm *Assembler) test(src IOperand, dst IOperand) {
	asm.emit2("test", src, dst)
}

func (asm *Assembler) add(src IOperand, dst IOperand) {
	asm.emit2("add", src, dst)
}

func (asm *Assembler) sub(src IOperand, dst IOperand) {
	asm.emit2("sub", src, dst)
}

func (asm *Assembler) mul(src IOperand, dst IOperand) {
	srcType := src.GetType()
	if !srcType.IsValid() {
		utils.ShouldNotReachHere()
	}
	// FIXME: Unify this
	if srcType == LIRTypeVector16D {
		// mulsd
		asm.emit2("mul", src, dst)
	} else if srcType == LIRTypeVector16S {
		utils.Unimplement()
	} else {
		// Signed full multiply of %rax by S
		// Result stored in %rdx:%rax
		asm.emit2("imul", src, dst)
	}
}

// Signed divide %rdx:%rax by S
// Quotient stored in %rax
// Remainder stored in %rdx
func (asm *Assembler) div(src IOperand) {
	source := asm.loadToScratchReg(src)
	// The Intel-syntax conversion instructions
	// cbw — sign-extend byte in %al to word in %ax,
	// cwde — sign-extend word in %ax to long in %eax,
	// cwd — sign-extend word in %ax to long in %dx:%ax,
	// cdq — sign-extend dword in %eax to quad in %edx:%eax,
	// cdqe — sign-extend dword in %eax to quad in %rax (x86-64 only),
	// cqo — sign-extend quad in %rax to octuple in %rdx:%rax (x86-64 only),
	// are called cbtw, cwtl, cwtd, cltd, cltq, and cqto in AT&T naming. as
	// accepts either naming for these instructions.
	sourceType := src.GetType()
	switch sourceType {
	case LIRTypeWord:
		asm.emit0("cwtd")
	case LIRTypeDWord:
		asm.emit0("cltd")
	case LIRTypeQWord:
		asm.emit0("cqto")
	default:
		utils.Unimplement()
	}
	asm.emit1("idiv", source)
}

func (asm *Assembler) call(res IOperand, target IOperand) {
	asm.emit1("call", target)
}

func (asm *Assembler) ret() {
	asm.emit0("ret")
}

func (asm *Assembler) label(name Label) {
	// Add per-function label prefix because all function labels are global
	// visiable and we need to distinguish them.
	asm.buf += fmt.Sprintf(".F%d_%s:\n", asm.funcIndex, name)
}

func (asm *Assembler) text(t Text) {
	asm.buf += fmt.Sprintf(".T_%d:\n", t.Id)
	switch t.Kind {
	case TextString:
		asm.buf += fmt.Sprintf("  .string \"%s\"\n", t.Value)
	case TextFloat:
		asm.buf += fmt.Sprintf("  .quad %s\n", t.Value)
	}
}

func (asm *Assembler) emit(instr *Instruction) {
	asm.comment(instr.Comment)
	switch instr.Op {
	case LIR_Mov:
		if instr.Result == NoReg {
			// For example, lir_ret may emit "mov res, noreg; ret"
			// mov a,a makes no sense
			return
		}
		asm.emit2("mov", instr.Args[0], instr.Result)
	case LIR_Jmp, LIR_Jle, LIR_Jlt,
		LIR_Jge, LIR_Jgt,
		LIR_Jeq, LIR_Jne:
		asm.jmp(instr.Op, instr.Result)
	case LIR_CmpLT, LIR_CmpLE, LIR_CmpGT, LIR_CmpGE, LIR_CmpEQ, LIR_CmpNE:
		asm.cmp(instr.Result, instr.Args[0], instr.Args[1], instr.Op)
	case LIR_Add:
		asm.add(instr.Args[0], instr.Args[1])
	case LIR_Sub:
		asm.sub(instr.Args[0], instr.Args[1])
	case LIR_Mul:
		asm.mul(instr.Args[0], instr.Args[1])
	case LIR_Div:
		asm.div(instr.Args[0])
	case LIR_And:
		asm.and(instr.Args[0], instr.Args[1])
	case LIR_Or:
		asm.or(instr.Args[0], instr.Args[1])
	case LIR_Xor:
		asm.xor(instr.Args[0], instr.Args[1])
	case LIR_Not:
		asm.not(instr.Args[0])
	case LIR_LShift:
		asm.sal(instr.Args[0], instr.Args[1])
	case LIR_RShift:
		asm.sar(instr.Args[0], instr.Args[1])
	case LIR_Ret:
		asm.comment("epilogue")
		asm.emitEpilogue()
	case LIR_Call:
		asm.call(instr.Result, instr.Args[0])
	case LIR_Test:
		asm.test(instr.Args[0], instr.Args[1])
	default:
		utils.Unimplement()
	}
}

// CodeGen translates LIR to x86_64 assembly code, it ought be nearly 1:1 mapping
// of LIR to assembly code. The only exception is register allocation, we don't
// have register allocation, so we use stack slot as virtual register. Also we
// have to use many temporary registers to construct x86 addressing mode.
func CodeGen(lirs []*LIR, debug bool) string {
	asm := NewAssembler()
	for i, lir := range lirs {
		// reset assembler before each function
		asm.stackOffset = -16
		asm.v2offset = make(map[int]int)
		asm.funcIndex = i

		// emit read-only data section
		asm.emitRoData(lir)

		// do assembly generation, order is important! entry appears first
		asm.emitPrologue(lir.Name)
		keys := make([]int, 0)
		for key := range lir.Instructions {
			keys = append(keys, key)
		}
		sort.SliceStable(keys, func(i, j int) bool {
			return keys[i] <= keys[j]
		})
		for _, idx := range keys {
			instrs := lir.Instructions[idx]
			label := lir.Labels[idx]
			asm.label(label)
			for _, instr := range instrs {
				asm.emit(instr)
			}
		}

		// Patch frame size now that we know it
		// Fixup frame size until all code was generated, we can not fix in
		// emitEpilogue because there are many ret instructions
		frameSize := utils.Abs(asm.stackOffset)
		// Align with 16 bytes
		frameSize = utils.Align16(frameSize)
		asm.patchSymbol(FrameSize, lir.NewImm(frameSize))

		// Print "register allocation" result for values
		// if debug {
		// 	asm.comment("stack layout")
		// 	for k, v := range asm.v2offset {
		// 		asm.buf += fmt.Sprintf("  # v%d => %d(%%rbp)\n", k, v)
		// 	}
		// }
	}
	return asm.buf
}
