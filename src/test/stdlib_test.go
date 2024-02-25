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
package test

import "testing"

func TestMath(t *testing.T) {
	source := `
	func main(){
		assert(min(1,2),1)
		assert(min(2,1),1)
		assert(max(1,2),2)
		assert(max(2,1),2)
		assert(min(1,1),1)
		assert(max(1,1),1)
		assert(clamp(1,2,3),2)
		assert(clamp(2,2,3),2)
		assert(clamp(3,2,3),3)
		assert(clamp(4,2,3),3)
		assert(clamp(-6,-5,-3),-5)
		assert(clamp(-4,-5,-3),-4)
		assert(clamp(-3,-5,-3),-3)
		assert(clamp(-2,-5,-3),-3)
		assert(clamp(0,-5,-3),-3)
	}
	`
	ExecExpect(source)
}
