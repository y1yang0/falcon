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
	"math"
)

type LSRA struct {
	lir    *LIR
	blocks []*ssa.Block

	instId2Block map[int]*ssa.Block // TODO: we should optimize this field out

	genKillMap   map[*ssa.Block]*GenKill
	liveInOutMap map[*ssa.Block]*LiveInOut

	vri2Interval map[int]*Interval // virtual register index to interval
	pri2Interval map[int]*Interval // physical register index to fixed interval

	nonFixedIntervals []*Interval

	workList *Interval
	current  *Interval

	actives  *utils.Set[*Interval]
	inactive *utils.Set[*Interval]

	spilled       bool
	nextStackSlot int // TODO: should we consider width?
}

type GenKill struct {
	gen  *utils.BitMap
	kill *utils.BitMap
}

type LiveInOut struct {
	in  *utils.BitMap
	out *utils.BitMap
}

func (ra *LSRA) allocateStackSlot() int {
	v := ra.nextStackSlot
	ra.nextStackSlot++
	return v
}

// used when building intervals
func (ra *LSRA) getOrCreateInterval(i int, virtual bool) *Interval {
	if !virtual {
		return ra.pri2Interval[i]
	}
	interval, ok := ra.vri2Interval[i]
	if !ok {
		interval = newInterval(i)
		ra.vri2Interval[i] = interval
		ra.nonFixedIntervals = append(ra.nonFixedIntervals, interval)
	}
	return interval
}

func (ra *LSRA) insertToWorkList(interval *Interval) {
	pos := &ra.workList

	for *pos != nil && (*pos).fistRange().from <= interval.fistRange().from {
		pos = &(*pos).next
	}

	interval.next = *pos
	*pos = interval
}

func (ra *LSRA) initOrder() {
	// TODO: A more appropriate order should be used.
	//       Order does not break correctness, but it is important for performance.
	//       For simplicity, we use the original order.
	ra.blocks = ra.lir.Blocks
}

func (ra *LSRA) numberInstructions() {
	ra.instId2Block = make(map[int]*ssa.Block)
	id := 0 // start from 0
	for _, b := range ra.blocks {
		is := ra.lir.Instructions[b.Id]
		for _, i := range is {
			i.Id = id
			ra.instId2Block[id] = b
			id++
		}
	}
}

