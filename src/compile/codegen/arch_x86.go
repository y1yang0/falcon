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
	RAX = Register{Index: -1, Virtual: false, Name: "rax", Type: LIRTypeQWord}
	RBX = Register{Index: -1, Virtual: false, Name: "rbx", Type: LIRTypeQWord}
	RCX = Register{Index: -1, Virtual: false, Name: "rcx", Type: LIRTypeQWord}
	RDX = Register{Index: -1, Virtual: false, Name: "rdx", Type: LIRTypeQWord}
	RSI = Register{Index: -1, Virtual: false, Name: "rsi", Type: LIRTypeQWord}
	RDI = Register{Index: -1, Virtual: false, Name: "rdi", Type: LIRTypeQWord}
	RSP = Register{Index: -1, Virtual: false, Name: "rsp", Type: LIRTypeQWord}
	RBP = Register{Index: -1, Virtual: false, Name: "rbp", Type: LIRTypeQWord}
	R8  = Register{Index: -1, Virtual: false, Name: "r8", Type: LIRTypeQWord}
	R9  = Register{Index: -1, Virtual: false, Name: "r9", Type: LIRTypeQWord}
	R10 = Register{Index: -1, Virtual: false, Name: "r10", Type: LIRTypeQWord}
	R11 = Register{Index: -1, Virtual: false, Name: "r11", Type: LIRTypeQWord}
	R12 = Register{Index: -1, Virtual: false, Name: "r12", Type: LIRTypeQWord}
	R13 = Register{Index: -1, Virtual: false, Name: "r13", Type: LIRTypeQWord}
	R14 = Register{Index: -1, Virtual: false, Name: "r14", Type: LIRTypeQWord}
	R15 = Register{Index: -1, Virtual: false, Name: "r15", Type: LIRTypeQWord}
	RIP = Register{Index: -1, Virtual: false, Name: "rip", Type: LIRTypeQWord}

	// 32-bit registers
	EAX  = Register{Index: -1, Virtual: false, Name: "eax", Type: LIRTypeDWord}
	EBX  = Register{Index: -1, Virtual: false, Name: "ebx", Type: LIRTypeDWord}
	ECX  = Register{Index: -1, Virtual: false, Name: "ecx", Type: LIRTypeDWord}
	EDX  = Register{Index: -1, Virtual: false, Name: "edx", Type: LIRTypeDWord}
	ESI  = Register{Index: -1, Virtual: false, Name: "esi", Type: LIRTypeDWord}
	EDI  = Register{Index: -1, Virtual: false, Name: "edi", Type: LIRTypeDWord}
	ESP  = Register{Index: -1, Virtual: false, Name: "esp", Type: LIRTypeDWord}
	EBP  = Register{Index: -1, Virtual: false, Name: "ebp", Type: LIRTypeDWord}
	R8D  = Register{Index: -1, Virtual: false, Name: "r8d", Type: LIRTypeDWord}
	R9D  = Register{Index: -1, Virtual: false, Name: "r9d", Type: LIRTypeDWord}
	R10D = Register{Index: -1, Virtual: false, Name: "r10d", Type: LIRTypeDWord}
	R11D = Register{Index: -1, Virtual: false, Name: "r11d", Type: LIRTypeDWord}
	R12D = Register{Index: -1, Virtual: false, Name: "r12d", Type: LIRTypeDWord}
	R13D = Register{Index: -1, Virtual: false, Name: "r13d", Type: LIRTypeDWord}
	R14D = Register{Index: -1, Virtual: false, Name: "r14d", Type: LIRTypeDWord}
	R15D = Register{Index: -1, Virtual: false, Name: "r15d", Type: LIRTypeDWord}

	// 16-bit registers
	AX   = Register{Index: -1, Virtual: false, Name: "ax", Type: LIRTypeWord}
	BX   = Register{Index: -1, Virtual: false, Name: "bx", Type: LIRTypeWord}
	CX   = Register{Index: -1, Virtual: false, Name: "cx", Type: LIRTypeWord}
	DX   = Register{Index: -1, Virtual: false, Name: "dx", Type: LIRTypeWord}
	SI   = Register{Index: -1, Virtual: false, Name: "si", Type: LIRTypeWord}
	DI   = Register{Index: -1, Virtual: false, Name: "di", Type: LIRTypeWord}
	SP   = Register{Index: -1, Virtual: false, Name: "sp", Type: LIRTypeWord}
	BP   = Register{Index: -1, Virtual: false, Name: "bp", Type: LIRTypeWord}
	R8W  = Register{Index: -1, Virtual: false, Name: "r8w", Type: LIRTypeWord}
	R9W  = Register{Index: -1, Virtual: false, Name: "r9w", Type: LIRTypeWord}
	R10W = Register{Index: -1, Virtual: false, Name: "r10w", Type: LIRTypeWord}
	R11W = Register{Index: -1, Virtual: false, Name: "r11w", Type: LIRTypeWord}
	R12W = Register{Index: -1, Virtual: false, Name: "r12w", Type: LIRTypeWord}
	R13W = Register{Index: -1, Virtual: false, Name: "r13w", Type: LIRTypeWord}
	R14W = Register{Index: -1, Virtual: false, Name: "r14w", Type: LIRTypeWord}
	R15W = Register{Index: -1, Virtual: false, Name: "r15w", Type: LIRTypeWord}

	// 8-bit registers
	AH   = Register{Index: -1, Virtual: false, Name: "ah", Type: LIRTypeByte}
	AL   = Register{Index: -1, Virtual: false, Name: "al", Type: LIRTypeByte}
	BH   = Register{Index: -1, Virtual: false, Name: "bh", Type: LIRTypeByte}
	BL   = Register{Index: -1, Virtual: false, Name: "bl", Type: LIRTypeByte}
	CH   = Register{Index: -1, Virtual: false, Name: "ch", Type: LIRTypeByte}
	CL   = Register{Index: -1, Virtual: false, Name: "cl", Type: LIRTypeByte}
	DH   = Register{Index: -1, Virtual: false, Name: "dh", Type: LIRTypeByte}
	DL   = Register{Index: -1, Virtual: false, Name: "dl", Type: LIRTypeByte}
	SIL  = Register{Index: -1, Virtual: false, Name: "sil", Type: LIRTypeByte}
	DIL  = Register{Index: -1, Virtual: false, Name: "dil", Type: LIRTypeByte}
	BPL  = Register{Index: -1, Virtual: false, Name: "bpl", Type: LIRTypeByte}
	SPL  = Register{Index: -1, Virtual: false, Name: "spl", Type: LIRTypeByte}
	R8B  = Register{Index: -1, Virtual: false, Name: "r8b", Type: LIRTypeByte}
	R9B  = Register{Index: -1, Virtual: false, Name: "r9b", Type: LIRTypeByte}
	R10B = Register{Index: -1, Virtual: false, Name: "r10b", Type: LIRTypeByte}
	R11B = Register{Index: -1, Virtual: false, Name: "r11b", Type: LIRTypeByte}
	R12B = Register{Index: -1, Virtual: false, Name: "r12b", Type: LIRTypeByte}
	R13B = Register{Index: -1, Virtual: false, Name: "r13b", Type: LIRTypeByte}
	R14B = Register{Index: -1, Virtual: false, Name: "r14b", Type: LIRTypeByte}
	R15B = Register{Index: -1, Virtual: false, Name: "r15b", Type: LIRTypeByte}

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
