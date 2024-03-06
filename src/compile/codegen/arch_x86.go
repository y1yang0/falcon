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
	"runtime"
)

// Reference
// https://web.stanford.edu/class/cs107/resources/x86-64-reference.pdf
// https://www.cs.cmu.edu/afs/cs/academic/class/15213-s20/www/recitations/x86-cheat-sheet.pdf

type ArchABI interface {
	ArgReg(idx int) Register
	CallerSaveRegs() []Register
	CalleeSaveRegs() []Register
}

var (
	BadReg = Register{Index: -1, Virtual: false, Name: "badreg", Type: LIRTypeVoid}
	NoReg  = Register{Index: -1, Virtual: false, Name: "noreg", Type: LIRTypeVoid}
	// 64-bit registers
	RAX = Register{Index: -1, Virtual: false, Name: "rax", Type: LIRTypeQWord, Affinity: 0}
	RBX = Register{Index: -1, Virtual: false, Name: "rbx", Type: LIRTypeQWord, Affinity: 1}
	RCX = Register{Index: -1, Virtual: false, Name: "rcx", Type: LIRTypeQWord, Affinity: 2}
	RDX = Register{Index: -1, Virtual: false, Name: "rdx", Type: LIRTypeQWord, Affinity: 3}
	RSI = Register{Index: -1, Virtual: false, Name: "rsi", Type: LIRTypeQWord, Affinity: 4}
	RDI = Register{Index: -1, Virtual: false, Name: "rdi", Type: LIRTypeQWord, Affinity: 5}
	RSP = Register{Index: -1, Virtual: false, Name: "rsp", Type: LIRTypeQWord, Affinity: 6}
	RBP = Register{Index: -1, Virtual: false, Name: "rbp", Type: LIRTypeQWord, Affinity: 7}
	R8  = Register{Index: -1, Virtual: false, Name: "r8", Type: LIRTypeQWord, Affinity: 8}
	R9  = Register{Index: -1, Virtual: false, Name: "r9", Type: LIRTypeQWord, Affinity: 9}
	R10 = Register{Index: -1, Virtual: false, Name: "r10", Type: LIRTypeQWord, Affinity: 10}
	R11 = Register{Index: -1, Virtual: false, Name: "r11", Type: LIRTypeQWord, Affinity: 11}
	R12 = Register{Index: -1, Virtual: false, Name: "r12", Type: LIRTypeQWord, Affinity: 12}
	R13 = Register{Index: -1, Virtual: false, Name: "r13", Type: LIRTypeQWord, Affinity: 13}
	R14 = Register{Index: -1, Virtual: false, Name: "r14", Type: LIRTypeQWord, Affinity: 14}
	R15 = Register{Index: -1, Virtual: false, Name: "r15", Type: LIRTypeQWord, Affinity: 15}
	RIP = Register{Index: -1, Virtual: false, Name: "rip", Type: LIRTypeQWord, Affinity: 16}

	// 32-bit registers
	EAX  = Register{Index: -1, Virtual: false, Name: "eax", Type: LIRTypeDWord, Affinity: 0}
	EBX  = Register{Index: -1, Virtual: false, Name: "ebx", Type: LIRTypeDWord, Affinity: 1}
	ECX  = Register{Index: -1, Virtual: false, Name: "ecx", Type: LIRTypeDWord, Affinity: 2}
	EDX  = Register{Index: -1, Virtual: false, Name: "edx", Type: LIRTypeDWord, Affinity: 3}
	ESI  = Register{Index: -1, Virtual: false, Name: "esi", Type: LIRTypeDWord, Affinity: 4}
	EDI  = Register{Index: -1, Virtual: false, Name: "edi", Type: LIRTypeDWord, Affinity: 5}
	ESP  = Register{Index: -1, Virtual: false, Name: "esp", Type: LIRTypeDWord, Affinity: 6}
	EBP  = Register{Index: -1, Virtual: false, Name: "ebp", Type: LIRTypeDWord, Affinity: 7}
	R8D  = Register{Index: -1, Virtual: false, Name: "r8d", Type: LIRTypeDWord, Affinity: 8}
	R9D  = Register{Index: -1, Virtual: false, Name: "r9d", Type: LIRTypeDWord, Affinity: 9}
	R10D = Register{Index: -1, Virtual: false, Name: "r10d", Type: LIRTypeDWord, Affinity: 10}
	R11D = Register{Index: -1, Virtual: false, Name: "r11d", Type: LIRTypeDWord, Affinity: 11}
	R12D = Register{Index: -1, Virtual: false, Name: "r12d", Type: LIRTypeDWord, Affinity: 12}
	R13D = Register{Index: -1, Virtual: false, Name: "r13d", Type: LIRTypeDWord, Affinity: 13}
	R14D = Register{Index: -1, Virtual: false, Name: "r14d", Type: LIRTypeDWord, Affinity: 14}
	R15D = Register{Index: -1, Virtual: false, Name: "r15d", Type: LIRTypeDWord, Affinity: 15}

	// 16-bit registers
	AX   = Register{Index: -1, Virtual: false, Name: "ax", Type: LIRTypeWord, Affinity: 0}
	BX   = Register{Index: -1, Virtual: false, Name: "bx", Type: LIRTypeWord, Affinity: 1}
	CX   = Register{Index: -1, Virtual: false, Name: "cx", Type: LIRTypeWord, Affinity: 2}
	DX   = Register{Index: -1, Virtual: false, Name: "dx", Type: LIRTypeWord, Affinity: 3}
	SI   = Register{Index: -1, Virtual: false, Name: "si", Type: LIRTypeWord, Affinity: 4}
	DI   = Register{Index: -1, Virtual: false, Name: "di", Type: LIRTypeWord, Affinity: 5}
	SP   = Register{Index: -1, Virtual: false, Name: "sp", Type: LIRTypeWord, Affinity: 6}
	BP   = Register{Index: -1, Virtual: false, Name: "bp", Type: LIRTypeWord, Affinity: 7}
	R8W  = Register{Index: -1, Virtual: false, Name: "r8w", Type: LIRTypeWord, Affinity: 8}
	R9W  = Register{Index: -1, Virtual: false, Name: "r9w", Type: LIRTypeWord, Affinity: 9}
	R10W = Register{Index: -1, Virtual: false, Name: "r10w", Type: LIRTypeWord, Affinity: 10}
	R11W = Register{Index: -1, Virtual: false, Name: "r11w", Type: LIRTypeWord, Affinity: 11}
	R12W = Register{Index: -1, Virtual: false, Name: "r12w", Type: LIRTypeWord, Affinity: 12}
	R13W = Register{Index: -1, Virtual: false, Name: "r13w", Type: LIRTypeWord, Affinity: 13}
	R14W = Register{Index: -1, Virtual: false, Name: "r14w", Type: LIRTypeWord, Affinity: 14}
	R15W = Register{Index: -1, Virtual: false, Name: "r15w", Type: LIRTypeWord, Affinity: 15}

	// 8-bit registers
	AH   = Register{Index: -1, Virtual: false, Name: "ah", Type: LIRTypeByte, Affinity: 0, IsHigh: true}
	AL   = Register{Index: -1, Virtual: false, Name: "al", Type: LIRTypeByte, Affinity: 0}
	BH   = Register{Index: -1, Virtual: false, Name: "bh", Type: LIRTypeByte, Affinity: 1, IsHigh: true}
	BL   = Register{Index: -1, Virtual: false, Name: "bl", Type: LIRTypeByte, Affinity: 1}
	CH   = Register{Index: -1, Virtual: false, Name: "ch", Type: LIRTypeByte, Affinity: 2, IsHigh: true}
	CL   = Register{Index: -1, Virtual: false, Name: "cl", Type: LIRTypeByte, Affinity: 2}
	DH   = Register{Index: -1, Virtual: false, Name: "dh", Type: LIRTypeByte, Affinity: 3, IsHigh: true}
	DL   = Register{Index: -1, Virtual: false, Name: "dl", Type: LIRTypeByte, Affinity: 3}
	SIL  = Register{Index: -1, Virtual: false, Name: "sil", Type: LIRTypeByte, Affinity: 4}
	DIL  = Register{Index: -1, Virtual: false, Name: "dil", Type: LIRTypeByte, Affinity: 5}
	BPL  = Register{Index: -1, Virtual: false, Name: "bpl", Type: LIRTypeByte, Affinity: 6}
	SPL  = Register{Index: -1, Virtual: false, Name: "spl", Type: LIRTypeByte, Affinity: 7}
	R8B  = Register{Index: -1, Virtual: false, Name: "r8b", Type: LIRTypeByte, Affinity: 8}
	R9B  = Register{Index: -1, Virtual: false, Name: "r9b", Type: LIRTypeByte, Affinity: 9}
	R10B = Register{Index: -1, Virtual: false, Name: "r10b", Type: LIRTypeByte, Affinity: 10}
	R11B = Register{Index: -1, Virtual: false, Name: "r11b", Type: LIRTypeByte, Affinity: 11}
	R12B = Register{Index: -1, Virtual: false, Name: "r12b", Type: LIRTypeByte, Affinity: 12}
	R13B = Register{Index: -1, Virtual: false, Name: "r13b", Type: LIRTypeByte, Affinity: 13}
	R14B = Register{Index: -1, Virtual: false, Name: "r14b", Type: LIRTypeByte, Affinity: 14}
	R15B = Register{Index: -1, Virtual: false, Name: "r15b", Type: LIRTypeByte, Affinity: 15}

	// 128-bit registers
	// single precision floating point
	XMM0S  = Register{Index: -1, Virtual: false, Name: "xmm0", Type: LIRTypeVector16S}
	XMM1S  = Register{Index: -1, Virtual: false, Name: "xmm1", Type: LIRTypeVector16S}
	XMM2S  = Register{Index: -1, Virtual: false, Name: "xmm2", Type: LIRTypeVector16S}
	XMM3S  = Register{Index: -1, Virtual: false, Name: "xmm3", Type: LIRTypeVector16S}
	XMM4S  = Register{Index: -1, Virtual: false, Name: "xmm4", Type: LIRTypeVector16S}
	XMM5S  = Register{Index: -1, Virtual: false, Name: "xmm5", Type: LIRTypeVector16S}
	XMM6S  = Register{Index: -1, Virtual: false, Name: "xmm6", Type: LIRTypeVector16S}
	XMM7S  = Register{Index: -1, Virtual: false, Name: "xmm7", Type: LIRTypeVector16S}
	XMM8S  = Register{Index: -1, Virtual: false, Name: "xmm8", Type: LIRTypeVector16S}
	XMM9S  = Register{Index: -1, Virtual: false, Name: "xmm9", Type: LIRTypeVector16S}
	XMM10S = Register{Index: -1, Virtual: false, Name: "xmm10", Type: LIRTypeVector16S}
	XMM11S = Register{Index: -1, Virtual: false, Name: "xmm11", Type: LIRTypeVector16S}
	XMM12S = Register{Index: -1, Virtual: false, Name: "xmm12", Type: LIRTypeVector16S}
	XMM13S = Register{Index: -1, Virtual: false, Name: "xmm13", Type: LIRTypeVector16S}
	XMM14S = Register{Index: -1, Virtual: false, Name: "xmm14", Type: LIRTypeVector16S}
	XMM15S = Register{Index: -1, Virtual: false, Name: "xmm15", Type: LIRTypeVector16S}
	// double precision floating point
	// TODO: Consolidate them
	XMM0D  = Register{Index: -1, Virtual: false, Name: "xmm0", Type: LIRTypeVector16D}
	XMM1D  = Register{Index: -1, Virtual: false, Name: "xmm1", Type: LIRTypeVector16D}
	XMM2D  = Register{Index: -1, Virtual: false, Name: "xmm2", Type: LIRTypeVector16D}
	XMM3D  = Register{Index: -1, Virtual: false, Name: "xmm3", Type: LIRTypeVector16D}
	XMM4D  = Register{Index: -1, Virtual: false, Name: "xmm4", Type: LIRTypeVector16D}
	XMM5D  = Register{Index: -1, Virtual: false, Name: "xmm5", Type: LIRTypeVector16D}
	XMM6D  = Register{Index: -1, Virtual: false, Name: "xmm6", Type: LIRTypeVector16D}
	XMM7D  = Register{Index: -1, Virtual: false, Name: "xmm7", Type: LIRTypeVector16D}
	XMM8D  = Register{Index: -1, Virtual: false, Name: "xmm8", Type: LIRTypeVector16D}
	XMM9D  = Register{Index: -1, Virtual: false, Name: "xmm9", Type: LIRTypeVector16D}
	XMM10D = Register{Index: -1, Virtual: false, Name: "xmm10", Type: LIRTypeVector16D}
	XMM11D = Register{Index: -1, Virtual: false, Name: "xmm11", Type: LIRTypeVector16D}
	XMM12D = Register{Index: -1, Virtual: false, Name: "xmm12", Type: LIRTypeVector16D}
	XMM13D = Register{Index: -1, Virtual: false, Name: "xmm13", Type: LIRTypeVector16D}
	XMM14D = Register{Index: -1, Virtual: false, Name: "xmm14", Type: LIRTypeVector16D}
	XMM15D = Register{Index: -1, Virtual: false, Name: "xmm15", Type: LIRTypeVector16D}

	// 256-bit registers
)

