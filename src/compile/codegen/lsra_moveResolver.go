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

import (
	"falcon/compile/ssa"
	"falcon/utils"
)

type MoveResolver struct {
	ra   *LSRA
	from *ssa.Block
	to   *ssa.Block

	pairs map[*Interval]*Interval

	stackSlot int

	cycleStart    *Interval
	foundNewCycle bool
}

func (mr *MoveResolver) record(from, to *Interval) {
	mr.pairs[from] = to
}

func (mr *MoveResolver) resolve() {
	utils.Assert(len(mr.from.Succs) > 1 && len(mr.to.Preds) > 1, "sanity check")

	// decide insertion position
	var b *ssa.Block
	var is []*Instruction
	var pos int
	if len(mr.from.Succs) == 1 {
		b = mr.from
		is = mr.ra.lir.Instructions[b.Id]
		l := len(is)
		if isBranchOp(is[l-1].Op) {
			pos = l - 1
		} else {
			pos = l
		}
	} else {
		b = mr.to
		is = mr.ra.lir.Instructions[b.Id]
		pos = 0
	}

	prIdToFrom := make(map[int]*Interval)
	for fi, _ := range mr.pairs {
		if fi.phyRegAssigned() {
			prIdToFrom[fi.phyRegIndex] = fi
		}
	}

	// process pairs
	processed := utils.NewSet[*Interval]()
	buffer := make([]*Instruction, 0)
	for fi, ti := range mr.pairs {
		buffer = mr.move(fi, ti, prIdToFrom, processed, buffer)
	}

	// insert new instructions
	nis := make([]*Instruction, 0)
	for i := 0; i < pos; i++ {
		nis = append(nis, is[i])
	}
	nis = append(nis, buffer...)
	for i := pos; i < len(is); i++ {
		nis = append(nis, is[i])
	}
	mr.ra.lir.Instructions[b.Id] = nis
}

func (mr *MoveResolver) move(fi *Interval, ti *Interval, prIdToFrom map[int]*Interval, processed *utils.Set[*Interval], buffer []*Instruction) []*Instruction {
	if processed.Contains(fi) {
		return buffer
	}

	if ti.stackSlotIndex != -1 {
		buffer = append(buffer, &Instruction{
			Op:     LIR_Mov,
			Result: &Addr{},
			Args:   []IOperand{&Register{}},
		})
	} else {
		ofi, ok := prIdToFrom[ti.phyRegIndex]
		if ok && !processed.Contains(ofi) {
			oti := mr.pairs[ofi]

			if mr.cycleStart != nil {
				if oti == mr.cycleStart {
					buffer = append(buffer, &Instruction{
						Op:     LIR_Mov,
						Result: &Addr{},
						Args:   []IOperand{&Register{}},
					})
					processed.Add(ofi)
					mr.foundNewCycle = true
				} else {
					buffer = mr.move(ofi, oti, prIdToFrom, processed, buffer)
				}
			} else {
				mr.cycleStart = fi
				buffer = mr.move(ofi, oti, prIdToFrom, processed, buffer)
				mr.cycleStart = nil
				if mr.foundNewCycle {
					mr.foundNewCycle = false
					buffer = append(buffer, &Instruction{
						Op:     LIR_Mov,
						Result: Addr{
							// TODO
						},
						Args: []IOperand{Register{
							// TODO
						}},
					})
				} else {
					buffer = append(buffer, &Instruction{
						Op:     LIR_Mov,
						Result: Addr{},
						Args:   []IOperand{Register{
							// TODO
						}},
					})
				}
			}
		} else {
			buffer = append(buffer, &Instruction{
				Op: LIR_Mov,
			})
		}
	}
	processed.Add(fi)
	return buffer
}

func newMoveResolver(ra *LSRA, from, to *ssa.Block) *MoveResolver {
	return &MoveResolver{
		ra:    ra,
		from:  from,
		to:    to,
		pairs: make(map[*Interval]*Interval),
	}
}
