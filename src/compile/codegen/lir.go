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
}

type LIR struct {
	vid          int                    // global virtual register id
	v2r          map[int]Register       // Value id to virtual register
	roid         int                    // global read-only section id
	Name         string                 // function Name
	Instructions map[int][]*Instruction // blocks of instructions, order of blocks is important
	Labels       map[int]Label          // labels for each block for continuation point
	Texts        []Text                 // read-only section literals(string/quads/longs)
}

type LIRTypeKind int

type LIRType struct {
	Width int // in bytes
}

var LIRTypeVoid = &LIRType{0}      // 0 byte, void
var LIRTypeByte = &LIRType{1}      // 1 byte, char, al/ah
var LIRTypeWord = &LIRType{2}      // 2 bytes, short, ax
var LIRTypeDWord = &LIRType{4}     // 4 bytes, int, eax
var LIRTypeQWord = &LIRType{8}     // 8 bytes, long, rax
var LIRTypeVector16 = &LIRType{16} // 16 bytes
var LIRTypeVector32 = &LIRType{32} // 32 bytes
var LIRTypeVector64 = &LIRType{64} // 64 bytes

type IOperand interface {
	String() string
}

type ITypedOperand interface {
	IOperand
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
	Type    *LIRType
	Index   int
	Name    string // mnemonic name
	Virtual bool   // virtual register, in fact almost all registers are virtual in this pass
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

func (x Register) GetType() *LIRType {
	return x.Type
}

func (x Addr) GetType() *LIRType {
	return x.Type
}

func (x Imm) GetType() *LIRType {
	return x.Type
}

// GetLIRType returns the LIRType for the given AST type
func GetLIRType(astType *ast.Type) *LIRType {
	if astType.Kind == ast.TypeArray {
		return LIRTypeQWord
	} else if astType.Kind == ast.TypeString {
		return LIRTypeQWord
	}
	switch astType {
	case ast.TLong:
		return LIRTypeQWord
	case ast.TInt:
		return LIRTypeDWord
	case ast.TShort:
		return LIRTypeWord
	case ast.TChar, ast.TBool, ast.TByte:
		return LIRTypeByte
	case ast.TVoid:
		return LIRTypeVoid
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

func (lir *LIR) SetResult(val *ssa.Value, result IOperand) {
	// if r, ok := result.(Register); ok && r.Virtual {
	lir.v2r[val.Id] = result.(Register)
	// }
}

func (lir *LIR) NewInstr(idx int, op LIROp, args ...IOperand) *Instruction {
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

func (lir *LIR) NewJmp(idx int, op LIROp, block *ssa.Block) *Instruction {
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
	switch v.(type) {
	case *ssa.Block:
		x.Comment = fmt.Sprintf("b%d", v.(*ssa.Block).Id)
	case *ssa.Value:
		x.Comment = fmt.Sprintf("%v", v)
	case string:
		x.Comment = v.(string)
	default:
		utils.Unimplement()
	}
}

func (x *LIR) verify() {
	// verify that all instructions have a result
	for _, instrs := range x.Instructions {
		for _, instr := range instrs {
			utils.Assert(instr.Result != nil, "miss result")
			utils.Assert(len(instr.Args) >= 0 && len(instr.Args) <= 2, "miss args")
		}
	}
}

func NewLIR(fn *ssa.Func) *LIR {
	return &LIR{
		vid:          0,
		roid:         0,
		v2r:          make(map[int]Register),
		Name:         fn.Name,
		Instructions: make(map[int][]*Instruction, len(fn.Blocks)), //order is important
		Labels:       make(map[int]Label),
		Texts:        make([]Text, 0),
	}
}