var AllRegisters = []Register{
	RAX, RBX, RCX, RDX, RSI, RDI, RSP, RBP, R8, R9, R10, R11, R12, R13, R14, R15, RIP,
	EAX, EBX, ECX, EDX, ESI, EDI, ESP, EBP, R8D, R9D, R10D, R11D, R12D, R13D, R14D, R15D,
	AX, BX, CX, DX, SI, DI, SP, BP, R8W, R9W, R10W, R11W, R12W, R13W, R14W, R15W,
	AH, AL, BH, BL, CH, CL, DH, DL, SIL, DIL, BPL, SPL, R8B, R9B, R10B, R11B, R12B, R13B, R14B, R15B,
	XMM0S, XMM1S, XMM2S, XMM3S, XMM4S, XMM5S, XMM6S, XMM7S, XMM8S, XMM9S, XMM10S, XMM11S, XMM12S, XMM13S, XMM14S, XMM15S,
	XMM0D, XMM1D, XMM2D, XMM3D, XMM4D, XMM5D, XMM6D, XMM7D, XMM8D, XMM9D, XMM10D, XMM11D, XMM12D, XMM13D, XMM14D, XMM15D,
}

// Cast a register to a specific type, i.e. RAX -> EAX
func (r Register) Cast(t *LIRType) Register {
	for _, reg := range AllRegisters {
		if reg.Affinity == r.Affinity && reg.Type == t &&
			!reg.IsHigh /*rax -> al*/ {
			return reg
		}
	}
	return NoReg
}

