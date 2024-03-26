// Copyright (c) 2024 The Falcon Programming Language
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
	"math"
	"sort"
)

// -----------------------------------------------------------------------------
// Linear Scan Register Allocation
//
// After lowering the IR to LIR, we need to allocate registers to the virtual
// registers. We use the linear scan register allocation algorithm to do this.
// The algorithm is based on the paper "Linear Scan Register Allocation for the
// Java HotSpot™ Client Compiler" by Christian Wimmer, et al.

type LSRA struct {
	lir    *LIR
	blocks []int

	genKillMap   map[int]*GenKill
	liveInOutMap map[int]*LiveInOut

	vreg2Interval map[int]*Interval // register index to interval

	workList []*Interval
	current  *Interval

	actives  []*Interval
	inactive []*Interval
	handled  []*Interval

	spilled       bool
	nextStackSlot int
}

// Interval represents a live interval, it contains a list of ranges and a list
// of use points. The ranges are sorted by the start position. The use points
// denote the instruction positions where the interval is used.
type Interval struct {
	index int

	// range is a keyword, use _range instead
	ranges []*Range
	uses   []*UsePoint

	phyRegIndex int
	fixed       bool
	sub         []*Interval
	parent      *Interval
}

func (i *Interval) String() string {
	str := "@"
	if i.fixed {
		str += "+"
	}
	for _, r := range i.ranges {
		str += fmt.Sprintf("[i%d,i%d)", r.from, r.to)
	}
	str += " @"
	for _, u := range i.uses {
		str += fmt.Sprintf("i%d ", u.id)
	}
	if len(i.sub) > 0 {
		str += " sub:"
		for _, s := range i.sub {
			str += fmt.Sprintf("i%d ", s.index)
		}
	}
	if i.parent != nil {
		str += fmt.Sprintf(" parent:i%d", i.parent.index)
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
}

type UseKind int

const (
	UKRead UseKind = iota
	UKWrite
)

func newInterval(vri int) *Interval {
	return &Interval{
		index:       vri,
		phyRegIndex: -1,
		fixed:       false,
		sub:         make([]*Interval, 0),
		parent:      nil,
	}
}

func (i *Interval) NumRanges() int {
	return len(i.ranges)
}

func (i *Interval) firstRange() *Range {
	return i.ranges[0]
}

func (i *Interval) lastRange() *Range {
	return i.ranges[len(i.ranges)-1]
}

func (i *Interval) cover(pos int) bool {
	for _, r := range i.ranges {
		if r.from <= pos && r.to >= pos {
			return true
		}
		r = r.next
	}
	return false
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (i *Interval) addRange(from, to int) {
	for _, r := range i.ranges {
		// Two ranges are overlapped
		if r.from <= from && r.to >= from {
			r.to = max(r.to, to)
			return
		} else if r.from <= to && r.to >= to {
			r.from = min(r.from, from)
			return
		}
	}
	// No overlapped range found, add a new range
	i.ranges = append(i.ranges, &Range{
		from: from,
		to:   to,
	})
}

func (i *Interval) addUsePoint(id int, kind UseKind) {
	i.uses = append(i.uses, &UsePoint{
		id:   id,
		kind: kind,
	})
}

func (i *Interval) intersect(k *Interval) int {
	for _, r1 := range i.ranges {
		for _, r2 := range k.ranges {
			if r1.from <= r2.to && r1.to >= r2.from {
				return min(r1.to, r2.to)
			}
		}
	}
	return -1
}

func (i *Interval) splitAt(pos int) *Interval {
	// split interval at given position into two sub intervals
	// the first sub interval is [from, pos], the second sub interval is [pos+1, to]
	// the second sub interval is returned
	for _, r := range i.ranges {
		if r.from <= pos && r.to >= pos {
			// pos is coverred by the interval, split it!
			ni := newInterval(i.index)
			ni.ranges = append(ni.ranges, &Range{
				from: pos + 1,
				to:   r.to,
			})
			r.to = pos
			i.sub = append(i.sub, ni)
			ni.parent = i
			return ni
		}
	}
	return nil
}

// func (i *Interval) splitAt(pos int) *Interval {
// 	// TODO: assert i.cover(pos) is true
// 	r := i._range
// 	for r.to < pos {
// 		r = r.next
// 	}
// 	ni := &Interval{
// 		index: i.index,
// 	}

// 	parent := i
// 	if i.parent != nil {
// 		parent = i.parent
// 	}

// 	ni.parent = parent

// 	cp := &parent.children
// 	for *cp != nil {
// 		cp = &(*cp).sibling
// 	}
// 	*cp = ni

// 	if r.from < pos {
// 		nr := &Range{
// 			from: pos,
// 			to:   r.to,
// 			next: r.next,
// 		}
// 		r.next = nil
// 		ni._range = nr
// 	} else {
// 		ni._range = r
// 		pr := i._range
// 		for pr.next != r {
// 			pr = pr.next
// 		}
// 		pr.next = nil
// 	}

// 	up := &i.usePoint
// 	for *up != nil && (*up).id < pos {
// 		up = &(*up).next
// 	}
// 	ni.usePoint = *up
// 	*up = nil
// 	return ni
// }

// func (i *Interval) nextUsePosAfter(pos int) int {
// 	u := i.usePoint
// 	for u != nil {
// 		if u.id > pos {
// 			return u.id
// 		}
// 		u = u.next
// 	}

// 	return math.MaxInt
// }

// func (i *Interval) at(pos int) *Interval {
// 	if i.cover(pos) {
// 		return i
// 	}
// 	c := i.children
// 	for c != nil {
// 		if c.cover(pos) {
// 			return c
// 		}
// 		c = c.sibling
// 	}
// 	// TODO: should not reach here
// 	return nil
// }

// func (i *Interval) stackSlotIndexAssigned() bool {
// 	if i.parent != nil {
// 		return i.parent.stackSlotIndex != -1
// 	}
// 	return i.stackSlotIndex != -1
// }

// func (i *Interval) assignStackSlot(index int) {
// 	if i.parent != nil {
// 		i.parent.stackSlotIndex = index
// 	} else {
// 		i.stackSlotIndex = index
// 	}
// }

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

type MoveResolver struct {
	ra   *LSRA
	from int
	to   int

	pairs map[*Interval]*Interval

	stackSlot int

	cycleStart    *Interval
	foundNewCycle bool
}

func (mr *MoveResolver) record(from, to *Interval) {
	mr.pairs[from] = to
}

func (mr *MoveResolver) resolve() {
	// utils.Assert(len(mr.from.Succs) > 1 && len(mr.to.Preds) > 1, "sanity check")

	// // decide insertion position
	// var b *ssa.Block
	// var is []*Instruction
	// var pos int
	// if len(mr.from.Succs) == 1 {
	// 	b = mr.from
	// 	is = mr.ra.lir.Instructions[b.Id]
	// 	l := len(is)
	// 	if isBranchOp(is[l-1].Op) {
	// 		pos = l - 1
	// 	} else {
	// 		pos = l
	// 	}
	// } else {
	// 	b = mr.to
	// 	is = mr.ra.lir.Instructions[b.Id]
	// 	pos = 0
	// }

	// prIdToFrom := make(map[int]*Interval)
	// for fi, _ := range mr.pairs {
	// 	if fi.phyRegAssigned() {
	// 		prIdToFrom[fi.phyRegIndex] = fi
	// 	}
	// }

	// // process pairs
	// processed := utils.NewSet[*Interval]()
	// buffer := make([]*Instruction, 0)
	// for fi, ti := range mr.pairs {
	// 	buffer = mr.move(fi, ti, prIdToFrom, processed, buffer)
	// }

	// // insert new instructions
	// nis := make([]*Instruction, 0)
	// for i := 0; i < pos; i++ {
	// 	nis = append(nis, is[i])
	// }
	// nis = append(nis, buffer...)
	// for i := pos; i < len(is); i++ {
	// 	nis = append(nis, is[i])
	// }
	// mr.ra.lir.Instructions[b.Id] = nis
}

// func (mr *MoveResolver) move(fi *Interval, ti *Interval, prIdToFrom map[int]*Interval, processed *utils.Set[*Interval], buffer []*Instruction) []*Instruction {
// 	if processed.Contains(fi) {
// 		return buffer
// 	}

// 	if ti.stackSlotIndex != -1 {
// 		buffer = append(buffer, &Instruction{
// 			Op:     LIR_Mov,
// 			Result: &Addr{},
// 			Args:   []IOperand{&Register{}},
// 		})
// 	} else {
// 		ofi, ok := prIdToFrom[ti.phyRegIndex]
// 		if ok && !processed.Contains(ofi) {
// 			oti := mr.pairs[ofi]

// 			if mr.cycleStart != nil {
// 				if oti == mr.cycleStart {
// 					buffer = append(buffer, &Instruction{
// 						Op:     LIR_Mov,
// 						Result: &Addr{},
// 						Args:   []IOperand{&Register{}},
// 					})
// 					processed.Add(ofi)
// 					mr.foundNewCycle = true
// 				} else {
// 					buffer = mr.move(ofi, oti, prIdToFrom, processed, buffer)
// 				}
// 			} else {
// 				mr.cycleStart = fi
// 				buffer = mr.move(ofi, oti, prIdToFrom, processed, buffer)
// 				mr.cycleStart = nil
// 				if mr.foundNewCycle {
// 					mr.foundNewCycle = false
// 					buffer = append(buffer, &Instruction{
// 						Op:     LIR_Mov,
// 						Result: Addr{
// 							// TODO
// 						},
// 						Args: []IOperand{Register{
// 							// TODO
// 						}},
// 					})
// 				} else {
// 					buffer = append(buffer, &Instruction{
// 						Op:     LIR_Mov,
// 						Result: Addr{},
// 						Args:   []IOperand{Register{
// 							// TODO
// 						}},
// 					})
// 				}
// 			}
// 		} else {
// 			buffer = append(buffer, &Instruction{
// 				Op: LIR_Mov,
// 			})
// 		}
// 	}
// 	processed.Add(fi)
// 	return buffer
// }

func newMoveResolver(ra *LSRA, from, to int) *MoveResolver {
	return &MoveResolver{
		ra:    ra,
		from:  from,
		to:    to,
		pairs: make(map[*Interval]*Interval),
	}
}

type GenKill struct {
	gen  *utils.BitMap
	kill *utils.BitMap
}

type LiveInOut struct {
	in  *utils.BitMap
	out *utils.BitMap
}

func (x *GenKill) String() string {
	return fmt.Sprintf("[gen:%s, kill:%s]", x.gen, x.kill)
}

func (x *LiveInOut) String() string {
	return fmt.Sprintf("[in:%s, out:%s]", x.in, x.out)
}

func (ra *LSRA) allocateStackSlot() int {
	v := ra.nextStackSlot
	ra.nextStackSlot++
	return v
}

// used when building intervals
func (ra *LSRA) getOrCreateInterval(i int) *Interval {
	if interval, ok := ra.vreg2Interval[i]; interval != nil && ok {
		return interval
	}
	interval := newInterval(i)
	ra.vreg2Interval[i] = interval
	return interval
}

func (ra *LSRA) initOrder() {
	// TODO: A more appropriate order should be used.
	//       Order does not break correctness, but it is important for performance.
	//       For simplicity, we use the original order.
	lir := ra.lir
	blocksOrder := make([]int, 0)
	for key := range lir.Instructions {
		blocksOrder = append(blocksOrder, key)
	}
	sort.SliceStable(blocksOrder, func(i, j int) bool {
		return blocksOrder[i] <= blocksOrder[j]
	})
	ra.blocks = blocksOrder
}

func (ra *LSRA) computeGenKillMap(nofVR int) {
	// Per-block liveness analysis
	m := make(map[int]*GenKill)
	for _, b := range ra.blocks {
		gk := GenKill{
			gen:  utils.NewBitMap(nofVR),
			kill: utils.NewBitMap(nofVR),
		}
		m[b] = &gk
		is := ra.lir.Instructions[b]
		for _, i := range is {
			// Instruction operands are all used after defined(say, in some preds),
			// i.e., generated
			for _, a := range i.Args {
				if r, ok := a.(Register); ok {
					if r.Virtual && !gk.kill.IsSet(r.Index) {
						gk.gen.Set(r.Index)
					}
				}
			}
			// Instruction result is defined, i.e., killed
			if r, ok := i.Result.(Register); ok {
				if r.Virtual {
					gk.kill.Set(r.Index)
				}
			}
		}
	}
	ra.genKillMap = m
}

func (ra *LSRA) computeLiveInOutMap(nofVR int) {
	// Global liveness analysis
	m := make(map[int]*LiveInOut)
	for _, b := range ra.blocks {
		m[b] = &LiveInOut{
			in:  utils.NewBitMap(nofVR),
			out: utils.NewBitMap(nofVR),
		}
	}
	changed := true
	for changed {
		for i := len(ra.blocks) - 1; i >= 0; i-- {
			b := ra.blocks[i]
			lio := m[b]
			// This is a backward data flow analysis, the rules are:
			// 1. LiveIn{b} = Gen{b} U (LiveOut{b} - Kill{b})
			// 2. LiveOut{b} = LiveIn{b} U LiveOut{succ1} U LiveOut{succ2} ...
			for _, s := range ra.lir.Edges[b] {
				lio2 := m[s]

				if lio.out.Unite(lio2.in) {
					changed = true
				}
			}

			in := lio.out.Copy()
			in.Remove(ra.genKillMap[b].kill)
			in.Unite(ra.genKillMap[b].gen)
			if lio.in.SetFrom(in) {
				changed = true
			}
		}
		changed = false
	}
	ra.liveInOutMap = m
}

func (ra *LSRA) buildIntervals() {
	ra.vreg2Interval = make(map[int]*Interval)

	for i := len(ra.blocks) - 1; i >= 0; i-- {
		b := ra.blocks[i]
		inOut := ra.liveInOutMap[b]
		out := inOut.out
		// For all instructions in the block, we build the initial intervals
		// which equals to the entire block, then try to shorten them.
		for i := 0; i < out.Size(); i++ {
			if out.IsSet(i) {
				is := ra.lir.Instructions[b]
				i := ra.getOrCreateInterval(i)
				i.addRange(is[0].Id, is[len(is)-1].Id)
			}
		}

		is := ra.lir.Instructions[b]
		for i := len(is) - 1; i >= 0; i-- {
			instruction := is[i]

			if instruction.Op == LIR_Call {
				regs := CallerSaveRegs(LIRTypeDWord)
				for _, i := range regs {
					i := ra.getOrCreateInterval(i.Index)
					i.fixed = true
					i.addRange(instruction.Id, instruction.Id)
				}
			}

			output := instruction.Result
			// Def point there, we need to update start position of the interval
			if r, ok := output.(Register); ok && output != NoReg {
				interval := ra.getOrCreateInterval(r.Index)
				if interval.NumRanges() > 0 {
					interval.firstRange().from = instruction.Id
				}
				interval.addUsePoint(instruction.Id, UKWrite)
			}
			// Use point there, we need to update end position of the interval
			// def is unknown, conservativly assume it starts at the beginning of
			// the block
			for _, input := range instruction.Args {
				if r, ok := input.(Register); ok && input != NoReg {
					blockFrom := is[0].Id
					interval := ra.getOrCreateInterval(r.Index)
					interval.addRange(blockFrom, instruction.Id)
					interval.addUsePoint(instruction.Id, UKRead)
				}
			}
		}
	}

	// TODO:Verify ranges in interval do not overlap
}

func sortWorklist(intervals []*Interval) {
	sort.SliceStable(intervals, func(i, j int) bool {
		return intervals[i].firstRange().from <= intervals[j].firstRange().from
	})
}

func (ra *LSRA) allocateRegisters() {
	for _, i := range ra.vreg2Interval {
		if i.ranges == nil {
			continue
		}
		ra.workList = append(ra.workList, i)
	}

	// cover pos and assigned a register
	actives := make([]*Interval, 0)
	// start before pos and end after pos, but do not cover pos
	inactives := make([]*Interval, 0)
	// end before pos or spilled to mem
	handled := make([]*Interval, 0)
	ra.actives = actives
	ra.inactive = inactives
	ra.handled = handled

	for len(ra.workList) > 0 {
		// Pick up lowest start position interval and process it
		sort.SliceStable(ra.workList, func(i, j int) bool {
			return ra.workList[i].firstRange().from <= ra.workList[j].firstRange().from
		})
		ra.current = ra.workList[0]
		ra.workList = ra.workList[1:]
		pos := ra.current.firstRange().from

		for i := len(actives) - 1; i >= 0; i-- {
			interval := actives[i]
			if interval.lastRange().to < pos {
				// Active interval does not overlap with pos, mark it as done
				// given that it is already processed
				actives = append(actives[:i], actives[i+1:]...)
				handled = append(handled, interval)
			} else if !interval.cover(pos) {
				// Active interval does not overlap with pos but not processed
				// yet, move it to inactive
				actives = append(actives[:i], actives[i+1:]...)
				inactives = append(inactives, interval)
			} else {
				// Any remaining intervals are really active
			}
		}

		for i := len(inactives) - 1; i >= 0; i-- {
			interval := inactives[i]
			if interval.lastRange().to < pos {
				// Inactive interval does not overlap with pos, move it to handled
				inactives = append(inactives[:i], inactives[i+1:]...)
				handled = append(handled, interval)
			} else if interval.cover(pos) {
				// Bad case, it becomes active again
				inactives = append(inactives[:i], inactives[i+1:]...)
				actives = append(actives, interval)
			} else {
				// Any remaining intervals are really inactive
			}
		}

		// Try to allocate physical register for current interval

		if !ra.tryAllocatePhyReg() {
			// ra.allocatePhyReg()
		}
		if ra.current.phyRegIndex != -1 {
			actives = append(actives, ra.current)
		}
		ra.actives = actives
		ra.inactive = inactives
		ra.handled = handled
	}
}

func (ra *LSRA) tryAllocatePhyReg() bool {
	freeRegPos := make(map[int]int)
	setFreePos := func(reg int, pos int) {
		if _, exist := freeRegPos[reg]; exist {
			freeRegPos[reg] = 0
		} else {
			utils.Assert(false, "why not otherwise")
		}
	}
	// All registers are entirely free
	for _, reg := range CallerSaveRegs(LIRTypeDWord) {
		freeRegPos[reg.Index] = math.MaxInt
	}

	// Remove the registers that are already assigned to active intervals
	for _, i := range ra.actives {
		setFreePos(i.phyRegIndex, 0)
	}
	// Inactive set is guaranteed to not cover start position of current interval
	// but MAY cover end position of current interval
	for _, i := range ra.inactive {
		if k := i.intersect(ra.current); k != -1 {
			// Bad case, inactive interval is intersecting with current interval
			// at position k, it frees until k
			setFreePos(i.phyRegIndex, k)
		}
	}

	// Pick up highest position from freeRegPos
	keys := make([]int, 0)
	for k := range freeRegPos {
		keys = append(keys, k)
	}
	sort.SliceStable(keys, func(i, j int) bool {
		return keys[i] <= keys[j]
	})
	index := -1
	pos := 0
	for _, k := range keys {
		if freeRegPos[k] > pos {
			index = k
			pos = freeRegPos[k]
		}
	}

	fmt.Printf("freePos: %s, interval %v\n", FindRegisterByIndex(index), ra.current)
	if freeRegPos[index] == -1 {
		// No free register
		return false
	} else if freeRegPos[index] > ra.current.lastRange().to {
		// The register is entire free across the current interval
		utils.Assert(ra.current.phyRegIndex == -1, "already assigned")
		ra.current.phyRegIndex = index
	} else {
		if !ra.current.cover(freeRegPos[index]) {
			fmt.Printf("xxfreePos: %s, interval %v\n", FindRegisterByIndex(index), ra.current)
		}
		utils.Assert(ra.current.cover(freeRegPos[index]), "should intersect")
		// The free register is not entirely free across the current interval
		// split current interval into two sub intervals, so that free register
		// can be used for the first sub interval, and second sub interval is
		// added to worklist for further processing.
		remaining := ra.current.splitAt(freeRegPos[index])
		utils.Assert(remaining != nil, "split interval failed")
		ra.workList = append(ra.workList, remaining)
	}
	return true
}

// func (ra *LSRA) spillInterval(interval *Interval) {
// 	if !interval.stackSlotIndexAssigned() {
// 		interval.assignStackSlot(ra.allocateStackSlot())
// 	}
// 	interval.spilled = true
// }

// func (ra *LSRA) allocatePhyReg() {
// 	l := nofAvailPhyReg()
// 	use := make([]int, l)
// 	block := make([]int, l)

// 	for i := 0; i < l; i++ {
// 		use[i] = math.MaxInt
// 		block[i] = math.MaxInt
// 	}

// 	ra.actives.ForEach(func(interval *Interval) {
// 		if !interval.fixed {
// 			use[interval.phyRegIndex] = min(use[interval.phyRegIndex], interval.nextUsePosAfter(ra.current.fistRange().from))
// 		}
// 	})

// 	ra.inactive.ForEach(func(interval *Interval) {
// 		if !interval.fixed {
// 			if ra.current.isIntersectingWith(interval) {
// 				use[interval.phyRegIndex] = min(use[interval.phyRegIndex], interval.nextUsePosAfter(ra.current.fistRange().from))
// 			}
// 		}
// 	})

// 	ra.actives.ForEach(func(interval *Interval) {
// 		if interval.fixed {
// 			use[interval.phyRegIndex] = 0
// 			block[interval.phyRegIndex] = 0
// 		}
// 	})

// 	ra.inactive.ForEach(func(interval *Interval) {
// 		if interval.fixed {
// 			if ra.current.isIntersectingWith(interval) {
// 				p := ra.current.intersectionPositionWith(interval)
// 				block[interval.phyRegIndex] = p
// 				use[interval.phyRegIndex] = min(p, use[interval.phyRegIndex])
// 			}
// 		}
// 	})

// 	index := 0
// 	pos := use[0]
// 	for i := 1; i < len(use); i++ {
// 		if use[i] > pos {
// 			index = i
// 			pos = use[i]
// 		}
// 	}

// 	if pos < ra.current.firstUsage() {
// 		ra.spillInterval(ra.current)
// 		u := ra.current.firstUsage()
// 		if u < math.MaxInt {
// 			ra.current.splitAt(u)
// 		}
// 	} else {
// 		ra.current.assignPhyReg(index)
// 		if block[index] > ra.current.lastRange().to {
// 			ra.insertToWorkList(ra.current.splitAt(block[index]))
// 		}
// 		ra.forEachActiveAndInactiveInterval(func(interval *Interval) {
// 			if !interval.fixed && interval.phyRegIndex == index {
// 				if ra.current.isIntersectingWith(interval) {
// 					f := ra.current.fistRange().from
// 					c := interval.splitAt(ra.current.fistRange().from)
// 					ra.spillInterval(c)
// 					n := c.nextUsePosAfter(f)
// 					if n != math.MaxInt {
// 						ra.insertToWorkList(c.splitAt(n))
// 					}
// 				}
// 			}
// 		})
// 	}
// }

// func (ra *LSRA) insertMoves() {
// 	for _, i := range ra.nonFixedIntervals {
// 		i.insertMoves(ra)
// 	}
// }

// func (ra *LSRA) resolveDataFlow() {
// 	for _, fb := range ra.blocks {
// 		for _, tb := range ra.lir.Edges[fb] {
// 			mr := newMoveResolver(ra, fb, tb)
// 			lives := ra.liveInOutMap[tb].in
// 			for i := 0; i < lives.Size(); i++ {
// 				if lives.IsSet(i) {
// 					interval := ra.vri2Interval[i]
// 					from := interval.at(0)
// 					to := interval.at(1)
// 					if from != to {
// 						mr.record(from, to)
// 					}
// 				}
// 			}
// 			mr.resolve()
// 		}
// 	}
// }

func (ra *LSRA) printGenKill() {
	fmt.Printf("===LiveGenKill==\n")
	for k, v := range ra.genKillMap {
		fmt.Printf("b%d: %s\n", k, v)
	}
}

func (ra *LSRA) printLiveInOut() {
	fmt.Printf("===LiveInOut==\n")
	for k, v := range ra.liveInOutMap {
		fmt.Printf("b%d: %s\n", k, v)
	}
}

func (ra *LSRA) printIntervals() {
	fmt.Printf("==Interval==\n")
	maxWidth := 0
	for _, i := range ra.vreg2Interval {
		last := i.lastRange().to
		if last > maxWidth {
			maxWidth = last
		}
	}
	for i := 0; i <= maxWidth; i++ {
		fmt.Printf("%d ", i)
	}
	fmt.Printf("\n")
	for k, i := range ra.vreg2Interval {
		from := i.firstRange().from
		to := i.lastRange().to
		var reg string
		if k >= 0 {
			reg = fmt.Sprintf("v%d", k)
		} else {
			reg = fmt.Sprintf("%s", FindRegisterByIndex(k))
		}
		s := 0
		fmt.Printf("|")

		for s < from {
			fmt.Printf("%s", "  ")
			s++
		}
		if from != 0 {
			fmt.Printf("|")
		} else {
			fmt.Printf("-")
		}
		for s <= to {
			fmt.Printf("%s", "--")
			s++
		}
		if to != maxWidth {
			fmt.Printf("|")
		} else {
			fmt.Printf("-")
		}
		for s <= maxWidth {
			fmt.Printf("%s", "  ")
			s++
		}

		fmt.Printf("| ")
		fmt.Printf("%s %v\n", reg, i)

	}
}

func (ra *LSRA) printRegAllocation() {
	fmt.Printf("==Assign==\n")
	for _, i := range ra.vreg2Interval {
		var regName string
		if i.index >= 0 {
			regName = fmt.Sprintf("v%d", i.index)
		} else {
			regName = FindRegisterByIndex(i.index).String()
		}
		fmt.Printf("interval:%s reg:%s\n", regName, FindRegisterByIndex(i.phyRegIndex).String())
	}
}

func (ra *LSRA) materialize() {
	// Replace virtual register with allocated physical register
	for _, block := range ra.blocks {
		is := ra.lir.Instructions[block]
		for _, instr := range is {
			for iarg, a := range instr.Args {
				if r, ok := a.(Register); ok {
					if r.Virtual {
						interval := ra.vreg2Interval[r.Index]
						instr.Args[iarg] = FindRegisterByIndex(interval.phyRegIndex)
					}
				}
			}
			if r, ok := instr.Result.(Register); ok {
				if r.Virtual {
					interval := ra.vreg2Interval[r.Index]
					instr.Result = FindRegisterByIndex(interval.phyRegIndex)
				}
			}
		}
	}
}

func (ra *LSRA) dumpInterval() {

}

func (ra *LSRA) allocate() {
	nofVR := ra.lir.vid

	// TODO: trace support
	ra.initOrder()
	ra.computeGenKillMap(nofVR)
	ra.printGenKill()
	ra.computeLiveInOutMap(nofVR)
	ra.printLiveInOut()
	ra.buildIntervals()

	ra.printIntervals()
	ra.allocateRegisters()
	ra.printRegAllocation()
	// ra.insertMoves()
	// ra.resolveDataFlow()
	ra.materialize()
}

// The entry
func lsra(lir *LIR) {
	ra := &LSRA{
		lir: lir,
	}
	ra.allocate()
	// TODO: Verify all virtual register are assigned
}