func (ra *LSRA) computeGenKillMap(nofVR int) {
	m := make(map[*ssa.Block]*GenKill)
	for _, b := range ra.blocks {
		gk := GenKill{
			gen:  utils.NewBitMap(nofVR),
			kill: utils.NewBitMap(nofVR),
		}
		m[b] = &gk
		is := ra.lir.Instructions[b.Id]
		for _, i := range is {
			for _, a := range i.Args {
				if r, ok := a.(Register); ok {
					if r.Virtual && !gk.kill.IsSet(r.Index) {
						gk.gen.Set(r.Index)
					}
				}
			}

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
	m := make(map[*ssa.Block]*LiveInOut)
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

			for _, s := range b.Succs {
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
	ra.pri2Interval = make(map[int]*Interval)
	ra.nonFixedIntervals = make([]*Interval, 0)

	for i := phyRegStart(); i <= phyRegEnd(); i++ {
		ra.pri2Interval[i] = newFixedInterval(i)
	}

	ra.vri2Interval = make(map[int]*Interval)

	for i := len(ra.blocks) - 1; i >= 0; i-- {
		b := ra.blocks[i]
		inOut := ra.liveInOutMap[b]
		out := inOut.out
		for i := 0; i < out.Size(); i++ {
			is := ra.lir.Instructions[b.Id]
			if out.IsSet(i) {
				i := ra.getOrCreateInterval(i, true)
				i.addRange(is[0].Id, is[len(is)-1].Id)
			}
		}

		is := ra.lir.Instructions[b.Id]
		for i := len(is) - 1; i >= 0; i-- {
			instruction := is[i]

			if instruction.Op == LIR_Call {
				cs := callerSaved()
				for _, i := range cs {
					ra.pri2Interval[i].addRange(instruction.Id, instruction.Id)
				}
			}

			output := instruction.Result
			if r, ok := output.(Register); ok {
				interval := ra.getOrCreateInterval(r.Index, r.Virtual)
				interval.updateFromForFistRange(instruction.Id)
				interval.addUsePoint(instruction.Id, UKWrite)
			}

			for _, input := range instruction.Args {
				if r, ok := input.(Register); ok {
					interval := ra.getOrCreateInterval(r.Index, r.Virtual)
					interval.addRange(is[0].Id, instruction.Id)
					interval.addUsePoint(instruction.Id, UKRead)
				}
			}
		}
	}
}

func (ra *LSRA) allocateRegisters() {
	for _, i := range ra.pri2Interval {
		ra.insertToWorkList(i)
	}
	for _, i := range ra.vri2Interval {
		ra.insertToWorkList(i)
	}

	actives := utils.NewSet[*Interval]()
	inactives := utils.NewSet[*Interval]()
	ra.actives = actives
	ra.inactive = inactives

	for ra.workList != nil {
		ra.current = ra.workList
		ra.workList = ra.workList.next

		pos := ra.current.fistRange().from

		inactives.ForEach(func(interval *Interval) {
			if interval.spilled {
				inactives.Remove(interval)
				return
			}
			lr := interval.lastRange()
			if lr.to < pos {
				inactives.Remove(interval)
			} else {
				if interval.cover(pos) {
					inactives.Remove(interval)
					actives.Add(interval)
				}
			}
		})

		actives.ForEach(func(interval *Interval) {
			if interval.spilled {
				actives.Remove(interval)
				return
			}
			lr := interval.lastRange()
			if lr.to <= pos {
				actives.Remove(interval)
			} else {
				if !interval.cover(pos) {
					actives.Remove(interval)
					inactives.Add(interval)
				}
			}
		})

		if ra.current.fixed {
			actives.Add(ra.current)
			continue
		}

		if !ra.tryAllocatePhyReg() {
			ra.allocatePhyReg()
		}

		if ra.current.phyRegAssigned() {
			actives.Add(ra.current)
		}
	}
}

func (ra *LSRA) tryAllocatePhyReg() bool {
	free := make([]int, nofAvailPhyReg())

	for i := 0; i < len(free); i++ {
		free[i] = math.MaxInt
	}

	ra.actives.ForEach(func(interval *Interval) {
		free[interval.phyRegIndex] = 0
	})

	ra.inactive.ForEach(func(interval *Interval) {
		if ra.current.isIntersectingWith(interval) {
			free[interval.phyRegIndex] = min(ra.current.intersectionPositionWith(interval), free[interval.phyRegIndex])
		}
	})

	index := 0
	pos := free[0]
	for i := 1; i < len(free); i++ {
		if free[i] > pos {
			index = i
			pos = free[i]
		}
	}

	if pos == 0 {
		return false
	}

	ra.current.assignPhyReg(index)
	if pos <= ra.current.lastRange().to {
		// TODO: should select the optimal position
		ra.insertToWorkList(ra.current.splitAt(pos))
	}
	return true
}

func (ra *LSRA) spillInterval(interval *Interval) {
	if !interval.stackSlotIndexAssigned() {
		interval.assignStackSlot(ra.allocateStackSlot())
	}
	interval.spilled = true
}

func (ra *LSRA) forEachActiveAndInactiveInterval(f func(interval *Interval)) {
	ra.actives.ForEach(func(interval *Interval) {
		f(interval)
	})

	ra.inactive.ForEach(func(interval *Interval) {
		f(interval)
	})
}

func (ra *LSRA) allocatePhyReg() {
	l := nofAvailPhyReg()
	use := make([]int, l)
	block := make([]int, l)

	for i := 0; i < l; i++ {
		use[i] = math.MaxInt
		block[i] = math.MaxInt
	}

	ra.actives.ForEach(func(interval *Interval) {
		if !interval.fixed {
			use[interval.phyRegIndex] = min(use[interval.phyRegIndex], interval.nextUsePosAfter(ra.current.fistRange().from))
		}
	})

	ra.inactive.ForEach(func(interval *Interval) {
		if !interval.fixed {
			if ra.current.isIntersectingWith(interval) {
				use[interval.phyRegIndex] = min(use[interval.phyRegIndex], interval.nextUsePosAfter(ra.current.fistRange().from))
			}
		}
	})

	ra.actives.ForEach(func(interval *Interval) {
		if interval.fixed {
			use[interval.phyRegIndex] = 0
			block[interval.phyRegIndex] = 0
		}
	})

	ra.inactive.ForEach(func(interval *Interval) {
		if interval.fixed {
			if ra.current.isIntersectingWith(interval) {
				p := ra.current.intersectionPositionWith(interval)
				block[interval.phyRegIndex] = p
				use[interval.phyRegIndex] = min(p, use[interval.phyRegIndex])
			}
		}
	})

	index := 0
	pos := use[0]
	for i := 1; i < len(use); i++ {
		if use[i] > pos {
			index = i
			pos = use[i]
		}
	}

	if pos < ra.current.firstUsage() {
		ra.spillInterval(ra.current)
		u := ra.current.firstUsage()
		if u < math.MaxInt {
			ra.current.splitAt(u)
		}
	} else {
		ra.current.assignPhyReg(index)
		if block[index] > ra.current.lastRange().to {
			ra.insertToWorkList(ra.current.splitAt(block[index]))
		}
		ra.forEachActiveAndInactiveInterval(func(interval *Interval) {
			if !interval.fixed && interval.phyRegIndex == index {
				if ra.current.isIntersectingWith(interval) {
					f := ra.current.fistRange().from
					c := interval.splitAt(ra.current.fistRange().from)
					ra.spillInterval(c)
					n := c.nextUsePosAfter(f)
					if n != math.MaxInt {
						ra.insertToWorkList(c.splitAt(n))
					}
				}
			}
		})
	}
}

func (ra *LSRA) insertMoves() {
	for _, i := range ra.nonFixedIntervals {
		i.insertMoves(ra)
	}
}

func (ra *LSRA) resolveDataFlow() {
	for _, fb := range ra.blocks {
		for _, tb := range fb.Succs {
			mr := newMoveResolver(ra, fb, tb)
			lives := ra.liveInOutMap[tb].in
			for i := 0; i < lives.Size(); i++ {
				if lives.IsSet(i) {
					interval := ra.vri2Interval[i]
					from := interval.at(0)
					to := interval.at(1)
					if from != to {
						mr.record(from, to)
					}
				}
			}
			mr.resolve()
		}
	}
}

func (ra *LSRA) allocate() {
	nofVR := ra.lir.vid

	// TODO: trace support
	ra.initOrder()
	ra.numberInstructions()
	ra.computeGenKillMap(nofVR)
	ra.computeLiveInOutMap(nofVR)
	ra.buildIntervals()
	ra.allocateRegisters()
	ra.insertMoves()
	ra.resolveDataFlow()
}

// The entry
func lsra(lir *LIR) {
	ra := &LSRA{
		lir: lir,
	}
	ra.allocate()
}