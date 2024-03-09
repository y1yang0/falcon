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
	"fmt"
	"math"
)

type Interval struct {
	index int

	// range is a keyword, use _range instead
	_range   *Range
	usePoint *UsePoint

	phyRegIndex    int
	stackSlotIndex int

	parent   *Interval
	children *Interval
	sibling  *Interval

	next *Interval

	fixed   bool
	spilled bool
}

func (i *Interval) String() string {
	str := "Interval@" + fmt.Sprintf("%d", i) + ":"
	r := i._range
	for r != nil {
		str += fmt.Sprintf(" [%d,%d)", i._range.from, i._range.to)
	}
	return str
}

type Range struct {
	// from instruction id, inclusive
	from int

	// to instruction id, inclusive
	to int

	next *Range
}

type UsePoint struct {
	id   int // instruction id
	kind UseKind

	next *UsePoint
}

type UseKind int

const (
	UKRead UseKind = iota
	UKWrite
)

func newInterval(vri int) *Interval {
	return &Interval{
		index:          vri,
		phyRegIndex:    -1,
		stackSlotIndex: -1,
	}
}

func newFixedInterval(pri int) *Interval {
	return &Interval{
		index:       -1,
		phyRegIndex: pri,
		fixed:       true,
	}
}

func (i *Interval) fistRange() *Range {
	return i._range
}

func (i *Interval) lastRange() *Range {
	r := i._range
	for r.next != nil {
		r = r.next
	}
	return r
}

func (i *Interval) firstUsage() int {
	if i.usePoint != nil {
		return i.usePoint.id
	}
	return math.MaxInt
}

func (i *Interval) cover(pos int) bool {
	r := i._range
	for r != nil {
		if r.from <= pos && r.to >= pos {
			return true
		}
		r = r.next
	}
	return false
}

func (i *Interval) addRange(from, to int) {
	if i._range == nil {
		i._range = &Range{
			from: from,
			to:   to,
		}
		return
	}

	if i._range.from <= to+1 {
		// merge
		i._range.from = min(i._range.from, from)
	} else {
		i._range = &Range{
			from: from,
			to:   to,
			next: i._range,
		}
	}
}

func (i *Interval) updateFromForFistRange(from int) {
	if i._range == nil {
		i._range = &Range{
			from: from,
			to:   from,
		}
	} else {
		i._range.from = from
	}
}

func (i *Interval) addUsePoint(id int, kind UseKind) {
	i.usePoint = &UsePoint{
		id:   id,
		kind: kind,
		next: i.usePoint,
	}
}

func (i *Interval) intersectionPositionWith(o *Interval) int {
	i1 := i._range
	i2 := o._range

	for i1 != nil && i2 != nil {
		if i2.from > i1.to {
			i1 = i1.next
		} else if i2.to < i1.from {
			i2 = i2.next
		} else {
			return min(i1.from, i2.from)
		}
	}
	return math.MaxInt
}

func (i *Interval) isIntersectingWith(o *Interval) bool {
	return i.intersectionPositionWith(o) != math.MaxInt
}

func (i *Interval) splitAt(pos int) *Interval {
	// TODO: assert i.cover(pos) is true
	r := i._range
	for r.to < pos {
		r = r.next
	}
	ni := &Interval{
		index: i.index,
	}

	parent := i
	if i.parent != nil {
		parent = i.parent
	}

	ni.parent = parent

	cp := &parent.children
	for *cp != nil {
		cp = &(*cp).sibling
	}
	*cp = ni

	if r.from < pos {
		nr := &Range{
			from: pos,
			to:   r.to,
			next: r.next,
		}
		r.next = nil
		ni._range = nr
	} else {
		ni._range = r
		pr := i._range
		for pr.next != r {
			pr = pr.next
		}
		pr.next = nil
	}

	up := &i.usePoint
	for *up != nil && (*up).id < pos {
		up = &(*up).next
	}
	ni.usePoint = *up
	*up = nil
	return ni
}

func (i *Interval) nextUsePosAfter(pos int) int {
	u := i.usePoint
	for u != nil {
		if u.id > pos {
			return u.id
		}
		u = u.next
	}

	return math.MaxInt
}

func (i *Interval) at(pos int) *Interval {
	if i.cover(pos) {
		return i
	}
	c := i.children
	for c != nil {
		if c.cover(pos) {
			return c
		}
		c = c.sibling
	}
	// TODO: should not reach here
	return nil
}

func (i *Interval) phyRegAssigned() bool {
	return i.phyRegIndex != -1
}

func (i *Interval) assignPhyReg(index int) {
	i.phyRegIndex = index
}

func (i *Interval) stackSlotIndexAssigned() bool {
	if i.parent != nil {
		return i.parent.stackSlotIndex != -1
	}
	return i.stackSlotIndex != -1
}

func (i *Interval) assignStackSlot(index int) {
	if i.parent != nil {
		i.parent.stackSlotIndex = index
	} else {
		i.stackSlotIndex = index
	}
}

func (i *Interval) insertMoves(ra *LSRA) {
	// left := i
	// right := i.children

	// for right != nil {
	// 	id1 := left.lastRange().to
	// 	id2 := right.fistRange().from
	// 	if id1+1 == id2 {
	// 		b1 := ra.instId2Block[id1]
	// 		b2 := ra.instId2Block[id2]
	// 		if b1 == b2 {
	// 			if left.spilled {
	// 				if !right.spilled {
	// 					insts := ra.lir.Instructions[b1.Id]
	// 					ra.lir.Instructions[b1.Id] = utils.InsertAt(
	// 						insts,
	// 						indexOfInst(insts, id2),
	// 						&Instruction{
	// 							Op: LIR_Mov,
	// 							Result: Register{
	// 								// TODO
	// 								Virtual: false,
	// 							},
	// 							Args: []IOperand{
	// 								Addr{
	// 									// TODO
	// 								},
	// 							},
	// 						},
	// 					)
	// 				}
	// 			} else if right.spilled {
	// 				if !left.spilled {
	// 					insts := ra.lir.Instructions[b1.Id]
	// 					ra.lir.Instructions[b1.Id] = utils.InsertAt(
	// 						insts,
	// 						indexOfInst(insts, id2),
	// 						&Instruction{
	// 							Op:     LIR_Mov,
	// 							Result: Addr{
	// 								// TODO
	// 							},
	// 							Args: []IOperand{
	// 								Register{
	// 									// TODO
	// 									Virtual: false,
	// 								},
	// 							},
	// 						},
	// 					)
	// 				}
	// 			} else if left.phyRegIndex != right.phyRegIndex {
	// 				insts := ra.lir.Instructions[b1.Id]
	// 				ra.lir.Instructions[b1.Id] = utils.InsertAt(
	// 					insts,
	// 					indexOfInst(insts, id2),
	// 					&Instruction{
	// 						Op: LIR_Mov,
	// 						Result: Register{
	// 							// TODO
	// 							Virtual: false,
	// 						},
	// 						Args: []IOperand{
	// 							Register{
	// 								// TODO
	// 								Virtual: false,
	// 							},
	// 						},
	// 					},
	// 				)
	// 			}
	// 		}
	// 	}
	// 	left = right
	// 	right = right.sibling
	// }
}

func indexOfInst(insts []*Instruction, id int) int {
	for i, e := range insts {
		if e.Id == id {
			return i
		}
	}
	// TODO: should not reach here
	return -1
}
