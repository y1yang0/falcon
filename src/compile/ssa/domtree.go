package ssa

import (
	"falcon/utils"
	"fmt"
)

// ------------------------------------------------------------------------------
// Dominator tree
//
// There are some general dominator definitions:
// * Dominators: a dom b if all paths from entry to block b include a
// * Strict Dominators: a sdom b if a dom b and a != b
// * Immediate Dominators: a idom b if a sdom b and there is no block c such that
// a sdom c sdom b
// * Post Dominators: a pdom b if all paths from block b to exit include a, this
// is reverse of dominators
//
// See "Graph-theoretic constructs for program flow analysis" for more information
// This is a iterative algorithm to compute dominator tree, it has O(n^2) time
// complexity
type DomTree struct {
	Func *Func
	Dom  map[*Block][]*Block
}

// a dom b if all paths from entry to block b include a
func (dt *DomTree) IsDominate(a, b *Block) bool {
	for _, dom := range dt.Dom[b] {
		if dom == a {
			return true
		}
	}
	return false
}

// a sdom b if a dom b and a != b
func (dt *DomTree) IsSDominate(a, b *Block) bool {
	return dt.IsDominate(a, b) && a != b
}

// a idom b if a sdom b and there is no block c such that a sdom c sdom b
func (dt *DomTree) IsIDominate(a, b *Block) bool {
	return dt.IsSDominate(a, b) && !dt.IsSDominate(b, a)
}

func intersect(a []*Block, b []*Block) []*Block {
	if len(a) > len(b) {
		a, b = b, a
	}
	res := make([]*Block, 0, len(a))
	for _, x := range a {
		for _, y := range b {
			if x == y {
				res = append(res, x)
				break
			}
		}
	}
	return res
}

func union(a []*Block, b []*Block) []*Block {
	m := make(map[*Block]bool)
	for _, x := range a {
		m[x] = true
	}
	for _, x := range b {
		m[x] = true
	}
	res := make([]*Block, 0, len(m))
	for x := range m {
		res = append(res, x)
	}
	return res
}

func (dt *DomTree) String() string {
	s := "== Dom Tree:\n"
	for block, doms := range dt.Dom {
		s += fmt.Sprintf("b%d", block.Id)
		s += ":"
		for _, dom := range doms {
			s += fmt.Sprintf(" b%d", dom.Id)
		}
		s += "\n"
	}
	return s
}

func BuildDomTree(fn *Func) *DomTree {
	dom := make(map[*Block][]*Block, len(fn.Blocks)) // block dominated by which blocks
	dom[fn.Entry] = []*Block{fn.Entry}
	for _, block := range fn.Blocks {
		if block == fn.Entry {
			continue
		}
		dom[block] = fn.Blocks
	}

	// Iteratively compute dom tree
	changed := true
	for changed {
		changed = false
		for _, block := range fn.Blocks {
			if block == fn.Entry {
				continue
			}
			var newdom []*Block
			if len(block.Preds) > 0 {
				newdom = dom[block.Preds[0]]
				for _, pred := range block.Preds[1:] {
					newdom = intersect(newdom, dom[pred])
				}
			}
			newdom = union(newdom, []*Block{block})
			if len(newdom) != len(dom[block]) {
				changed = true
				dom[block] = newdom
			}
		}
	}
	// Computes dom tree
	return &DomTree{Func: fn, Dom: dom}
}

// Verify the dominance relationship of a function
func VerifyDom(fn *Func) {
	domTree := BuildDomTree(fn)
	for _, block := range fn.Blocks {
		for _, val := range block.Values {
			for _, use := range val.Uses {
				if use.Op == OpPhi {
					// If use is a phi, argument block must dominate the pred
					// block of use block. For example,
					// b1: v1 = phi(v2, v3) ;; pred b4, b5
					// b2: v2 = add(v4, v5)
					// b3: v3 = add(v6, v7)
					// b2 and b3 must dominate b4 and b5 respectively
					for ipred, pred := range use.Block.Preds {
						phiArg := use.Args[ipred]
						if !domTree.IsDominate(phiArg.Block, pred) {
							DumpSSAToDotFile(fn)
							fmt.Printf("%v\n", fn)
							fmt.Printf("%v\n", domTree)
							utils.Fatal("b%v does not dominate b%d",
								phiArg.Block.Id, pred.Id)
						}
					}
					continue
				}
				if !domTree.IsDominate(val.Block, use.Block) {
					fmt.Printf("%v", fn)
					DumpSSAToDotFile(fn)
					utils.Fatal("def v%d(b%d) does not dominate its use v%d(b%d)",
						val.Id, val.Block.Id, use.Id, use.Block.Id)
				}
			}
		}
	}
}
