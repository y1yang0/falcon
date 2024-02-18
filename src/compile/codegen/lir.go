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
	"falcon/ast"
	"falcon/compile/ssa"
	"falcon/utils"
	"fmt"
)

// ------------------------------------------------------------------------------
// Low-level Intermediate Representation (LIR)
//
// See Linear Scan Register Allocation for the Java HotSpot™ Client Compiler for
// more details about LIR design.
// LIR is a three-operand form for operators, withthe first operand being the
// result of the operation. The second and third operands are the arguments to
// the operation. x86-64 employs a two-operand form for most instructions, the
// right operand is equal to the result. For example, when lowering the add Value
// "add v1, v2", the v2 is the result, so we need to generate a new virtual register
// v3 for the result and move left operand v1 to the result register, then add
// the right operand v2 to the result register, i.e.
// mov v3, v1, v3
// add v3, v2, v3
// It's a bit of a misnomer on x86-64, but it's a good representation for other
// architectures such as aarch64.
type LIROp int

const (
	LIR_Add LIROp = iota
	LIR_Sub
	LIR_Mul
	LIR_Div
	LIR_And
	LIR_Or
	LIR_Xor
	LIR_Not
	LIR_LShift
	LIR_RShift
	LIR_CmpLE
	LIR_CmpLT
	LIR_CmpGE
	LIR_CmpGT
	LIR_CmpEQ
	LIR_CmpNE
	LIR_Mov
	LIR_Ret
	LIR_Jmp
	LIR_Jle
	LIR_Jlt
	LIR_Jeq
	LIR_Jne
	LIR_Jz
	LIR_Jnz
	LIR_Jge
	LIR_Jgt
	LIR_Test
	LIR_Call
)

type Instruction struct {
	Op      LIROp
	Result  IOperand
	Args    []IOperand // two-operand form
	Comment string

	Id int
}

type LIRTypeKind int

type LIRType struct {
	Width           int // in bytes
	SinglePrecision bool
}

var LIRTypeBottom = &LIRType{-1, false}    // not even a type
var LIRTypeVoid = &LIRType{0, false}       // 0 byte, void
var LIRTypeByte = &LIRType{1, false}       // 1 byte, char, al/ah
var LIRTypeWord = &LIRType{2, false}       // 2 bytes, short, ax
var LIRTypeDWord = &LIRType{4, false}      // 4 bytes, int, eax
var LIRTypeQWord = &LIRType{8, false}      // 8 bytes, long, rax
var LIRTypeVector16S = &LIRType{16, false} // 16 bytes, single-precision float
var LIRTypeVector16D = &LIRType{16, true}  // 16 bytes, double-precision float
var LIRTypeVector32 = &LIRType{32, false}  // 32 bytes
var LIRTypeVector64 = &LIRType{64, false}  // 64 bytes

func (x *LIRType) IsValid() bool {
	return x != LIRTypeBottom
}

type IOperand interface {
	String() string
	GetType() *LIRType
}

// mangleable label name, e.g. L0, L1, L2
type Label struct {
	Name string
}

// un-mangleable symbol name, e.g. function name
type Symbol struct {
	Name string
}

// register, either physical or virtual, e.g. %rax, %rbp, v0, v1
type Register struct {
	Type     *LIRType
	Index    int
	Name     string // mnemonic name
	Virtual  bool   // virtual register, in fact almost all registers are virtual in this pass
	Affinity int
	IsHigh   bool
}

type TextKind int

const (
	TextString TextKind = iota
	TextFloat
)

// read-only section literal
type Text struct {
	Id    int
	Kind  TextKind
	Value string
}

// immediate value, e.g. mov $123, %rax => $123
type Imm struct {
	Type  *LIRType
	Value interface{}
}

// operand offset, e.g. 8(%rbp) => 8
type Offset struct {
	Value int
}

// memory address, e.g. 8(%rbp) or .quad_0(%rbp, %rax, 8)
type Addr struct {
	Type  *LIRType
	Base  Register
	Index Register
	Scale int
	Disp  IOperand // int or Symbol, e.g. 8(%rbp) or .quad_0(%rbp, %rax, 8)
}

func (x Register) GetType() *LIRType { return x.Type }

func (x Addr) GetType() *LIRType { return x.Type }

func (x Imm) GetType() *LIRType { return x.Type }

func (x Offset) GetType() *LIRType { return LIRTypeBottom }

func (x Label) GetType() *LIRType { return LIRTypeBottom }

func (x Symbol) GetType() *LIRType { return LIRTypeBottom }

func (x Text) GetType() *LIRType { return LIRTypeBottom }

// GetLIRType returns the LIRType for the given AST type
func GetLIRType(astType *ast.Type) *LIRType {
	switch {
	case astType.IsLong():
		return LIRTypeQWord
	case astType.IsInt():
		return LIRTypeDWord
	case astType.IsShort():
		return LIRTypeWord
	case astType.IsChar(), astType.IsBool(), astType.IsByte():
		return LIRTypeByte
	case astType.IsVoid():
		return LIRTypeVoid
	case astType.IsString():
		return LIRTypeQWord
	case astType.IsArray():
		return LIRTypeQWord
	case astType.IsFloat():
		return LIRTypeVector16S
	case astType.IsDouble():
		return LIRTypeVector16D
	default:
		utils.Unimplement()
	}
	return nil
}

func (x LIROp) String() string {
	switch x {
	case LIR_Add:
		return "add"
	case LIR_Sub:
		return "sub"
	case LIR_Mul:
		return "mul"
	case LIR_Div:
		return "div"
	case LIR_And:
		return "and"
	case LIR_Or:
		return "or"
	case LIR_Xor:
		return "xor"
	case LIR_LShift:
		return "lshift"
	case LIR_RShift:
		return "rshift"
	case LIR_CmpLE:
		return "cmple"
	case LIR_CmpLT:
		return "cmplt"
	case LIR_CmpGE:
		return "cmpge"
	case LIR_CmpGT:
		return "cmpgt"
	case LIR_CmpEQ:
		return "cmpeq"
	case LIR_CmpNE:
		return "cmpne"
	case LIR_Mov:
		return "mov"
	case LIR_Ret:
		return "ret"
	case LIR_Jmp:
		return "jmp"
	case LIR_Jle:
		return "jle"
	case LIR_Jlt:
		return "jl"
	case LIR_Jeq:
		return "je"
	case LIR_Jne:
		return "jne"
	case LIR_Jz:
		return "jz"
	case LIR_Jnz:
		return "jnz"
	case LIR_Jge:
		return "jge"
	case LIR_Jgt:
		return "jg"
	case LIR_Test:
		return "test"
	case LIR_Call:
		return "call"
	default:
		utils.Unimplement()
	}
	return ""
}

func (x Register) String() string {
	if x.Virtual {
		return fmt.Sprintf("v%d", x.Index)
	}
	return x.Name
}

func (x Imm) String() string {
	return fmt.Sprintf("$%d", x.Value)
}

func (x Offset) String() string {
	return fmt.Sprintf("%d", x.Value)
}

func (x Addr) String() string {
	return fmt.Sprintf("%s[%s]+%v", x.Base, x.Index, x.Disp)
}

func (x Label) String() string {
	return x.Name
}

func (x Symbol) String() string {
	return x.Name
}

func (x Text) String() string {
	return x.Value
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
