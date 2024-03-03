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

package ssa

import (
	"falcon/utils"
	"fmt"
)

// == Code conjured by yyang, Feb, 2024 ==

// -----------------------------------------------------------------------------
// Loop Optimizations
//
// Let's do some fancy and interesting optimizations. I don't expect it to have
// any practical significance, it's just for showing off the power of Falcon IR.
type Loop struct {
	Header      *Block
	Tail        *Block
	Body        []*Block
	Parent      *Loop
	Childrens   []*Loop
	Irreducible bool
}

func (lp *Loop) String() string {
	return fmt.Sprintf("Loop@b%d", lp.Header.Id)
}

func (lp *Loop) FullString() string {
	body := ""
	for _, b := range lp.Body {
		body += fmt.Sprintf("b%d ", b.Id)
	}
	tail := ""
	if lp.Tail != nil {
		tail = fmt.Sprintf("%d", lp.Tail.Id)
	}
	return fmt.Sprintf("Loop@b%d(Tb%s): Parent:%v Childrens:%v Body:%v",
		lp.Header.Id, tail, lp.Parent, lp.Childrens, body)
}

func (lp *Loop) IsRotated() bool {
	// A loop is rotated if its tail determins whether to continue the loop
	// i.e. it has more than one successor
	return lp.Tail != nil && len(lp.Tail.Succs) > 1
}

func (lp *Loop) IsNormal() bool {
	// A loop is normal if its header determins whether to continue the loop
	// i.e. it has only one successor
	if !lp.IsRotated() {
		return true
	}
	return false
}

type LoopTree struct {
	fn    *Func
	Loops []*Loop
}

func (lt *LoopTree) FullString() string {
	loops := ""
	for _, loop := range lt.Loops {
		loops += loop.FullString() + "\n"
	}
	return fmt.Sprintf("[[LoopTree]]\n%v", loops)
}

func NewLoopTree(fn *Func) *LoopTree {
	return &LoopTree{fn, make([]*Loop, 0)}
}

func (lt *LoopTree) GetLoopById(id int) *Loop {
	for _, loop := range lt.Loops {
		if loop.Header.Id == id {
			return loop
		}
	}
	return nil
}

// ------------------------------------------------------------------------------
// Loop Detection
// Adopt a novel algorithm from "A New Algorithm for Identifying Loops in Decompilation"
// to find loop nesting. It runs in almost O(n) time complexity and is very
// efficient. It also able to detect irreducible loops.
const TraceLoopNesting = true

type LoopBuilder struct {
	visited     map[*Block]bool   // visited flag
	dfsp        map[*Block]int    // DFS spanning position
	iheader     map[*Block]*Block // innermost loop header of block
	headers     []*Block          // loop headers
	irreducible map[*Block]bool   // irreducible loop headers
}

// Tagging h as loop header of b
func (lb *LoopBuilder) taggingHeader(b, h *Block) {
	if b == h || h == nil {
		return
	}
	cur1, cur2 := b, h
	for lb.iheader[cur1] != nil {
		ih := lb.iheader[cur1]
		if ih == cur2 {
			return
		}
		if lb.dfsp[ih] < lb.dfsp[cur2] {
			lb.iheader[cur1] = cur2
			cur1 = cur2
			cur2 = ih
		} else {
			cur1 = ih
		}
	}
	lb.iheader[cur1] = cur2
}

func (lb *LoopBuilder) traverse(b0 *Block, DFSPPos int) *Block {
	lb.visited[b0] = true
	lb.dfsp[b0] = DFSPPos
	// p: starting from h0, the path to b0(if b0 is traversed)
	for _, b := range b0.Succs {
		if !lb.visited[b] {
			// case a: b is not traversed, traverse it; if then b is found in
			// loop body, tag b's innermost loop header as b0's header
			nh := lb.traverse(b, DFSPPos+1)
			lb.taggingHeader(b0, nh)
			continue
		}
		// b is traversed, denote "p" as the current path from entry to b0
		if lb.dfsp[b] > 0 {
			// case b: b is in p, tag b as b0's header
			lb.headers = append(lb.headers, b)
			lb.taggingHeader(b0, b)
		} else if lb.iheader[b] == nil {
			// case c: b is not in p nor in loop body, do nothing
		} else {
			h := lb.iheader[b] // h is b's innermost loop header
			if lb.dfsp[h] > 0 {
				// case d: b is not in p but its innermost loop header h is in p
				// tag h as b0's header
				lb.taggingHeader(b0, h)
			} else {
				// case e, b is not in p and its innermost loop header h is not in p
				// mark h and its ancestors as irreducible because h is entered
				// from either b0 or its loop entry
				lb.irreducible[h] = true
				for lb.iheader[h] != nil {
					h = lb.iheader[h]
					if lb.dfsp[h] > 0 {
						lb.taggingHeader(b0, h)
						break
					}
					// mark loop h irreducible
					lb.irreducible[h] = true
				}
			}
		}
	}
	lb.dfsp[b0] = 0
	return lb.iheader[b0]
}

