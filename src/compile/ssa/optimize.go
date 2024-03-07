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

type Optimizer struct {
	Func  *Func
	Debug bool
}

// Ideal iteratively performs peephole optimizations on the HIR until no more
// changes are made.
func (opt *Optimizer) Ideal() {
	changed := 1
	round := 0
	for changed == 1 {
		changed = 0
		changed |= opt.simplifyPhi()
		changed |= opt.simplifyCFG()
		changed |= opt.dce()
		round++
	}
	if opt.Debug {
		fmt.Printf("%d round ideal optimization\n", round)
	}
}

// -----------------------------------------------------------------------------
// Phi Simplification
// This pass simplifies phi nodes in the CFG. It replaces phi nodes with a single
// argument with the argument itself. It also replaces phi nodes with the same
// argument in all incoming edges.

func (opt *Optimizer) simplifyPhi() int {
	fn := opt.Func
	changed := 0
	for _, block := range fn.Blocks {
		// Reverse order because we may remove phi value
		for i := len(block.Values) - 1; i >= 0; i-- {
			val := block.Values[i]
			if val.Op == OpPhi {
				if len(val.Args) == 1 {
					fmt.Printf("Simplify %v to %v\n", val, val.Args[0])
					changed = 1
					val.ReplaceUses(val.Args[0])
					block.RemoveValue(val)
				} else if len(val.Args) == 0 {
					panic("Phi node with no arguments")
				} else {
					// all phi's incoming args are same
					// replace phi(v1,v1,v1) with v1
					same := true
					for _, arg := range val.Args {
						if arg != val.Args[0] {
							same = false
							break
						}
					}
					if same {
						fmt.Printf("Simplify %v to %v\n", val, val.Args[0])
						changed = 1
						val.ReplaceUses(val.Args[0])
						block.RemoveValue(val)
						continue
					}
					// all phi's incoming args are itself+one different value
					var one *Value
					for _, arg := range val.Args {
						if arg != val {
							if one == nil {
								one = arg
							} else {
								one = nil
								break
							}
						}
					}
					if one != nil {
						fmt.Printf("Simplify %v to %v\n", val, one)
						changed = 1
						val.ReplaceUses(one)
						block.RemoveValue(val)
						continue
					}
				}
			}
		}
	}
	return changed
}

// -----------------------------------------------------------------------------
// Dead Code Elimination
// This pass eliminates dead code from the CFG. It removes values that are not
// used and are not pinned.

func isPinned(val *Value) bool {
	switch val.Op {
	// TODO: For memory access, we need memory SSA to determine if it is dead,
	// now we conservatively assume it is live.
	case OpParam, OpCall, OpLoad, OpLoadIndex, OpStoreIndex, OpStore:
		return true
	}
	return false
}

func findReachableBlocksRecursively(block *Block, reachable map[*Block]bool) {
	if reachable[block] {
		return
	}
	reachable[block] = true
	for _, succ := range block.Succs {
		findReachableBlocksRecursively(succ, reachable)
	}
}

func FindReachableBlocks(block *Block) map[*Block]bool {
	reachable := make(map[*Block]bool)
	findReachableBlocksRecursively(block, reachable)
	return reachable
}

func (opt *Optimizer) dce() int {
	fn := opt.Func
	changed := 0

	// Find reachable blocks from entry
	reachable := FindReachableBlocks(fn.Entry)
	if opt.Debug {
		str := ""
		for block := range reachable {
			str += fmt.Sprintf("b%d ", block.Id)
		}
		fmt.Printf("Reachable blocks: %s\n", str)
	}

	// Remove dead values from reachable blocks, we do not make futile efforts
	// on unreachable blocks, they will be removed later.
	for block, _ := range reachable {
		// Reverse order because we may remove dead value
		for i := len(block.Values) - 1; i >= 0; i-- {
			val := block.Values[i]
			if len(val.Uses) == 0 && len(val.UseBlock) == 0 && !isPinned(val) {
				if opt.Debug {
					fmt.Printf("Dead value %v\n", val)
				}
				block.RemoveValue(val)
				changed = 1
			}
		}
	}

	// Remove unreachable blocks and its values, this ensures def-use relationship
	// is correct
	for i := len(fn.Blocks) - 1; i >= 0; i-- {
		block := fn.Blocks[i]
		if !reachable[block] {
			utils.Assert(block.Hint != HintEntry, "entry always reachable")
			if opt.Debug {
				fmt.Printf("Dead block %d\n", block.Id)
			}
			for _, succ := range block.Succs {
				if len(succ.Preds) > 1 {
					// possibly there are some phi values, remove correspdoning
					// arguments and uses
					for ipred, pred := range succ.Preds {
						if pred == block {
							for _, val := range succ.Values {
								if val.Op == OpPhi {
									def := val.Args[ipred]
									def.RemoveUseOnce(val)
									val.Args = append(val.Args[:ipred], val.Args[ipred+1:]...)
									break
								}
							}
							break
						}
					}
				}
			}
			for _, succ := range block.Succs {
				succ.RemovePred(block)
			}
			fn.RemoveBlock(block)
			changed = 1
		}
	}
	return changed
}

