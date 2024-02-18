// Copyright (c) 2024 The Sprite Programming Language
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

type PhyReg struct {
	index int
}

var RAX_ = defPhyReg(0)
var RBX_ = defPhyReg(1)
var RCX_ = defPhyReg(2)
var RDX_ = defPhyReg(3)
var RSI_ = defPhyReg(4)
var RDI_ = defPhyReg(5)
var R8_ = defPhyReg(6)
var R9_ = defPhyReg(7)
var R10_ = defPhyReg(8)
var R11_ = defPhyReg(9)
var R12_ = defPhyReg(10)
var R13_ = defPhyReg(11)
var R14_ = defPhyReg(12)
var R15_ = defPhyReg(13)

var RBP_ = defPhyReg(14)
var RSP_ = defPhyReg(15)

func defPhyReg(index int) *PhyReg {
	return &PhyReg{
		index: index,
	}
}

func phyRegStart() int {
	return RAX_.index
}

func phyRegEnd() int {
	return R15_.index
}

func nofAvailPhyReg() int {
	return phyRegEnd() - phyRegStart() + 1
}

func callerSaved() []int {
	return []int{RAX_.index, RCX_.index, RDX_.index, R8_.index, R9_.index, R10_.index, R11_.index}
}