func (lt *LoopTree) BuildLoopTree() {
	builder := &LoopBuilder{
		visited:     make(map[*Block]bool),
		dfsp:        make(map[*Block]int),
		iheader:     make(map[*Block]*Block),
		headers:     make([]*Block, 0),
		irreducible: make(map[*Block]bool),
	}
	// Traverse the CFG to find loop headers
	builder.traverse(lt.fn.Entry, 0)

	// Build loop tree
	domTree := BuildDomTree(lt.fn)
	for _, header := range builder.headers {
		loop := &Loop{Header: header}
		for _, pred := range header.Preds {
			if domTree.IsDominate(header, pred) {
				loop.Tail = pred
				break
			}
		}
		if builder.irreducible[header] {
			loop.Irreducible = true
		}
		lt.Loops = append(lt.Loops, loop)
	}
	for _, block := range lt.fn.Blocks {
		if header := builder.iheader[block]; header != nil {
			loop := lt.GetLoopById(header.Id)
			loop.Body = append(loop.Body, block)
			if innerLoop := lt.GetLoopById(block.Id); innerLoop != nil {
				innerLoop.Parent = loop
				loop.Childrens = append(loop.Childrens, innerLoop)
			}
		}
	}
	fmt.Printf("%v\n", lt.FullString())
}

// Induction variable refers to a variable that increments or decrements by a
// constant value in each iteration of a loop. If induction variable is in form
// of iv = base*factor + disp, then it's a basic induction variable. We detect
// basic induction variables and use them to optimize loop.
type InductionVar struct {
	base   *Value
	factor int
	disp   int
}

func (lt *LoopTree) FindIV() {
	ivs := make(map[*Value]*InductionVar)
	for _, loop := range lt.Loops {
		if loop.IsNormal() {
			// Collect all phis in the loop header as IV candidates
			for _, value := range loop.Header.Values {
				if value.Op != OpPhi {
					continue
				}
				ivs[value] = &InductionVar{
					base:   value,
					factor: 1,
					disp:   0,
				}
			}
			// Scan loop body to find factor and disp which in turns to find IV
			size := 0
			for size != len(ivs) {
				size = len(ivs)
				// Repeatly scan loop body to find IV
				for _, block := range loop.Body {
					for _, value := range block.Values {
						if value.Op == OpAdd {
							// TODO: implement OpSub
							arg0 := value.Args[0]
							arg1 := value.Args[1]
							var iv *InductionVar
							var c int
							if ivs[arg0] != nil && arg1.Op == OpConst {
								iv = ivs[arg0]
								c = arg1.Sym.(int)
							} else if ivs[arg1] != nil && arg0.Op == OpConst {
								iv = ivs[arg1]
								c = arg0.Sym.(int)
							}
							if iv == nil {
								continue
							}
							// #1 value = iv + c where iv = base*factor + disp
							// add value to iv such that value = base*factor + disp + c
							ivs[value] = &InductionVar{
								base:   iv.base,
								factor: iv.factor,
								disp:   iv.disp + c,
							}
						} else if value.Op == OpMul {
							arg0 := value.Args[0]
							arg1 := value.Args[1]
							var iv *InductionVar
							var f int
							if ivs[arg0] != nil && arg1.Op == OpConst {
								iv = ivs[arg0]
								f = arg1.Sym.(int)
							} else if ivs[arg1] != nil && arg0.Op == OpConst {
								iv = ivs[arg1]
								f = arg0.Sym.(int)
							}
							if iv == nil {
								continue
							}
							// #2 value = iv * f where iv = base*factor + disp
							// add value to iv such that value = base*factor*f + disp*f
							ivs[value] = &InductionVar{
								base:   iv.base,
								factor: iv.factor * f,
								disp:   iv.disp * f,
							}
						}
					}
				}
			}

		} else if loop.IsRotated() {
			// TODO: Implement it
		} else {
			// Unknown loop shape!
			utils.ShouldNotReachHere()
		}
		fmt.Printf("[[Induction Variables]]\n")
		for iv, indVar := range ivs {
			fmt.Printf("v%v: v%v*%v+%v\n", iv.Id, indVar.base.Id, indVar.factor, indVar.disp)
		}
	}
}

func OptimizeLoop(fn *Func) {
	lt := NewLoopTree(fn)
	lt.BuildLoopTree()
	lt.FindIV()
}