// -----------------------------------------------------------------------------
// CFG Simplification
// This pass simplifies the control flow graph. It removes unnecessary jumps and
// merges intermediate blocks.
func isConstBool(val *Value) bool {
	return val.Op == OpConst && val.Type.IsBool()
}

func (opt *Optimizer) simplifyCFG() int {
	fn := opt.Func
	changed := 0
	// If block always jumps to the same block, jump directly to the target
	for _, block := range fn.Blocks {
		if block.Kind == BlockIf {
			ctrl := block.Ctrl
			if isConstBool(ctrl) {
				taken := 0
				if ctrl.Sym.(bool) == false {
					taken = 1
				}

				fmt.Printf("Simplify If b%d to b%d\n", block.Id, block.Succs[taken].Id)
				// kill path from block to notTaken
				notTaken := block.Succs[1-taken]
				if len(notTaken.Preds) > 1 {
					// possibly there are some phi values, remove correspdoning
					// arguments and uses
					for ipred, pred := range notTaken.Preds {
						if pred == block {
							for _, val := range notTaken.Values {
								if val.Op == OpPhi {
									def := val.Args[ipred]
									def.RemoveUseOnce(val)
									val.Args = append(val.Args[:ipred], val.Args[ipred+1:]...)
									break
								}
							}
							break
						}
					}
				}
				block.Kind = BlockGoto
				ctrl.RemoveUseBlock(block)
				block.RemoveSucc(notTaken)
				notTaken.RemovePred(block)

				utils.Assert(len(block.Succs) == 1,
					"block has only one successor now")
				changed = 1
			}
		}
	}
	// If block is an imtermidiate block with only one predecessor and one
	// successor, merge it with the predecessor
	for _, block := range fn.Blocks {
		if block.Kind == BlockGoto &&
			len(block.Preds) == 1 &&
			len(block.Succs) == 1 &&
			len(block.Values) == 0 {
			pred := block.Preds[0]
			succ := block.Succs[0]
			if len(pred.Succs) == 1 && len(succ.Preds) == 1 {
				if opt.Debug {
					fmt.Printf("Merge block b%d into b%d\n", block.Id, pred.Id)
				}
				// remove intermediate block
				block.RemoveSucc(succ)
				block.RemovePred(pred)
				pred.RemoveSucc(block)
				succ.RemovePred(block)

				// merge values
				pred.WireTo(succ)
				pred.Values = append(pred.Values, block.Values...)
				block.Values = nil
				changed = 1
			}
		}
	}
	return changed
}

// -----------------------------------------------------------------------------
// Value Numbering
// This is a local value numbering, it creates hash table per-block and found
// if there is a same value in current block.
const BadHashValue = 0

func hash(nums ...int) int {
	switch len(nums) {
	case 1:
		return nums[0]
	case 2:
		return (nums[0] << 7) ^ nums[1]
	case 3:
		return ((nums[0] << 7) ^ nums[1]<<7) ^ nums[2]
	case 4:
		return (((nums[0] << 7) ^ nums[1]<<7) ^ nums[2]<<7) ^ nums[3]
	default:
		return BadHashValue
	}
}

const EnableLoopOpts = false

func OptimizeHIR(fn *Func, debug bool) {
	opt := &Optimizer{
		Func:  fn,
		Debug: debug,
	}
	opt.Ideal()

	if EnableLoopOpts {
		OptimizeLoop(fn)
	}
}