func ReturnReg(t *LIRType) Register {
	switch t {
	case LIRTypeQWord:
		return RAX
	case LIRTypeDWord:
		return EAX
	case LIRTypeWord:
		return AX
	case LIRTypeByte:
		return AL
	case LIRTypeVoid:
		// no return value
		return NoReg
	default:
		utils.ShouldNotReachHere()
	}
	return BadReg

}

func CallerSaveRegs(t *LIRType) []Register {
	switch t {
	case LIRTypeQWord:
		return []Register{RAX, RCX, RDX, RSI, RDI, R8, R9, R10, R11}
	case LIRTypeDWord:
		return []Register{EAX, ECX, EDX, ESI, EDI, R8D, R9D, R10D, R11D}
	case LIRTypeWord:
		return []Register{AX, CX, DX, SI, DI, R8W, R9W, R10W, R11W}
	case LIRTypeByte:
		return []Register{AL, CL, DL, SIL, DIL, R8B, R9B, R10B, R11B}
	case LIRTypeVector16S:
		// all %xmm are volatile, i.e. caller-save
		return []Register{XMM0S, XMM1S, XMM2S, XMM3S, XMM4S, XMM5S, XMM6S, XMM7S, XMM8S, XMM9S, XMM10S, XMM11S, XMM12S, XMM13S, XMM14S, XMM15S}
	case LIRTypeVector16D:
		// ditto
		return []Register{XMM0D, XMM1D, XMM2D, XMM3D, XMM4D, XMM5D, XMM6D, XMM7D, XMM8D, XMM9D, XMM10D, XMM11D, XMM12D, XMM13D, XMM14D, XMM15D}
	default:
		utils.ShouldNotReachHere()
	}
	return nil
}

func CalleeSaveRegs(t *LIRType) []Register {
	switch t {
	case LIRTypeQWord:
		return []Register{RBX, RBP, R12, R13, R14, R15}
	case LIRTypeDWord:
		return []Register{EBX, EBP, R12D, R13D, R14D, R15D}
	case LIRTypeWord:
		return []Register{BX, BP, R12W, R13W, R14W, R15W}
	case LIRTypeByte:
		return []Register{BL, BPL, R12B, R13B, R14B, R15B}
	default:
		utils.ShouldNotReachHere()
	}
	return nil
}

// Calling Convention
func ArgReg(idx int, t *LIRType) Register {
	var argReg64, argReg32, argReg16, argReg8 []Register
	if runtime.GOOS == "windows" {
		// Windows-specific fastcall calling convention
		if idx > 4 {
			utils.Unimplement()
		}
		argReg64 = []Register{RCX, RDX, R8, R9}
		argReg32 = []Register{ECX, EDX, R8D, R9D}
		argReg16 = []Register{CX, DX, R8W, R9W}
		argReg8 = []Register{CL, DL, R8B, R9B}
	} else {
		//System V AMD64 ABI
		if idx > 5 {
			utils.Unimplement()
		}
		argReg64 = []Register{RDI, RSI, RDX, RCX, R8, R9}
		argReg32 = []Register{EDI, ESI, EDX, ECX, R8D, R9D}
		argReg16 = []Register{DI, SI, DX, CX, R8W, R9W}
		argReg8 = []Register{DIL, SIL, DL, CL, R8B, R9B}
	}

	switch t {
	case LIRTypeQWord:
		return argReg64[idx]
	case LIRTypeDWord:
		return argReg32[idx]
	case LIRTypeWord:
		return argReg16[idx]
	case LIRTypeByte:
		return argReg8[idx]
	case LIRTypeVector16S:
		return []Register{XMM0S, XMM1S, XMM2S, XMM3S, XMM4S, XMM5S, XMM6S, XMM7S}[idx]
	case LIRTypeVector16D:
		return []Register{XMM0D, XMM1D, XMM2D, XMM3D, XMM4D, XMM5D, XMM6D, XMM7D}[idx]
	default:
		utils.ShouldNotReachHere()
	}
	return BadReg
}
